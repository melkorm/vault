package locksutil

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test_CreateLocks(t *testing.T) {
	locks := CreateLocks()
	if len(locks) != 256 {
		t.Fatalf("bad: len(locks): expected:256 actual:%d", len(locks))
	}
}

func TestManyLocks(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	locks := CreateLocks()
	for i := 0; i < 100; i++ {
		lockName := fmt.Sprintf("%d", r.Int())
		lock := LockForKey(locks, lockName)
		lock.Lock()
	}
}
