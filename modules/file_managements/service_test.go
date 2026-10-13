package file_managements

import (
	"errors"
	"reflect"
	"testing"

	"go-api/constant"
	"go-api/modules/models"
)

func TestCreateQueuesJobWithInputMeta(t *testing.T) {
	repo, db := newTestRepo(t)
	publisher := &fakePublisher{}
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), publisher, testUserID)
	storageModel := createStorage(t, db, testUserID, "application/pdf", "report.pdf")

	detail, created, err := svc.Create(newGinContext(), CreateFileManagementRequest{
		StorageID: storageModel.ID.Uint64(),
		Action:    ActionPdfSummary,
		Params:    map[string]any{"length": "short"},
	})
	if err != nil || !created {
		t.Fatalf("create = %v, created %v", err, created)
	}

	if detail.Status != constant.FileManagementNew || detail.StatusLabel != "New" || detail.CreatedBy != testUserID {
		t.Errorf("detail = %+v", detail.FileManagementItem)
	}
	if !reflect.DeepEqual(detail.AvailableEvents, []string{EventCancel}) {
		t.Errorf("available events = %v", detail.AvailableEvents)
	}
	if detail.Storage == nil || detail.Storage.OriginalFilename != "report.pdf" {
		t.Errorf("storage = %+v", detail.Storage)
	}
	if got := detail.Meta[MetaGroupInput][MetaTypeParams]; !reflect.DeepEqual(got, map[string]string{"length": "short"}) {
		t.Errorf("input meta = %v", got)
	}
	if ids := publisher.ids(); !reflect.DeepEqual(ids, []uint64{detail.ID.Uint64()}) {
		t.Errorf("published = %v", ids)
	}
}

func TestCreateReturnsExistingActiveJob(t *testing.T) {
	repo, db := newTestRepo(t)
	publisher := &fakePublisher{}
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), publisher, testUserID)
	storageModel := createStorage(t, db, testUserID, "application/pdf", "report.pdf")

	input := CreateFileManagementRequest{StorageID: storageModel.ID.Uint64(), Action: ActionPdfSummary, Params: map[string]any{"length": "short"}}
	first, _, err := svc.Create(newGinContext(), input)
	if err != nil {
		t.Fatal(err)
	}

	again, created, err := svc.Create(newGinContext(), input)
	if err != nil || created || again.ID != first.ID {
		t.Fatalf("second create = %v, created %v, id %d want %d", err, created, again.ID, first.ID)
	}

	// different params is a different job
	other, created, err := svc.Create(newGinContext(), CreateFileManagementRequest{
		StorageID: storageModel.ID.Uint64(), Action: ActionPdfSummary, Params: map[string]any{"length": "long"},
	})
	if err != nil || !created || other.ID == first.ID {
		t.Fatalf("create with other params = %v, created %v", err, created)
	}

	// once the first one is finished, the same input creates a new job
	if _, err := svc.Cancel(newGinContext(), first.ID.Uint64()); err != nil {
		t.Fatal(err)
	}
	third, created, err := svc.Create(newGinContext(), input)
	if err != nil || !created || third.ID == first.ID {
		t.Fatalf("create after cancel = %v, created %v", err, created)
	}

	if got := len(publisher.ids()); got != 3 {
		t.Errorf("published %d jobs, want 3", got)
	}
}

func TestCreateValidation(t *testing.T) {
	repo, db := newTestRepo(t)
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, testUserID)

	pdf := createStorage(t, db, testUserID, "application/pdf", "report.pdf")
	image := createStorage(t, db, testUserID, "image/png", "photo.png")
	othersPdf := createStorage(t, db, otherUserID, "application/pdf", "secret.pdf")
	deletedPdf := createStorage(t, db, testUserID, "application/pdf", "old.pdf")
	db.Model(&deletedPdf).Update("status", constant.StatusDeleted)

	cases := []struct {
		name  string
		input CreateFileManagementRequest
		code  string
	}{
		{"unknown action", CreateFileManagementRequest{StorageID: pdf.ID.Uint64(), Action: "ocr"}, "action_not_available"},
		{"missing storage", CreateFileManagementRequest{StorageID: 999, Action: ActionPdfSummary}, "storage_not_found"},
		{"other user's storage", CreateFileManagementRequest{StorageID: othersPdf.ID.Uint64(), Action: ActionPdfSummary}, "storage_not_found"},
		{"deleted storage", CreateFileManagementRequest{StorageID: deletedPdf.ID.Uint64(), Action: ActionPdfSummary}, "storage_not_found"},
		{"wrong file type", CreateFileManagementRequest{StorageID: image.ID.Uint64(), Action: ActionPdfSummary}, "unsupported_file_type"},
		{"invalid params", CreateFileManagementRequest{StorageID: pdf.ID.Uint64(), Action: ActionPdfSummary, Params: map[string]any{"length": "huge"}}, "invalid_params"},
	}

	for _, c := range cases {
		_, _, err := svc.Create(newGinContext(), c.input)
		t.Run(c.name, func(t *testing.T) { assertServiceError(t, err, c.code) })
	}

	var total int64
	db.Model(&models.FileManagements{}).Count(&total)
	if total != 0 {
		t.Errorf("%d jobs created by invalid requests", total)
	}
}

