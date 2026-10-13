package file_managements

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type testResponse struct {
	StatusCode int             `json:"status_code"`
	Data       json.RawMessage `json:"data"`
	Error      json.RawMessage `json:"error"`
}

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo, db := newTestRepo(t)
	registry := NewRegistry(fakeProcessor{})
	svc := newTestService(repo, registry, &fakePublisher{}, testUserID)
	storageModel := createStorage(t, db, testUserID, "application/pdf", "report.pdf")

	app := gin.New()
	authCalled := 0
	NewRegistrator(app.Group(""), svc, func(ctx *gin.Context) { authCalled++ })

	do := func(method, path, body string) (int, testResponse) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		app.ServeHTTP(recorder, request)

		var response testResponse
		_ = json.Unmarshal(recorder.Body.Bytes(), &response)
		return recorder.Code, response
	}

	code, response := do(http.MethodGet, "/file-managements/actions", "")
	if code != http.StatusOK || !strings.Contains(string(response.Data), `"action":"pdf_summary"`) {
		t.Fatalf("actions = %d %s", code, response.Data)
	}

	createBody := fmt.Sprintf(`{"storage_id":%d,"action":"pdf_summary","params":{"length":"short"}}`, storageModel.ID)
	code, response = do(http.MethodPost, "/file-managements", createBody)
	if code != http.StatusCreated {
		t.Fatalf("create = %d %s", code, response.Error)
	}

	var created FileManagementDetail
	_ = json.Unmarshal(response.Data, &created)

	code, _ = do(http.MethodPost, "/file-managements", createBody)
	if code != http.StatusOK {
		t.Errorf("duplicate create = %d, want 200", code)
	}

	code, response = do(http.MethodPost, "/file-managements", `{"action":"pdf_summary"}`)
	if code != http.StatusBadRequest {
		t.Errorf("create without storage = %d", code)
	}

	code, response = do(http.MethodPost, "/file-managements", `{"storage_id":999,"action":"pdf_summary"}`)
	if code != http.StatusNotFound || !strings.Contains(string(response.Error), "storage_not_found") {
		t.Errorf("create with unknown storage = %d %s", code, response.Error)
	}

	path := fmt.Sprintf("/file-managements/%d", created.ID)
	if code, _ = do(http.MethodGet, path, ""); code != http.StatusOK {
		t.Errorf("detail = %d", code)
	}
	if code, _ = do(http.MethodGet, "/file-managements", ""); code != http.StatusOK {
		t.Errorf("list = %d", code)
	}
	if code, _ = do(http.MethodGet, "/file-managements/abc", ""); code != http.StatusBadRequest {
		t.Errorf("detail with bad id = %d", code)
	}
	if code, _ = do(http.MethodPost, path+"/retry", ""); code != http.StatusConflict {
		t.Errorf("retry new job = %d, want 409", code)
	}
	if code, _ = do(http.MethodDelete, path, ""); code != http.StatusConflict {
		t.Errorf("delete active job = %d, want 409", code)
	}
	if code, _ = do(http.MethodPost, path+"/cancel", ""); code != http.StatusOK {
		t.Errorf("cancel = %d", code)
	}
	if code, _ = do(http.MethodDelete, path, ""); code != http.StatusOK {
		t.Errorf("delete = %d", code)
	}
	if code, _ = do(http.MethodGet, path, ""); code != http.StatusNotFound {
		t.Errorf("detail after delete = %d", code)
	}

	if authCalled == 0 {
		t.Error("auth middleware was not applied")
	}
}
