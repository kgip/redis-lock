package redis_lock

import (
	"context"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"log"
	"strconv"
	"time"
)

const (
	defaultTimeout = 35 * time.Second //Default Timeout for acquiring locks
	defaultExpire  = 30 * time.Second //Default lock expiration time
	contextKey     = "LockContextIDKey"
	//redis lock lua script
	lockScript = `if redis.call('exists',KEYS[1]) == 1 then
					  local r = redis.call('hget', KEYS[1],ARGV[1])
					  if r then
						  r = r + 1
						  if redis.call('hset', KEYS[1], ARGV[1], r) then
							  if redis.call('expire', KEYS[1], ARGV[2]) == 1 then
								  return 1
                              else
                                  redis.call('del', KEYS[1])
                              end
                          end
					  end
				  elseif redis.call('hset', KEYS[1], ARGV[1], 1) then
					  if redis.call('expire', KEYS[1], ARGV[2]) == 1 then
						  return 1
                      else
                          redis.call('del', KEYS[1])
					  end
				  end
				  return 0`
	//redis unlock lua script
	unlockScript = `if redis.call('exists',KEYS[1]) == 1 then
						local r = redis.call('hget', KEYS[1],ARGV[1])
						if r then
                            r = r - 1 
                            if r == 0 then
                            	if redis.call('del', KEYS[1]) == 1 then
									return 1
								end
                            elseif redis.call('hset', KEYS[1], ARGV[1], r) then
                                return 1
							end
						end
                    else 
                        return 1
                    end
                    return 0`
)

type Locker interface {
	Lock(ctx LockContext)
	TryLock(ctx LockContext, time time.Duration) bool //Attempt to lock for a fixed duration
	Unlock() bool
}

// LockContext goroutine lock context
type LockContext context.Context

// Context
//Obtain the goroutine lock context, which is required when locking.
//When using reentrant locks, the context of multiple locking must be the same.
func Context() LockContext {
	return context.WithValue(context.Background(), contextKey, uuid.NewV4().String())
}

// RedisLock Redis distributed lock structure, implementing Locker interface
type RedisLock struct {
	key            string        //lock key
	client         redis.Cmdable //redis connect client interface
	timeout        time.Duration //lock Timeout for acquiring locks
	expire         time.Duration //lock expiration time
	expireSecond   int           //The number of seconds in which the lock expires
	enableWatchdog bool          //Whether to enable the watchdog to renew the lock
	lockSignal     chan byte     //lock signal
	unlockSignal   chan byte     //unlock signal
	ctx            LockContext   //current lock goroutine context
}

// NewRedisLock
//lockKey and redisClient are required parameters, and options are optional parameters.
//Among them, lockKey is the key stored in redis of the distributed lock, redisClient
//is the operation client of redis, and the optional parameters of options include the
//expiration time of the lock, Expire, the timeout time of acquiring the lock, Timeout ,
//and the identifier of whether to start the watchdog renewal mechanism, EnableWatchdog.
//When passing parameters, a form such as Expire(20 * time.Second) can be used.
func NewRedisLock(lockKey string, redisClient redis.Cmdable, options ...option) *RedisLock {
	redisLock := &RedisLock{timeout: defaultTimeout, expire: defaultExpire, expireSecond: int(defaultExpire) / int(time.Second)}
	options = append(options, key(lockKey), client(redisClient)) //Add required parameters
	for _, parameterize := range options {
		parameterize(redisLock)
	}
	if redisLock.enableWatchdog {
		redisLock.lockSignal = make(chan byte)
		redisLock.unlockSignal = make(chan byte)
		ticker := time.NewTicker(redisLock.expire / 2)
		//Start watchdog renewal goroutine
		go func() {
			for range redisLock.lockSignal {
				ticker.Reset(redisLock.expire / 2)
				for {
					select {
					case <-redisLock.unlockSignal:
						log.Println("end renew expire ")
						break
					case <-ticker.C:
						redisLock.client.Expire(redisLock.key, redisLock.expire)
						log.Printf("renew expire -------------------------------->%ds", redisLock.expireSecond)
					}
				}
			}
		}()
	}
	return redisLock
}

// Lock ctx is the context of the current goroutine, which can be obtained through Context()
func (r *RedisLock) Lock(ctx LockContext) {
	if !r.TryLock(ctx, r.timeout) {
		panic("acquire lock timeout")
	}
}

// TryLock ctx is the context of the current goroutine, which can be obtained through Context(). timeout is the maximum wait time to acquire a lock.
func (r *RedisLock) TryLock(ctx LockContext, timeout time.Duration) bool {
	t := time.NewTimer(timeout)
	for {
		select {
		case <-t.C:
			log.Println("acquire lock timeout")
			return false
		default:
			if r.client.Eval(lockScript, []string{r.key}, ctx.Value(contextKey), r.expireSecond).Val().(int64) == 1 { //The lock was acquired successfully
				r.ctx = ctx
				if r.enableWatchdog && r.getLockTimes() <= 1 {
					r.lockSignal <- 0
				}
				return true
			}
		}
		//Re-acquire the lock after waiting for 1 second
		time.Sleep(time.Second)
	}
}

func (r *RedisLock) Unlock() bool {
	if r.client.Eval(unlockScript, []string{r.key}, r.ctx.Value(contextKey)).Val().(int64) == 1 {
		if r.enableWatchdog && r.getLockTimes() <= 0 {
			r.unlockSignal <- 0
		}
		return true
	}
	return false
}

//Get the number of locks
func (r *RedisLock) getLockTimes() int64 {
	val := r.client.HGet(r.key, r.ctx.Value(contextKey).(string)).Val()
	if val == "" {
		return 0
	}
	times, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Panicf("Redis Error! Get lock times failed : %s!", err.Error())
	}
	return times
}
