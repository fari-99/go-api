package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"

	"go-api/modules/configs"
	"go-api/modules/models"
)

const accessTokenIndex = "access_token"
const refreshTokenIndex = "refresh_token"
const plusInfinite = "+inf"
const negativeInfinite = "-inf"

type SessionToken struct {
	Uuid             string    `json:"uuid"`
	AccessExpiredAt  time.Time `json:"access_expired_at"`
	RefreshExpiredAt time.Time `json:"refresh_expired_at"`
}

type KeyRedisSessionData struct {
	KeyAccess       string
	KeyRefresh      string
	KeyTotalAccess  string
	KeyTotalRefresh string
	KeyFamily       string
}

type SessionData struct {
	Token SessionToken

	UserID        string      `json:"user_id"`
	UserDetails   interface{} `json:"user_details"`
	Authorization bool        `json:"authorization"`
}

type FamilyCheck struct {
	OldUuid  string `json:"old_uuid"`
	NewUuid  string `json:"new_uuid"`
	Username string `json:"username"`
}

func getKeyRedis(username string, uuid string) KeyRedisSessionData {
	return KeyRedisSessionData{
		KeyAccess:       fmt.Sprintf("%s:%s", uuid, accessTokenIndex),      // uuid:access_token, value: device and user data
		KeyRefresh:      fmt.Sprintf("%s:%s", uuid, refreshTokenIndex),     // uuid:refresh_token, value: device and user data
		KeyTotalAccess:  fmt.Sprintf("%s:%s", username, accessTokenIndex),  // username:access_token, value: collection of uuid
		KeyTotalRefresh: fmt.Sprintf("%s:%s", username, refreshTokenIndex), // username:refresh_token, value: collection of uuid
		KeyFamily:       fmt.Sprintf("%s:%s", uuid, "family"),              // uuid:family, value: collection of old uuid
	}
}

// why using ZADD and not SADD for storing session?
// ZADD has a feature to add scoring system to sort their member, which SADD not.
// after that we use that scoring system to expired time
// then we sort the member by their score if the score less than expected value (time now unix), then they expired

func removeExpiredToken(ctx context.Context, redisSession redis.UniversalClient, username string) (err error) {
	keyRedis := getKeyRedis(username, "")
	timeNow := cast.ToString(time.Now().Unix())

	// get all expired access token
	accessTokenUuids, err := redisSession.ZRangeByScore(ctx, keyRedis.KeyTotalAccess, &redis.ZRangeBy{
		Min: negativeInfinite,
		Max: timeNow,
	}).Result()
	if err != nil {
		return err
	}

	// get all expired
	refreshTokenUuids, err := redisSession.ZRangeByScore(ctx, keyRedis.KeyTotalRefresh, &redis.ZRangeBy{
		Min: negativeInfinite,
		Max: timeNow,
	}).Result()
	if err != nil {
		return err
	}

	if len(accessTokenUuids) > 0 {
		for _, accessTokenUuid := range accessTokenUuids {
			_, _ = RemoveRedisSession(ctx, username, accessTokenUuid)
		}
	}

	if len(refreshTokenUuids) > 0 {
		for _, refreshTokenUuid := range refreshTokenUuids {
			_, _ = RemoveRedisSession(ctx, username, refreshTokenUuid)
		}
	}

	return nil
}

func getTotalLogin(ctx context.Context, redisSession redis.UniversalClient, username string) (totalLoginAccessToken int64, totalLoginRefreshToken int64, err error) {
	keyRedis := getKeyRedis(username, "")

	err = removeExpiredToken(ctx, redisSession, username)
	if err != nil {
		return 0, 0, err
	}

	totalLoginAccessToken, err = redisSession.ZCard(ctx, keyRedis.KeyTotalAccess).Result()
	if err != nil {
		return 0, 0, err
	}

	totalLoginRefreshToken, err = redisSession.ZCard(ctx, keyRedis.KeyTotalRefresh).Result()
	if err != nil {
		return 0, 0, err
	}

	return totalLoginAccessToken, totalLoginRefreshToken, nil
}

func getAllUuid(ctx context.Context, username string) (accessUuids []string, refreshUuids []string, err error) {
	keyRedis := getKeyRedis(username, "")
	redisSession := configs.GetRedisSessionConfig()

	err = removeExpiredToken(ctx, redisSession, username)
	if err != nil {
		return nil, nil, err
	}

	accessUuids, err = redisSession.ZRange(ctx, keyRedis.KeyTotalAccess, 0, -1).Result()
	if err != nil {
		return nil, nil, err
	}

	refreshUuids, err = redisSession.ZRange(ctx, keyRedis.KeyTotalRefresh, 0, -1).Result()
	if err != nil {
		return nil, nil, err
	}

	return accessUuids, refreshUuids, nil
}

