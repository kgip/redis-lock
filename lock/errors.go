package lock

import "errors"

var (
	KeyEmptyError       = errors.New("lock key is empty")
	RedisClientNilError = errors.New("redis client is nil")
	TimeoutError        = errors.New("acquire lock timeout")
	NotFoundLockError   = errors.New("lock is not found")
)
