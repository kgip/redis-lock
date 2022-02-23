package lock

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type RedisLockOperator struct {
	locks   *sync.Map //Save the mapping of lock keys to lock objects
	mutex   *sync.Mutex
	client  RedisClientAdapter //redis connect client interface
	options []option
}

func NewRedisLockOperator(client RedisClientAdapter, options ...option) *RedisLockOperator {
	return &RedisLockOperator{locks: &sync.Map{}, mutex: &sync.Mutex{}, client: client, options: options}
}

func handleError() error {
	if e := recover(); e != nil {
		if msg, ok := e.(string); ok {
			return errors.New(msg)
		} else if err, ok := e.(error); ok {
			return err
		} else {
			return errors.New(fmt.Sprintf("unknown error：%v", e))
		}
	}
	return nil
}

//tryGetLock Get the lock object from locks and return it. If it does not exist, return nil.
func (operator *RedisLockOperator) tryGetLock(key string) Locker {
	if lockItem, ok := operator.locks.Load(key); ok {
		if lock, ok := lockItem.(Locker); ok {
			return lock
		} else {
			panic(fmt.Sprintf("unknown object: %v", lock))
		}
	}
	return nil
}

//GetLock Get the lock object from locks and return it.
//If it does not exist, first create the lock object according to the key,
//then save the lock object to locks, and finally return the lock object.
func (operator *RedisLockOperator) GetLock(key string, options ...option) (lock Locker) {
	if lock = operator.tryGetLock(key); lock == nil {
		operator.mutex.Lock()
		defer operator.mutex.Unlock()
		if lock = operator.tryGetLock(key); lock == nil {
			lock = NewRedisLock(key, operator.client, options...)
			operator.locks.Store(key, lock)
		}
	}
	return lock
}

func (operator *RedisLockOperator) Lock(key string, ctx LockContext) (err error) {
	defer func() { err = handleError() }()
	lock := operator.GetLock(key, operator.options...)
	lock.Lock(ctx)
	return nil
}

func (operator *RedisLockOperator) TryLock(key string, ctx LockContext, timeout time.Duration) bool {
	defer func() {
		if err := handleError(); err != nil {
			log.Println(err)
		}
	}()
	lock := operator.GetLock(key, operator.options...)
	return lock.TryLock(ctx, timeout)
}

func (operator *RedisLockOperator) Unlock(key string) bool {
	defer func() {
		if err := handleError(); err != nil {
			log.Println(err)
		}
	}()
	if lock := operator.tryGetLock(key); lock != nil {
		return lock.Unlock()
	}
	return false
}
