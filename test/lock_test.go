package test

import (
	"github.com/go-redis/redis"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/kgip/redis-lock/adapters"
	"github.com/kgip/redis-lock/lock"
	uuid "github.com/satori/go.uuid"
	"sync"
	"testing"
	"time"
)

var V8Client lock.RedisClientAdapter = adapters.NewGoRedisV8Adapter(redisV8.NewClient(&redisV8.Options{Addr: "192.168.1.10:6379"}))

var Client lock.RedisClientAdapter = adapters.NewGoRedisAdapter(redis.NewClient(&redis.Options{Addr: "192.168.1.10:6379"}))

func TestUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(uuid.NewV4().String())
	}
}

func TestRedisOptions(t *testing.T) {
	lock := lock.NewRedisLock("key", adapters.NewGoRedisV8Adapter(&redisV8.Client{}), lock.Timeout(5*time.Second))
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
		t.Log(Client.Eval(lock.Context(), "if redis.call('hset', KEYS[1], ARGV[1], 1) then return 1 else return 0 end", []string{"aaaa"}, "bbbb").Val.(int64))
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
		eval := Client.Eval(lock.Context(), LockScript, []string{"key1"}, "iheaifheoi", 1000)
		if eval.Val.(int64) == 0 {
			t.Error(eval)
		} else {
			t.Log(eval)
		}
	}
}

var ctx = lock.Context()

func TestLock(t *testing.T) {
	lock := lock.NewRedisLock("key1", Client)
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
		eval := Client.Eval(lock.Context(), UnlockScript, []string{"key1"}, "iheaifheoi")
		if eval.Val.(int64) == 0 {
			t.Error(eval)
		} else {
			t.Log(eval)
		}
	}
}

func TestUnlock(t *testing.T) {
	lock := lock.NewRedisLock("key1", Client)
	lock.Lock(ctx)
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		t.Log("waiting!!!!!!!!!!!")
	}
	lock.Unlock()
}

func TestWatchDog(t *testing.T) {
	lock := lock.NewRedisLock("key1", Client, lock.EnableWatchdog(true), lock.Expire(10*time.Second))
	lock.Lock(ctx)
	for i := 0; i < 50; i++ {
		time.Sleep(time.Second)
		t.Log("waiting!!!!!!!!!!!")
	}
	lock.Unlock()
}

func TestAdapter(t *testing.T) {
	val := Client.HGet(lock.Context(), "aaaa", "bbbb").Val
	if val == "" {
		t.Error("nil")
	} else {
		t.Log(val)
	}
}

func TestReentrant(t *testing.T) {
	locker := lock.NewRedisLock("counter", Client)
	ctx := lock.Context()
	locker.Lock(ctx)
	locker.Lock(ctx)
	t.Log("lock success")
	locker.Unlock()
	locker.Unlock()
}

var wg = sync.WaitGroup{}

func TestBatchSpeed(t *testing.T) {
	locker := lock.NewRedisLock("counter", Client)
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			locker.Lock(lock.Context())
			time.Sleep(time.Millisecond)
			locker.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestRedisLock_Unlock(t *testing.T) {
	for i := 0; i < 10; i++ {
		n := 0
		locker := lock.NewRedisLock("counter", Client)
		wg.Add(20)
		for i := 0; i < 10; i++ {
			go func() {
				ctx := lock.Context()
				locker.Lock(ctx)
				locker.Lock(ctx)
				n++
				locker.Unlock()
				locker.Unlock()
				wg.Done()
			}()
		}

		for i := 0; i < 10; i++ {
			go func() {
				ctx := lock.Context()
				locker.Lock(ctx)
				n--
				locker.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()
		t.Log(n)
	}
}

func TestRedisLockOperator(t *testing.T) {
	lockOperator := lock.NewRedisLockOperator(V8Client)
	var key = "aaa"
	lockOperator.Lock(key, lock.Context())
	for i := 0; i < 10; i++ {
		t.Log("lock operator test")
		time.Sleep(10 * time.Second)
	}
	lockOperator.Unlock(key)
}
