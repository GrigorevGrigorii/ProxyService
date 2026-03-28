# ProxyService

This project was created **for educational purposes only**, as a playground to learn **Go** (HTTP routing, PostgreSQL integration, Redis caching, background task scheduling, and basic proxying with retries/timeouts).

## What it is

`ProxyService` is a lightweight API gateway that:

- Exposes a proxy API under `/api/proxy/v1/...`
- Allows proxying **only** to services/paths stored in PostgreSQL
- Caches successful `GET` responses in Redis to reduce upstream load
- Uses a background tasks to maintain cache freshness
- Currently proxies **HTTP GET** with timeout, retries, and caching; other HTTP methods are scaffolded and return `not_implemented_error`

For local testing, the repo includes `mock-api-server`.
It also includes an `admin-api-server` for managing allowed services/targets in PostgreSQL.

## Architecture Overview

- **Proxy Server** – handles client requests, caches responses, and forwards to upstream services.
- **Admin Server** – CRUD for services and targets.
- **Mock Server** – simple upstream for testing.
- **PostgreSQL** – stores service definitions, targets, and configuration.
- **Redis** – caches successful GET responses and stores background task data.
- **Background Tasks** – periodically refreshes cache entries (e.g., for frequently accessed endpoints).

## API

### Proxy server

Base: `http://localhost:8080`

- `GET /ping` – health check
- `GET /api/proxy/v1/:service/*path`
  - Tries to get response from Redis cache
  - Forwards the request to `service.scheme://service.host/:path` (query string preserved)
  - Response status, headers, and body are passed through
  - Only allowed if both `:service` and the `(method, path, query)` exist in PostgreSQL
- `POST|PUT|DELETE /api/proxy/v1/:service/*path` – currently returns `not_implemented_error`

### Admin server

Base: `http://localhost:8082`

- `GET /ping` – health check
- Service CRUD:
  - `GET /api/admin/v1/service`
  - `GET /api/admin/v1/service/:name`
  - `POST /api/admin/v1/service`
  - `PUT /api/admin/v1/service/:name`
  - `DELETE /api/admin/v1/service/:name`

### Mock server

Base: `http://localhost:8081`

- `GET /ping` – health check
- `GET|POST|PUT|DELETE /mock` – returns JSON with `{ method, query, body }`; used as upstream for testing

## Configuration

Configuration is defined in `configs/`. The files include settings for the proxy, admin, mock servers, PostgreSQL connections, Redis, and background tasks.

## Caching and Background Tasks

### How caching works

- On a successful `GET` request, the proxy stores the response in Redis.
- Subsequent identical requests are served from Redis.

### Background tasks

The scheduler periodically refreshes cache entries for services that have targets with `cache_interval is not null`. It:
- Queries PostgreSQL for services with caching enabled
- Fetches fresh responses from upstream and stores them in Redis
- This ensures cache hits even when the upstream is temporarily slow or down.

## Retries / Timeout behavior (GET)

The proxy uses an internal HTTP client for GET requests:

- Applies `timeout` via `http.Client{Timeout: ...}`
- Retries when:
  - request returns an error, or
  - upstream responds with `502`, `503`, or `504`
- Retries up to `retry_count` times with `retry_interval` between attempts

## Database and Migrations

PostgreSQL schema is defined in `internal/database/migrations`. To initialize the database locally:

```bash
make migrate-local
```

## How to run
### Local run (Docker Compose)
```bash
make start-containers-with-build
```

This starts:

- `proxy` on `127.0.0.1:8080`
- `mock` on `127.0.0.1:8081`
- `admin` on `127.0.0.1:8082`
- `postgres` on `5432`
- `redis` on `6379`

## Testing
Run unit tests:
```bash
make test
```
