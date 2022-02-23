package cache

import (
	"context"
	"time"
)

// Cache We only use 3 base function in redis
type Cache interface {

	// Set put the initial value
	Set(ctx context.Context, key string, value []byte, exp time.Duration) error

	// Get the value
	Get(ctx context.Context, key string) (string, error)

	// Del Delete the value
	Del(ctx context.Context, key string) error

	// Exist Check existing key
	Exist(ctx context.Context, key string) (bool, error)
}
