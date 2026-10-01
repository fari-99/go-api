package storages

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"

	wsauth "github.com/fari-99/go-helper/ws_auth"
	"github.com/fari-99/go-helper/wsstorage"
	"go-api/constant"
	"go-api/modules/models"
	"go-api/modules/ws_auth"
)

// Custom close codes (4000-4999 are reserved for applications).
const (
	closeUnauthorized = websocket.StatusCode(4401) // session revoked or refresh token reused
	closeTokenExpired = websocket.StatusCode(4408) // access token expired and not renewed
)

const chunkQueueSize = 4

type wsController struct {
	service Service
	wsAuth  wsauth.Service
	cfg     wsConfig
	storage wsstorage.Config
}

func newWSController(service Service, wsAuth wsauth.Service) wsController {
	return wsController{
		service: service,
		wsAuth:  wsAuth,
		cfg:     wsConfigFromEnv(),
		storage: wsstorage.ConfigFromEnv(),
	}
}

// controlMessage is every client text frame. Binary frames are: 4 byte big endian upload id + payload.
type controlMessage struct {
	Type         string `json:"type"`
	ID           uint32 `json:"id"`
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	Size         int64  `json:"size"`
	RefreshToken string `json:"refresh_token"`
}

// UploadAction upgrades to a WebSocket (already authenticated by ws_auth.Middleware) and serves uploads.
func (c wsController) UploadAction(ctx *gin.Context) {
	sessionValue, _ := ctx.Get(ws_auth.ContextSession)
	session, ok := sessionValue.(*wsauth.Session)
	if !ok {
		ctx.AbortWithStatus(401)
		return
	}

	expiresValue, _ := ctx.Get(ws_auth.ContextAccessExpiresAt)
	accessExpiresAt, _ := expiresValue.(time.Time)

	conn, err := websocket.Accept(ctx.Writer, ctx.Request, &websocket.AcceptOptions{
		OriginPatterns: c.cfg.OriginPatterns,
	})
	if err != nil {
		return // Accept already wrote the HTTP error
	}
	defer conn.CloseNow()

	// a chunk frame is 4 bytes id + payload
	conn.SetReadLimit(c.cfg.MaxChunkBytes + 4)

	connCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	wc := &wsConn{
		controller:      c,
		ginCtx:          ctx,
		conn:            conn,
		ctx:             connCtx,
		cancel:          cancel,
		session:         session,
		accessExpiresAt: accessExpiresAt,
		uploads:         make(map[uint32]*wsUpload),
	}

	go wc.watch()
	wc.readLoop()

	cancel()
	wc.abortAll("connection closed")
	wc.wg.Wait()
}

type wsConn struct {
	controller wsController
	ginCtx     *gin.Context
	conn       *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	mu              sync.Mutex // guards everything below
	session         *wsauth.Session
	accessExpiresAt time.Time
	uploads         map[uint32]*wsUpload
	totalBytes      int64
	closeOnce       sync.Once
}

type wsUpload struct {
	id       uint32
	fileName string
	fileType string
	declared int64
	received int64 // read loop only
	ended    bool  // read loop only
	lastProg time.Time

	chunks chan []byte
	cancel context.CancelFunc
	done   chan struct{}

	abortMu   sync.Mutex
	abortCode string
	abortMsg  string
}

func (u *wsUpload) abort(code, message string) {
	u.abortMu.Lock()
	if u.abortCode == "" {
		u.abortCode, u.abortMsg = code, message
	}
	u.abortMu.Unlock()
	u.cancel()
}

func (u *wsUpload) abortReason() (string, string) {
	u.abortMu.Lock()
	defer u.abortMu.Unlock()
	return u.abortCode, u.abortMsg
}

func (w *wsConn) send(payload map[string]any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	writeCtx, cancel := context.WithTimeout(w.ctx, 10*time.Second)
	defer cancel()
	_ = w.conn.Write(writeCtx, websocket.MessageText, data)
}

func (w *wsConn) sendError(id uint32, code, message string) {
	w.send(map[string]any{"type": "error", "id": id, "code": code, "message": message})
}

// closeWith closes the socket once; the read loop then exits.
func (w *wsConn) closeWith(code websocket.StatusCode, reason string) {
	w.closeOnce.Do(func() {
		_ = w.conn.Close(code, reason)
		w.cancel()
	})
}

func (w *wsConn) readLoop() {
	cfg := w.controller.cfg

	for {
		readCtx, cancel := context.WithTimeout(w.ctx, cfg.IdleTimeout)
		messageType, data, err := w.conn.Read(readCtx)
		cancel()
		if err != nil {
			return
		}

		switch messageType {
		case websocket.MessageText:
			var msg controlMessage
			if err = json.Unmarshal(data, &msg); err != nil {
				w.sendError(0, "bad_message", "invalid JSON control message")
				continue
			}

			if !w.handleControl(msg) {
				return
			}
		case websocket.MessageBinary:
			if !w.handleChunk(data) {
				return
			}
		}
	}
}

