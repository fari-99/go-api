package file_processor_helper

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUploadSendsFileAndFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/pdf/summary" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("authorization header = %q", got)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		content, _ := io.ReadAll(file)
		if string(content) != "pdf bytes" || header.Filename != "report.pdf" {
			t.Errorf("file = %q (%s)", content, header.Filename)
		}
		if got := r.FormValue("length"); got != "short" {
			t.Errorf("length field = %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"summary":"ok"}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL + "/", APIKey: "secret"})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Upload(context.Background(), UploadRequest{
		Endpoint: "pdf/summary",
		FileName: "report.pdf",
		File:     strings.NewReader("pdf bytes"),
		Fields:   map[string]string{"length": "short"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || string(resp.Body) != `{"summary":"ok"}` {
		t.Errorf("response = %d %s", resp.StatusCode, resp.Body)
	}
}

func TestUploadHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("bad file"))
	}))
	defer server.Close()

	client, _ := New(Config{BaseURL: server.URL})
	resp, err := client.Upload(context.Background(), UploadRequest{Endpoint: "/x", FileName: "a", File: strings.NewReader("a")})

	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity || apiErr.Body != "bad file" {
		t.Fatalf("err = %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("response should be returned with the error, got %+v", resp)
	}
}

func TestUploadTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	client, _ := New(Config{BaseURL: server.URL, Timeout: 50 * time.Millisecond})
	_, err := client.Upload(context.Background(), UploadRequest{Endpoint: "/x", FileName: "a", File: strings.NewReader("a")})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestNewRequiresBaseURL(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error without base url")
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("FILE_PROCESSOR_BASE_URL", " http://localhost:9000 ")
	t.Setenv("FILE_PROCESSOR_TIMEOUT_SECONDS", "12")
	t.Setenv("FILE_PROCESSOR_DEBUG", "true")

	config := ConfigFromEnv()
	if config.BaseURL != "http://localhost:9000" || config.Timeout != 12*time.Second || !config.Debug {
		t.Errorf("config = %+v", config)
	}
}
