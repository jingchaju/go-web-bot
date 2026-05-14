package redisclient

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Init(addr, password string, db, poolSize int) error {
	client = redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db, PoolSize: poolSize})
	return client.Ping(context.Background()).Err()
}
func Get() *redis.Client { return client }

type Queue struct {
	Name string
	TTL  time.Duration
}

func RegisterQueue(name string, ttl time.Duration) Queue { return Queue{Name: name, TTL: ttl} }
func (q Queue) Push(ctx context.Context, payload string) error {
	return client.RPush(ctx, q.Name, payload).Err()
}
func (q Queue) Pop(ctx context.Context, timeout time.Duration) (string, error) {
	items, err := client.BLPop(ctx, timeout, q.Name).Result()
	if err != nil {
		return "", err
	}
	if len(items) < 2 {
		return "", redis.Nil
	}
	return items[1], nil
}
