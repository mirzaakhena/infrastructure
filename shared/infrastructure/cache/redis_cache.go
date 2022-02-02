package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	Client *redis.Client
}

// Set receive key and value as input and return error
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, exp time.Duration) error {
	err := c.Client.Set(ctx, key, value, exp).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get receive key and bytes (return object from passing by reference)
// return isExist flag and error
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := c.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("Redis key %s not found", key)
	}

	if err != nil {
		return "", err
	}
	return result, nil
}

// Del deletes by key
func (c *RedisCache) Del(ctx context.Context, key string) error {

	err := c.Client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil

}

// Reset receive key and value as input and return error
func (c *RedisCache) Reset(ctx context.Context, key string, value []byte) error {
	err := c.Client.Set(ctx, key, value, redis.KeepTTL).Err()
	if err != nil {
		return err
	}
	return nil
}
