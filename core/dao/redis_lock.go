package dao

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
)

// 尝试获取分布式锁
func AcquireLock(rdb *redis.Client, lockKey string, lockValue string, expiration time.Duration) (bool, error) {
	// 使用 SETNX 尝试获取锁，并设置锁的过期时间
	success, err := rdb.SetNX(context.Background(), lockKey, lockValue, expiration).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

// 释放锁
func ReleaseLock(rdb *redis.Client, lockKey string, lockValue string) error {
	// 使用 Lua 脚本保证只有持有锁的客户端才能释放锁
	luaScript := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end`
	_, err := rdb.Eval(context.Background(), luaScript, []string{lockKey}, lockValue).Result()
	return err
}
