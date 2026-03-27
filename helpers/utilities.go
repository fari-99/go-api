package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/goodsign/monday"
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

func ToLocale(t *time.Time, format string) (string, error) {
	if t == nil {
		return "", fmt.Errorf("time is null")
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return "", fmt.Errorf("failed to load location, %w", err)
	}

	locale := monday.LocaleIdID
	formatted := monday.Format(t.In(loc), format, monday.Locale(locale))

	return formatted, nil
}
