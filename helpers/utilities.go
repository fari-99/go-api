package helpers

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
)

func ParamsDefault(ctx *gin.Context, key string, defaultValue string) string {
	value, exists := ctx.Params.Get(key)
	if !exists {
		value = defaultValue
	}

	return value
}

func LoggingMessage(message string, data interface{}) {
	if data == nil {
		return
	}

	dataMarshal, _ := json.Marshal(data)
	log.Printf("%s, Data := %s", message, string(dataMarshal))
}

func Recover(message string) {
	if r := recover(); r != nil {
		LoggingMessage(message, r)
	}

	return
}
