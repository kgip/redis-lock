package test

import (
	"github.com/go-redis/redis"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/kgip/redis-lock/adapters"
	"github.com/kgip/redis-lock/lock"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var (
	V8Client lock.RedisClientAdapter = adapters.NewGoRedisV8Adapter(redisV8.NewClient(&redisV8.Options{Addr: "192.168.31.34:6379"}))
	Client   lock.RedisClientAdapter = adapters.NewGoRedisAdapter(redis.NewClient(&redis.Options{Addr: "192.168.31.34:6379"}))
	ctx                              = lock.Context()
	wg                               = sync.WaitGroup{}
)

func TestLock(t *testing.T) {
	lock, _ := lock.NewRedisLock("key1", Client, nil)
	err := lock.Lock(ctx)
	assert.Equal(t, err, nil)
	defer func() {
		ok, _ := lock.Unlock()
		assert.Equal(t, ok, true)
	}()
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		t.Log("waiting!!!!!!!!!!!")
	}
}

func TestWatchDog(t *testing.T) {
	lock, _ := lock.NewRedisLock("testWatchDogLockKey", Client, &lock.Config{EnableWatchdog: true, Expire: 10 * time.Second})
	err := lock.Lock(ctx)
	assert.Equal(t, err, nil)
	defer func() {

	}()
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		t.Log("waiting!!!!!!!!!!!")
	}
	ok, _ := lock.Unlock()
	assert.Equal(t, ok, true)
	t.Log("complete!")
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
	}
}

func TestReentrant(t *testing.T) {
	locker, _ := lock.NewRedisLock("reentrantLockKey", Client, nil)
	err := locker.Lock(ctx)
	assert.Equal(t, err, nil)
	err = locker.Lock(ctx)
	assert.Equal(t, err, nil)
	t.Log("lock success")
	ok, _ := locker.Unlock()
	assert.Equal(t, ok, true)
	ok, _ = locker.Unlock()
	assert.Equal(t, ok, true)
}

func TestRedisLock_Unlock(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := 0
		locker, _ := lock.NewRedisLock("key", Client, nil)
		wg.Add(20)
		for j := 0; j < 10; j++ {
			go func() {
				ctx := lock.Context()
				locker.Lock(ctx)
				locker.Lock(ctx)
				n++
				locker.Unlock()
				locker.Unlock()
				wg.Done()
			}()
		}

		for j := 0; j < 10; j++ {
			go func() {
				ctx := lock.Context()
				locker.Lock(ctx)
				n--
				locker.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		t.Log(n)
		assert.Equal(t, n, 0)
	}
}

func TestRedisLockOperator(t *testing.T) {
	lockOperator, _ := lock.NewRedisLockOperator(V8Client, lock.Config{})
	key := "key"
	for i := 0; i < 100; i++ {
		n := 0
		wg.Add(20)
		for j := 0; j < 10; j++ {
			go func() {
				ctx := lock.Context()
				lockOperator.Lock(key, ctx)
				lockOperator.Lock(key, ctx)
				n++
				lockOperator.Unlock(key)
				lockOperator.Unlock(key)
				wg.Done()
			}()
		}

		for j := 0; j < 10; j++ {
			go func() {
				ctx := lock.Context()
				locker, _ := lockOperator.GetLock(key, lock.Config{})
				locker.Lock(ctx)
				n--
				locker, _ = lockOperator.GetLock(key, lock.Config{})
				locker.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		t.Log(n)
		assert.Equal(t, n, 0)
	}
}
