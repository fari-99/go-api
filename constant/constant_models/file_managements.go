package constant_models

import "go-api/constant"

func GetFileManagementStatuses() map[int]string {
	fileManagementStatus := make(map[int]string)

	fileManagementStatus[constant.FileManagementNew] = "New"
	fileManagementStatus[constant.FileManagementOnProgress] = "On Progress"
	fileManagementStatus[constant.FileManagementFailed] = "Failed"
	fileManagementStatus[constant.FileManagementCancelled] = "Cancelled"
	fileManagementStatus[constant.FileManagementCompleted] = "Completed"

	return fileManagementStatus
}
