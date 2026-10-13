package file_managements

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"net/http"
	"time"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"github.com/looplab/fsm"

	"go-api/constant"
	"go-api/constant/constant_models"
	"go-api/helpers"
	"go-api/modules/models"
)

// Error is an expected failure with the HTTP status and code returned to the client.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

var (
	ErrNotFound          = &Error{Status: http.StatusNotFound, Code: "file_management_not_found", Message: "file management not found"}
	ErrStorageNotFound   = &Error{Status: http.StatusNotFound, Code: "storage_not_found", Message: "storage not found"}
	ErrActionUnavailable = &Error{Status: http.StatusUnprocessableEntity, Code: "action_not_available", Message: "action is not available"}
	ErrUnsupportedFile   = &Error{Status: http.StatusUnprocessableEntity, Code: "unsupported_file_type", Message: "file type is not supported by this action"}
	ErrTooManyActive     = &Error{Status: http.StatusTooManyRequests, Code: "too_many_active_jobs", Message: "too many file managements in progress, please wait until one finishes"}
	ErrInvalidTransition = &Error{Status: http.StatusConflict, Code: "invalid_status_transition", Message: "this action is not allowed in the current status"}
)

// systemActorID is the activity log actor for writes made by the worker
const systemActorID = 0

type Service interface {
	GetActions(ctx *gin.Context) []ActionInfo
	// Create returns created=false when an identical job is already new / on progress
	Create(ctx *gin.Context, input CreateFileManagementRequest) (detail *FileManagementDetail, created bool, err error)
	GetList(ctx *gin.Context, filter ListFilter) ([]FileManagementItem, *paginator.Pagination, error)
	GetDetail(ctx *gin.Context, fileManagementID uint64) (*FileManagementDetail, error)
	Cancel(ctx *gin.Context, fileManagementID uint64) (*FileManagementDetail, error)
	Retry(ctx *gin.Context, fileManagementID uint64) (*FileManagementDetail, error)
	Delete(ctx *gin.Context, fileManagementID uint64) error
}

type service struct {
	repo          Repository
	registry      *Registry
	publisher     Publisher
	maxActive     int64
	currentUserID func(ctx *gin.Context) (uint64, error)
}

func NewService(repo Repository, registry *Registry, publisher Publisher) Service {
	return service{
		repo:          repo,
		registry:      registry,
		publisher:     publisher,
		maxActive:     int64(envInt("FILE_MANAGEMENT_MAX_ACTIVE_PER_USER", 5)),
		currentUserID: currentUserID,
	}
}

// requestContext is the context handed to the repository. gin reuses *gin.Context after the
// request ends, so it must not be passed to database/sql, which can read it later.
func requestContext(ctx *gin.Context) context.Context {
	if ctx.Request != nil {
		return ctx.Request.Context()
	}
	return context.Background()
}

func currentUserID(ctx *gin.Context) (uint64, error) {
	uuid, _ := ctx.Get("uuid")
	uuidString, _ := uuid.(string)

	currentUser, err := helpers.GetCurrentUser(ctx, uuidString)
	if err != nil {
		return 0, err
	}

	return currentUser.ID.Uint64(), nil
}

func (s service) GetActions(ctx *gin.Context) []ActionInfo {
	actions := []ActionInfo{}
	for _, processor := range s.registry.List() {
		actions = append(actions, ActionInfo{
			Action:             processor.Action(),
			AcceptedMimes:      processor.AcceptedMimes(),
			AcceptedExtensions: processor.AcceptedExtensions(),
			MaxFileBytes:       processor.MaxFileBytes(),
			Params:             processor.Params(),
		})
	}
	return actions
}

func (s service) Create(ctx *gin.Context, input CreateFileManagementRequest) (*FileManagementDetail, bool, error) {
	userID, err := s.currentUserID(ctx)
	if err != nil {
		return nil, false, err
	}

	processor, ok := s.registry.Get(input.Action)
	if !ok {
		return nil, false, ErrActionUnavailable
	}

	// someone else's file is reported as not found so ids can't be probed
	storageModel, notFound, err := s.repo.GetStorage(requestContext(ctx), input.StorageID)
	if err != nil {
		return nil, false, err
	} else if notFound || storageModel.CreatedBy.Uint64() != userID || storageModel.Status != constant.StatusActive {
		return nil, false, ErrStorageNotFound
	}

	if !acceptsFile(processor, *storageModel) {
		return nil, false, ErrUnsupportedFile
	}

	params, err := processor.ValidateParams(input.Params)
	if err != nil {
		return nil, false, &Error{Status: http.StatusUnprocessableEntity, Code: "invalid_params", Message: err.Error()}
	}

	existing, err := s.findDuplicate(ctx, userID, input.StorageID, input.Action, params)
	if err != nil {
		return nil, false, err
	} else if existing != nil {
		detail, err := s.buildDetail(ctx, *existing, storageModel)
		return detail, false, err
	}

	totalActive, err := s.repo.CountActive(requestContext(ctx), userID)
	if err != nil {
		return nil, false, err
	} else if totalActive >= s.maxActive {
		return nil, false, ErrTooManyActive
	}

	fileManagement := models.FileManagements{
		StorageID: models.IDType(input.StorageID),
		Action:    input.Action,
		Status:    constant.FileManagementNew,
		CreatedBy: models.IDType(userID),
	}

	metas := buildMetas(0, MetaGroupInput, MetaTypeParams, params, userID)
	if err := s.repo.Create(requestContext(ctx), &fileManagement, metas, userID); err != nil {
		return nil, false, err
	}

	s.publish(ctx, &fileManagement, userID)

	detail, err := s.buildDetail(ctx, fileManagement, storageModel)
	return detail, true, err
}