func TestCreateActiveCap(t *testing.T) {
	repo, db := newTestRepo(t)
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, testUserID)
	svc.maxActive = 2

	for i := 0; i < 2; i++ {
		storageModel := createStorage(t, db, testUserID, "application/pdf", "a.pdf")
		if _, _, err := svc.Create(newGinContext(), CreateFileManagementRequest{StorageID: storageModel.ID.Uint64(), Action: ActionPdfSummary}); err != nil {
			t.Fatal(err)
		}
	}

	storageModel := createStorage(t, db, testUserID, "application/pdf", "a.pdf")
	_, _, err := svc.Create(newGinContext(), CreateFileManagementRequest{StorageID: storageModel.ID.Uint64(), Action: ActionPdfSummary})
	assertServiceError(t, err, "too_many_active_jobs")
}

func TestCreatePublishFailureMarksFailed(t *testing.T) {
	repo, db := newTestRepo(t)
	publisher := &fakePublisher{err: errors.New("broker down")}
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), publisher, testUserID)
	storageModel := createStorage(t, db, testUserID, "application/pdf", "report.pdf")

	detail, created, err := svc.Create(newGinContext(), CreateFileManagementRequest{StorageID: storageModel.ID.Uint64(), Action: ActionPdfSummary})
	if err != nil || !created {
		t.Fatalf("create = %v", err)
	}

	if detail.Status != constant.FileManagementFailed || detail.FinishedAt == nil {
		t.Errorf("status = %d, finished_at %v", detail.Status, detail.FinishedAt)
	}
	if detail.Error == nil || detail.Error.Code != "queue_publish_failed" {
		t.Errorf("error = %+v", detail.Error)
	}

	// retry once the broker is back
	publisher.err = nil
	retried, err := svc.Retry(newGinContext(), detail.ID.Uint64())
	if err != nil || retried.Status != constant.FileManagementNew || retried.Error != nil || retried.FinishedAt != nil {
		t.Fatalf("retry = %v, %+v", err, retried)
	}
	if ids := publisher.ids(); !reflect.DeepEqual(ids, []uint64{detail.ID.Uint64()}) {
		t.Errorf("published = %v", ids)
	}
}

func TestCancelRetryDeleteRules(t *testing.T) {
	repo, db := newTestRepo(t)
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, testUserID)
	storageModel := createStorage(t, db, testUserID, "application/pdf", "report.pdf")

	detail, _, err := svc.Create(newGinContext(), CreateFileManagementRequest{StorageID: storageModel.ID.Uint64(), Action: ActionPdfSummary})
	if err != nil {
		t.Fatal(err)
	}
	id := detail.ID.Uint64()

	_, err = svc.Retry(newGinContext(), id)
	assertServiceError(t, err, "invalid_status_transition")

	err = svc.Delete(newGinContext(), id)
	assertServiceError(t, err, "file_management_active")

	cancelled, err := svc.Cancel(newGinContext(), id)
	if err != nil || cancelled.Status != constant.FileManagementCancelled || cancelled.FinishedAt == nil {
		t.Fatalf("cancel = %v, %+v", err, cancelled)
	}

	_, err = svc.Cancel(newGinContext(), id)
	assertServiceError(t, err, "invalid_status_transition")

	// other users can't see, cancel or delete it
	otherSvc := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, otherUserID)
	_, err = otherSvc.GetDetail(newGinContext(), id)
	assertServiceError(t, err, "file_management_not_found")
	err = otherSvc.Delete(newGinContext(), id)
	assertServiceError(t, err, "file_management_not_found")

	if err := svc.Delete(newGinContext(), id); err != nil {
		t.Fatal(err)
	}
	_, err = svc.GetDetail(newGinContext(), id)
	assertServiceError(t, err, "file_management_not_found")

	var metaCount int64
	db.Model(&models.FileManagementMeta{}).Where("file_management_id = ?", id).Count(&metaCount)
	if metaCount != 0 {
		t.Errorf("%d metas left after delete", metaCount)
	}
}

func TestGetListOnlyOwnJobs(t *testing.T) {
	repo, db := newTestRepo(t)
	svc := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, testUserID)
	otherSvc := newTestService(repo, NewRegistry(fakeProcessor{}), &fakePublisher{}, otherUserID)

	mine := createStorage(t, db, testUserID, "application/pdf", "a.pdf")
	theirs := createStorage(t, db, otherUserID, "application/pdf", "b.pdf")
	first, _, _ := svc.Create(newGinContext(), CreateFileManagementRequest{StorageID: mine.ID.Uint64(), Action: ActionPdfSummary})
	_, _, _ = svc.Create(newGinContext(), CreateFileManagementRequest{StorageID: mine.ID.Uint64(), Action: ActionPdfSummary, Params: map[string]any{"length": "long"}})
	_, _, _ = otherSvc.Create(newGinContext(), CreateFileManagementRequest{StorageID: theirs.ID.Uint64(), Action: ActionPdfSummary})
	_, _ = svc.Cancel(newGinContext(), first.ID.Uint64())

	items, _, err := svc.GetList(newGinContext(), ListFilter{Page: 1, Limit: 20})
	if err != nil || len(items) != 2 {
		t.Fatalf("list = %v, %d items", err, len(items))
	}

	items, _, err = svc.GetList(newGinContext(), ListFilter{Page: 1, Limit: 20, Status: constant.FileManagementCancelled})
	if err != nil || len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("filtered list = %v, %+v", err, items)
	}
}
