package helpers

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
		log.Printf(message)
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

func PasswordAuth(password string, inputPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(inputPassword))
	return err
}
