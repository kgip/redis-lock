package redis_lock

import (
	"time"
)

type Locker interface {
	Lock(ctx LockContext) bool
	Unlock(ctx LockContext) bool
	TryLock(ctx LockContext, time time.Duration) bool //Attempt to lock for a fixed duration
}
