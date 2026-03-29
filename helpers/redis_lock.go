// helpers/redis_lock.go

package helpers

import (
	"context"

	"github.com/go-redsync/redsync/v4"

	"go-api/modules/configs"
)

type RedisLockHelper struct {
	ctx   context.Context
	mutex *redsync.Mutex
}

func RedisLock() *RedisLockHelper {
	return &RedisLockHelper{}
}

func (r *RedisLockHelper) SetContext(ctx context.Context) *RedisLockHelper {
	r.ctx = ctx
	return r
}

func (r *RedisLockHelper) Lock(key string) error {
	redLock := configs.GetRedisLock()
	r.mutex = redLock.NewMutex(key)

	if r.ctx != nil {
		return r.mutex.LockContext(r.ctx)
	}

	return r.mutex.Lock()
}

func (r *RedisLockHelper) Unlock() error {
	if r.mutex == nil {
		return nil
	}

	var err error
	if r.ctx != nil {
		_, err = r.mutex.UnlockContext(r.ctx)
	} else {
		_, err = r.mutex.Unlock()
	}

	return err
}
