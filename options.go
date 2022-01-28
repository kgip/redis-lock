package redis_lock

import (
	"github.com/go-redis/redis"
	"time"
)

type option func(o interface{}) interface{}

type optionHandler func(parameter interface{}) option

var (
	Key optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(string); ok {
				r.key = param
			} else {
				panic("Key is not of type string")
			}
			return o
		}
	}
	Expire optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(time.Duration); ok {
				r.expireSecond = int(param) / int(time.Second)
				if r.expireSecond <= 0 {
					panic("The lock expiration time is less than 1 second")
				}
				r.expire = param
			} else {
				panic("Expire is not of type time.Duration")
			}
			return o
		}
	}
	Timeout optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(time.Duration); ok {
				r.timeout = param
			} else {
				panic("Timeout is not of type time.Duration")
			}
			return o
		}
	}
	Client optionHandler = func(parameter interface{}) option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(redis.Cmdable); ok {
				r.client = param
			} else {
				panic("Context is not of type redis.Cmdable")
			}
			return o
		}
	}
)