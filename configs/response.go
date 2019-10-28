package configs

import "github.com/kataras/iris"

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
		StatusCode: iris.StatusInternalServerError,
		Success:    false,
		Error: iris.Map{
			"message": errMsg,
		},
	}

	return defaultError
}

func NewResponse(ctx iris.Context, statusCode int, data interface{}) (int, error) {
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

	// adding status code
	ctx.StatusCode(statusCode)

	// send data as json
	n, err := ctx.JSON(response)
	if err != nil {
		n, err = ctx.JSON(getDefaultError(err))
	}
	return n, err
}