// handleControl returns false when the connection must end.
func (w *wsConn) handleControl(msg controlMessage) bool {
	switch msg.Type {
	case "ping":
		w.send(map[string]any{"type": "pong"})
	case "start":
		w.handleStart(msg)
	case "end":
		w.handleEnd(msg)
	case "abort":
		if upload := w.getUpload(msg.ID); upload != nil {
			upload.abort("aborted", "upload aborted by client")
		}
	case "renew":
		return w.handleRenew(msg)
	default:
		w.sendError(msg.ID, "unknown_type", "unknown message type")
	}

	return true
}

func (w *wsConn) getUpload(id uint32) *wsUpload {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.uploads[id]
}

func (w *wsConn) handleStart(msg controlMessage) {
	cfg := w.controller.cfg

	w.mu.Lock()
	expired := time.Now().After(w.accessExpiresAt)
	_, duplicate := w.uploads[msg.ID]
	session := w.session
	w.mu.Unlock()

	switch {
	case expired:
		w.sendError(msg.ID, "token_expired", "access token expired, send renew")
		return
	case duplicate:
		w.sendError(msg.ID, "duplicate_id", "an upload with this id is already in progress")
		return
	case !wsstorage.ValidFileType(msg.FileType):
		w.sendError(msg.ID, "invalid_file_type", "file_type is invalid")
		return
	case msg.FileName == "":
		w.sendError(msg.ID, "invalid_file_name", "file_name is required")
		return
	case msg.Size < 0 || msg.Size > cfg.MaxFileBytes:
		w.sendError(msg.ID, "file_too_large", "file exceeds the maximum allowed size")
		return
	}

	release, ok, err := w.controller.wsAuth.AcquireUpload(w.ctx, session.UserID)
	if err != nil {
		log.Printf("ws upload: acquire slot failed, err := %s", err.Error())
		w.sendError(msg.ID, "internal_error", "failed to start upload")
		return
	} else if !ok {
		w.sendError(msg.ID, "too_many_uploads", "too many concurrent uploads")
		return
	}

	uploadCtx, cancel := context.WithCancel(w.ctx)
	upload := &wsUpload{
		id:       msg.ID,
		fileName: msg.FileName,
		fileType: msg.FileType,
		declared: msg.Size,
		chunks:   make(chan []byte, chunkQueueSize),
		cancel:   cancel,
		done:     make(chan struct{}),
	}

	w.mu.Lock()
	w.uploads[msg.ID] = upload
	w.mu.Unlock()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer close(upload.done)
		defer release()
		defer cancel()
		defer func() {
			w.mu.Lock()
			delete(w.uploads, upload.id)
			w.mu.Unlock()
		}()

		w.runUpload(uploadCtx, upload, session)
	}()

	w.send(map[string]any{"type": "started", "id": msg.ID})
}

func (w *wsConn) runUpload(ctx context.Context, upload *wsUpload, session *wsauth.Session) {
	controller := w.controller

	result, err := wsstorage.Upload(ctx, controller.storage, upload.fileType, upload.fileName,
		&chunkReader{ctx: ctx, chunks: upload.chunks}, controller.cfg.MaxFileBytes)
	if err != nil {
		if code, message := upload.abortReason(); code != "" {
			w.sendError(upload.id, code, message)
			return
		}

		switch {
		case errors.Is(err, wsstorage.ErrFileTooLarge):
			w.sendError(upload.id, "file_too_large", "file exceeds the maximum allowed size")
		case errors.Is(err, wsstorage.ErrEmptyFile):
			w.sendError(upload.id, "empty_file", "file is empty")
		case ctx.Err() != nil:
			// connection is gone, nobody to tell
		default:
			log.Printf("ws upload failed, id := %d, err := %s", upload.id, err.Error())
			w.sendError(upload.id, "upload_failed", "failed to upload file")
		}
		return
	}

	createdBy, _ := strconv.ParseUint(session.UserID, 10, 64)
	saved, err := controller.service.CreateStorage(w.ginCtx, models.Storages{
		Type:             result.Type,
		Path:             result.Path,
		Filename:         result.Filename,
		Mime:             result.Mime,
		OriginalFilename: result.OriginalFilename,
		Status:           constant.StatusActive,
		CreatedBy:        models.IDType(createdBy),
	})
	if err != nil {
		log.Printf("ws upload: save storage record failed, id := %d, err := %s", upload.id, err.Error())
		w.sendError(upload.id, "upload_failed", "failed to save file")
		return
	}

	w.send(map[string]any{"type": "ack", "id": upload.id, "storage": saved})
}

