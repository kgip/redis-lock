package lock

import (
	"context"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
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
						else 
                            return 1
						end
                    else 
                        return 1
                    end
                    return 0`
)

type Locker interface {
	Lock(ctx LockContext) error
	TryLock(ctx LockContext, time time.Duration) error //Attempt to lock for a fixed duration
	Unlock() (bool, error)
}

// LockContext goroutine lock context
type LockContext context.Context

// Context
//Obtain the goroutine lock context, which is required when locking.
//When using reentrant locks, the context of multiple locking must be the same.
func Context() LockContext {
	return context.WithValue(context.Background(), contextKey, uuid.NewV4().String())
}

type Config struct {
	Timeout        time.Duration //lock Timeout for acquiring locks
	Expire         time.Duration //lock expiration time
	EnableWatchdog bool          //Whether to enable the watchdog to renew the lock
	Logger         Logger
}

// RedisLock Redis distributed lock structure, implementing Locker interface
type RedisLock struct {
	key          string             //lock key
	client       RedisClientAdapter //redis connect client interface
	config       *Config
	lockSignal   chan byte   //lock signal
	unlockSignal chan byte   //unlock signal
	ctx          LockContext //current lock goroutine context
}

// NewRedisLock
//lockKey and redisClient are required parameters, and options are optional parameters.
func NewRedisLock(lockKey string, redisClient RedisClientAdapter, config *Config) (*RedisLock, error) {
	if lockKey == "" {
		return nil, KeyEmptyError
	}
	if redisClient == nil {
		return nil, RedisClientNilError
	}
	if config == nil {
		config = &Config{
			Timeout: defaultTimeout,
			Expire:  defaultExpire,
			Logger:  &DefaultLogger{Level: Info, Log: Zap()},
		}
	} else {
		if config.Timeout <= 0 {
			config.Timeout = defaultTimeout
		}
		if config.Expire <= 0 {
			config.Expire = defaultExpire
		}
		if config.Logger == nil {
			config.Logger = &DefaultLogger{Level: Info, Log: Zap()}
		}
	}
	redisLock := &RedisLock{key: lockKey, client: redisClient, config: config}
	if redisLock.config.EnableWatchdog {
		redisLock.lockSignal = make(chan byte)
		redisLock.unlockSignal = make(chan byte)
		ticker := time.NewTicker(redisLock.config.Expire / 2)
		//Start watchdog renewal goroutine
		go func() {
			for range redisLock.lockSignal {
				ticker.Reset(redisLock.config.Expire / 2)
				for {
					select {
					case <-redisLock.unlockSignal:
						redisLock.config.Logger.Info("Redis lock watchdog end renew expire")
						return
					case <-ticker.C:
						if err := redisLock.client.Expire(redisLock.ctx, redisLock.key, redisLock.config.Expire).Err; err != nil {
							redisLock.config.Logger.Error(fmt.Sprintf("Redis lock watchdog renew expire failed: %v", err))
							continue
						}
						redisLock.config.Logger.Info(fmt.Sprintf("Redis lock watchdog renew expire %ds", redisLock.config.Expire/time.Second))
					}
				}
			}
		}()
	}
	return redisLock, nil
}

// Lock ctx is the context of the current goroutine, which can be obtained through Context()
func (r *RedisLock) Lock(ctx LockContext) error {
	return r.TryLock(ctx, r.config.Timeout)
}

// TryLock ctx is the context of the current goroutine, which can be obtained through Context(). timeout is the maximum wait time to acquire a lock.
func (r *RedisLock) TryLock(ctx LockContext, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	t := time.NewTimer(timeout)
	for {
		select {
		case <-t.C:
			return TimeoutError
		default:
			eval := r.client.Eval(ctx, lockScript, []string{r.key}, ctx.Value(contextKey), int(r.config.Expire/time.Second))
			if eval.Err != nil {
				return eval.Err
			}
			if eval.Val != nil && eval.Val.(int64) == 1 { //The lock was acquired successfully
				r.ctx = ctx
				lockTimes, err := r.getLockTimes()
				if err != nil {
					return err
				}
				if r.config.EnableWatchdog && lockTimes <= 1 {
					r.lockSignal <- 0
				}
				return nil
			}
		}
		//Re-acquire the lock after waiting for 5 millisecond
		time.Sleep(5 * time.Millisecond)
	}
}

func (r *RedisLock) Unlock() (bool, error) {
	if r.client.Eval(r.ctx, unlockScript, []string{r.key}, r.ctx.Value(contextKey)).Val.(int64) == 1 {
		lockTimes, err := r.getLockTimes()
		if err != nil {
			return true, err
		}
		if r.config.EnableWatchdog && lockTimes <= 0 {
			r.unlockSignal <- 0
		}
		return true, nil
	}
	return false, nil
}

//Get the number of locks
func (r *RedisLock) getLockTimes() (int64, error) {
	val := r.client.HGet(r.ctx, r.key, r.ctx.Value(contextKey).(string)).StringVal()
	if val == "" {
		return 0, nil
	}
	times, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Redis Error! Get lock times failed : %s!", err.Error()))
	}
	return times, nil
}
