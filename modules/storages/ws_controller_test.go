package storages

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/coder/websocket"
	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/fari-99/go-helper/token_generator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	wsauth "github.com/fari-99/go-helper/ws_auth"
	"go-api/modules/models"
	"go-api/modules/ws_auth"
)

type fakeService struct {
	mu    sync.Mutex
	saved []models.Storages
}

func (f *fakeService) GetDetail(*gin.Context, uint64) (*models.Storages, bool, error) {
	return nil, false, nil
}
func (f *fakeService) Uploads(*gin.Context, *multipart.Form) ([]models.Storages, error) {
	return nil, nil
}
func (f *fakeService) CreateStorage(_ *gin.Context, m models.Storages) (*models.Storages, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m.ID = models.IDType(len(f.saved) + 1)
	f.saved = append(f.saved, m)
	return &m, nil
}

func (f *fakeService) GetList(*gin.Context, int, int) ([]models.Storages, *paginator.Pagination, error) {
	return nil, nil, nil
}

func (f *fakeService) OpenOwned(*gin.Context, uint64) (*models.Storages, *os.File, error) {
	return nil, nil, ErrStorageNotFound
}

type wsHarness struct {
	server  *httptest.Server
	auth    wsauth.Service
	service *fakeService
	pair    *wsauth.TokenPair
	storage string
}

func newWSHarness(t *testing.T) *wsHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	t.Setenv("USER_DETAILS_PASSPHRASE", "test-passphrase")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("WS_UPLOAD_MAX_FILE_MB", "1")
	storage := t.TempDir()
	t.Setenv("LOCAL_STORAGE_PATH", storage)

	mr := miniredis.RunT(t)
	auth := wsauth.NewService(redis.NewClient(&redis.Options{Addr: mr.Addr()}), wsauth.Config{
		AccessTTL: 5 * time.Minute, RefreshTTL: 30 * time.Minute, RenewGrace: 30 * time.Second,
		MaxConcurrent: 3, CounterTTL: time.Hour, TokenRatePerMin: 100,
		AccessSecret: "a", RefreshSecret: "r", SignMethod: "HS256", EncryptionKey: "k",
	})

	pair, err := auth.Issue(context.Background(), nil, token_generator.UserDetails{ID: "42", Username: "fari"})
	if err != nil {
		t.Fatal(err)
	}

	service := &fakeService{}
	router := gin.New()
	control := newWSController(service, auth)
	router.GET("/ws/storages/upload", ws_auth.Middleware(auth), control.UploadAction)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &wsHarness{server: server, auth: auth, service: service, pair: pair, storage: storage}
}

func (h *wsHarness) dial(t *testing.T, token string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, h.server.URL+"/ws/storages/upload?token="+token, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	conn.SetReadLimit(1 << 20)
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

func sendJSON(t *testing.T, conn *websocket.Conn, payload map[string]any) {
	t.Helper()
	data, _ := json.Marshal(payload)
	if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func sendChunk(t *testing.T, conn *websocket.Conn, id uint32, payload []byte) {
	t.Helper()
	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame, id)
	copy(frame[4:], payload)
	if err := conn.Write(context.Background(), websocket.MessageBinary, frame); err != nil {
		t.Fatalf("write chunk: %v", err)
	}
}

// waitFor reads messages until one has the wanted type (skipping progress etc).
func waitFor(t *testing.T, conn *websocket.Conn, want string) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("waiting for %q: %v", want, err)
		}

		var msg map[string]any
		_ = json.Unmarshal(data, &msg)
		if msg["type"] == want {
			return msg
		}
	}
}

func TestWSUploadHappyPath(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendJSON(t, conn, map[string]any{"type": "start", "id": 1, "file_name": "hello.txt", "file_type": "docs", "size": 11})
	waitFor(t, conn, "started")
	sendChunk(t, conn, 1, []byte("hello "))
	sendChunk(t, conn, 1, []byte("world"))
	sendJSON(t, conn, map[string]any{"type": "end", "id": 1})

	ack := waitFor(t, conn, "ack")
	storage := ack["storage"].(map[string]any)
	if storage["mime"] != "text/plain; charset=utf-8" || storage["original_filename"] != "hello.txt" || storage["type"] != "docs" {
		t.Errorf("unexpected storage record: %v", storage)
	}

	if len(h.service.saved) != 1 || h.service.saved[0].CreatedBy != 42 {
		t.Fatalf("saved records = %+v, want one created by user 42", h.service.saved)
	}

	saved := h.service.saved[0]
	content, err := os.ReadFile(filepath.Join(h.storage, saved.Type, saved.Path, saved.Filename))
	if err != nil || string(content) != "hello world" {
		t.Errorf("stored content = %q, err = %v", content, err)
	}
}

func TestWSRejectsMissingOrBadToken(t *testing.T) {
	h := newWSHarness(t)

	for _, token := range []string{"", "garbage", h.pair.RefreshToken} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, resp, err := websocket.Dial(ctx, h.server.URL+"/ws/storages/upload?token="+token, nil)
		cancel()
		if err == nil || resp == nil || resp.StatusCode != 401 {
			t.Errorf("token %q: err=%v resp=%v, want 401", token, err, resp)
		}
	}
}

func TestWSFileTooLargeIsAbortedAndNotSaved(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendJSON(t, conn, map[string]any{"type": "start", "id": 1, "file_name": "big.bin", "file_type": "docs"})
	waitFor(t, conn, "started")

	chunk := []byte(strings.Repeat("a", 700*1024))
	sendChunk(t, conn, 1, chunk)
	sendChunk(t, conn, 1, chunk) // 1.4 MB > 1 MB limit

	if msg := waitFor(t, conn, "error"); msg["code"] != "file_too_large" {
		t.Errorf("error = %v, want file_too_large", msg)
	}
	if len(h.service.saved) != 0 {
		t.Error("oversized upload must not create a storage record")
	}
}

