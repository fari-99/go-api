package helpers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	StatusCode int         `json:"status"`
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      interface{} `json:"error,omitempty"`
}

func isSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func getDefaultError(err error) Response {
	errMsg := "Error send response data"
	if err != nil {
		errMsg = err.Error()
	}

	defaultError := Response{
		StatusCode: http.StatusInternalServerError,
		Success:    false,
		Error: gin.H{
			"message": errMsg,
		},
	}

	return defaultError
}

func NewResponse(ctx *gin.Context, statusCode int, data interface{}) {
	response := Response{
		StatusCode: statusCode,
		Success:    isSuccess(statusCode),
	}

	// check if data is error or success
	if response.Success {
		response.Data = data
	} else {
		response.Error = data
	}

	// send data as json
	ctx.JSON(statusCode, response)
}
