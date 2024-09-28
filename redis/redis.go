package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

var Pool *redis.Client

func init() {
	Pool = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		PoolSize: 1000,
	})
	_, err := Pool.Ping(context.Background()).Result()
	if err != nil {
		log.Println("redis connect error")
	}
}
