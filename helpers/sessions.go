package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/cast"

	"go-api/modules/configs"
	"go-api/modules/models"
)

const accessTokenIndex = "access_token"
const refreshTokenIndex = "refresh_token"
const plusInfinite = "+inf"
const negativeInfinite = "-inf"

type SessionToken struct {
	Uuid             string `json:"uuid"`
	AccessExpiredAt  int64  `json:"access_expired_at"`
	RefreshExpiredAt int64  `json:"refresh_expired_at"`
}

type KeyRedisSessionData struct {
	KeyAccess       string
	KeyRefresh      string
	KeyTotalAccess  string
	KeyTotalRefresh string
}

type SessionData struct {
	Token SessionToken

	UserID        uint64      `json:"user_id"`
	UserDetails   interface{} `json:"user_details"`
	Authorization bool        `json:"authorization"`
}

func getKeyRedis(username string, uuid string) KeyRedisSessionData {
	return KeyRedisSessionData{
		KeyAccess:       fmt.Sprintf("%s:%s", uuid, accessTokenIndex),      // uuid:access_token
		KeyRefresh:      fmt.Sprintf("%s:%s", uuid, refreshTokenIndex),     // uuid:refresh_token
		KeyTotalAccess:  fmt.Sprintf("%s:%s", username, accessTokenIndex),  // username:access_token
		KeyTotalRefresh: fmt.Sprintf("%s:%s", username, refreshTokenIndex), // username:refresh_token
	}
}

func removeExpiredToken(redisSession *redis.Client, username string) (err error) {
	keyRedis := getKeyRedis(username, "")
	timeNow := cast.ToString(time.Now().Unix())

	// get all expired access token
	accessTokenUuids, err := redisSession.ZRangeByScore(keyRedis.KeyTotalAccess, redis.ZRangeBy{
		Min: negativeInfinite,
		Max: timeNow,
	}).Result()
	if err != nil {
		return err
	}

	// get all expired
	refreshTokenUuids, err := redisSession.ZRangeByScore(keyRedis.KeyTotalRefresh, redis.ZRangeBy{
		Min: negativeInfinite,
		Max: timeNow,
	}).Result()
	if err != nil {
		return err
	}

	if len(accessTokenUuids) > 0 {
		for _, accessTokenUuid := range accessTokenUuids {
			_, _ = RemoveRedisSession(username, accessTokenUuid)
		}
	}

	if len(refreshTokenUuids) > 0 {
		for _, refreshTokenUuid := range refreshTokenUuids {
			_, _ = RemoveRedisSession(username, refreshTokenUuid)
		}
	}

	return nil
}

func getTotalLogin(redisSession *redis.Client, username string) (totalLoginAccessToken int64, totalLoginRefreshToken int64, err error) {
	keyRedis := getKeyRedis(username, "")

	err = removeExpiredToken(redisSession, username)
	if err != nil {
		return 0, 0, err
	}

	totalLoginAccessToken, err = redisSession.ZCard(keyRedis.KeyTotalAccess).Result()
	if err != nil {
		return 0, 0, err
	}

	totalLoginRefreshToken, err = redisSession.ZCard(keyRedis.KeyTotalRefresh).Result()
	if err != nil {
		return 0, 0, err
	}

	return totalLoginAccessToken, totalLoginRefreshToken, nil
}

func getAllUuid(username string) (accessUuids []string, refreshUuids []string, err error) {
	keyRedis := getKeyRedis(username, "")
	redisSession := configs.GetRedisSessionConfig()

	err = removeExpiredToken(redisSession, username)
	if err != nil {
		return nil, nil, err
	}

	accessUuids, err = redisSession.ZRange(keyRedis.KeyTotalAccess, 0, -1).Result()
	if err != nil {
		return nil, nil, err
	}

	refreshUuids, err = redisSession.ZRange(keyRedis.KeyTotalRefresh, 0, -1).Result()
	if err != nil {
		return nil, nil, err
	}

	return accessUuids, refreshUuids, nil
}

