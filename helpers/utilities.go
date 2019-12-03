package helpers

import (
	"encoding/json"
	"errors"
	"go-api/models"
	"log"
	"os"
	"reflect"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// Check JSON string type
func IsJSONString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil
}

// Check JSON type
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// GeneratePassword generates bcrypt hash string of the given plaintext password
func GeneratePassword(password string) (hash string, err error) {
	passwordCost, err := strconv.ParseInt(os.Getenv("PASSWORD_COST"), 10, 8)
	if err != nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), int(passwordCost))
	if err != nil {
		return
	}

	hash = string(hashedPassword)
	return
}

func AuthenticatePassword(customerModel *models.Customers, password string) error {
	errCheck := bcrypt.CompareHashAndPassword([]byte(customerModel.Password), []byte(password))
	if errCheck != nil {
		return errCheck
	}

	return nil
}

// InArray checks whether needle is in haystack.
func InArray(needle interface{}, haystack interface{}) (bool, int, error) {
	haystackValue := reflect.ValueOf(haystack)
	haystackType := haystackValue.Type()

	if haystackType.Kind() != reflect.Array && haystackType.Kind() != reflect.Slice {
		err := errors.New("Parameter 2 is not an array or slice")
		return false, -1, err
	}

	switch haystackType.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < haystackValue.Len(); i++ {
			hayVal := haystackValue.Index(i).Interface()

			if reflect.DeepEqual(hayVal, needle) {
				return true, i, nil
			}
		}
	}

	return false, -1, nil
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
