package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/urlshortener/internal/models"
	_ "github.com/lib/pq"
)

// PostgresRepo implements the URL repository interface
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo creates a new PostgreSQL repository
func NewPostgresRepo(dsn string) (*PostgresRepo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepo{db: db}, nil
}

// Close closes the database connection
func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// CreateURL creates a new short URL
func (r *PostgresRepo) CreateURL(ctx context.Context, url *models.ShortURL) error {
	query := `
		INSERT INTO short_urls (code, long_url, expire_at, custom_alias, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		url.Code, url.LongURL, url.ExpireAt, url.CustomAlias, url.CreatedBy, url.Metadata,
	).Scan(&url.ID, &url.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

// GetURLByCode retrieves a URL by its short code
func (r *PostgresRepo) GetURLByCode(ctx context.Context, code string) (*models.ShortURL, error) {
	query := `
		SELECT id, code, long_url, created_at, expire_at, is_deleted, custom_alias, created_by, metadata
		FROM short_urls
		WHERE code = $1 AND is_deleted = false`

	url := &models.ShortURL{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&url.ID, &url.Code, &url.LongURL, &url.CreatedAt, &url.ExpireAt,
		&url.IsDeleted, &url.CustomAlias, &url.CreatedBy, &url.Metadata,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	// Check if URL has expired
	if url.ExpireAt != nil && time.Now().After(*url.ExpireAt) {
		return nil, ErrURLExpired
	}

	return url, nil
}

// GetURLMetadata retrieves URL metadata including click statistics
func (r *PostgresRepo) GetURLMetadata(ctx context.Context, code string) (*models.URLMetadata, error) {
	query := `
		SELECT 
			s.code, s.long_url, s.created_at, s.expire_at, s.is_deleted,
			COALESCE(cs.total_clicks, 0) as total_clicks,
			cs.last_access_at
		FROM short_urls s
		LEFT JOIN click_stats cs ON s.code = cs.code
		WHERE s.code = $1 AND s.is_deleted = false`

	metadata := &models.URLMetadata{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&metadata.Code, &metadata.LongURL, &metadata.CreatedAt, &metadata.ExpireAt,
		&metadata.IsDeleted, &metadata.TotalClicks, &metadata.LastAccessAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL metadata: %w", err)
	}

	// Check if URL has expired
	if metadata.ExpireAt != nil && time.Now().After(*metadata.ExpireAt) {
		return nil, ErrURLExpired
	}

	return metadata, nil
}

// DeleteURL soft deletes a URL
func (r *PostgresRepo) DeleteURL(ctx context.Context, code string) error {
	query := `UPDATE short_urls SET is_deleted = true WHERE code = $1`
	
	result, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrURLNotFound
	}

	return nil
}

// RecordClick records a click event
func (r *PostgresRepo) RecordClick(ctx context.Context, event *models.ClickEvent) error {
	query := `
		INSERT INTO click_events (code, user_agent, ip_address, referer, country, device_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, ts`

	err := r.db.QueryRowContext(ctx, query,
		event.Code, event.UserAgent, event.IPAddress, event.Referer, event.Country, event.DeviceType,
	).Scan(&event.ID, &event.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	return nil
}

// GetExpiredURLs gets URLs that have expired
func (r *PostgresRepo) GetExpiredURLs(ctx context.Context, limit int) ([]string, error) {
	query := `
		SELECT code FROM short_urls
		WHERE expire_at IS NOT NULL AND expire_at < NOW() AND is_deleted = false
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired URLs: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("failed to scan expired URL code: %w", err)
		}
		codes = append(codes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired URLs: %w", err)
	}

	return codes, nil
}

// MarkURLsAsDeleted marks multiple URLs as deleted
func (r *PostgresRepo) MarkURLsAsDeleted(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}

	// Build query with placeholders
	query := `UPDATE short_urls SET is_deleted = true WHERE code = ANY($1)`
	
	_, err := r.db.ExecContext(ctx, query, codes)
	if err != nil {
		return fmt.Errorf("failed to mark URLs as deleted: %w", err)
	}

	return nil
}

// GetURLsByUser gets URLs created by a specific user
func (r *PostgresRepo) GetURLsByUser(ctx context.Context, user string, page, pageSize int) (*models.URLListResponse, error) {
	offset := (page - 1) * pageSize

	// Get total count
	countQuery := `SELECT COUNT(*) FROM short_urls WHERE created_by = $1 AND is_deleted = false`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, user).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL count: %w", err)
	}

	// Get URLs
	query := `
		SELECT 
			s.code, s.long_url, s.created_at, s.expire_at, s.is_deleted,
			COALESCE(cs.total_clicks, 0) as total_clicks,
			cs.last_access_at
		FROM short_urls s
		LEFT JOIN click_stats cs ON s.code = cs.code
		WHERE s.created_by = $1 AND s.is_deleted = false
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, user, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.URLMetadata
	for rows.Next() {
		var url models.URLMetadata
		err := rows.Scan(
			&url.Code, &url.LongURL, &url.CreatedAt, &url.ExpireAt,
			&url.IsDeleted, &url.TotalClicks, &url.LastAccessAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating URLs: %w", err)
	}

	return &models.URLListResponse{
		URLs: urls,
		Pagination: models.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
		Total: total,
	}, nil
}

// Custom errors
var (
	ErrURLNotFound = fmt.Errorf("URL not found")
	ErrURLExpired  = fmt.Errorf("URL has expired")
)
