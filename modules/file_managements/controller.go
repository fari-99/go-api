package file_managements

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
)

type controller struct {
	service Service
}

// respondError answers expected errors with their status and code, anything else with a generic 500.
func respondError(ctx *gin.Context, err error, fallbackMessage string) {
	var serviceErr *Error
	if errors.As(err, &serviceErr) {
		helpers.NewResponse(ctx, serviceErr.Status, gin.H{
			"error_code":    serviceErr.Code,
			"error_message": serviceErr.Message,
		})
		return
	}

	log.Printf("file managements: %s, err := %s", fallbackMessage, err.Error())
	helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
		"error_message": fallbackMessage,
	})
}

func fileManagementIDParam(ctx *gin.Context) (uint64, bool) {
	fileManagementID, err := strconv.ParseUint(ctx.Param("fileManagementID"), 10, 64)
	if err != nil || fileManagementID == 0 {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{"error_message": "fileManagementID is not valid"})
		return 0, false
	}
	return fileManagementID, true
}

// ActionsAction godoc
// @Summary      List available file management actions
// @Tags         file-managements
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  helpers.Response
// @Router       /file-managements/actions [get]
func (c controller) ActionsAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"items": c.service.GetActions(ctx),
	})
}

// CreateAction godoc
// @Summary      Create a file management job for an uploaded storage file
// @Description  Returns 201 for a new job, or 200 with the existing job when an identical one is still new / on progress.
// @Tags         file-managements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateFileManagementRequest  true  "storage_id, action and params"
// @Success      200  {object}  helpers.Response
// @Success      201  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Failure      422  {object}  helpers.Response
// @Failure      429  {object}  helpers.Response
// @Router       /file-managements [post]
func (c controller) CreateAction(ctx *gin.Context) {
	var input CreateFileManagementRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "invalid request body",
		})
		return
	}

	if err := input.Validate(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err,
			"error_message": "invalid request body",
		})
		return
	}

	detail, created, err := c.service.Create(ctx, input)
	if err != nil {
		respondError(ctx, err, "failed to create file management")
		return
	}

	statusCode := http.StatusOK
	if created {
		statusCode = http.StatusCreated
	}

	helpers.NewResponse(ctx, statusCode, detail)
}

// ListAction godoc
// @Summary      List file management jobs of the current user
// @Tags         file-managements
// @Produce      json
// @Security     BearerAuth
// @Param        page  query  int  false  "Page (default 1)"
// @Param        limit  query  int  false  "Page size (default 20, max 100)"
// @Param        status  query  int  false  "1 new, 2 on progress, 3 failed, 4 cancelled, 5 completed"
// @Param        action  query  string  false  "Action name"
// @Param        storage_id  query  int  false  "Storage ID"
// @Success      200  {object}  helpers.Response
// @Failure      500  {object}  helpers.Response
// @Router       /file-managements [get]
func (c controller) ListAction(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status, _ := strconv.ParseInt(ctx.Query("status"), 10, 8)
	storageID, _ := strconv.ParseUint(ctx.Query("storage_id"), 10, 64)

	items, paginatorData, err := c.service.GetList(ctx, ListFilter{
		Page:      page,
		Limit:     limit,
		Status:    int8(status),
		Action:    ctx.Query("action"),
		StorageID: storageID,
	})
	if err != nil {
		respondError(ctx, err, "failed to list file managements")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"paginator": paginatorData,
		"items":     items,
	})
}

// DetailAction godoc
// @Summary      Get a file management job with its input, output and error
// @Tags         file-managements
// @Produce      json
// @Security     BearerAuth
// @Param        fileManagementID  path  int  true  "File management ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /file-managements/{fileManagementID} [get]
func (c controller) DetailAction(ctx *gin.Context) {
	fileManagementID, ok := fileManagementIDParam(ctx)
	if !ok {
		return
	}

	detail, err := c.service.GetDetail(ctx, fileManagementID)
	if err != nil {
		respondError(ctx, err, "failed to get file management")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
}

// CancelAction godoc
// @Summary      Cancel a new or on progress file management job
// @Tags         file-managements
// @Produce      json
// @Security     BearerAuth
// @Param        fileManagementID  path  int  true  "File management ID"
// @Success      200  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Failure      409  {object}  helpers.Response
// @Router       /file-managements/{fileManagementID}/cancel [post]
func (c controller) CancelAction(ctx *gin.Context) {
	fileManagementID, ok := fileManagementIDParam(ctx)
	if !ok {
		return
	}

	detail, err := c.service.Cancel(ctx, fileManagementID)
	if err != nil {
		respondError(ctx, err, "failed to cancel file management")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
}

// RetryAction godoc
// @Summary      Retry a failed or cancelled file management job
// @Tags         file-managements
// @Produce      json
// @Security     BearerAuth
// @Param        fileManagementID  path  int  true  "File management ID"
// @Success      200  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Failure      409  {object}  helpers.Response
// @Failure      429  {object}  helpers.Response
// @Router       /file-managements/{fileManagementID}/retry [post]
func (c controller) RetryAction(ctx *gin.Context) {
	fileManagementID, ok := fileManagementIDParam(ctx)
	if !ok {
		return
	}

	detail, err := c.service.Retry(ctx, fileManagementID)
	if err != nil {
		respondError(ctx, err, "failed to retry file management")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
}

// DeleteAction godoc
// @Summary      Delete a completed, failed or cancelled file management job
// @Tags         file-managements
// @Produce      json
// @Security     BearerAuth
// @Param        fileManagementID  path  int  true  "File management ID"
// @Success      200  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Failure      409  {object}  helpers.Response
// @Router       /file-managements/{fileManagementID} [delete]
func (c controller) DeleteAction(ctx *gin.Context) {
	fileManagementID, ok := fileManagementIDParam(ctx)
	if !ok {
		return
	}

	if err := c.service.Delete(ctx, fileManagementID); err != nil {
		respondError(ctx, err, "failed to delete file management")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{"message": "file management deleted"})
}
