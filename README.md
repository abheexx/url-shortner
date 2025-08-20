# URL Shortener Service

A production-grade, distributed URL shortener service built with Go, designed to handle 50K+ requests per second with p95 latency under 100ms.

## Features

- **High Performance**: Optimized for read-heavy workloads with Redis caching
- **Scalable**: Horizontal scaling with stateless API design
- **Production Ready**: Comprehensive observability, rate limiting, and security
- **Developer Friendly**: Docker Compose setup, comprehensive testing, and CI/CD ready

## Architecture

- **Language**: Go 1.22+
- **Web Framework**: Gin with graceful shutdown and middleware
- **Storage**: PostgreSQL (primary), Redis (cache), Kafka (event log - optional)
- **ID Generation**: ULID with base62 encoding
- **Observability**: OpenTelemetry, Prometheus metrics, structured logging
- **Security**: Rate limiting, input validation, CORS, security headers

## Performance Targets

- **Throughput**: 50K+ requests/second
- **Latency**: p95 < 100ms (cache hit), p95 < 200ms (cache miss)
- **Cache Hit Ratio**: ≥ 90% on hot data
- **Availability**: 99.9% uptime

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### 1. Clone and Setup

```bash
git clone <repository-url>
cd urlshortener
make deps
```

### 2. Run with Docker Compose

```bash
make docker-run
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- API service on port 8080
- pgAdmin on port 5050 (admin/admin)

### 3. Run Database Migrations

```bash
make migrate-up
```

### 4. Test the Service

```bash
# Health check
curl http://localhost:8080/api/v1/healthz

# Create a short URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com"}'

# Redirect to long URL
curl -L http://localhost:8080/YOUR_CODE_HERE
```

## Development

### Local Development

```bash
# Install dependencies
make deps

# Run locally (requires PostgreSQL and Redis)
make dev

# Run tests
make test

# Run linter
make lint

# Run benchmarks
make bench
```

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make test          # Run tests with coverage
make dev           # Run locally
make docker-run    # Start with Docker Compose
make docker-stop   # Stop Docker Compose services
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make load-test     # Run k6 load tests
```

## API Documentation

### Endpoints

#### Create Short URL
```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://www.example.com",
  "custom_alias": "my-link",  // optional
  "expire_at": "2024-12-31T23:59:59Z",  // optional
  "created_by": "user123"     // optional
}
```

#### Redirect to Long URL
```http
GET /:code
# Returns 301 redirect to long URL
```

#### Get URL Metadata
```http
GET /api/v1/urls/:code
# Returns URL metadata and click statistics
```

#### Delete URL
```http
DELETE /api/v1/urls/:code
# Soft deletes the URL
```

#### Health Checks
```http
GET /api/v1/healthz  # Health check
GET /api/v1/readyz   # Readiness check
GET /metrics          # Prometheus metrics
```

### Response Examples

#### Create URL Response
```json
{
  "code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "long_url": "https://www.example.com",
  "expire_at": null,
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### URL Metadata Response
```json
{
  "code": "abc123",
  "long_url": "https://www.example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "expire_at": null,
  "total_clicks": 42,
  "last_access_at": "2024-01-01T12:00:00Z",
  "is_deleted": false
}
```

## Configuration

Configuration is handled via environment variables with sensible defaults:

```bash
# Server
URLSHORTENER_SERVER_PORT=8080
URLSHORTENER_SERVER_READ_TIMEOUT=30s
URLSHORTENER_SERVER_WRITE_TIMEOUT=30s

# Database
URLSHORTENER_DATABASE_HOST=localhost
URLSHORTENER_DATABASE_PORT=5432
URLSHORTENER_DATABASE_USER=urlshortener
URLSHORTENER_DATABASE_PASSWORD=password
URLSHORTENER_DATABASE_DBNAME=urlshortener

# Redis
URLSHORTENER_REDIS_HOST=localhost
URLSHORTENER_REDIS_PORT=6379
URLSHORTENER_REDIS_PASSWORD=
URLSHORTENER_REDIS_TTL=24h

# Rate Limiting
URLSHORTENER_RATE_LIMIT_GLOBAL_RPS=100
URLSHORTENER_RATE_LIMIT_PER_IP_RPS=10

# Logging
URLSHORTENER_LOGGING_LEVEL=info
URLSHORTENER_LOGGING_FORMAT=json
```

## Testing

### Unit Tests
```bash
make test
```

### Load Testing
```bash
# Install k6
# macOS: brew install k6
# Linux: https://k6.io/docs/getting-started/installation/

make load-test
```

The load test simulates:
- 95% GET requests (redirects)
- 5% POST requests (create URLs)
- Ramp up to 500 concurrent users
- Performance thresholds: p95 < 100ms

### Test Coverage
Tests generate HTML coverage reports in `coverage.html`.

## Monitoring & Observability

### Metrics
- **Prometheus endpoint**: `/metrics`
- **Key metrics**: Request duration, cache hit ratio, database latency, rate limit counts

### Logging
- **Structured JSON logging** with Zap
- **Request correlation** with request IDs
- **Configurable levels** (debug, info, warn, error)

### Tracing
- **OpenTelemetry integration** for distributed tracing
- **Request flow tracking** across services

## Deployment

### Docker
```bash
# Build image
make docker-build

# Run container
docker run -p 8080:8080 urlshortener:latest
```

### Kubernetes
Kubernetes manifests and Helm charts are available in `deploy/k8s/` and `deploy/helm/`.

### Production Considerations
- Use managed PostgreSQL and Redis services
- Configure proper resource limits and requests
- Set up horizontal pod autoscaling
- Use ingress controllers for load balancing
- Configure proper monitoring and alerting

## Security Features

- **Rate limiting** (global + per-IP)
- **Input validation** and sanitization
- **CORS configuration**
- **Security headers** (XSS protection, content type options)
- **URL allowlist/blocklist** support
- **Graceful degradation** on cache failures

## Performance Tuning

### Redis Optimization
```bash
# Redis configuration for high throughput
maxmemory 1gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### Go Runtime Tuning
```bash
export GOMAXPROCS=4  # Adjust based on CPU cores
export GOGC=100       # Garbage collection tuning
```

### Database Optimization
- Connection pooling configured
- Proper indexing on frequently queried columns
- Regular VACUUM and ANALYZE

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Ensure all tests pass
6. Submit a pull request

### Development Guidelines
- Follow Go best practices and idioms
- Maintain test coverage above 80%
- Use conventional commit messages
- Update documentation for API changes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: Create GitHub issues for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check the [docs/](docs/) directory for detailed guides

## Roadmap

- [ ] Kafka integration for event streaming
- [ ] Advanced analytics dashboard
- [ ] Multi-tenant support
- [ ] API versioning
- [ ] GraphQL endpoint
- [ ] Mobile SDKs
- [ ] Advanced caching strategies
- [ ] Geographic distribution

---

**Built with Go and modern DevOps practices**
