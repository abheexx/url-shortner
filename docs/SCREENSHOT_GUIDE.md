# Screenshot Guide for URL Shortener README

This guide helps you add real screenshots to make the README more engaging and easier to understand.

## Required Screenshots

### 1. Architecture Diagram
**File**: `docs/images/architecture-diagram.png`
**Size**: 800x600px recommended
**Content**: System architecture showing all components and their relationships

**How to create**:
- Use tools like [draw.io](https://draw.io), [Lucidchart](https://lucidchart.com), or [Miro](https://miro.com)
- Export as PNG with transparent background
- Include: Client, Load Balancer, API Gateway, Go Service, Redis, PostgreSQL, Monitoring

### 2. API Flow Diagram
**File**: `docs/images/api-flow.png`
**Size**: 800x600px recommended
**Content**: Request flow from client to response

**How to create**:
- Show the complete request lifecycle
- Include: Request → Validation → Processing → Cache → Database → Response
- Use arrows and flow indicators

### 3. Performance Metrics Dashboard
**File**: `docs/images/performance-metrics.png`
**Size**: 1000x600px recommended
**Content**: Real-time performance metrics

**How to capture**:
1. Start your service: `make docker-run`
2. Open Prometheus metrics: http://localhost:8080/metrics
3. Take screenshot of key metrics
4. Or use Grafana if you have it configured

### 4. Load Test Results
**File**: `docs/images/load-test-results.png`
**Size**: 1000x700px recommended
**Content**: k6 load test execution and results

**How to capture**:
1. Run load test: `make load-test`
2. Take screenshot during execution
3. Capture final results summary
4. Show performance thresholds being met

## Optional Screenshots

### 5. Service Health Check
**File**: `docs/images/health-check.png`
**Size**: 600x400px
**Content**: Health check endpoint response

**How to capture**:
```bash
curl http://localhost:8080/api/v1/healthz
# Take screenshot of terminal or browser response
```

### 6. API Response Examples
**File**: `docs/images/api-examples.png`
**Size**: 800x500px
**Content**: Sample API requests and responses

**How to capture**:
1. Use tools like Postman or Insomnia
2. Show successful URL creation
3. Show redirect response
4. Show error handling

### 7. Docker Containers Running
**File**: `docs/images/docker-containers.png`
**Size**: 800x400px
**Content**: Docker containers status

**How to capture**:
```bash
docker ps
# Take screenshot of running containers
```

## Screenshot Best Practices

### Image Quality
- **Resolution**: Minimum 72 DPI, 144 DPI for high-quality displays
- **Format**: PNG for diagrams, JPG for photos, SVG for scalable graphics
- **Size**: Keep under 500KB for better loading

### Content Guidelines
- **Clear and readable**: Text should be legible
- **Focused**: Show only relevant information
- **Consistent**: Use similar styling across all images
- **Professional**: Clean, uncluttered appearance

### Technical Requirements
- **File naming**: Use lowercase with hyphens (e.g., `architecture-diagram.png`)
- **Directory**: Save all images in `docs/images/`
- **Git**: Add images to version control
- **README**: Reference images with relative paths

## Tools for Creating Screenshots

### Diagram Creation
- [draw.io](https://draw.io) - Free, web-based diagram tool
- [Lucidchart](https://lucidchart.com) - Professional diagramming
- [Miro](https://miro.com) - Collaborative whiteboarding
- [Figma](https://figma.com) - Design and prototyping

### Screenshot Capture
- **macOS**: Cmd+Shift+4 (area selection)
- **Windows**: Snipping Tool or Win+Shift+S
- **Linux**: GNOME Screenshot or Flameshot
- **Browser**: Browser dev tools for web screenshots

### Image Editing
- [GIMP](https://gimp.org) - Free Photoshop alternative
- [Canva](https://canva.com) - Online design tool
- [Pixlr](https://pixlr.com) - Web-based image editor

## Example Screenshot Workflow

1. **Plan your screenshots** - Decide what to show
2. **Set up your environment** - Get service running
3. **Take screenshots** - Capture at appropriate moments
4. **Edit and optimize** - Crop, resize, enhance
5. **Save with proper names** - Use consistent naming
6. **Add to README** - Update image references
7. **Test locally** - Verify images display correctly
8. **Commit and push** - Add to version control

## Quick Start Commands

```bash
# Start service for screenshots
make docker-run

# Check service health
curl http://localhost:8080/api/v1/healthz

# View metrics
curl http://localhost:8080/metrics

# Run load test
make load-test

# Check Docker status
docker ps
docker logs urlshortener-api
```

## Need Help?

If you need assistance creating specific screenshots:
1. Check the existing text-based diagrams for reference
2. Use the architecture documentation in `ADR-001-architecture.md`
3. Run the service and explore the endpoints
4. Take screenshots of what you see

Remember: Good screenshots make documentation much more engaging and easier to understand!