// findDuplicate returns the user's new / on progress job with the same storage, action and params.
func (s service) findDuplicate(ctx *gin.Context, userID, storageID uint64, action string, params map[string]string) (*models.FileManagements, error) {
	candidates, err := s.repo.GetActive(requestContext(ctx), userID, storageID, action)
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		metas, err := s.repo.GetMetas(requestContext(ctx), candidate.ID.Uint64())
		if err != nil {
			return nil, err
		}

		if maps.Equal(metaValues(metas, MetaGroupInput, MetaTypeParams), params) {
			return &candidate, nil
		}
	}

	return nil, nil
}

// publish queues a new job. When queueing fails the job is marked failed so the user can retry it.
func (s service) publish(ctx *gin.Context, fileManagement *models.FileManagements, userID uint64) {
	publishErr := s.publisher.Publish(fileManagement.ID.Uint64())
	if publishErr == nil {
		return
	}

	log.Printf("file managements: failed to queue %d, err := %s", fileManagement.ID, publishErr.Error())

	to, err := nextStatus(ctx, fileManagement.Status, EventFail)
	if err != nil {
		return
	}

	now := time.Now()
	metas := errorMetas(fileManagement.ID.Uint64(), MetaTypeQueuePublish, &ProcessError{
		Code:     "queue_publish_failed",
		Message:  "failed to queue the file for processing, please retry",
		Internal: publishErr.Error(),
	})

	err = s.repo.Transition(requestContext(ctx), fileManagement.ID.Uint64(), fileManagement.Status, map[string]interface{}{
		"status":      to,
		"finished_at": now,
	}, metas, userID)
	if err != nil {
		log.Printf("file managements: failed to mark %d as failed, err := %s", fileManagement.ID, err.Error())
		return
	}

	fileManagement.Status = to
	fileManagement.FinishedAt = &now
}

func (s service) GetList(ctx *gin.Context, filter ListFilter) ([]FileManagementItem, *paginator.Pagination, error) {
	userID, err := s.currentUserID(ctx)
	if err != nil {
		return nil, nil, err
	}

	filter.CreatedBy = userID
	fileManagements, pagination, err := s.repo.GetList(requestContext(ctx), filter)
	if err != nil {
		return nil, nil, err
	}

	items := make([]FileManagementItem, 0, len(fileManagements))
	for _, fileManagement := range fileManagements {
		items = append(items, buildItem(fileManagement))
	}

	return items, pagination, nil
}

func (s service) GetDetail(ctx *gin.Context, fileManagementID uint64) (*FileManagementDetail, error) {
	fileManagement, err := s.getOwned(ctx, fileManagementID)
	if err != nil {
		return nil, err
	}

	return s.buildDetail(ctx, *fileManagement, nil)
}

func (s service) Cancel(ctx *gin.Context, fileManagementID uint64) (*FileManagementDetail, error) {
	return s.transition(ctx, fileManagementID, EventCancel, func(updates map[string]interface{}) {
		updates["finished_at"] = time.Now()
	}, false)
}

func (s service) Retry(ctx *gin.Context, fileManagementID uint64) (*FileManagementDetail, error) {
	userID, err := s.currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	fileManagement, err := s.getOwned(ctx, fileManagementID)
	if err != nil {
		return nil, err
	}

	// the action may have been removed since the job was created
	if _, ok := s.registry.Get(fileManagement.Action); !ok {
		return nil, ErrActionUnavailable
	}

	totalActive, err := s.repo.CountActive(requestContext(ctx), userID)
	if err != nil {
		return nil, err
	} else if totalActive >= s.maxActive {
		return nil, ErrTooManyActive
	}

	return s.transition(ctx, fileManagementID, EventRetry, func(updates map[string]interface{}) {
		updates["started_at"] = nil
		updates["finished_at"] = nil
	}, true)
}