func setRedisSession(ctx context.Context, username string, data SessionData) error {
	redisSession := configs.GetRedisSessionConfig()
	dataMarshal, _ := json.Marshal(data.UserDetails) // TODO : Adding device details

	keyRedis := getKeyRedis(username, data.Token.Uuid)

	err := redisSession.Set(ctx, keyRedis.KeyAccess, string(dataMarshal), getTimeDuration(data.Token.AccessExpiredAt)).Err() // automatically expired
	if err != nil {
		return fmt.Errorf("error set redis session access token, err := %s", err.Error())
	}

	err = redisSession.Set(ctx, keyRedis.KeyRefresh, string(dataMarshal), getTimeDuration(data.Token.RefreshExpiredAt)).Err() // automatically expired
	if err != nil {
		return fmt.Errorf("error set redis session refresh token, err := %s", err.Error())
	}

	err = redisSession.ZAdd(ctx, keyRedis.KeyTotalAccess, redis.Z{
		Score:  cast.ToFloat64(data.Token.AccessExpiredAt), // as expired time, set on env (default 1 day)
		Member: data.Token.Uuid,
	}).Err()
	if err != nil {
		return err
	}

	err = redisSession.ZAdd(ctx, keyRedis.KeyTotalRefresh, redis.Z{
		Score:  cast.ToFloat64(data.Token.RefreshExpiredAt), // as expired time, set on env (default 30 day)
		Member: data.Token.Uuid,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetAllSessions(ctx context.Context, username string) ([]models.Users, error) {
	_, refreshUuids, err := getAllUuid(ctx, username)
	if err != nil {
		return nil, err
	}

	var users []models.Users
	for _, refreshUuid := range refreshUuids {
		user, err := GetCurrentUser(ctx, refreshUuid)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return users, nil
}

func CheckToken(ctx context.Context, username, uuid string) (isExistAccess, isExistRefresh bool, err error) {
	keyRedis := getKeyRedis(username, uuid)
	redisSession := configs.GetRedisSessionConfig()

	resultAccess, err := redisSession.Exists(ctx, keyRedis.KeyAccess).Result()
	if err != nil {
		return false, false, err
	}

	resultRefresh, err := redisSession.Exists(ctx, keyRedis.KeyRefresh).Result()
	if err != nil {
		return false, false, err
	}

	return resultAccess > 0, resultRefresh > 0, nil
}

func SetupLoginSession(ctx context.Context, username string, data SessionData) (totalLogin int64, err error) {
	// check if login session > total session that allowed to login
	// if more, then return error that new session can't be created because you have device already connected
	// if less, then
	// 1. set redis session uuid (set uuid:access_token data) (set uuid:refresh_token data)
	// 2. put redis session uuid using zadd (zadd username:access_token uuid) (zadd username:refresh_token uuid)
	// 3. get total login using zcard (zcard username:access_token) (zcard username:refresh_token)
	// 4. return total login

	redisSession := configs.GetRedisSessionConfig()
	totalLoginAccessToken, _, err := getTotalLogin(ctx, redisSession, username)
	if err != nil {
		return 0, err
	}

	redisSession.Exists(ctx)

	if totalLoginAccessToken >= cast.ToInt64(os.Getenv("TOTAL_LOGIN_SESSION")) {
		return totalLogin, fmt.Errorf("total login session are more than allowed, logout one of your session from one of your device, or delete all sessions")
	}

	err = setRedisSession(ctx, username, data)
	if err != nil {
		return 0, err
	}

	totalLoginAccessToken, _, err = getTotalLogin(ctx, redisSession, username)
	if err != nil {
		return 0, err
	}

	return totalLoginAccessToken, err
}

func RemoveRedisSession(ctx context.Context, username, uuid string) (totalLogin int64, err error) {
	// 1. delete redis session uuid (del uuid:access_token) (del uuid:refresh_token)
	// 2. delete redis member using zrem (zrem username:access_token uuid) (zrem username:refresh_token uuid)
	// 3. get total login using zcard (zcard username:access_token) (zcard username:refresh_token)
	// 4. return total login

	keyRedis := getKeyRedis(username, uuid)

	redisSession := configs.GetRedisSessionConfig()
	err = redisSession.Del(ctx, keyRedis.KeyAccess).Err() // delete access token redis
	if err != nil {
		return 0, err
	}

	err = redisSession.Del(ctx, keyRedis.KeyRefresh).Err() // delete refresh token redis
	if err != nil {
		return 0, err
	}

	err = redisSession.ZRem(ctx, keyRedis.KeyTotalAccess, uuid).Err() // delete access token member redis
	if err != nil {
		return 0, err
	}

	err = redisSession.ZRem(ctx, keyRedis.KeyTotalRefresh, uuid).Err() // delete refresh token member redis
	if err != nil {
		return 0, err
	}

	totalLoginAccessToken, _, err := getTotalLogin(ctx, redisSession, username)
	if err != nil {
		return 0, err
	}

	return totalLoginAccessToken, nil
}

func DeleteAllSession(ctx context.Context, username string, uuid string) (err error) {
	// 1. get all members using zmembers (zmembers username:access_token) (zmembers username:refresh_token)
	// 2. delete redis session by -looping- uuid from zmembers using del (del uuid:access_token) (del uuid:refresh_token)
	// 3. delete redis member using del (del username:access_token) (del username:refresh_token)
	// 4. get total login using scard (scard username:access_token) (scard username:refresh_token)
	accessUuids, refreshUuids, err := getAllUuid(ctx, username)

	for _, accessUuid := range accessUuids {
		if accessUuid == uuid { // exclude current session
			continue
		}

		_, err = RemoveRedisSession(ctx, username, accessUuid)
		if err != nil {
			return err
		}
	}

	for _, refreshUuid := range refreshUuids {
		if refreshUuid == uuid { // exclude current session
			continue
		}

		_, err = RemoveRedisSession(ctx, username, refreshUuid)
		if err != nil {
			return err
		}
	}

	return err
}

// using family as refresh token rotation check
// Flow:
// 1. User-A login and got his AT (Access Token) 1 and RT (Refresh Token) 1
// 2. User-B got RT 1
// 3. User-B use RT 1 to get new AT and RT, which System return AT 2 and RT 2
// 4. User-A try to access something using AT 1, which System return error, because AT 1 already refreshed
// 5. User-A use RT 1 to get new AT and RT, which System return error, because RT 1 already refreshed
// 5.1. because RT 1 already used to refresh token, and someone else using it again, System will flag this token
// 5.2. System delete RT 1 family token, which deleting AT 2 and RT 2 token from System
// 5.3. System ask User-A to re-authenticate (re-login) to get new token
// 6. User-A login and got his new AT 3 and RT 3
// On the System:
// 1. if refresh token used, check if refresh uuid already in the family
// 1.a. if it already in the family, then delete new Access Token, Refresh Token, and family member
// 1.b. if not already in the family, then ok :thumbs:!

// SetFamily set family
func SetFamily(ctx context.Context, username, oldUuid, newUuid string, expiration time.Time) (err error) {
	// set uuid to family using set (set old_uuid:family new_uuid)
	dataFamily := FamilyCheck{
		OldUuid:  oldUuid,
		NewUuid:  newUuid,
		Username: username,
	}

	dataMarshal, _ := json.Marshal(dataFamily)

	keyRedis := getKeyRedis(username, oldUuid)
	redisSession := configs.GetRedisSessionConfig()
	_, err = redisSession.Set(ctx, keyRedis.KeyFamily, string(dataMarshal), getTimeDuration(expiration)).Result()
	if err != nil {
		return err
	}

	return nil
}

// CheckFamily check family
func CheckFamily(ctx context.Context, username, oldUuid string) (isUsed bool, err error) {
	// check if old_uuid already used, by set (set old_uuid:family old_uuid)
	keyRedis := getKeyRedis(username, oldUuid)
	redisSession := configs.GetRedisSessionConfig()

	// check if old_uuid is in the family that already refreshed
	familyDataMarshal, err := redisSession.Get(ctx, keyRedis.KeyFamily).Result()
	if err != nil && err != redis.Nil {
		return false, err
	} else if err == redis.Nil {
		// if not found, then old_uuid is a new_uuid
		return false, nil
	}

	var familyData FamilyCheck
	_ = json.Unmarshal([]byte(familyDataMarshal), &familyData)

	// because there is old uuid, delete new access token and refresh token
	_, err = RemoveRedisSession(ctx, familyData.Username, familyData.NewUuid)
	if err != nil {
		return true, err
	}

	// delete current family
	redisSession.Del(ctx, keyRedis.KeyFamily)

	return true, nil
}

// GetCurrentUser get current user session from cookie uuid, uuid already set when jwt claim already set.
// you can access it by -> uuid, _ := ctx.Get("uuid")
// TODO: change return to users and device login details
func GetCurrentUser(ctx context.Context, uuidIdentifier string) (*models.Users, error) {
	keyRedis := getKeyRedis("", uuidIdentifier)

	redisSession := configs.GetRedisSessionConfig()
	redisData, err := redisSession.Get(ctx, keyRedis.KeyAccess).Result()
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
func GetCurrentUserRefresh(ctx context.Context, uuidIdentifier string) (*models.Users, error) {
	keyRedis := getKeyRedis("", uuidIdentifier)

	redisSession := configs.GetRedisSessionConfig()
	redisData, err := redisSession.Get(ctx, keyRedis.KeyRefresh).Result()
	if err != nil {
		return nil, err
	}

	var userData models.Users
	_ = json.Unmarshal([]byte(redisData), &userData)

	return &userData, nil
}

func getTimeDuration(input interface{}) time.Duration {
	var target time.Time

	switch v := input.(type) {
	case int64:
		target = time.Unix(v, 0)
	case time.Time:
		target = v
	default:
		panic(fmt.Sprintf("unsupported type: %T", input))
	}

	return target.Sub(time.Now())
}
