package redis_lock

import (
	"context"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"log"
	"sync"
	"time"
)

const (
	defaultTimeout = 10 * time.Second //Default Timeout for acquiring locks
	defaultExpire  = 10 * time.Second //Default lock expiration time
	contextKey     = "LockContextKey"
	//redis lock lua script
	lockScript = `if redis.call('exists',KEYS[1]) then
					  local r = redis.call('hget', KEYS[1],ARGV[1])
					  if r then 
						  r = r + 1
						  if redis.call('hset', KEYS[1], ARGV[1], r) then
							  if redis.call('expire', KEYS[1], ARGV[2]) then
								  return 1
                              end
                          end
					  end
				  else
					  if redis.call('hset', KEYS[1], ARGV[1], r) then
                          if redis.call('expire', KEYS[1], ARGV[2]) then
                              return 1
                          end
                      end
				  end
				  return 0`
	//redis unlock lua script
	unlockScript = `if redis.call('exists',KEYS[1]) then
						local r = redis.call('hget', KEYS[1],ARGV[1])
						if r then
                            r = r - 1 
                            if r == 0 then
                            	if redis.call('del', KEYS[1]) then
									return 1
								end
                            else 
                            	if redis.call('hset', KEYS[1], ARGV[1], r) then
                                	return 1
								end
							end
						end
                    else 
                        return 1
                    end
                    return 0`
)

type Locker interface {
	Lock(ctx LockContext) bool
	TryLock(ctx LockContext, time time.Duration) bool //Attempt to lock for a fixed duration
	Unlock(ctx LockContext) bool
}

type LockContext context.Context

func Context() LockContext {
	return context.WithValue(context.Background(), contextKey, uuid.NewV4().String())
}

type RedisLock struct {
	client         redis.Cmdable //redis connect client interface
	key            string        //lock key
	timeout        time.Duration //lock Timeout for acquiring locks
	expire         time.Duration //lock expiration time
	expireSecond   int           //The number of seconds in which the lock expires
	enableWatchdog bool          //Whether to enable the watchdog to renew the lock
	lock           *sync.Mutex
	ctx            context.Context
}

func NewRedisLock(options ...option) *RedisLock {
	redisLock := &RedisLock{timeout: defaultTimeout, expire: defaultExpire, lock: &sync.Mutex{}}
	for _, parameterize := range options {
		parameterize(redisLock)
	}
	return redisLock
}

func (r *RedisLock) Timeout(timeout time.Duration) *RedisLock {
	r.timeout = timeout
	return r
}

func (r *RedisLock) Lock(ctx LockContext) bool {
	return r.TryLock(ctx, r.timeout)
}

// TryLock time: Maximum wait time to acquire a lock
func (r *RedisLock) TryLock(ctx LockContext, timeout time.Duration) bool {
	if r.key == "" {
		log.Println("lock key cannot be empty")
		return false
	}
	if r.client == nil {
		log.Println("redis client cannot be nil")
		return false
	}
	t := time.NewTimer(timeout)
	for {
		select {
		case <-t.C:
			log.Println("acquire lock timeout")
			return false
		default:
			if r.client.Eval(lockScript, []string{r.key}, ctx.Value(contextKey), r.expireSecond).Val().(int64) == 1 { //The lock was acquired successfully
				return true
			}
		}
		//Re-acquire the lock after waiting for 1 second
		time.Sleep(time.Second)
	}
}

func (r *RedisLock) Unlock(ctx LockContext) bool {
	return r.client.Eval(unlockScript, []string{r.key}, ctx.Value(contextKey)).Val().(int64) == 1
}
