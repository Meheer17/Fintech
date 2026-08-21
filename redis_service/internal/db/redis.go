package db

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisDatabase wraps the go-redis Client.
type RedisDatabase struct {
	Client *redis.Client
}

// ConnectRedis initializes a connection to a Redis server.
func ConnectRedis(ctx context.Context, addr, password string, dbNum int) (*RedisDatabase, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})

	// Ping to verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := client.Ping(pingCtx).Err()
	if err != nil {
		return nil, err
	}

	return &RedisDatabase{Client: client}, nil
}

// GetNamespacedKey helper formats keys as "namespace:key" or "default:key"
func GetNamespacedKey(namespace, key string) string {
	if namespace == "" {
		return "default:" + key
	}
	return namespace + ":" + key
}

// Set stores binary data with an optional TTL.
func (r *RedisDatabase) Set(ctx context.Context, namespacedKey string, data []byte, ttl time.Duration) error {
	return r.Client.Set(ctx, namespacedKey, data, ttl).Err()
}

// Get retrieves binary data by key. Returns the bytes, a boolean indicating if found, and any error.
func (r *RedisDatabase) Get(ctx context.Context, namespacedKey string) ([]byte, bool, error) {
	val, err := r.Client.Get(ctx, namespacedKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return val, true, nil
}

// Delete removes a key from the cache. Returns true if the key was deleted.
func (r *RedisDatabase) Delete(ctx context.Context, namespacedKey string) (bool, error) {
	deletedCount, err := r.Client.Del(ctx, namespacedKey).Result()
	if err != nil {
		return false, err
	}
	return deletedCount > 0, nil
}

// InvalidateNamespace deletes all keys matching namespace:* safely using SCAN to avoid blocking Redis.
func (r *RedisDatabase) InvalidateNamespace(ctx context.Context, namespace string) (int64, error) {
	if namespace == "" {
		namespace = "default"
	}
	pattern := namespace + ":*"
	var count int64
	var cursor uint64

	for {
		// Scan keys matching the namespace pattern in batches of 100
		keys, nextCursor, err := r.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return count, err
		}

		if len(keys) > 0 {
			deleted, err := r.Client.Del(ctx, keys...).Result()
			if err != nil {
				return count, err
			}
			count += deleted
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}
