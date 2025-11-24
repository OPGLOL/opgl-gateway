# OPGL Gateway

API Gateway that orchestrates requests across OPGL microservices.

## Features

- Routes requests to appropriate microservices
- Orchestrates complex operations (analyze player)
- Single entry point for frontend

## Architecture

```
Client → Gateway (8080) → opgl-data (8081)
                       → opgl-cortex-engine (8082)
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Gateway health check |
| `/api/v1/summoner/{region}/{summonerName}` | GET | Get summoner (→ opgl-data) |
| `/api/v1/matches/{region}/{puuid}` | GET | Get matches (→ opgl-data) |
| `/api/v1/analyze/{region}/{summonerName}` | GET | Full analysis (orchestrates both services) |

## How it Works

**Analyze Endpoint Flow**:
1. Gateway receives analyze request
2. Calls opgl-data to get summoner info
3. Calls opgl-data to get match history
4. Sends data to opgl-cortex-engine for analysis
5. Returns complete analysis to client

## Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Set service URLs if different from defaults
   ```

3. **Run the service**:
   ```bash
   go run main.go
   ```

Service runs on port **8080** by default.

## Environment Variables

- `PORT` - Gateway port (default: 8080)
- `OPGL_DATA_URL` - Data service URL (default: http://localhost:8081)
- `OPGL_CORTEX_URL` - Cortex engine URL (default: http://localhost:8082)

## Running All Services

See docker-compose.yml in the root project directory to run all services together.

## Testing

Use Bruno collection at `bruno-collections/opgl/` to test all endpoints through the gateway.
