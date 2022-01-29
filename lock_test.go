package redis_lock

import (
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"testing"
	"time"
)

var Client = redis.NewClient(&redis.Options{Addr: "192.168.32.36:6379"})

func TestUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(uuid.NewV4().String())
	}
}

func TestRedisOptions(t *testing.T) {
	lock := NewRedisLock("key", &redis.Client{}, Timeout(5*time.Second))
	t.Log(lock)
}

func TestDuration(t *testing.T) {
	timeout := 5 * time.Second
	ticker := time.NewTicker(timeout / 2)
	for {
		select {
		case signal := <-ticker.C:
			t.Log(signal)
		default:
			time.Sleep(time.Millisecond)
			//t.Log("waiting!!!")
		}
	}
}

func TestEval(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Log(Client.Eval("if redis.call('hset', KEYS[1], ARGV[1], 1) then return 1 else return 0 end", []string{"aaaa"}, "bbbb").Val().(int64))
	}
}

var LockScript = `if redis.call('exists',KEYS[1]) == 1 then
					  local r = redis.call('hget', KEYS[1],ARGV[1])
					  if r then
						  r = r + 1
						  if redis.call('hset', KEYS[1], ARGV[1], r) then
							  if redis.call('expire', KEYS[1], ARGV[2]) == 1 then
								  return 1
                              else
                                  redis.call('del', KEYS[1])
                              end
                          end
					  end
				  elseif redis.call('hset', KEYS[1], ARGV[1], 1) then
					  if redis.call('expire', KEYS[1], ARGV[2]) == 1 then
						  return 1
                      else
                          redis.call('del', KEYS[1])
					  end
				  end
				  return 0`

func TestLockEval(t *testing.T) {
	for i := 0; i < 100; i++ {
		eval := Client.Eval(LockScript, []string{"key1"}, "iheaifheoi", 1000)
		if eval.Val().(int64) == 0 {
			t.Error(eval)
		} else {
			t.Log(eval)
		}
	}
}

var ctx = Context()

func TestLock(t *testing.T) {
	lock := NewRedisLock("key1", Client)
	for i := 0; i < 2; i++ {
		lock.Lock(ctx)
	}
}

var UnlockScript = `if redis.call('exists',KEYS[1]) == 1 then
						local r = redis.call('hget', KEYS[1],ARGV[1])
						if r then
                            r = r - 1 
                            if r == 0 then
                            	if redis.call('del', KEYS[1]) == 1 then
									return 1
								end
                            elseif redis.call('hset', KEYS[1], ARGV[1], r) then
                                return 1
							end
						end
                    else 
                        return 1
                    end
                    return 0`

func TestUnlockEval(t *testing.T) {
	for i := 0; i < 99; i++ {
		eval := Client.Eval(UnlockScript, []string{"key1"}, "iheaifheoi")
		if eval.Val().(int64) == 0 {
			t.Error(eval)
		} else {
			t.Log(eval)
		}
	}
}

func TestUnlock(t *testing.T) {
	lock := NewRedisLock("key1", Client)
	lock.Lock(ctx)
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		t.Log("waiting!!!!!!!!!!!")
	}
	lock.Unlock()
}

func TestWatchDog(t *testing.T) {
	lock := NewRedisLock("key1", Client, EnableWatchdog(true), Expire(10*time.Second))
	lock.Lock(ctx)
	for i := 0; i < 50; i++ {
		time.Sleep(time.Second)
		t.Log("waiting!!!!!!!!!!!")
	}
	lock.Unlock()
}

func TestHGet(t *testing.T) {
	val := Client.HGet("aaaa", "bbbb").Val()
	if val == "" {
		t.Error("nil")
	} else {
		t.Log(val)
	}
}
