package core

import (
	"github.com/go-redis/redis"
	"time"
)

type option func(o interface{}) interface{}

type optionHandler func(parameter interface{}) option

var (
	Key optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			if r, ok := o.(*RedisLock); ok {
				param := parameter.(string)
				r.key = param
			} else {
				panic("Key is not of type string")
			}
			return o
		}
	}
	Timeout optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			if r, ok := o.(*RedisLock); ok {
				param := parameter.(time.Duration)
				r.timeout = param
			} else {
				panic("Timeout is not of type time.Duration")
			}
			return o
		}
	}
	Client optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			if r, ok := o.(*RedisLock); ok {
				param := parameter.(redis.Cmdable)
				r.client = param
			} else {
				panic("Context is not of type redis.Cmdable")
			}
			return o
		}
	}
)
