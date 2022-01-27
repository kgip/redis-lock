package core

import (
	"context"
	"github.com/go-redis/redis"
	"log"
	lock "redis-lock"
	"time"
)

const (
	defaultTimeout = 5 * time.Second //Default lock expiration time
	script = `if redis.call('exists',KEYS[1]) == 1 then
                  local r = redis.call('hget', KEYS[1],ARGV[1])
                  if r then 
                      r = r + 1
                      redis.call('hset', KEYS[1], ARGV[1], r)
                      redis.call('expire', KEYS[1], ARGV[2])
                  else
                      return 0
                  end
              else
                  redis.call('hset', KEYS[1], ARGV[1], 1)
                  redis.call('expire', KEYS[1], ARGV[1])
              end
              return 1
             `
)


type RedisLock struct {
	client  redis.Cmdable
	key     string        //lock key
	timeout time.Duration //lock timeout
	ctx     context.Context
}

func NewRedisLock(options ...option) *RedisLock {
	redisLock := &RedisLock{timeout: defaultTimeout}
	for _, item := range options {
		item(redisLock)
	}
	return redisLock
}

func (r *RedisLock) Lock(ctx lock.LockContext) bool {
	if r.key == "" {
		log.Println("lock key cannot be empty")
		return false
	}
	if r.client == nil {
		log.Println("redis client cannot be nil")
		return false
	}
	for {
		r.client.Eval(script, []string{r.key}, ctx.Value(lock.ContextKey), r.timeout).Val()
	}
}

// TryLock time: Maximum wait time to acquire a lock
func (r *RedisLock) TryLock(ctx lock.LockContext, time time.Duration) bool {
	panic("implement me")
}

func (r *RedisLock) Unlock(ctx lock.LockContext) bool {
	panic("implement me")
}
