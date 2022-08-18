package lock

import (
	"time"
)

type Option func(o interface{}) interface{}

type optionHandler func(parameter interface{}) Option

var (
	key optionHandler = func(parameter interface{}) Option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(string); ok {
				if param == "" {
					panic("lock key cannot be empty")
				}
				r.key = param
			} else {
				panic("Key is not of type string")
			}
			return o
		}
	}
	client optionHandler = func(parameter interface{}) Option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(RedisClientAdapter); ok {
				r.client = param
			} else {
				panic("Context is not of type redis.Cmdable")
			}
			return o
		}
	}
	Expire optionHandler = func(parameter interface{}) Option {
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
	Timeout optionHandler = func(parameter interface{}) Option {
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
	EnableWatchdog optionHandler = func(parameter interface{}) Option {
		return func(o interface{}) interface{} {
			r := o.(*RedisLock)
			if param, ok := parameter.(bool); ok {
				r.enableWatchdog = param
			} else {
				panic("EnableWatchdog is not of type bool")
			}
			return o
		}
	}
)
