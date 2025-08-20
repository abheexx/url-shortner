package cache

import (
	"context"

	"github.com/urlshortener/internal/models"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a URL from cache
	Get(ctx context.Context, code string) (*models.ShortURL, error)

	// Set stores a URL in cache
	Set(ctx context.Context, code string, url *models.ShortURL) error

	// SetNegative sets a negative cache entry for not-found URLs
	SetNegative(ctx context.Context, code string) error

	// Delete removes a URL from cache
	Delete(ctx context.Context, code string) error

	// InvalidateExpired removes expired URLs from cache
	InvalidateExpired(ctx context.Context, codes []string) error

	// GetStats retrieves cache statistics
	GetStats(ctx context.Context) (map[string]interface{}, error)

	// Ping tests the cache connection
	Ping(ctx context.Context) error

	// Flush clears all cache entries
	Flush(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}
