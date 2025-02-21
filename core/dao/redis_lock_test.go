package dao

import (
	"context"
	"testing"
	"time"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core"
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

func TestCreateUserMap(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL: "root:abcd1234@tcp(120.79.221.36:3306)/titan_explorer?charset=utf8mb4&parseTime=True&loc=Local",
	}

	mdb, err := NewDbMgr(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = mdb.CreateUserMap(context.Background(), "1661628099@qq.com", &core.User{
		Account: "titan1lnchypve35pd69pdkvgz3pu2cl2kjx280wu0fn",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetServiceNumsByStatus(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL: "root:abcd1234@tcp(120.79.221.36:3306)/titan_explorer?charset=utf8mb4&parseTime=True&loc=Local",
	}

	mdb, err := NewDbMgr(cfg)
	if err != nil {
		t.Fatal(err)
	}

	infos, err := mdb.GetServiceNumsByStatus("titan1lnchypve35pd69pdkvgz3pu2cl2kjx280wu0fn", []core.OrderStatus{
		core.OrderStatusDone, core.OrderStatusExpired, core.OrderStatusTermination})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(infos)
}
