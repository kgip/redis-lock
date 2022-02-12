package lock

import (
	"context"
	"time"
)

type Result struct {
	Val interface{}
	Err error
}

type BoolResult struct {
	Result
}

func (r *BoolResult) BoolVal() bool {
	if val, ok := r.Val.(bool); ok {
		return val
	}
	return false
}

type StringResult struct {
	Result
}

func (r *StringResult) StringVal() string {
	if val, ok := r.Val.(string); ok {
		return val
	}
	return ""
}

type RedisOps interface {
	Expire(ctx context.Context, key string, expiration time.Duration) *BoolResult
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Result
	HGet(ctx context.Context, key, field string) *StringResult
}

type RedisClientAdapter interface {
	RedisOps
}
