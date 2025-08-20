# ADR-001: URL Shortener Service Architecture

## Status

Accepted

## Context

We need to build a production-grade URL shortener service that can handle 50K+ requests per second with p95 latency under 100ms. The service must be horizontally scalable, highly available, and maintainable.

## Key Requirements

- **Performance**: 50K+ RPS, p95 < 100ms
- **Scalability**: Horizontal scaling capability
- **Availability**: 99.9% uptime
- **Read-heavy workload**: 95% GET requests, 5% POST requests
- **Production ready**: Observability, security, monitoring

## Architecture Decisions

### 1. Language Choice: Go

**Decision**: Use Go 1.22+ as the primary language.

**Rationale**:
- **Performance**: Excellent performance characteristics for HTTP services
- **Concurrency**: Built-in goroutines and channels for efficient concurrency
- **Memory efficiency**: Low memory footprint and predictable GC behavior
- **Production maturity**: Strong ecosystem for production services
- **Team expertise**: Good familiarity with Go in the team

**Alternatives considered**:
- **Node.js**: Good performance but higher memory usage
- **Rust**: Excellent performance but longer development time
- **Java**: Good performance but higher resource usage

### 2. Web Framework: Gin

**Decision**: Use Gin as the HTTP framework.

**Rationale**:
- **Performance**: One of the fastest Go web frameworks
- **Middleware support**: Rich ecosystem of middleware
- **Validation**: Built-in request validation
- **Maturity**: Production-tested and widely adopted
- **Documentation**: Excellent documentation and examples

**Alternatives considered**:
- **Fiber**: Good performance but newer ecosystem
- **Echo**: Good performance but less middleware
- **Standard library**: Maximum performance but more boilerplate

### 3. Storage Strategy: PostgreSQL + Redis

**Decision**: Use PostgreSQL as primary storage with Redis as a read-through cache.

**Rationale**:
- **PostgreSQL**: ACID compliance, reliability, rich querying
- **Redis**: Sub-millisecond read performance, persistence options
- **Cache strategy**: Read-through + write-through for consistency
- **Fallback**: Graceful degradation when Redis is unavailable

**Cache Strategy Details**:
- **TTL**: 24 hours for positive cache, 5 minutes for negative cache
- **Invalidation**: On delete, update, and expiration
- **Fallback**: Direct database queries when cache fails

**Alternatives considered**:
- **MongoDB**: Good performance but eventual consistency
- **Cassandra**: Excellent scalability but operational complexity
- **Pure Redis**: Good performance but limited durability

### 4. ID Generation: ULID + Base62

**Decision**: Use ULID with base62 encoding for short URLs.

**Rationale**:
- **ULID**: Time-ordered, unique, URL-safe
- **Base62**: Shorter than base64, URL-friendly
- **Collision resistance**: Extremely low collision probability
- **Sortability**: Time-based ordering for analytics
- **Custom aliases**: Support for user-defined short codes

**Alternatives considered**:
- **Snowflake**: Good performance but requires coordination
- **UUID**: Good uniqueness but longer strings
- **Sequential IDs**: Simple but predictable and security concerns

### 5. Rate Limiting: Token Bucket

**Decision**: Implement token bucket rate limiting with global and per-IP limits.

**Rationale**:
- **Token bucket**: Smooths traffic spikes, allows bursts
- **Global limits**: Protects overall system capacity
- **Per-IP limits**: Prevents abuse from individual sources
- **Configurable**: Adjustable limits for different environments
- **Performance**: Efficient implementation with minimal overhead

**Configuration**:
- Global: 100 RPS with 20 burst
- Per-IP: 10 RPS with 5 burst
- Window: 1 second sliding window

**Alternatives considered**:
- **Leaky bucket**: Simpler but less flexible
- **Fixed window**: Simpler but allows traffic spikes
- **Sliding window**: More accurate but higher complexity

### 6. Observability: OpenTelemetry + Prometheus + Zap

**Decision**: Use OpenTelemetry for tracing, Prometheus for metrics, and Zap for structured logging.

**Rationale**:
- **OpenTelemetry**: Vendor-neutral, future-proof tracing
- **Prometheus**: Industry standard for metrics, excellent querying
- **Zap**: High-performance structured logging
- **Integration**: All components work well together
- **Production ready**: Battle-tested in production environments

**Metrics collected**:
- HTTP request duration, count, and status
- Cache hit/miss ratios
- Database operation latency
- Rate limiting statistics

**Alternatives considered**:
- **Jaeger**: Good tracing but vendor lock-in
- **StatsD**: Good metrics but less powerful than Prometheus
- **Logrus**: Good logging but slower than Zap

### 7. Security: Defense in Depth

