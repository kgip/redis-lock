package adapters

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/kgip/redis-lock/lock"
	"time"
)

type GoRedisAdapter struct {
	client *redis.Client
}

func NewGoRedisAdapter(client *redis.Client) *GoRedisAdapter {
	if client == nil {
		panic("client can not be nil")
	}
	return &GoRedisAdapter{client: client}
}

func (g *GoRedisAdapter) Expire(ctx context.Context, key string, expiration time.Duration) *lock.BoolResult {
	r := g.client.Expire(key, expiration)
	return &lock.BoolResult{Result: lock.Result{Val: r.Val(), Err: r.Err()}}
}

func (g *GoRedisAdapter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *lock.Result {
	r := g.client.Eval(script, keys, args...)
	return &lock.Result{Val: r.Val(), Err: r.Err()}
}

func (g *GoRedisAdapter) HGet(ctx context.Context, key, field string) *lock.StringResult {
	r := g.client.HGet(key, field)
	return &lock.StringResult{Result: lock.Result{Val: r.Val(), Err: r.Err()}}
}
