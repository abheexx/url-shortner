package obs

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Tracer provides tracing functionality
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new tracer instance
func NewTracer() *Tracer {
	// For now, use a no-op tracer
	// In production, you would configure a real tracer (Jaeger, Zipkin, etc.)
	tracer := noop.NewTracerProvider().Tracer("urlshortener")
	
	return &Tracer{
		tracer: tracer,
	}
}

// TracingMiddleware creates a Gin middleware for tracing
func TracingMiddleware(tracer *Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		
		// Create span for the request
		spanName := c.FullPath()
		if spanName == "" {
			spanName = c.Request.URL.Path
		}
		
		ctx, span := tracer.tracer.Start(ctx, spanName)
		defer span.End()
		
		// Set trace context in request
		c.Request = c.Request.WithContext(ctx)
		
		// Process request
		c.Next()
		
		// Add response status to span (simplified for now)
		_ = span
	}
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name)
}

// GetTracer returns the underlying tracer
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}
