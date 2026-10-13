package file_managements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fari-99/go-helper/rabbitmq"

	"go-api/constant"
	"go-api/modules/models"
	"go-api/modules/storages"
)

// FileOpener opens the bytes of a storage record; close must always be called.
type FileOpener func(storageModel models.Storages) (file *os.File, close func(), err error)

// Worker processes queued jobs, see docs/prd/file-managements.md §5.7
type Worker struct {
	repo     Repository
	registry *Registry
	openFile FileOpener
	timeout  time.Duration
}

func NewWorker(repo Repository, registry *Registry) *Worker {
	return &Worker{
		repo:     repo,
		registry: registry,
		openFile: storages.OpenStoredFile,
		timeout:  envDuration("FILE_MANAGEMENT_PROCESS_TIMEOUT_SECONDS", 300*time.Second),
	}
}

// Handle is the RabbitMQ consumer handler. Errors are logged and the message is acked: a job
// that can't be processed is either marked failed or left for the user to cancel and retry.
func (w *Worker) Handle(body rabbitmq.ConsumerHandlerData) {
	if body.EventType != constant.EventFileManagementProcess {
		log.Printf("file managements: there is no event [%s], ignoring event...", body.EventType)
		return
	}

	var payload QueuePayload
	dataMarshal, _ := json.Marshal(body.Data)
	if err := json.Unmarshal(dataMarshal, &payload); err != nil || payload.FileManagementID == 0 {
		log.Printf("file managements: invalid payload %s", dataMarshal)
		return
	}

	if err := w.Process(context.Background(), payload.FileManagementID); err != nil {
		log.Printf("file managements: error processing %d, err := %s", payload.FileManagementID, err.Error())
	}
}

// Process runs one job. A job that is no longer new (cancelled, deleted, taken by another
// worker) is skipped without error.
func (w *Worker) Process(ctx context.Context, fileManagementID uint64) error {
	fileManagement, notFound, err := w.repo.GetDetail(ctx, fileManagementID)
	if err != nil {
		return err
	} else if notFound {
		log.Printf("file managements: %d not found, skipping", fileManagementID)
		return nil
	}

	to, err := nextStatus(ctx, fileManagement.Status, EventProcess)
	if err != nil {
		log.Printf("file managements: %d is not new (status %d), skipping", fileManagementID, fileManagement.Status)
		return nil
	}

	attempt := fileManagement.Attempt + 1
	err = w.repo.Transition(ctx, fileManagementID, fileManagement.Status, map[string]interface{}{
		"status":      to,
		"attempt":     attempt,
		"started_at":  time.Now(),
		"finished_at": nil,
	}, nil, systemActorID)
	if errors.Is(err, ErrTransitionLost) {
		log.Printf("file managements: %d changed before processing, skipping", fileManagementID)
		return nil
	} else if err != nil {
		return err
	}

	fileManagement.Status = to
	fileManagement.Attempt = attempt

	result, processErr := w.run(ctx, *fileManagement)
	if processErr != nil {
		return w.finish(ctx, *fileManagement, EventFail, result, processErr)
	}

	return w.finish(ctx, *fileManagement, EventComplete, result, nil)
}

func (w *Worker) run(ctx context.Context, fileManagement models.FileManagements) (*Result, *ProcessError) {
	processor, ok := w.registry.Get(fileManagement.Action)
	if !ok {
		return nil, &ProcessError{Code: "action_not_available", Message: "action is not available"}
	}

	storageModel, notFound, err := w.repo.GetStorage(ctx, fileManagement.StorageID.Uint64())
	if err != nil {
		return nil, &ProcessError{Code: "internal_error", Message: "failed to process the file", Internal: err.Error()}
	} else if notFound || storageModel.Status != constant.StatusActive {
		return nil, &ProcessError{Code: "storage_unavailable", Message: "the file is no longer available"}
	}

	metas, err := w.repo.GetMetas(ctx, fileManagement.ID.Uint64())
	if err != nil {
		return nil, &ProcessError{Code: "internal_error", Message: "failed to process the file", Internal: err.Error()}
	}

	file, closeFile, err := w.openFile(*storageModel)
	if err != nil {
		return nil, &ProcessError{Code: "storage_unavailable", Message: "the file is no longer available", Internal: err.Error()}
	}
	defer closeFile()

	stat, err := file.Stat()
	if err != nil {
		return nil, &ProcessError{Code: "internal_error", Message: "failed to process the file", Internal: err.Error()}
	} else if stat.Size() > processor.MaxFileBytes() {
		return nil, &ProcessError{
			Code:    "file_too_large",
			Message: fmt.Sprintf("the file is larger than %d MB", processor.MaxFileBytes()>>20),
		}
	}

	processCtx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	result, err := processor.Process(processCtx, Job{
		FileManagement: fileManagement,
		Storage:        *storageModel,
		File:           file,
		Params:         metaValues(metas, MetaGroupInput, MetaTypeParams),
	})
	if err != nil {
		var processErr *ProcessError
		if !errors.As(err, &processErr) {
			processErr = &ProcessError{Code: "internal_error", Message: "failed to process the file", Internal: err.Error()}
		}
		return result, processErr
	}

	return result, nil
}

// finish writes the outcome. When the job is no longer on progress (cancelled meanwhile) the
// result is dropped.
func (w *Worker) finish(ctx context.Context, fileManagement models.FileManagements, event string, result *Result, processErr *ProcessError) error {
	to, err := nextStatus(ctx, fileManagement.Status, event)
	if err != nil {
		return err
	}

	id := fileManagement.ID.Uint64()
	var metas []models.FileManagementMeta
	if result != nil {
		metas = append(metas, buildMetas(id, MetaGroupRequest, MetaTypeThirdParty, result.Request, systemActorID)...)
	}

	if processErr != nil {
		metas = append(metas, errorMetas(id, attemptMetaType(fileManagement.Attempt), processErr)...)
	} else if result != nil {
		processor, _ := w.registry.Get(fileManagement.Action)
		metas = append(metas, buildMetas(id, MetaGroupOutput, processor.OutputType(), result.Output, systemActorID)...)
	}

	err = w.repo.Transition(ctx, id, fileManagement.Status, map[string]interface{}{
		"status":      to,
		"finished_at": time.Now(),
	}, metas, systemActorID)
	if errors.Is(err, ErrTransitionLost) {
		log.Printf("file managements: %d is no longer on progress, dropping the result", id)
		return nil
	} else if err != nil {
		return err
	}

	if processErr != nil {
		log.Printf("file managements: %d failed on attempt %d, err := %s", id, fileManagement.Attempt, processErr.Error())
	}

	return nil
}