// handleChunk returns false when the connection must end.
func (w *wsConn) handleChunk(data []byte) bool {
	cfg := w.controller.cfg

	if len(data) < 4 {
		w.sendError(0, "bad_message", "binary frame is missing the upload id")
		return true
	}

	id := binary.BigEndian.Uint32(data[:4])
	payload := data[4:]

	upload := w.getUpload(id)
	if upload == nil || upload.ended {
		w.sendError(id, "unknown_upload", "no active upload with this id")
		return true
	}

	upload.received += int64(len(payload))
	if upload.received > cfg.MaxFileBytes {
		upload.abort("file_too_large", "file exceeds the maximum allowed size")
		return true
	}

	if cfg.MaxConnectionBytes > 0 {
		w.mu.Lock()
		w.totalBytes += int64(len(payload))
		exceeded := w.totalBytes > cfg.MaxConnectionBytes
		w.mu.Unlock()

		if exceeded {
			upload.abort("connection_limit", "connection upload limit reached")
			w.closeWith(websocket.StatusMessageTooBig, "connection upload limit reached")
			return false
		}
	}

	// bounded queue: a slow backend applies backpressure to the client, but not forever
	stall := time.NewTimer(cfg.StallTimeout)
	defer stall.Stop()

	select {
	case upload.chunks <- payload:
	case <-upload.done:
		return true
	case <-stall.C:
		upload.abort("upload_stalled", "storage backend is too slow")
		return true
	case <-w.ctx.Done():
		return false
	}

	if time.Since(upload.lastProg) > 250*time.Millisecond {
		upload.lastProg = time.Now()
		w.send(map[string]any{"type": "progress", "id": id, "received": upload.received})
	}

	return true
}

func (w *wsConn) handleEnd(msg controlMessage) {
	upload := w.getUpload(msg.ID)
	if upload == nil || upload.ended {
		w.sendError(msg.ID, "unknown_upload", "no active upload with this id")
		return
	}

	if upload.declared > 0 && upload.received != upload.declared {
		upload.abort("size_mismatch", "received size does not match the declared size")
		return
	}

	upload.ended = true
	close(upload.chunks)
}

func (w *wsConn) handleRenew(msg controlMessage) bool {
	w.mu.Lock()
	current := w.session
	w.mu.Unlock()

	pair, session, err := w.controller.wsAuth.Renew(w.ctx, w.ginCtx.Request, msg.RefreshToken)
	if errors.Is(err, wsauth.ErrInvalidToken) || errors.Is(err, wsauth.ErrTokenReused) {
		w.sendError(0, "unauthorized", "unauthorized")
		w.closeWith(closeUnauthorized, "unauthorized")
		return false
	} else if err != nil {
		log.Printf("ws upload: renew failed, err := %s", err.Error())
		w.sendError(0, "internal_error", "failed to renew token")
		return true
	}

	// a refresh token from another user/session must not take over this connection
	if session.FamilyID != current.FamilyID || session.UserID != current.UserID {
		w.sendError(0, "unauthorized", "unauthorized")
		w.closeWith(closeUnauthorized, "unauthorized")
		return false
	}

	w.mu.Lock()
	w.session = session
	w.accessExpiresAt = pair.AccessExpiredAt
	w.mu.Unlock()

	w.send(map[string]any{
		"type":               "renewed",
		"access_token":       pair.AccessToken,
		"refresh_token":      pair.RefreshToken,
		"access_expires_at":  pair.AccessExpiredAt,
		"refresh_expires_at": pair.RefreshExpiredAt,
	})
	return true
}

// watch pings the client and re-checks the session while the connection is open, so a revoked
// session can't keep uploading. Expiry only closes an idle connection: an in-flight file may finish.
func (w *wsConn) watch() {
	grace := w.controller.wsAuth.Config().RenewGrace
	pingTicker := time.NewTicker(30 * time.Second)
	checkTicker := time.NewTicker(10 * time.Second)
	defer pingTicker.Stop()
	defer checkTicker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-pingTicker.C:
			pingCtx, cancel := context.WithTimeout(w.ctx, 10*time.Second)
			err := w.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				w.cancel()
				return
			}
		case <-checkTicker.C:
			w.mu.Lock()
			familyID := w.session.FamilyID
			expiredLong := time.Now().After(w.accessExpiresAt.Add(grace))
			active := len(w.uploads)
			w.mu.Unlock()

			if !w.controller.wsAuth.SessionAlive(w.ctx, familyID) {
				w.closeWith(closeUnauthorized, "session revoked")
				return
			}

			if expiredLong && active == 0 {
				w.closeWith(closeTokenExpired, "token expired")
				return
			}
		}
	}
}

func (w *wsConn) abortAll(reason string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, upload := range w.uploads {
		upload.abort("aborted", reason)
	}
}

// chunkReader feeds queued binary frames to the storage uploader as one stream.
type chunkReader struct {
	ctx    context.Context
	chunks <-chan []byte
	buf    []byte
}

func (r *chunkReader) Read(p []byte) (int, error) {
	for len(r.buf) == 0 {
		select {
		case chunk, ok := <-r.chunks:
			if !ok {
				return 0, io.EOF
			}
			r.buf = chunk
		case <-r.ctx.Done():
			return 0, r.ctx.Err()
		}
	}

	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}
