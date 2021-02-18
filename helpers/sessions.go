package helpers

import (
	"encoding/json"
	"fmt"
	"go-api/configs"
	"go-api/models"
	"log"
	"os"
	"time"
)

type SessionToken struct {
	AccessExpiredAt  int64  `json:"access_expired_at"`
	AccessUuid       string `json:"access_uuid"`
	RefreshExpiredAt int64  `json:"refresh_expired_at"`
	RefreshUuid      string `json:"refresh_uuid"`
}

type SessionData struct {
	Token SessionToken

	UserID        int64       `json:"user_id"`
	UserDetails   interface{} `json:"user_details"`
	Authorization bool        `json:"authorization"`
}

func SetRedisSession(data SessionData) error {
	redisSession := configs.GetRedisSessionConfig()
	dataMarshal, _ := json.Marshal(data.UserDetails)

	accessExpired := time.Unix(data.Token.AccessExpiredAt, 0)
	refreshExpired := time.Unix(data.Token.RefreshExpiredAt, 0)
	now := time.Now()

	err := redisSession.Set(data.Token.AccessUuid, string(dataMarshal), accessExpired.Sub(now)).Err()
	if err != nil {
		return fmt.Errorf("error set redis session access token, err := %s", err.Error())
	}

	err = redisSession.Set(data.Token.RefreshUuid, string(dataMarshal), refreshExpired.Sub(now)).Err()
	if err != nil {
		return fmt.Errorf("error set redis session refresh token, err := %s", err.Error())
	}

	return nil
}

func GetCurrentUser(uuidIdentifier string) (models.Customers, error) {
	redisSession := configs.GetRedisSessionConfig()

	redisData, err := redisSession.Get(uuidIdentifier).Result()
	if err != nil {
		return models.Customers{}, err
	}

	var userData models.Customers
	_ = json.Unmarshal([]byte(redisData), &userData)

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
