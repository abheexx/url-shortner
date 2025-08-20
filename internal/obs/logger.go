package obs

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps the Zap logger
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger creates a new logger instance
func NewLogger(level, format string) (*Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var config zap.Config
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger.Sugar()}, nil
}

// LoggingMiddleware creates a Gin middleware for request logging
func LoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Log after request is processed
		logger.Infow("HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// RecoveryMiddleware creates a Gin middleware for panic recovery
func RecoveryMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorw("Panic recovered",
					"panic", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"client_ip", c.ClientIP(),
				)

				c.JSON(500, gin.H{
					"error":   "internal_error",
					"message": "Internal server error",
				})
			}
		}()
		
		c.Next()
	}
}
