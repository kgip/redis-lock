package lock

import (
	"fmt"
	"sync"
	"time"
)

type LockOperator interface {
	GetLock(key string, config Config) (lock Locker)
	Lock(key string, ctx LockContext) (err error)
	TryLock(key string, ctx LockContext, timeout time.Duration) bool
	Unlock(key string) bool
}

type RedisLockOperator struct {
	locks  map[string]Locker //Save the mapping of lock keys to lock objects
	mutex  *sync.Mutex
	client RedisClientAdapter //redis connect client interface
	config *Config
}

func NewRedisLockOperator(client RedisClientAdapter, config Config) (*RedisLockOperator, error) {
	if config.Logger == nil {
		config.Logger = &DefaultLogger{Level: Info, Log: Zap()}
	}
	if client == nil {
		return nil, RedisClientNilError
	}
	return &RedisLockOperator{locks: make(map[string]Locker), mutex: &sync.Mutex{}, client: client, config: &config}, nil
}

//GetLock Get the lock object from locks and return it.
//If it does not exist, first create the lock object according to the key,
//then save the lock object to locks, and finally return the lock object.
func (operator *RedisLockOperator) GetLock(key string, config Config) (lock Locker, err error) {
	if lock = operator.locks[key]; lock == nil {
		operator.mutex.Lock()
		defer operator.mutex.Unlock()
		if lock = operator.locks[key]; lock == nil {
			operator.config.Logger.Info(fmt.Sprintf("Create a new lock with key: %s", key))
			lock, err = NewRedisLock(key, operator.client, &config)
			if err != nil {
				return nil, err
			}
			operator.locks[key] = lock
		}
	}
	return lock, nil
}

func (operator *RedisLockOperator) Lock(key string, ctx LockContext) error {
	if lock, err := operator.GetLock(key, *operator.config); err != nil {
		return err
	} else {
		return lock.Lock(ctx)
	}
}

func (operator *RedisLockOperator) TryLock(key string, ctx LockContext, timeout time.Duration) error {
	if lock, err := operator.GetLock(key, *operator.config); err != nil {
		return err
	} else {
		return lock.TryLock(ctx, timeout)
	}
}

func (operator *RedisLockOperator) Unlock(key string) (bool, error) {
	if lock := operator.locks[key]; lock != nil {
		return lock.Unlock()
	} else {
		return false, NotFoundLockError
	}
}