**Decision**: Implement multiple layers of security controls.

**Rationale**:
- **Input validation**: Prevents injection attacks
- **Rate limiting**: Prevents abuse and DoS
- **CORS**: Controls cross-origin access
- **Security headers**: Protects against common web vulnerabilities
- **URL filtering**: Blocks malicious domains

**Security measures**:
- Input length limits (max 2048 characters)
- Protocol allowlist (HTTP/HTTPS only)
- Domain blocklist support
- Security headers (XSS protection, content type options)
- CORS configuration

### 8. Deployment: Docker + Kubernetes

**Decision**: Use Docker for containerization and Kubernetes for orchestration.

**Rationale**:
- **Docker**: Industry standard, excellent tooling
- **Kubernetes**: Production-grade orchestration, auto-scaling
- **Portability**: Consistent deployment across environments
- **Ecosystem**: Rich ecosystem of tools and services
- **Scalability**: Built-in horizontal scaling and load balancing

**Production considerations**:
- Resource limits and requests
- Horizontal pod autoscaling
- Pod disruption budgets
- Liveness and readiness probes
- Rolling update strategy

## Architecture Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Load Balancer │    │   Load Balancer │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
┌─────────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
│   API Pod 1     │    │   API Pod 2     │    │   API Pod N     │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Gin HTTP  │ │    │ │   Gin HTTP  │ │    │ │   Gin HTTP  │ │
│ │   Server    │ │    │ │   Server    │ │    │ │   Server    │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │  Business   │ │    │ │  Business   │ │    │ │  Business   │ │
│ │   Logic     │ │    │ │   Logic     │ │    │ │   Logic     │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Cache     │ │    │ │   Cache     │ │    │ │   Cache     │ │
│ │  Interface  │ │    │ │  Interface  │ │    │ │  Interface  │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │        Redis Cluster      │
                    │     (Read/Write Cache)    │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      PostgreSQL          │
                    │    (Primary Storage)     │
                    └─────────────────────────┘
```

## Performance Characteristics

### Latency Breakdown (Target: p95 < 100ms)

**Cache Hit Path**:
- Redis GET: ~1ms
- Response serialization: ~5ms
- Network overhead: ~10ms
- **Total: ~16ms**

**Cache Miss Path**:
- Redis GET (miss): ~1ms
- PostgreSQL query: ~10-50ms
- Cache SET: ~2ms
- Response serialization: ~5ms
- Network overhead: ~10ms
- **Total: ~28-68ms**

### Throughput Considerations

**Bottlenecks**:
- **Network I/O**: Limited by network capacity
- **Redis**: Can handle 100K+ ops/sec per instance
- **PostgreSQL**: Connection pool limits, query optimization
- **Go runtime**: GOMAXPROCS, garbage collection tuning

**Scaling strategies**:
- Horizontal scaling of API pods
- Redis cluster for higher cache capacity
- PostgreSQL read replicas for analytics
- Connection pooling optimization

## Trade-offs and Risks

### Trade-offs Made

1. **Complexity vs Performance**: Added Redis complexity for performance gains
2. **Consistency vs Latency**: Eventual consistency in cache for lower latency
3. **Flexibility vs Security**: Strict input validation limits some use cases
4. **Development Speed vs Production Readiness**: More upfront work for production deployment

### Risks and Mitigations

1. **Cache Failure**
   - Risk: Performance degradation
   - Mitigation: Graceful fallback to database, circuit breakers

2. **Database Bottleneck**
   - Risk: Single point of failure
   - Mitigation: Connection pooling, read replicas, query optimization

3. **Memory Leaks**
   - Risk: Service instability
   - Mitigation: Memory profiling, monitoring, proper cleanup

4. **Rate Limiting Bypass**
   - Risk: Service abuse
   - Mitigation: Multiple rate limiting layers, IP validation

## Future Considerations

### Short Term (3-6 months)
- Add comprehensive test coverage
- Implement Kafka for event streaming
- Add advanced monitoring dashboards
- Performance optimization and tuning

### Medium Term (6-12 months)
- Multi-region deployment
- Advanced caching strategies
- GraphQL API endpoint
- Mobile SDKs

### Long Term (12+ months)
- Machine learning for URL analytics
- Advanced security features
- Multi-tenant architecture
- Geographic distribution

## Conclusion

The chosen architecture provides a solid foundation for a high-performance, scalable URL shortener service. The combination of Go, Gin, PostgreSQL, and Redis offers excellent performance characteristics while maintaining operational simplicity. The observability stack ensures we can monitor and debug issues in production, while the security measures protect against common threats.

The architecture is designed to evolve over time, with clear upgrade paths for additional features and scaling requirements.
