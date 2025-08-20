package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/urlshortener/internal/models"
)

// RedisCache implements the cache interface using Redis
type RedisCache struct {
	client     *redis.Client
	ttl        time.Duration
	negativeTTL time.Duration
}

// CachedURL represents a URL stored in cache
type CachedURL struct {
	LongURL   string     `json:"long_url"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`
	IsDeleted bool       `json:"is_deleted"`
	CreatedAt time.Time  `json:"created_at"`
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int, ttl, negativeTTL time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10,
		MinIdleConns: 5,
		MaxRetries: 3,
	})

	return &RedisCache{
		client:      client,
		ttl:         ttl,
		negativeTTL: negativeTTL,
	}
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Get retrieves a URL from cache
func (c *RedisCache) Get(ctx context.Context, code string) (*models.ShortURL, error) {
	key := fmt.Sprintf("url:%s", code)
	
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var cached CachedURL
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached URL: %w", err)
	}

	// Check if URL is deleted
	if cached.IsDeleted {
		return nil, ErrURLDeleted
	}

	// Check if URL has expired
	if cached.ExpireAt != nil && time.Now().After(*cached.ExpireAt) {
		return nil, ErrURLExpired
	}

	return &models.ShortURL{
		Code:      code,
		LongURL:   cached.LongURL,
		CreatedAt: cached.CreatedAt,
		ExpireAt:  cached.ExpireAt,
		IsDeleted: cached.IsDeleted,
	}, nil
}

// Set stores a URL in cache
func (c *RedisCache) Set(ctx context.Context, code string, url *models.ShortURL) error {
	key := fmt.Sprintf("url:%s", code)
	
	cached := CachedURL{
		LongURL:   url.LongURL,
		ExpireAt:  url.ExpireAt,
		IsDeleted: url.IsDeleted,
		CreatedAt: url.CreatedAt,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("failed to marshal URL for cache: %w", err)
	}

	// Calculate TTL
	ttl := c.ttl
	if url.ExpireAt != nil {
		// If URL has expiration, use the shorter of cache TTL or time until expiration
		timeUntilExpiry := time.Until(*url.ExpireAt)
		if timeUntilExpiry < ttl {
			ttl = timeUntilExpiry
		}
	}

	// Add some buffer to TTL to avoid edge cases
	if ttl > 0 {
		ttl += time.Minute
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// SetNegative sets a negative cache entry for not-found URLs
func (c *RedisCache) SetNegative(ctx context.Context, code string) error {
	key := fmt.Sprintf("url:%s", code)
	
	// Store a special marker for negative cache
	negative := CachedURL{
		IsDeleted: true,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(negative)
	if err != nil {
		return fmt.Errorf("failed to marshal negative cache: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.negativeTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set negative cache: %w", err)
	}

	return nil
}

// Delete removes a URL from cache
func (c *RedisCache) Delete(ctx context.Context, code string) error {
	key := fmt.Sprintf("url:%s", code)
	
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// InvalidateExpired removes expired URLs from cache
func (c *RedisCache) InvalidateExpired(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}

	// Build pipeline for batch deletion
	pipe := c.client.Pipeline()
	for _, code := range codes {
		key := fmt.Sprintf("url:%s", code)
		pipe.Del(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate expired URLs: %w", err)
	}

	return nil
}

// GetStats retrieves cache statistics
func (c *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info := c.client.Info(ctx, "stats").Val()
	
	stats := make(map[string]interface{})
	stats["info"] = info
	
	// Get memory usage
	memory := c.client.Info(ctx, "memory").Val()
	stats["memory"] = memory
	
	// Get client count
	clientList := c.client.ClientList(ctx).Val()
	stats["clients"] = len(clientList)
	
	return stats, nil
}

// Ping tests the Redis connection
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Flush clears all cache entries (use with caution)
func (c *RedisCache) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Custom errors
var (
	ErrCacheMiss   = fmt.Errorf("cache miss")
	ErrURLDeleted  = fmt.Errorf("URL is deleted")
	ErrURLExpired  = fmt.Errorf("URL has expired")
)
