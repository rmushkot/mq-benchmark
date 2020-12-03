package redis

import (
	"fmt"

	"github.com/go-redis/redis/v8"
)

const redisPort = 6379

func newClient(host string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, redisPort),
	})
}