func GetAllSessions(username string) ([]models.Users, error) {
	_, refreshUuids, err := getAllUuid(username)
	if err != nil {
		return nil, err
	}

	var users []models.Users
	for _, refreshUuid := range refreshUuids {
		user, err := GetCurrentUser(refreshUuid)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

func CheckToken(username, uuid string) (isExistAccess, isExistRefresh bool, err error) {
	keyRedis := getKeyRedis(username, uuid)
	redisSession := configs.GetRedisSessionConfig()

	resultAccess, err := redisSession.Exists(keyRedis.KeyAccess).Result()
	if err != nil {
		return false, false, err
	}

	resultRefresh, err := redisSession.Exists(keyRedis.KeyRefresh).Result()
	if err != nil {
		return false, false, err
	}

	return resultAccess > 0, resultRefresh > 0, nil
}

func SetupLoginSession(username string, data SessionData) (totalLogin int64, err error) {
	// check if login session > total session that allowed to login
	// if more, then return error that new session can't be created because you have device already connected
	// if less, then
	// 1. set redis session uuid (set uuid:access_token data) (set uuid:refresh_token data)
	// 2. put redis session uuid using zadd (zadd username:access_token uuid) (zadd username:refresh_token uuid)
	// 3. get total login using zcard (zcard username:access_token) (zcard username:refresh_token)
	// 4. return total login

	redisSession := configs.GetRedisSessionConfig()
	totalLoginAccessToken, _, err := getTotalLogin(redisSession, username)
	if err != nil {
		return 0, err
	}

	redisSession.Exists()

	if totalLoginAccessToken >= cast.ToInt64(os.Getenv("TOTAL_LOGIN_SESSION")) {
		return totalLogin, fmt.Errorf("total login session are more than allowed, logout one of your session from one of your device, or delete all sessions")
	}

	err = setRedisSession(username, data)
	if err != nil {
		return 0, err
	}

	totalLoginAccessToken, _, err = getTotalLogin(redisSession, username)
	if err != nil {
		return 0, err
	}

	return totalLoginAccessToken, err
}

func RemoveRedisSession(username, uuid string) (totalLogin int64, err error) {
	// 1. delete redis session uuid (del uuid:access_token) (del uuid:refresh_token)
	// 2. delete redis member using zrem (zrem username:access_token uuid) (zrem username:refresh_token uuid)
	// 3. get total login using zcard (zcard username:access_token) (zcard username:refresh_token)
	// 4. return total login

	keyRedis := getKeyRedis(username, uuid)

	redisSession := configs.GetRedisSessionConfig()
	err = redisSession.Del(keyRedis.KeyAccess).Err() // delete access token redis
	if err != nil {
		return 0, err
	}

	err = redisSession.Del(keyRedis.KeyRefresh).Err() // delete refresh token redis
	if err != nil {
		return 0, err
	}

	err = redisSession.ZRem(keyRedis.KeyTotalAccess, uuid).Err() // delete access token member redis
	if err != nil {
		return 0, err
	}

	err = redisSession.ZRem(keyRedis.KeyTotalRefresh, uuid).Err() // delete refresh token member redis
	if err != nil {
		return 0, err
	}

	totalLoginAccessToken, _, err := getTotalLogin(redisSession, username)
	if err != nil {
		return 0, err
	}

	return totalLoginAccessToken, nil
}

func DeleteAllSession(username string, uuid string) (err error) {
	// 1. get all members using zmembers (zmembers username:access_token) (zmembers username:refresh_token)
	// 2. delete redis session by -looping- uuid from smembers using del (del uuid:access_token) (del uuid:refresh_token)
	// 3. delete redis member using del (del username:access_token) (del username:refresh_token)
	// 4. get total login using scard (scard username:access_token) (scard username:refresh_token)
	accessUuids, refreshUuids, err := getAllUuid(username)

	for _, accessUuid := range accessUuids {
		if accessUuid == uuid { // exclude current session
			continue
		}

		_, err = RemoveRedisSession(username, accessUuid)
		if err != nil {
			return err
		}
	}

	for _, refreshUuid := range refreshUuids {
		if refreshUuid == uuid { // exclude current session
			continue
		}

		_, err = RemoveRedisSession(username, refreshUuid)
		if err != nil {
			return err
		}
	}

	return err
}

func setRedisSession(username string, data SessionData) error {
	redisSession := configs.GetRedisSessionConfig()
	dataMarshal, _ := json.Marshal(data.UserDetails) // TODO : Adding device details

	accessExpired := time.Unix(data.Token.AccessExpiredAt, 0)
	refreshExpired := time.Unix(data.Token.RefreshExpiredAt, 0)
	now := time.Now()

	keyRedis := getKeyRedis(username, data.Token.Uuid)

	err := redisSession.Set(keyRedis.KeyAccess, string(dataMarshal), accessExpired.Sub(now)).Err() // automatically expired
	if err != nil {
		return fmt.Errorf("error set redis session access token, err := %s", err.Error())
	}

	err = redisSession.Set(keyRedis.KeyRefresh, string(dataMarshal), refreshExpired.Sub(now)).Err() // automatically expired
	if err != nil {
		return fmt.Errorf("error set redis session refresh token, err := %s", err.Error())
	}

	err = redisSession.ZAdd(keyRedis.KeyTotalAccess, redis.Z{
		Score:  cast.ToFloat64(data.Token.AccessExpiredAt), // as expired time, set on env (default 1 day)
		Member: data.Token.Uuid,
	}).Err()
	if err != nil {
		return err
	}

	err = redisSession.ZAdd(keyRedis.KeyTotalRefresh, redis.Z{
		Score:  cast.ToFloat64(data.Token.RefreshExpiredAt), // as expired time, set on env (default 30 day)
		Member: data.Token.Uuid,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetCurrentUser get current user session from cookie uuid, uuid already set when jwt claim already set.
// you can access it by -> uuid, _ := ctx.Get("uuid")
// TODO: change return to users and device login details
func GetCurrentUser(uuidIdentifier string) (*models.Users, error) {
	keyRedis := getKeyRedis("", uuidIdentifier)

	redisSession := configs.GetRedisSessionConfig()
	redisData, err := redisSession.Get(keyRedis.KeyAccess).Result()
	if err != nil {
		return nil, err
	}

	var userData models.Users
	_ = json.Unmarshal([]byte(redisData), &userData)

	return &userData, nil
}

// GetCurrentUserRefresh get current user session from cookie uuid, uuid already set when jwt claim already set.
// you can access it by -> uuid, _ := ctx.Get("uuid")
// TODO: change return to users and device login details
func GetCurrentUserRefresh(uuidIdentifier string) (*models.Users, error) {
	keyRedis := getKeyRedis("", uuidIdentifier)

	redisSession := configs.GetRedisSessionConfig()
	redisData, err := redisSession.Get(keyRedis.KeyRefresh).Result()
	if err != nil {
		return nil, err
	}

	var userData models.Users
	_ = json.Unmarshal([]byte(redisData), &userData)

	return &userData, nil
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
