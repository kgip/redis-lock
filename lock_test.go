package redis_lock

import (
	"fmt"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"testing"
	"time"
)

func TestUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(uuid.NewV4().String())
	}
}

func TestRedisOptions(t *testing.T) {
	lock := NewRedisLock(Key("key"), Client(&redis.Client{}), Timeout(5*time.Second))
	t.Log(lock)
}

func TestDuration(t *testing.T) {
	timeout := 5 * time.Minute
	fmt.Println(int(timeout) / int(time.Second))
}

func TestEval(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "192.168.32.36:6379"})
	t.Log(client.Eval("if redis.call('hget', KEYS[1], ARGV[1]) then return 1 else return 0 end", []string{"aaaa"}, "bbbb").Val().(int64))
}
