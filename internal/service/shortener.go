package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/urlshortener/internal/cache"
	"github.com/urlshortener/internal/id"
	"github.com/urlshortener/internal/models"
	"github.com/urlshortener/internal/repo"
)

// ShortenerService provides URL shortening business logic
type ShortenerService struct {
	repo   repo.URLRepository
	cache  cache.Cache
	idGen  *id.Generator
	config Config
}

// Config holds service configuration
type Config struct {
	BaseURL      string
	CodeLength   int
	MaxURLLength int
	AllowedHosts []string
	BlockedHosts []string
}

// NewShortenerService creates a new shortener service
func NewShortenerService(repo repo.URLRepository, cache cache.Cache, config Config) *ShortenerService {
	return &ShortenerService{
		repo:   repo,
		cache:  cache,
		idGen:  id.NewGenerator(config.CodeLength),
		config: config,
	}
}

// CreateShortURL creates a new short URL
func (s *ShortenerService) CreateShortURL(ctx context.Context, req *models.CreateURLRequest) (*models.CreateURLResponse, error) {
	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		return nil, err
	}

	// Generate or validate custom alias
	var code string
	var customAlias bool
	if req.CustomAlias != nil && *req.CustomAlias != "" {
		code = s.idGen.GenerateCustomCode(*req.CustomAlias)
		customAlias = true
		
		// Check if custom code already exists
		if exists, _ := s.codeExists(ctx, code); exists {
			return nil, fmt.Errorf("custom alias already exists")
		}
	} else {
		// Generate unique code
		for i := 0; i < 10; i++ { // Retry up to 10 times
			code = s.idGen.GenerateCode()
			if exists, _ := s.codeExists(ctx, code); !exists {
				break
			}
		}
		customAlias = false
	}

	// Create short URL
	shortURL := &models.ShortURL{
		Code:        code,
		LongURL:     req.URL,
		ExpireAt:    req.ExpireAt,
		CustomAlias: customAlias,
		CreatedBy:   req.CreatedBy,
		Metadata:    req.Metadata,
	}

	// Save to database
	if err := s.repo.CreateURL(ctx, shortURL); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Warm cache
	if err := s.cache.Set(ctx, code, shortURL); err != nil {
		// Log error but don't fail the request
		// In production, you might want to send this to a monitoring system
	}

	// Build response
	shortURLStr := fmt.Sprintf("%s/%s", s.config.BaseURL, code)
	response := &models.CreateURLResponse{
		Code:      code,
		ShortURL:  shortURLStr,
		LongURL:   req.URL,
		ExpireAt:  req.ExpireAt,
		CreatedAt: shortURL.CreatedAt,
	}

	return response, nil
}

// GetLongURL retrieves the long URL for a given code
func (s *ShortenerService) GetLongURL(ctx context.Context, code string, userAgent, ipAddress, referer string) (*models.ShortURL, error) {
	// Try cache first
	url, err := s.cache.Get(ctx, code)
	if err == nil {
		// Cache hit - record click asynchronously
		go s.recordClickAsync(context.Background(), code, userAgent, ipAddress, referer)
		return url, nil
	}

	// Cache miss - check if it's a negative cache hit
	if err == cache.ErrURLDeleted || err == cache.ErrURLExpired {
		return nil, err
	}

	// Fallback to database
	url, err = s.repo.GetURLByCode(ctx, code)
	if err != nil {
		// Set negative cache for not found
		if err == repo.ErrURLNotFound {
			s.cache.SetNegative(ctx, code)
		}
		return nil, err
	}

	// Warm cache
	if err := s.cache.Set(ctx, code, url); err != nil {
		// Log error but continue
	}

	// Record click
	if err := s.recordClick(ctx, code, userAgent, ipAddress, referer); err != nil {
		// Log error but don't fail the request
	}

	return url, nil
}