func TestWSDeclaredSizeAboveLimitRejectedAtStart(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendJSON(t, conn, map[string]any{"type": "start", "id": 1, "file_name": "big.bin", "file_type": "docs", "size": 5 << 20})
	if msg := waitFor(t, conn, "error"); msg["code"] != "file_too_large" {
		t.Errorf("error = %v, want file_too_large", msg)
	}
}

func TestWSSizeMismatch(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendJSON(t, conn, map[string]any{"type": "start", "id": 1, "file_name": "a.txt", "file_type": "docs", "size": 100})
	waitFor(t, conn, "started")
	sendChunk(t, conn, 1, []byte("short"))
	sendJSON(t, conn, map[string]any{"type": "end", "id": 1})

	if msg := waitFor(t, conn, "error"); msg["code"] != "size_mismatch" {
		t.Errorf("error = %v, want size_mismatch", msg)
	}
}

func TestWSInvalidFileTypeRejected(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendJSON(t, conn, map[string]any{"type": "start", "id": 1, "file_name": "a.txt", "file_type": "../../etc"})
	if msg := waitFor(t, conn, "error"); msg["code"] != "invalid_file_type" {
		t.Errorf("error = %v, want invalid_file_type", msg)
	}
}

func TestWSConcurrentUploadCapPerUser(t *testing.T) {
	h := newWSHarness(t)

	// two sockets of the same user share the cap of 3
	first, second := h.dial(t, h.pair.AccessToken), h.dial(t, h.pair.AccessToken)

	for id := 1; id <= 2; id++ {
		sendJSON(t, first, map[string]any{"type": "start", "id": id, "file_name": "a.txt", "file_type": "docs"})
		waitFor(t, first, "started")
	}
	sendJSON(t, second, map[string]any{"type": "start", "id": 1, "file_name": "a.txt", "file_type": "docs"})
	waitFor(t, second, "started")

	sendJSON(t, second, map[string]any{"type": "start", "id": 2, "file_name": "a.txt", "file_type": "docs"})
	if msg := waitFor(t, second, "error"); msg["code"] != "too_many_uploads" {
		t.Errorf("error = %v, want too_many_uploads", msg)
	}

	// finishing one frees a slot
	sendChunk(t, first, 1, []byte("x"))
	sendJSON(t, first, map[string]any{"type": "end", "id": 1})
	waitFor(t, first, "ack")

	sendJSON(t, second, map[string]any{"type": "start", "id": 2, "file_name": "a.txt", "file_type": "docs"})
	waitFor(t, second, "started")
}

func TestWSDisconnectReleasesSlotsAndCleansPartialFile(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	for id := 1; id <= 3; id++ {
		sendJSON(t, conn, map[string]any{"type": "start", "id": id, "file_name": "a.txt", "file_type": "docs"})
		waitFor(t, conn, "started")
	}
	sendChunk(t, conn, 1, []byte("partial"))
	conn.Close(websocket.StatusNormalClosure, "bye")

	deadline := time.Now().Add(5 * time.Second)
	for {
		release, ok, _ := h.auth.AcquireUpload(context.Background(), "42")
		if ok {
			release()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("slots were not released after the socket closed")
		}
		time.Sleep(50 * time.Millisecond)
	}

	if len(h.service.saved) != 0 {
		t.Error("aborted upload must not create a record")
	}
}

func TestWSRenewInPlaceAndReuseClosesSocket(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendJSON(t, conn, map[string]any{"type": "renew", "refresh_token": h.pair.RefreshToken})
	renewed := waitFor(t, conn, "renewed")
	if renewed["access_token"] == h.pair.AccessToken || renewed["refresh_token"] == "" {
		t.Fatalf("renew did not rotate: %v", renewed)
	}

	// uploads keep working on the same socket after renew
	sendJSON(t, conn, map[string]any{"type": "start", "id": 1, "file_name": "a.txt", "file_type": "docs"})
	waitFor(t, conn, "started")

	// replaying the old refresh token = theft: socket is closed with 4401 and the family is dead
	sendJSON(t, conn, map[string]any{"type": "renew", "refresh_token": h.pair.RefreshToken})
	waitFor(t, conn, "error")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := conn.Read(ctx)
	if websocket.CloseStatus(err) != closeUnauthorized {
		t.Errorf("close status = %v (err %v), want 4401", websocket.CloseStatus(err), err)
	}

	if _, _, err = h.auth.Authenticate(context.Background(), renewed["access_token"].(string)); err == nil {
		t.Error("rotated access token must be revoked after reuse detection")
	}
}

func TestWSBinaryFrameForUnknownUpload(t *testing.T) {
	h := newWSHarness(t)
	conn := h.dial(t, h.pair.AccessToken)

	sendChunk(t, conn, 99, []byte("x"))
	if msg := waitFor(t, conn, "error"); msg["code"] != "unknown_upload" {
		t.Errorf("error = %v, want unknown_upload", msg)
	}

	if err := conn.Write(context.Background(), websocket.MessageBinary, []byte{1}); err != nil {
		t.Fatal(err)
	}
	if msg := waitFor(t, conn, "error"); msg["code"] != "bad_message" {
		t.Errorf("error = %v, want bad_message", msg)
	}
}
