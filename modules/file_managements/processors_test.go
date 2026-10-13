package file_managements

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"go-api/modules/models"
	fileprocessor "go-api/pkg/file_processor_helper"
)

func TestValidateParams(t *testing.T) {
	specs := []ParamSpec{
		{Name: "language", Type: paramTypeString, MinLen: 2, MaxLen: 10},
		{Name: "length", Type: paramTypeString, Options: []string{"short", "long"}},
		{Name: "timestamps", Type: paramTypeBool},
	}

	got, err := validateParams(specs, map[string]any{"language": " en ", "length": "short", "timestamps": "1"})
	want := map[string]string{"language": "en", "length": "short", "timestamps": "true"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v (%v), want %v", got, err, want)
	}

	if got, err := validateParams(specs, nil); err != nil || len(got) != 0 {
		t.Errorf("nil params = %v (%v)", got, err)
	}

	invalid := []map[string]any{
		{"unknown": "x"},
		{"language": "e"},
		{"language": 12},
		{"length": "medium"},
		{"timestamps": "maybe"},
		{"timestamps": 1},
	}
	for _, params := range invalid {
		if _, err := validateParams(specs, params); err == nil {
			t.Errorf("params %v should be rejected", params)
		}
	}
}

func TestAcceptsFile(t *testing.T) {
	processor := fakeProcessor{}

	cases := []struct {
		mime, filename string
		want           bool
	}{
		{"application/pdf", "a.pdf", true},
		{"application/pdf; charset=binary", "a.bin", true},
		{"application/octet-stream", "Report.PDF", true},
		{"", "a.pdf", true},
		{"application/octet-stream", "a.exe", false},
		{"text/html; charset=utf-8", "a.pdf", false},
	}

	for _, c := range cases {
		got := acceptsFile(processor, models.Storages{Mime: c.mime, OriginalFilename: c.filename})
		if got != c.want {
			t.Errorf("acceptsFile(%q, %q) = %v, want %v", c.mime, c.filename, got, c.want)
		}
	}
}

type fakeUploader struct {
	got  fileprocessor.UploadRequest
	body string
	resp *fileprocessor.UploadResponse
	err  error
}

func (u *fakeUploader) Upload(ctx context.Context, input fileprocessor.UploadRequest) (*fileprocessor.UploadResponse, error) {
	u.got = input
	content, _ := io.ReadAll(input.File)
	u.body = string(content)
	return u.resp, u.err
}

func TestUploadProcessorSuccess(t *testing.T) {
	uploader := &fakeUploader{resp: &fileprocessor.UploadResponse{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"summary":"hello","pages":12,"sections":[{"title":"a"}],"empty":null}`),
		Latency:    1500 * time.Millisecond,
	}}
	processor := uploadProcessor{action: ActionPdfSummary, outputType: "summary", endpoint: "/pdf/summary", client: uploader}

	result, err := processor.Process(context.Background(), Job{
		Storage: models.Storages{OriginalFilename: "report.pdf"},
		File:    strings.NewReader("pdf"),
		Params:  map[string]string{"length": "short"},
	})
	if err != nil {
		t.Fatal(err)
	}

	wantOutput := map[string]string{"summary": "hello", "pages": "12", "sections": `[{"title":"a"}]`, "empty": ""}
	if !reflect.DeepEqual(result.Output, wantOutput) {
		t.Errorf("output = %v, want %v", result.Output, wantOutput)
	}
	wantRequest := map[string]string{"endpoint": "/pdf/summary", "status_code": "200", "latency_ms": "1500"}
	if !reflect.DeepEqual(result.Request, wantRequest) {
		t.Errorf("request = %v, want %v", result.Request, wantRequest)
	}
	if uploader.got.FileName != "report.pdf" || uploader.body != "pdf" || uploader.got.Fields["length"] != "short" {
		t.Errorf("upload request = %+v body %q", uploader.got, uploader.body)
	}
}

func TestUploadProcessorErrors(t *testing.T) {
	cases := []struct {
		name     string
		uploader *fakeUploader
		code     string
	}{
		{
			name: "http error",
			uploader: &fakeUploader{
				resp: &fileprocessor.UploadResponse{StatusCode: http.StatusBadGateway},
				err:  &fileprocessor.Error{StatusCode: http.StatusBadGateway, Body: "down"},
			},
			code: "third_party_error",
		},
		{name: "timeout", uploader: &fakeUploader{err: context.DeadlineExceeded}, code: "third_party_timeout"},
		{name: "unreachable", uploader: &fakeUploader{err: errors.New("connection refused")}, code: "third_party_unavailable"},
		{
			name:     "not json",
			uploader: &fakeUploader{resp: &fileprocessor.UploadResponse{StatusCode: http.StatusOK, Body: []byte("ok")}},
			code:     "invalid_response",
		},
		{
			name:     "json array",
			uploader: &fakeUploader{resp: &fileprocessor.UploadResponse{StatusCode: http.StatusOK, Body: []byte(`["a"]`)}},
			code:     "invalid_response",
		},
	}

	for _, c := range cases {
		processor := uploadProcessor{endpoint: "/x", client: c.uploader}
		result, err := processor.Process(context.Background(), Job{File: strings.NewReader("a")})

		var processErr *ProcessError
		if !errors.As(err, &processErr) || processErr.Code != c.code {
			t.Errorf("%s: err = %v, want %s", c.name, err, c.code)
			continue
		}
		if processErr.Message == "" || result == nil || result.Request["endpoint"] != "/x" {
			t.Errorf("%s: message %q, result %+v", c.name, processErr.Message, result)
		}
	}
}

func TestRegistryFromEnv(t *testing.T) {
	t.Setenv("FILE_PROCESSOR_BASE_URL", "")
	if got := RegistryFromEnv().List(); len(got) != 0 {
		t.Errorf("without base url got %d actions", len(got))
	}

	t.Setenv("FILE_PROCESSOR_BASE_URL", "http://localhost:9000")
	t.Setenv("FILE_MANAGEMENT_PDF_SUMMARY_MAX_MB", "3")
	registry := RegistryFromEnv()

	var actions []string
	for _, processor := range registry.List() {
		actions = append(actions, processor.Action())
	}
	if !reflect.DeepEqual(actions, []string{ActionPdfSummary, ActionAudioTranscription}) {
		t.Errorf("actions = %v", actions)
	}

	pdf, _ := registry.Get(ActionPdfSummary)
	if pdf.MaxFileBytes() != 3<<20 {
		t.Errorf("pdf max bytes = %d", pdf.MaxFileBytes())
	}
}
