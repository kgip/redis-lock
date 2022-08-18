package adapters

import (
	"context"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/kgip/redis-lock/lock"
	"time"
)

type GoRedisV8Adapter struct {
	client redisV8.Cmdable
}

func NewGoRedisV8Adapter(client redisV8.Cmdable) *GoRedisV8Adapter {
	if client == nil {
		panic("client can not be nil")
	}
	return &GoRedisV8Adapter{client: client}
}

func (g *GoRedisV8Adapter) Expire(ctx context.Context, key string, expiration time.Duration) *lock.BoolResult {
	r := g.client.Expire(ctx, key, expiration)
	return &lock.BoolResult{Result: lock.Result{Val: r.Val(), Err: r.Err()}}
}

func (g *GoRedisV8Adapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *lock.Result {
	r := g.client.Eval(ctx, script, keys, args...)
	return &lock.Result{Val: r.Val(), Err: r.Err()}
}

func (g *GoRedisV8Adapter) HGet(ctx context.Context, key, field string) *lock.StringResult {
	r := g.client.HGet(ctx, key, field)
	return &lock.StringResult{Result: lock.Result{Val: r.Val(), Err: r.Err()}}
}
