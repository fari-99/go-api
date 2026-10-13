package file_managements

import (
	"fmt"
	"sort"

	"go-api/modules/models"
)

// Meta groups and types, see docs/prd/file-managements.md §5.2
const (
	MetaGroupInput   = "input"
	MetaGroupOutput  = "output"
	MetaGroupRequest = "request" // internal, never returned to users
	MetaGroupError   = "error"

	MetaTypeParams       = "params"
	MetaTypeThirdParty   = "third_party"
	MetaTypeQueuePublish = "queue_publish"

	metaKeyErrorCode     = "code"
	metaKeyErrorMessage  = "message"
	metaKeyErrorInternal = "internal_message" // internal, never returned to users
)

// visibleMetaGroups are returned as-is in the detail response
var visibleMetaGroups = map[string]bool{MetaGroupInput: true, MetaGroupOutput: true}

func attemptMetaType(attempt int) string {
	return fmt.Sprintf("attempt_%d", attempt)
}

// buildMetas turns values into meta rows of one group and type, sorted by key.
func buildMetas(fileManagementID uint64, group, metaType string, values map[string]string, createdBy uint64) []models.FileManagementMeta {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	metas := make([]models.FileManagementMeta, 0, len(keys))
	for _, key := range keys {
		metas = append(metas, models.FileManagementMeta{
			FileManagementID: models.IDType(fileManagementID),
			Group:            group,
			Type:             metaType,
			Key:              key,
			Value:            values[key],
			CreatedBy:        models.IDType(createdBy),
		})
	}

	return metas
}

func errorMetas(fileManagementID uint64, metaType string, processErr *ProcessError) []models.FileManagementMeta {
	values := map[string]string{
		metaKeyErrorCode:    processErr.Code,
		metaKeyErrorMessage: processErr.Message,
	}
	if processErr.Internal != "" {
		values[metaKeyErrorInternal] = processErr.Internal
	}

	return buildMetas(fileManagementID, MetaGroupError, metaType, values, systemActorID)
}

// groupVisibleMetas builds group -> type -> key -> value for the groups users may see.
func groupVisibleMetas(metas []models.FileManagementMeta) map[string]map[string]map[string]string {
	grouped := map[string]map[string]map[string]string{}
	for _, meta := range metas {
		if !visibleMetaGroups[meta.Group] {
			continue
		}

		if grouped[meta.Group] == nil {
			grouped[meta.Group] = map[string]map[string]string{}
		}
		if grouped[meta.Group][meta.Type] == nil {
			grouped[meta.Group][meta.Type] = map[string]string{}
		}
		grouped[meta.Group][meta.Type][meta.Key] = meta.Value
	}

	return grouped
}

func metaValues(metas []models.FileManagementMeta, group, metaType string) map[string]string {
	values := map[string]string{}
	for _, meta := range metas {
		if meta.Group == group && meta.Type == metaType {
			values[meta.Key] = meta.Value
		}
	}
	return values
}

// latestError returns the code and message of the most recently written error.
func latestError(metas []models.FileManagementMeta) *ErrorInfo {
	var latestType string
	var latest models.FileManagementMeta
	for _, meta := range metas {
		if meta.Group != MetaGroupError || meta.Key != metaKeyErrorCode {
			continue
		}
		if latestType == "" || meta.UpdatedAt.After(latest.UpdatedAt) ||
			meta.UpdatedAt.Equal(latest.UpdatedAt) && meta.ID > latest.ID {
			latestType, latest = meta.Type, meta
		}
	}

	if latestType == "" {
		return nil
	}

	values := metaValues(metas, MetaGroupError, latestType)
	return &ErrorInfo{Code: values[metaKeyErrorCode], Message: values[metaKeyErrorMessage]}
}
