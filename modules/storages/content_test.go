package storages

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"go-api/modules/models"
)

type contentService struct {
	fakeService
	model *models.Storages
	path  string
}

func (c *contentService) OpenOwned(_ *gin.Context, id uint64) (*models.Storages, *os.File, error) {
	if c.model == nil || id != 1 {
		return nil, nil, ErrStorageNotFound
	}
	file, err := os.Open(c.path)
	return c.model, file, err
}

func newContentRouter(t *testing.T, mime string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	path := filepath.Join(t.TempDir(), "f.bin")
	if err := os.WriteFile(path, []byte("0123456789"), 0644); err != nil {
		t.Fatal(err)
	}

	service := &contentService{model: &models.Storages{Mime: mime, OriginalFilename: "a b.bin"}, path: path}

	// the real registrator also registers the public routes; make sure the paths don't clash
	router := gin.New()
	NewRegistrator(router.Group(""), service, func(ctx *gin.Context) {}, nil)
	return router
}

func get(router *gin.Engine, target string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestContentInlineMediaAndRange(t *testing.T) {
	router := newContentRouter(t, "video/mp4")

	rec := get(router, "/storages/1/content", map[string]string{"Range": "bytes=2-5"})
	if rec.Code != http.StatusPartialContent || rec.Body.String() != "2345" {
		t.Fatalf("range: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("content-type = %q", rec.Header().Get("Content-Type"))
	}
	if got := rec.Header().Get("Content-Disposition"); got != `inline; filename="a b.bin"` {
		t.Errorf("disposition = %q", got)
	}
}

func TestContentDownloadAndNonMediaForceAttachment(t *testing.T) {
	media := newContentRouter(t, "audio/mpeg")
	if got := get(media, "/storages/1/content?download=1", nil).Header().Get("Content-Disposition"); got[:10] != "attachment" {
		t.Errorf("download disposition = %q", got)
	}

	page := newContentRouter(t, "text/html; charset=utf-8")
	rec := get(page, "/storages/1/content", nil)
	if got := rec.Header().Get("Content-Disposition"); got[:10] != "attachment" {
		t.Errorf("html must not render inline, got %q", got)
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Error("missing CSP")
	}
}

func TestContentNotFoundAndBadID(t *testing.T) {
	router := newContentRouter(t, "image/png")
	if code := get(router, "/storages/2/content", nil).Code; code != http.StatusNotFound {
		t.Errorf("missing = %d", code)
	}
	if code := get(router, "/storages/abc/content", nil).Code; code != http.StatusBadRequest {
		t.Errorf("bad id = %d", code)
	}
}

type detailService struct {
	fakeService
	model    *models.Storages
	notFound bool
}

func (d *detailService) GetDetail(*gin.Context, uint64) (*models.Storages, bool, error) {
	return d.model, d.notFound, nil
}

func TestDetailActionFoundAndNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	found := gin.New()
	NewRegistrator(found.Group(""), &detailService{model: &models.Storages{Filename: "x.png"}}, func(*gin.Context) {}, nil)
	rec := get(found, "/storages/7", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "x.png") {
		t.Errorf("existing record: code=%d body=%s", rec.Code, rec.Body.String())
	}

	missing := gin.New()
	NewRegistrator(missing.Group(""), &detailService{notFound: true}, func(*gin.Context) {}, nil)
	if code := get(missing, "/storages/7", nil).Code; code != http.StatusNotFound {
		t.Errorf("missing record = %d, want 404", code)
	}
}