// GetURLMetadata retrieves metadata for a URL
func (s *ShortenerService) GetURLMetadata(ctx context.Context, code string) (*models.URLMetadata, error) {
	// Try cache first for basic info
	_, err := s.cache.Get(ctx, code)
	if err == nil {
		// Get full metadata from database
		metadata, err := s.repo.GetURLMetadata(ctx, code)
		if err != nil {
			return nil, err
		}
		return metadata, nil
	}

	// Fallback to database
	metadata, err := s.repo.GetURLMetadata(ctx, code)
	if err != nil {
		return nil, err
	}

	// Warm cache with basic info
	shortURL := &models.ShortURL{
		Code:      metadata.Code,
		LongURL:   metadata.LongURL,
		CreatedAt: metadata.CreatedAt,
		ExpireAt:  metadata.ExpireAt,
		IsDeleted: metadata.IsDeleted,
	}
	
	if err := s.cache.Set(ctx, code, shortURL); err != nil {
		// Log error but continue
	}

	return metadata, nil
}

// DeleteURL deletes a URL
func (s *ShortenerService) DeleteURL(ctx context.Context, code string) error {
	// Delete from database
	if err := s.repo.DeleteURL(ctx, code); err != nil {
		return err
	}

	// Invalidate cache
	if err := s.cache.Delete(ctx, code); err != nil {
		// Log error but don't fail the request
	}

	return nil
}

// CleanupExpiredURLs removes expired URLs
func (s *ShortenerService) CleanupExpiredURLs(ctx context.Context) error {
	// Get expired URLs from database
	codes, err := s.repo.GetExpiredURLs(ctx, 100) // Process in batches
	if err != nil {
		return fmt.Errorf("failed to get expired URLs: %w", err)
	}

	if len(codes) == 0 {
		return nil
	}

	// Mark as deleted in database
	if err := s.repo.MarkURLsAsDeleted(ctx, codes); err != nil {
		return fmt.Errorf("failed to mark URLs as deleted: %w", err)
	}

	// Invalidate from cache
	if err := s.cache.InvalidateExpired(ctx, codes); err != nil {
		return fmt.Errorf("failed to invalidate expired URLs from cache: %w", err)
	}

	return nil
}

// validateURL validates the input URL
func (s *ShortenerService) validateURL(longURL string) error {
	// Check length
	if len(longURL) > s.config.MaxURLLength {
		return fmt.Errorf("URL too long (max %d characters)", s.config.MaxURLLength)
	}

	// Parse URL
	parsed, err := url.Parse(longURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS URLs are allowed")
	}

	// Check host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	// Check blocked hosts
	for _, blocked := range s.config.BlockedHosts {
		if strings.Contains(parsed.Host, blocked) {
			return fmt.Errorf("URL host is blocked")
		}
	}

	// Check allowed hosts if specified
	if len(s.config.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range s.config.AllowedHosts {
			if strings.Contains(parsed.Host, allowedHost) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("URL host is not in allowed list")
		}
	}

	return nil
}

// codeExists checks if a code already exists
func (s *ShortenerService) codeExists(ctx context.Context, code string) (bool, error) {
	// Try cache first
	_, err := s.cache.Get(ctx, code)
	if err == nil {
		return true, nil
	}

	// Check database
	_, err = s.repo.GetURLByCode(ctx, code)
	if err == repo.ErrURLNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// recordClick records a click event
func (s *ShortenerService) recordClick(ctx context.Context, code, userAgent, ipAddress, referer string) error {
	event := &models.ClickEvent{
		Code:      code,
		UserAgent: &userAgent,
		IPAddress: &ipAddress,
		Referer:   &referer,
	}

	return s.repo.RecordClick(ctx, event)
}

// recordClickAsync records a click event asynchronously
func (s *ShortenerService) recordClickAsync(ctx context.Context, code, userAgent, ipAddress, referer string) {
	// Use a separate context with timeout for async operations
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_ = s.recordClick(ctx, code, userAgent, ipAddress, referer)
}
