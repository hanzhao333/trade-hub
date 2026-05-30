package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(addr string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{Addr: addr})
}

func Ping(ctx context.Context, c *goredis.Client) error {
	return c.Ping(ctx).Err()
}

// 第 3 周学习：在此封装 SetPoolTicker / GetPoolTicker 等缓存方法