// transition fires event on an owned job. setUpdates adds the columns to change besides status.
func (s service) transition(ctx *gin.Context, fileManagementID uint64, event string, setUpdates func(map[string]interface{}), publish bool) (*FileManagementDetail, error) {
	userID, err := s.currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	fileManagement, err := s.getOwned(ctx, fileManagementID)
	if err != nil {
		return nil, err
	}

	to, err := nextStatus(ctx, fileManagement.Status, event)
	if err != nil {
		return nil, transitionError(err)
	}

	updates := map[string]interface{}{"status": to}
	setUpdates(updates)

	err = s.repo.Transition(requestContext(ctx), fileManagementID, fileManagement.Status, updates, nil, userID)
	if errors.Is(err, ErrTransitionLost) {
		return nil, ErrInvalidTransition
	} else if err != nil {
		return nil, err
	}

	fileManagement, err = s.getOwned(ctx, fileManagementID)
	if err != nil {
		return nil, err
	}

	if publish {
		s.publish(ctx, fileManagement, userID)
	}

	return s.buildDetail(ctx, *fileManagement, nil)
}

func (s service) Delete(ctx *gin.Context, fileManagementID uint64) error {
	userID, err := s.currentUserID(ctx)
	if err != nil {
		return err
	}

	fileManagement, err := s.getOwned(ctx, fileManagementID)
	if err != nil {
		return err
	}

	if !isTerminalStatus(fileManagement.Status) {
		return &Error{Status: http.StatusConflict, Code: "file_management_active", Message: "cancel the file management before deleting it"}
	}

	err = s.repo.Delete(requestContext(ctx), fileManagementID, userID)
	if errors.Is(err, ErrTransitionLost) {
		return ErrInvalidTransition
	}
	return err
}

// getOwned loads a job of the current user. Someone else's job is reported as not found.
func (s service) getOwned(ctx *gin.Context, fileManagementID uint64) (*models.FileManagements, error) {
	userID, err := s.currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	fileManagement, notFound, err := s.repo.GetDetail(requestContext(ctx), fileManagementID)
	if err != nil {
		return nil, err
	} else if notFound || fileManagement.CreatedBy.Uint64() != userID {
		return nil, ErrNotFound
	}

	return fileManagement, nil
}

// buildDetail adds the storage summary and the metas users may see. storageModel is loaded when nil.
func (s service) buildDetail(ctx *gin.Context, fileManagement models.FileManagements, storageModel *models.Storages) (*FileManagementDetail, error) {
	metas, err := s.repo.GetMetas(requestContext(ctx), fileManagement.ID.Uint64())
	if err != nil {
		return nil, err
	}

	if storageModel == nil {
		storageModel, _, err = s.repo.GetStorage(requestContext(ctx), fileManagement.StorageID.Uint64())
		if err != nil {
			return nil, err
		}
	}

	detail := &FileManagementDetail{
		FileManagementItem: buildItem(fileManagement),
		Meta:               groupVisibleMetas(metas),
	}

	if storageModel != nil {
		detail.Storage = &StorageSummary{
			ID:               storageModel.ID,
			OriginalFilename: storageModel.OriginalFilename,
			Mime:             storageModel.Mime,
		}
	}

	if fileManagement.Status == constant.FileManagementFailed {
		detail.Error = latestError(metas)
	}

	return detail, nil
}

func buildItem(fileManagement models.FileManagements) FileManagementItem {
	return FileManagementItem{
		ID:              fileManagement.ID,
		StorageID:       fileManagement.StorageID,
		Action:          fileManagement.Action,
		Status:          fileManagement.Status,
		StatusLabel:     constant_models.GetFileManagementStatuses()[int(fileManagement.Status)],
		AvailableEvents: userEvents(fileManagement.Status),
		Attempt:         fileManagement.Attempt,
		StartedAt:       fileManagement.StartedAt,
		FinishedAt:      fileManagement.FinishedAt,
		CreatedBy:       fileManagement.CreatedBy,
		CreatedAt:       fileManagement.CreatedAt,
		UpdatedAt:       fileManagement.UpdatedAt,
	}
}

// userEvents are the events the user can fire from status (process/complete/fail belong to the worker).
func userEvents(status int8) []string {
	events := []string{}
	for _, event := range availableEvents(status) {
		if event == EventCancel || event == EventRetry {
			events = append(events, event)
		}
	}
	return events
}

func transitionError(err error) error {
	var invalidEvent fsm.InvalidEventError
	if errors.As(err, &invalidEvent) {
		return ErrInvalidTransition
	}
	return fmt.Errorf("file management transition: %w", err)
}
