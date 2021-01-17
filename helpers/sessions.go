package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"go-api/configs"
	"go-api/models"
	"log"
	"os"
	"time"
)

type SessionData struct {
	AccessUuid  string `json:"access_uuid"`
	RefreshUuid string `json:"refresh_uuid"`

	UserID        int64       `json:"user_id"`
	UserDetails   interface{} `json:"user_details"`
	Authorization bool        `json:"authorization"`
}

func SetRedisSession(data SessionData, ctx iris.Context) (redisSession *sessions.Session) {
	sessionConfig := configs.GetRedisSessionConfig()
	dataMarshal, _ := json.Marshal(data.UserDetails)

	s := sessionConfig.Start(ctx)
	s.Set(fmt.Sprintf("%s", data.AccessUuid), string(dataMarshal))
	s.Set(fmt.Sprintf("%s", data.RefreshUuid), string(dataMarshal))

	return s
}

func GetCurrentUser(uuidIdentifier string, ctx iris.Context) (models.Customers, error) {
	sessionConfig := configs.GetRedisSessionConfig()

	s := sessionConfig.Start(ctx)
	dataCookie := s.Get(fmt.Sprintf("%s", uuidIdentifier))

	if dataCookie == nil {
		return models.Customers{}, fmt.Errorf("cookie is not valid")
	}

	var userData models.Customers
	_ = json.Unmarshal([]byte(dataCookie.(string)), &userData)

	return userData, nil
}

func GetSessionDuration(lifetime int64) time.Duration {
	lifetimeType := os.Getenv("REDIS_SESSION_LIFETIME_TYPE")

	timeDuration := time.Duration(lifetime)
	switch lifetimeType {
	case "second":
		return timeDuration * time.Second
	case "minute":
		return timeDuration * time.Minute
	case "hour":
		return timeDuration * time.Hour
	default:
		log.Fatal(fmt.Sprintf("session '%s' is not supported", lifetimeType))
	}

	return 0
}
