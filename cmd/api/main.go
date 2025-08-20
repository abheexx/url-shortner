package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/urlshortener/internal/cache"
	"github.com/urlshortener/internal/config"
	httphandler "github.com/urlshortener/internal/http"
	"github.com/urlshortener/internal/obs"
	"github.com/urlshortener/internal/rate"
	"github.com/urlshortener/internal/repo"
	"github.com/urlshortener/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := obs.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting URL Shortener service")

	// Initialize database
	db, err := repo.NewPostgresRepo(cfg.GetDSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(
		cfg.GetRedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.TTL,
		cfg.Redis.NegativeTTL,
	)
	defer redisCache.Close()

	// Initialize rate limiter
	rateLimiter := rate.NewLimiter(rate.Config{
		GlobalRPS:  cfg.RateLimit.GlobalRPS,
		PerIPRPS:   cfg.RateLimit.PerIPRPS,
		BurstSize:  cfg.RateLimit.BurstSize,
		WindowSize: cfg.RateLimit.WindowSize,
	})
	defer rateLimiter.Close()

	// Initialize service
	serviceConfig := service.Config{
		BaseURL:      fmt.Sprintf("http://localhost:%s", cfg.Server.Port),
		CodeLength:   8,
		MaxURLLength: 2048,
		AllowedHosts: cfg.Security.AllowedHosts,
		BlockedHosts: cfg.Security.BlockedDomains,
	}

	shortenerService := service.NewShortenerService(db, redisCache, serviceConfig)

	// Initialize HTTP handler
	handler := httphandler.NewHandler(shortenerService, serviceConfig.BaseURL)

	// Initialize observability
	metrics := obs.NewMetrics()
	tracer := obs.NewTracer()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(
		obs.LoggingMiddleware(logger),
		obs.RecoveryMiddleware(logger),
		obs.CORSMiddleware(cfg.Security.AllowedOrigins),
		rate.RateLimitMiddleware(rateLimiter),
		obs.MetricsMiddleware(metrics),
		obs.TracingMiddleware(tracer),
	)

	// Health check endpoints
	router.GET("/api/v1/healthz", handler.HealthCheck)
	router.GET("/api/v1/readyz", handler.ReadinessCheck)

	// Metrics endpoint
	router.GET("/metrics", obs.MetricsHandler(metrics))

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/shorten", handler.CreateShortURL)
		api.GET("/urls/:code", handler.GetURLMetadata)
		api.DELETE("/urls/:code", handler.DeleteURL)
		api.GET("/users/:user/urls", handler.GetUserURLs)
	}

	// Admin routes
	admin := router.Group("/api/v1/admin")
	{
		admin.POST("/cleanup", handler.CleanupExpired)
	}

	// Redirect route (must be last to avoid conflicts)
	router.GET("/:code", handler.RedirectToLongURL)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Start background workers
	go startBackgroundWorkers(context.Background(), shortenerService, logger)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

// startBackgroundWorkers starts background tasks
func startBackgroundWorkers(ctx context.Context, service *service.ShortenerService, logger *obs.Logger) {
	// Cleanup expired URLs every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := service.CleanupExpiredURLs(ctx); err != nil {
				logger.Error("Failed to cleanup expired URLs", "error", err)
			} else {
				logger.Info("Cleanup of expired URLs completed")
			}
		}
	}
}
