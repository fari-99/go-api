package redis_helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go-api/modules/configs"
)

const CounterTypeTotp = "counter_totp"
const CounterTypeRecoveryCode = "counter_recovery_code"

const CounterTotpKey = "COUNTER_TOTP:%d:%s"
const CounterRecoveryCodeKey = "COUNTER_RECOVERY_CODE:%d:%s"

const TwoFAActionTypeCreate = "create"                // first time create
const TwoFAActionTypeLogin = "login"                  // when login
const TwoFAActionTypeResetPassword = "reset-password" // when reset password

// const TwoFAActionType = "type"

var ActionType = map[string]bool{
	TwoFAActionTypeCreate:        true,
	TwoFAActionTypeLogin:         true,
	TwoFAActionTypeResetPassword: true,
}

type CounterConfig struct {
	Ctx    context.Context
	Type   string
	UserID uint64
	Action string

	MaxExpMinutes int64
	MaxCounter    int64
}

func (conf *CounterConfig) Count() error {
	if conf.UserID == 0 || conf.Action == "" {
		return errors.New("user_id or action is required")
	}

	if _, ok := ActionType[conf.Action]; !ok {
		return fmt.Errorf("invalid action: %s", conf.Action)
	}

	var key string
	maxExpMinutes := conf.MaxExpMinutes
	maxCounter := conf.MaxCounter

	switch conf.Type {
	case CounterTypeTotp:
		key = fmt.Sprintf(CounterTotpKey, conf.UserID, conf.Action)
		if maxExpMinutes <= 0 {
			maxExpMinutes, _ = strconv.ParseInt(os.Getenv("MAX_EXPIRED_COUNTER_TOTP"), 10, 64)
			if maxExpMinutes == 0 {
				maxExpMinutes = 5
			}
		}

		if maxCounter <= 0 {
			maxCounter, _ = strconv.ParseInt(os.Getenv("MAX_COUNTER_VALUE_TOTP"), 10, 64)
			if maxCounter == 0 {
				maxCounter = 3
			}
		}

	case CounterTypeRecoveryCode:
		key = fmt.Sprintf(CounterRecoveryCodeKey, conf.UserID, conf.Action)
		if maxExpMinutes <= 0 {
			maxExpMinutes, _ = strconv.ParseInt(os.Getenv("MAX_EXPIRED_COUNTER_RECOVERY_CODE"), 10, 64)
			if maxExpMinutes == 0 {
				maxExpMinutes = 60
			}
		}

		if maxCounter <= 0 {
			maxCounter, _ = strconv.ParseInt(os.Getenv("MAX_COUNTER_VALUE_RECOVERY_CODE"), 10, 64)
			if maxCounter == 0 {
				maxCounter = 3
			}
		}

	default:
		return fmt.Errorf("invalid type [%s]", conf.Type)
	}

	ctx := conf.Ctx
	return handleCounter(ctx, key, maxExpMinutes, maxCounter)
}

func (conf *CounterConfig) Reset() error {
	if conf.UserID == 0 || conf.Action == "" {
		return errors.New("user_id or action is required")
	}

	if _, ok := ActionType[conf.Action]; !ok {
		return fmt.Errorf("invalid action: %s", conf.Action)
	}

	var key string
	switch conf.Type {
	case CounterTypeTotp:
		key = fmt.Sprintf(CounterTotpKey, conf.UserID, conf.Action)

	case CounterTypeRecoveryCode:
		key = fmt.Sprintf(CounterRecoveryCodeKey, conf.UserID, conf.Action)
	default:
		return fmt.Errorf("invalid type [%s]", conf.Type)
	}

	ctx := conf.Ctx
	return handleReset(ctx, key)
}

func handleCounter(ctx context.Context, key string, maxExp, maxCounter int64) error {
	rdb := configs.GetRedis(configs.REDIS_COUNTING_PREFIX)
	expiry := time.Duration(maxExp) * time.Minute

	// Increment first
	val, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	// On first increment, set the TTL
	if val == 1 {
		if err := rdb.Expire(ctx, key, expiry).Err(); err != nil {
			return err
		}
	}

	// Check AFTER increment
	if val > maxCounter {
		ttl, _ := rdb.TTL(ctx, key).Result()
		return fmt.Errorf("too many attempts (%d/%d), retry after %.0f seconds", maxCounter, maxCounter, ttl.Seconds())
	}

	return nil
}

func handleReset(ctx context.Context, key string) error {
	rdb := configs.GetRedis(configs.REDIS_COUNTING_PREFIX)
	return rdb.Del(ctx, key).Err()
}
