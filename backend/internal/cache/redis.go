package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/parse-companies/backend/internal/domain"
)

// redisCache is the production Cache backed by Redis.
type redisCache struct {
	rdb *redis.Client
}

// NewRedis builds a Cache from a redis:// URL.
func NewRedis(redisURL string) (Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache: parse url: %w", err)
	}
	return &redisCache{rdb: redis.NewClient(opt)}, nil
}

// NewRedisFromClient wraps an existing client (used in tests with miniredis).
func NewRedisFromClient(c *redis.Client) Cache { return &redisCache{rdb: c} }

func (r *redisCache) Get(ctx context.Context, key string) (Entry, bool, error) {
	b, err := r.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, fmt.Errorf("cache: get: %w", err)
	}
	entry, err := decode(b)
	if err != nil {
		return Entry{}, false, fmt.Errorf("cache: decode: %w", err)
	}
	return entry, true, nil
}

func (r *redisCache) Set(ctx context.Context, key string, companies []domain.Company, ttl time.Duration) error {
	b, err := encode(Entry{Companies: companies, FetchedAt: time.Now()})
	if err != nil {
		return fmt.Errorf("cache: encode: %w", err)
	}
	if err := r.rdb.Set(ctx, key, b, ttl).Err(); err != nil {
		return fmt.Errorf("cache: set: %w", err)
	}
	return nil
}
