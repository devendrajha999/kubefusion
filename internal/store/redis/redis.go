package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func Connect(ctx context.Context, addr, password string) (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: 0})
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return c, nil
}
