# opgl-gateway-service

Pure API Gateway service that routes and orchestrates communication between OPGL microservices.

## Service Overview

- **Port**: 8080 (default)
- **Purpose**: Routes requests and orchestrates between data, cortex, and auth services
- **Framework**: Go with gorilla/mux router
- **Database**: None (lightweight gateway!)

## Architecture

```
Client → Gateway → Data Service (summoner/matches)
                → Cortex Service (analysis)
                → Auth Service (token validation, rate limiting)
```

The gateway is a pure proxy/router with no database. Authentication and rate limiting are delegated to opgl-auth-service.

## Project Structure

```
opgl-gateway-service/
├── main.go                      # Application entry point
├── internal/
│   ├── api/
│   │   ├── router.go            # Route definitions
│   │   ├── handlers.go          # HTTP request handlers
│   │   └── handlers_test.go     # Handler unit tests
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware for preflight requests
│   │   ├── logging.go           # Request/response logging middleware
│   │   ├── auth.go              # Auth middleware (calls auth service)
│   │   └── ratelimit.go         # Rate limit middleware (calls auth service)
│   ├── errors/
│   │   └── errors.go            # Error types and responses
│   ├── models/
│   │   └── models.go            # Shared data models
│   ├── proxy/
│   │   ├── interface.go         # ServiceProxyInterface for dependency injection
│   │   └── proxy.go             # Service proxy implementation
│   └── validation/
│       └── validation.go        # Request validation
├── Makefile                     # Build, test, and run commands
├── Dockerfile                   # Docker containerization
└── .env.example                 # Environment variable template
```

## Endpoints

All endpoints use **POST** method (per project guidelines):

| Endpoint | Description | Rate Limited |
|----------|-------------|--------------|
| `POST /health` | Health check | No |
| `POST /api/v1/summoner` | Proxy to opgl-data-service | Yes |
| `POST /api/v1/matches` | Proxy to opgl-data-service | Yes |
| `POST /api/v1/analyze` | Orchestrated analysis (data + cortex) | Yes |

Rate limiting requires `X-API-Key` header.

## Request Body Format

All endpoints use Riot ID format:

```json
{
  "region": "na",
  "gameName": "Newyenn",
  "tagLine": "GGEZ"
}
```

For matches endpoint, optional `count` parameter (defaults to 20):

```json
{
  "region": "na",
  "gameName": "Newyenn",
  "tagLine": "GGEZ",
  "count": 10
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `OPGL_DATA_URL` | http://localhost:8081 | opgl-data-service URL |
| `OPGL_CORTEX_URL` | http://localhost:8082 | opgl-cortex-engine-service URL |
| `OPGL_AUTH_URL` | http://localhost:8083 | opgl-auth-service URL |

## Development Commands

```bash
# Run locally
make run

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Build binary
make build

# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Lint code (requires golangci-lint)
make lint
```

## Key Implementation Details

### Handler Pattern
- Handlers receive requests, validate input, call proxy methods, and return JSON responses
- All handlers validate required fields: region, gameName, tagLine
- Error responses use structured JSON with error codes

### Service Proxy Pattern
- `ServiceProxy` handles all HTTP communication with downstream services
- Uses POST requests with JSON bodies for all service calls
- `GetMatchesByPUUID` method exists for internal optimization (avoids redundant lookups)

### Middleware Stack
1. **CORS Middleware** - Handles preflight OPTIONS requests
2. **Logging Middleware** - Logs incoming requests and response status codes
3. **Rate Limit Middleware** - Calls auth service to check API key rate limits

### Rate Limiting
- Gateway calls `POST /api/v1/ratelimit/check` on auth service
- Requires `X-API-Key` header on rate-limited endpoints
- Returns rate limit headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### Analysis Flow (POST /api/v1/analyze)
1. Check rate limit via auth service
2. Fetch summoner data from opgl-data-service using Riot ID
3. Fetch match history from opgl-data-service using PUUID (efficiency optimization)
4. Send summoner + matches to opgl-cortex-engine-service for analysis
5. Return analysis result to client

## Testing

Tests use interfaces for dependency injection:
- `ServiceProxyInterface` allows mocking proxy calls in handler tests
- Run `make test` for unit tests with race detection

## Dependencies

- `github.com/gorilla/mux` - HTTP router
- `github.com/rs/zerolog` - Structured logging
- `github.com/google/uuid` - UUID parsing (for auth context)
