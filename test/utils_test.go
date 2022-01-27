package test

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestUUID(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(uuid.NewV4().String())
	}
}

func TestRedisOptions(t *testing.T) {

}