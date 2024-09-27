package dao

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
)

func TestLock(t *testing.T) {
	RedisCache = redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		PoolSize:     100,
		ReadTimeout:  10 * time.Second,
		DialTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	if f, err := AcquireLock(RedisCache, "lock-test", "1", 15*time.Second); err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	} else {
		if f {
			t.Logf("Lock successfully acquired")
			if err := ReleaseLock(RedisCache, "lock-test", "1"); err != nil {
				t.Fatalf("Failed to release lock: %v", err)
			} else {
				t.Logf("Lock successfully released")
			}
		} else {
			t.Fatalf("Failed to acquire lock: %v", err)
		}
	}
}
