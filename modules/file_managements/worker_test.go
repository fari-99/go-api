package file_managements

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/fari-99/go-helper/rabbitmq"

	"go-api/constant"
	"go-api/modules/models"
)

// createJob creates a queued pdf_summary job through the service.
func createJob(t *testing.T, repo Repository, registry *Registry, storageModel models.Storages) uint64 {
	t.Helper()

	svc := newTestService(repo, registry, &fakePublisher{}, testUserID)
	detail, _, err := svc.Create(newGinContext(), CreateFileManagementRequest{
		StorageID: storageModel.ID.Uint64(),
		Action:    ActionPdfSummary,
		Params:    map[string]any{"length": "short"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return detail.ID.Uint64()
}

func getDetail(t *testing.T, repo Repository, registry *Registry, id uint64) *FileManagementDetail {
	t.Helper()

	detail, err := newTestService(repo, registry, &fakePublisher{}, testUserID).GetDetail(newGinContext(), id)
	if err != nil {
		t.Fatal(err)
	}
	return detail
}

func TestWorkerCompletesJob(t *testing.T) {
	repo, db := newTestRepo(t)

	var gotJob Job
	var gotContent string
	registry := NewRegistry(fakeProcessor{process: func(ctx context.Context, job Job) (*Result, error) {
		gotJob = job
		content, _ := io.ReadAll(job.File)
		gotContent = string(content)
		return &Result{
			Output:  map[string]string{"summary": "short text", "pages": "3"},
			Request: map[string]string{"endpoint": "/fake", "status_code": "200"},
		}, nil
	}})

	id := createJob(t, repo, registry, createStorage(t, db, testUserID, "application/pdf", "report.pdf"))
	worker := newTestWorker(t, repo, registry, "pdf bytes")

	// through the queue handler, with the payload as it arrives from JSON
	worker.Handle(rabbitmq.ConsumerHandlerData{
		EventType: constant.EventFileManagementProcess,
		Data:      map[string]interface{}{"file_management_id": float64(id)},
	})

	if gotContent != "pdf bytes" || !reflect.DeepEqual(gotJob.Params, map[string]string{"length": "short"}) {
		t.Errorf("job content %q params %v", gotContent, gotJob.Params)
	}

	detail := getDetail(t, repo, registry, id)
	if detail.Status != constant.FileManagementCompleted || detail.Attempt != 1 || detail.StartedAt == nil || detail.FinishedAt == nil {
		t.Errorf("detail = %+v", detail.FileManagementItem)
	}
	if got := detail.Meta[MetaGroupOutput]["summary"]; !reflect.DeepEqual(got, map[string]string{"summary": "short text", "pages": "3"}) {
		t.Errorf("output meta = %v", got)
	}
	if _, ok := detail.Meta[MetaGroupRequest]; ok {
		t.Error("request meta must not be returned to users")
	}
	if len(detail.AvailableEvents) != 0 || detail.Error != nil {
		t.Errorf("events %v error %+v", detail.AvailableEvents, detail.Error)
	}

	metas, _ := repo.GetMetas(context.Background(), id)
	if got := metaValues(metas, MetaGroupRequest, MetaTypeThirdParty); got["endpoint"] != "/fake" {
		t.Errorf("request meta = %v", got)
	}

	// a redelivered message is a no-op
	if err := worker.Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if detail := getDetail(t, repo, registry, id); detail.Attempt != 1 || detail.Status != constant.FileManagementCompleted {
		t.Errorf("after redelivery = %+v", detail.FileManagementItem)
	}
}

func TestWorkerFailureKeepsErrorPerAttempt(t *testing.T) {
	repo, db := newTestRepo(t)

	calls := 0
	registry := NewRegistry(fakeProcessor{process: func(ctx context.Context, job Job) (*Result, error) {
		calls++
		if calls == 1 {
			return &Result{Request: map[string]string{"status_code": "502"}},
				&ProcessError{Code: "third_party_error", Message: "could not process", Internal: "HTTP 502: upstream secret"}
		}
		return nil, errors.New("plain error")
	}})

	id := createJob(t, repo, registry, createStorage(t, db, testUserID, "application/pdf", "report.pdf"))
	worker := newTestWorker(t, repo, registry, "pdf")

	if err := worker.Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	detail := getDetail(t, repo, registry, id)
	if detail.Status != constant.FileManagementFailed || detail.Error == nil || detail.Error.Code != "third_party_error" || detail.Error.Message != "could not process" {
		t.Fatalf("detail = %+v error %+v", detail.FileManagementItem, detail.Error)
	}
	if !reflect.DeepEqual(detail.AvailableEvents, []string{EventRetry}) {
		t.Errorf("available events = %v", detail.AvailableEvents)
	}

	// retry, fail again with a non ProcessError
	if _, err := newTestService(repo, registry, &fakePublisher{}, testUserID).Retry(newGinContext(), id); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond) // updated_at decides the latest error
	if err := worker.Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	detail = getDetail(t, repo, registry, id)
	if detail.Attempt != 2 || detail.Error == nil || detail.Error.Code != "internal_error" {
		t.Fatalf("after retry = %+v error %+v", detail.FileManagementItem, detail.Error)
	}

	metas, _ := repo.GetMetas(context.Background(), id)
	first := metaValues(metas, MetaGroupError, attemptMetaType(1))
	if first[metaKeyErrorCode] != "third_party_error" || first[metaKeyErrorInternal] != "HTTP 502: upstream secret" {
		t.Errorf("attempt 1 error meta = %v", first)
	}
	if second := metaValues(metas, MetaGroupError, attemptMetaType(2)); second[metaKeyErrorCode] != "internal_error" {
		t.Errorf("attempt 2 error meta = %v", second)
	}
}

func TestWorkerDropsResultWhenCancelledDuringProcessing(t *testing.T) {
	repo, db := newTestRepo(t)

	var id uint64
	registry := NewRegistry(fakeProcessor{process: func(ctx context.Context, job Job) (*Result, error) {
		if _, err := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, testUserID).Cancel(newGinContext(), id); err != nil {
			t.Errorf("cancel during processing: %v", err)
		}
		return &Result{Output: map[string]string{"summary": "late"}}, nil
	}})

	id = createJob(t, repo, registry, createStorage(t, db, testUserID, "application/pdf", "report.pdf"))
	if err := newTestWorker(t, repo, registry, "pdf").Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	detail := getDetail(t, repo, registry, id)
	if detail.Status != constant.FileManagementCancelled {
		t.Errorf("status = %d, want cancelled", detail.Status)
	}
	if _, ok := detail.Meta[MetaGroupOutput]; ok {
		t.Errorf("late output was stored: %v", detail.Meta)
	}
}

func TestWorkerSkipsCancelledJob(t *testing.T) {
	repo, db := newTestRepo(t)

	called := false
	registry := NewRegistry(fakeProcessor{process: func(ctx context.Context, job Job) (*Result, error) {
		called = true
		return &Result{}, nil
	}})

	id := createJob(t, repo, registry, createStorage(t, db, testUserID, "application/pdf", "report.pdf"))
	if _, err := newTestService(repo, registry, &fakePublisher{}, testUserID).Cancel(newGinContext(), id); err != nil {
		t.Fatal(err)
	}

	if err := newTestWorker(t, repo, registry, "pdf").Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if called || getStatus(t, db, id) != constant.FileManagementCancelled {
		t.Errorf("cancelled job was processed (called %v)", called)
	}

	// unknown ids are skipped too
	if err := newTestWorker(t, repo, registry, "pdf").Process(context.Background(), 999); err != nil {
		t.Error(err)
	}
}

func TestWorkerFailsUnavailableStorageAndLargeFile(t *testing.T) {
	repo, db := newTestRepo(t)
	registry := NewRegistry(fakeProcessor{maxBytes: 4})

	storageModel := createStorage(t, db, testUserID, "application/pdf", "report.pdf")
	deletedID := createJob(t, repo, registry, storageModel)
	db.Model(&storageModel).Update("status", constant.StatusDeleted)

	if err := newTestWorker(t, repo, registry, "pdf").Process(context.Background(), deletedID); err != nil {
		t.Fatal(err)
	}
	if detail := getDetail(t, repo, registry, deletedID); detail.Error == nil || detail.Error.Code != "storage_unavailable" {
		t.Errorf("deleted storage error = %+v", detail.Error)
	}

	largeID := createJob(t, repo, registry, createStorage(t, db, testUserID, "application/pdf", "big.pdf"))
	if err := newTestWorker(t, repo, registry, "more than four bytes").Process(context.Background(), largeID); err != nil {
		t.Fatal(err)
	}
	if detail := getDetail(t, repo, registry, largeID); detail.Error == nil || detail.Error.Code != "file_too_large" {
		t.Errorf("large file error = %+v", detail.Error)
	}
}

func TestWorkerTimeout(t *testing.T) {
	repo, db := newTestRepo(t)
	registry := NewRegistry(fakeProcessor{process: func(ctx context.Context, job Job) (*Result, error) {
		<-ctx.Done()
		return nil, &ProcessError{Code: "third_party_timeout", Message: "too long", Internal: ctx.Err().Error()}
	}})

	id := createJob(t, repo, registry, createStorage(t, db, testUserID, "application/pdf", "report.pdf"))
	worker := newTestWorker(t, repo, registry, "pdf")
	worker.timeout = 20 * time.Millisecond

	if err := worker.Process(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if detail := getDetail(t, repo, registry, id); detail.Error == nil || detail.Error.Code != "third_party_timeout" {
		t.Errorf("error = %+v", detail.Error)
	}
}
