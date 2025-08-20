package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/urlshortener/internal/models"
	"github.com/urlshortener/internal/service"
)

// Handler provides HTTP handlers for the URL shortener API
type Handler struct {
	service *service.ShortenerService
	baseURL string
}

// NewHandler creates a new HTTP handler
func NewHandler(service *service.ShortenerService, baseURL string) *Handler {
	return &Handler{
		service: service,
		baseURL: baseURL,
	}
}

// CreateShortURL handles POST /api/v1/shorten
func (h *Handler) CreateShortURL(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Extract user info from headers (could be JWT token in production)
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		req.CreatedBy = &userID
	}

	// Create short URL
	response, err := h.service.CreateShortURL(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "internal_error"
		
		if strings.Contains(err.Error(), "custom alias already exists") {
			status = http.StatusConflict
			errorCode = "alias_exists"
		} else if strings.Contains(err.Error(), "invalid URL") {
			status = http.StatusBadRequest
			errorCode = "invalid_url"
		} else if strings.Contains(err.Error(), "URL too long") {
			status = http.StatusBadRequest
			errorCode = "url_too_long"
		} else if strings.Contains(err.Error(), "blocked") {
			status = http.StatusForbidden
			errorCode = "url_blocked"
		}

		c.JSON(status, models.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectToLongURL handles GET /:code
func (h *Handler) RedirectToLongURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_code",
			Message: "URL code is required",
		})
		return
	}

	// Extract request information for analytics
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	referer := c.GetHeader("Referer")

	// Get long URL
	url, err := h.service.GetLongURL(c.Request.Context(), code, userAgent, ipAddress, referer)
	if err != nil {
		status := http.StatusNotFound
		errorCode := "url_not_found"
		
		if strings.Contains(err.Error(), "expired") {
			status = http.StatusGone
			errorCode = "url_expired"
		} else if strings.Contains(err.Error(), "deleted") {
			status = http.StatusGone
			errorCode = "url_deleted"
		}

		c.JSON(status, models.ErrorResponse{
			Error:   errorCode,
			Message: "URL not found or no longer available",
		})
		return
	}

	// Redirect to long URL
	c.Redirect(http.StatusMovedPermanently, url.LongURL)
}

// GetURLMetadata handles GET /api/v1/urls/:code
func (h *Handler) GetURLMetadata(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_code",
			Message: "URL code is required",
		})
		return
	}

	// Get metadata
	metadata, err := h.service.GetURLMetadata(c.Request.Context(), code)
	if err != nil {
		status := http.StatusNotFound
		errorCode := "url_not_found"
		
		if strings.Contains(err.Error(), "expired") {
			status = http.StatusGone
			errorCode = "url_expired"
		}

		c.JSON(status, models.ErrorResponse{
			Error:   errorCode,
			Message: "URL not found or no longer available",
		})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// DeleteURL handles DELETE /api/v1/urls/:code
func (h *Handler) DeleteURL(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_code",
			Message: "URL code is required",
		})
		return
	}

	// TODO: Add authentication/authorization check
	// For now, allow deletion (in production, check if user owns the URL)

	// Delete URL
	if err := h.service.DeleteURL(c.Request.Context(), code); err != nil {
		status := http.StatusNotFound
		errorCode := "url_not_found"
		
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}

		c.JSON(status, models.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
		"code":    code,
	})
}

// HealthCheck handles GET /api/v1/healthz
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services: map[string]string{
			"api": "healthy",
		},
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck handles GET /api/v1/readyz
func (h *Handler) ReadinessCheck(c *gin.Context) {
	// Check if service is ready to handle requests
	// This could include checking database and cache connections
	
	response := models.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
		Services: map[string]string{
			"api": "ready",
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetUserURLs handles GET /api/v1/users/:user/urls
func (h *Handler) GetUserURLs(c *gin.Context) {
	user := c.Param("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user",
			Message: "User parameter is required",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get URLs for user (this would need to be added to the service interface)
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "User URL listing not yet implemented",
	})
	return
}

// CleanupExpired handles POST /api/v1/admin/cleanup (admin only)
func (h *Handler) CleanupExpired(c *gin.Context) {
	// TODO: Add admin authentication
	// For now, allow the operation (in production, verify admin privileges)

	if err := h.service.CleanupExpiredURLs(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "cleanup_failed",
			Message: "Failed to cleanup expired URLs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cleanup completed successfully",
		"timestamp": time.Now(),
	})
}
