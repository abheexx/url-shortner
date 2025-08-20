package models

import (
	"time"
)

// ShortURL represents a shortened URL in the database
type ShortURL struct {
	ID          int64      `json:"id" db:"id"`
	Code        string     `json:"code" db:"code"`
	LongURL     string     `json:"long_url" db:"long_url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpireAt    *time.Time `json:"expire_at,omitempty" db:"expire_at"`
	IsDeleted   bool       `json:"is_deleted" db:"is_deleted"`
	CustomAlias bool       `json:"custom_alias" db:"custom_alias"`
	CreatedBy   *string    `json:"created_by,omitempty" db:"created_by"`
	Metadata    *string    `json:"metadata,omitempty" db:"metadata"`
}

// CreateURLRequest represents the request to create a short URL
type CreateURLRequest struct {
	URL         string     `json:"url" binding:"required,url"`
	CustomAlias *string    `json:"custom_alias,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
	CreatedBy   *string    `json:"created_by,omitempty"`
	Metadata    *string    `json:"metadata,omitempty"`
}

// CreateURLResponse represents the response after creating a short URL
type CreateURLResponse struct {
	Code      string     `json:"code"`
	ShortURL  string     `json:"short_url"`
	LongURL   string     `json:"long_url"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// URLMetadata represents the metadata for a short URL
type URLMetadata struct {
	Code         string     `json:"code"`
	LongURL      string     `json:"long_url"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpireAt     *time.Time `json:"expire_at,omitempty"`
	TotalClicks  int64      `json:"total_clicks"`
	LastAccessAt *time.Time `json:"last_access_at,omitempty"`
	IsDeleted    bool       `json:"is_deleted"`
}

// ClickEvent represents a click event for analytics
type ClickEvent struct {
	ID         int64      `json:"id" db:"id"`
	Code       string     `json:"code" db:"code"`
	Timestamp  time.Time  `json:"timestamp" db:"ts"`
	UserAgent  *string    `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress  *string    `json:"ip_address,omitempty" db:"ip_address"`
	Referer    *string    `json:"referer,omitempty" db:"referer"`
	Country    *string    `json:"country,omitempty" db:"country"`
	DeviceType *string    `json:"device_type,omitempty" db:"device_type"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// URLListResponse represents a paginated list of URLs
type URLListResponse struct {
	URLs       []URLMetadata `json:"urls"`
	Pagination Pagination    `json:"pagination"`
	Total      int64         `json:"total"`
}
