package redis_lock

import (
	"context"
	"github.com/satori/go.uuid"
)

const (
	ContextKey = "LockContextKey"
)

type LockContext context.Context

func Context() LockContext {
	return context.WithValue(context.Background(), ContextKey, uuid.NewV4().String())
}
