package repo

import (
	"context"

	"github.com/urlshortener/internal/models"
)

// URLRepository defines the interface for URL storage operations
type URLRepository interface {
	// CreateURL creates a new short URL
	CreateURL(ctx context.Context, url *models.ShortURL) error

	// GetURLByCode retrieves a URL by its short code
	GetURLByCode(ctx context.Context, code string) (*models.ShortURL, error)

	// GetURLMetadata retrieves URL metadata including click statistics
	GetURLMetadata(ctx context.Context, code string) (*models.URLMetadata, error)

	// DeleteURL soft deletes a URL
	DeleteURL(ctx context.Context, code string) error

	// RecordClick records a click event
	RecordClick(ctx context.Context, event *models.ClickEvent) error

	// GetExpiredURLs gets URLs that have expired
	GetExpiredURLs(ctx context.Context, limit int) ([]string, error)

	// MarkURLsAsDeleted marks multiple URLs as deleted
	MarkURLsAsDeleted(ctx context.Context, codes []string) error

	// GetURLsByUser gets URLs created by a specific user
	GetURLsByUser(ctx context.Context, user string, page, pageSize int) (*models.URLListResponse, error)

	// Close closes the repository connection
	Close() error
}
