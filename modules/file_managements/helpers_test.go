package file_managements

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-api/constant"
	"go-api/helpers/activitylog"
	"go-api/modules/configs"
	"go-api/modules/models"
)

const testUserID = 7
const otherUserID = 8

// newTestRepo opens a throw-away sqlite file with the activity-log callbacks registered, like the
// real connection, so conditional updates and upserts are checked against them too.
func newTestRepo(t *testing.T) (Repository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(&models.Storages{}, &models.FileManagements{}, &models.FileManagementMeta{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	err = db.Exec("CREATE UNIQUE INDEX file_management_meta_un ON file_management_meta (file_management_id, `group`, `type`, `key`)").Error
	if err != nil {
		t.Fatalf("create index: %v", err)
	}

	if err := activitylog.RegisterCallbacks(db); err != nil {
		t.Fatalf("register activity log: %v", err)
	}

	return NewRepository(&configs.DI{DB: db}), db
}

func createStorage(t *testing.T, db *gorm.DB, createdBy uint64, mime, originalFilename string) models.Storages {
	t.Helper()

	storageModel := models.Storages{
		Type:             "documents",
		Path:             "2026/10/",
		Filename:         "stored-name",
		Mime:             mime,
		OriginalFilename: originalFilename,
		Status:           constant.StatusActive,
		CreatedBy:        models.IDType(createdBy),
	}
	if err := db.Create(&storageModel).Error; err != nil {
		t.Fatalf("create storage: %v", err)
	}

	return storageModel
}

type fakePublisher struct {
	mu        sync.Mutex
	published []uint64
	err       error
}

func (p *fakePublisher) Publish(fileManagementID uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, fileManagementID)
	return nil
}

func (p *fakePublisher) ids() []uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]uint64(nil), p.published...)
}

// fakeProcessor accepts PDFs and runs process (or returns a fixed output).
type fakeProcessor struct {
	action   string
	maxBytes int64
	process  func(ctx context.Context, job Job) (*Result, error)
}

func (p fakeProcessor) Action() string {
	if p.action == "" {
		return ActionPdfSummary
	}
	return p.action
}
func (p fakeProcessor) OutputType() string           { return "summary" }
func (p fakeProcessor) AcceptedMimes() []string      { return []string{"application/pdf"} }
func (p fakeProcessor) AcceptedExtensions() []string { return []string{".pdf"} }
func (p fakeProcessor) Params() []ParamSpec {
	return []ParamSpec{{Name: "length", Type: paramTypeString, Options: []string{"short", "long"}}}
}
func (p fakeProcessor) MaxFileBytes() int64 {
	if p.maxBytes == 0 {
		return 1 << 20
	}
	return p.maxBytes
}
func (p fakeProcessor) ValidateParams(params map[string]any) (map[string]string, error) {
	return validateParams(p.Params(), params)
}
func (p fakeProcessor) Process(ctx context.Context, job Job) (*Result, error) {
	if p.process != nil {
		return p.process(ctx, job)
	}
	return &Result{
		Output:  map[string]string{"summary": "short text"},
		Request: map[string]string{"endpoint": "/fake", "status_code": "200"},
	}, nil
}

func newTestService(repo Repository, registry *Registry, publisher Publisher, userID uint64) service {
	return service{
		repo:          repo,
		registry:      registry,
		publisher:     publisher,
		maxActive:     5,
		currentUserID: func(ctx *gin.Context) (uint64, error) { return userID, nil },
	}
}

func newGinContext() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	return ctx
}

// newTestWorker opens a temp file with content instead of reading from storages.
func newTestWorker(t *testing.T, repo Repository, registry *Registry, content string) *Worker {
	t.Helper()

	path := filepath.Join(t.TempDir(), "file.pdf")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return &Worker{
		repo:     repo,
		registry: registry,
		timeout:  time.Minute,
		openFile: func(models.Storages) (*os.File, func(), error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, nil, err
			}
			return file, func() { _ = file.Close() }, nil
		},
	}
}

func assertServiceError(t *testing.T, err error, code string) {
	t.Helper()

	var serviceErr *Error
	if !errors.As(err, &serviceErr) || serviceErr.Code != code {
		t.Fatalf("err = %v, want code %s", err, code)
	}
}

func getStatus(t *testing.T, db *gorm.DB, fileManagementID uint64) int8 {
	t.Helper()

	var fileManagement models.FileManagements
	if err := db.Unscoped().First(&fileManagement, fileManagementID).Error; err != nil {
		t.Fatal(err)
	}
	return fileManagement.Status
}
