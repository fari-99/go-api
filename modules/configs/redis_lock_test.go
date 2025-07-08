package configs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

var (
	keyLock        = "test-redsync"
	countLockCheck = 1
	countLock      = 0

	keySet        = "test-set"
	countSetCheck = 1
	countSet      = 0
)

func getRedis() *redis.Client {
	redisOption := redis.Options{
		Network:      "tcp",
		Addr:         fmt.Sprintf("%s:%s", "redis", "6379"),
		DB:           cast.ToInt(2),
		MaxRetries:   cast.ToInt(3),
		PoolFIFO:     true,
		MinIdleConns: cast.ToInt(1),
	}

	client := redis.NewClient(&redisOption)

	err := client.Ping(context.Background()).Err()
	if err != nil {
		panic(err.Error())
	}

	return client
}

func redisLock() error {
	client := getRedis()
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			mutex := rs.NewMutex(keyLock)
			// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond) // minimum 2 seconds
			// defer cancel()

			ctx := context.Background()

			if err := mutex.LockContext(ctx); err != nil {
				return
			}

			// unlock when everything is done
			defer mutex.UnlockContext(ctx)

			// DO SOMETHING HERE THAT NEED LOCKED

			// check if count already set or not
			if countLock > 0 {
				return
			}

			countLock++

			// extend timeout mutex, redsync mutex default expired is 8 second
			// NOTE: a process would typically only extend the lock
			// If it knows it will need more time to complete its operation.
			// If it's unsure, it would be better to release the lock as soon as it's done with its operation,
			// so other processes can acquire the lock.
			_, err := mutex.ExtendContext(ctx)
			if errors.Is(err, context.DeadlineExceeded) {
				return
			} else if err != nil && strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
				// in goroutine the error is
				// node #0: context deadline exceeded
				return
			} else if err != nil {
				panic(err.Error())
			}
		}()
	}

	wg.Wait()

	if countLock != countLockCheck {
		return fmt.Errorf("count [total := %d] is failed", countLock)
	}

	return nil
}

func redisSetLock() error {
	client := getRedis()

	getRedis := func(client *redis.Client) (exists bool, err error) {
		err = client.Get(context.Background(), keySet).Err()
		if errors.Is(err, redis.Nil) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		return true, nil
	}

	setRedis := func(client *redis.Client) error {
		err := client.Set(context.Background(), keySet, "true", 5*time.Minute).Err()
		if err != nil {
			return err
		}

		return nil
	}

	delRedis := func(client *redis.Client) error {
		err := client.Del(context.Background(), keySet).Err()
		if err != nil {
			return err
		}

		return nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			defer func() {
				_ = delRedis(client)
			}()

			for {
				exists, err := getRedis(client)
				if err != nil {
					panic(err.Error())
				}

				if exists {
					continue
				}

				err = setRedis(client)
				if err != nil {
					panic(err.Error())
				}

				// DO SOMETHING HERE THAT NEED LOCKED
				if countSet > 0 {
					break
				}

				countSet++

				break
			}

		}()
	}

	wg.Wait()

	if countSet != countSetCheck {
		return fmt.Errorf("count [total := %d] is failed", countSet)
	}

	return nil
}

func TestRedisLock(t *testing.T) {
	err := redisLock()
	if err != nil {
		t.Fail()
		t.Log(err.Error())
		return
	}
	t.Log(fmt.Sprintf("success lock count to %d", countLock))
}

func TestRedisSet(b *testing.T) {
	err := redisSetLock()
	if err != nil {
		b.Fail()
		b.Log(err.Error())
		return
	}
	b.Log(fmt.Sprintf("success set lock count to %d", countSet))
}
