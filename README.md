# ProxyService

This project was created **for educational purposes only**, as a small playground to learn **Go** (HTTP routing, PostgreSQL integration, and basic proxying with retries/timeouts).

## What it is
`ProxyService` is a lightweight API gateway:

- It exposes a proxy API under `/api/proxy/v1/...`
- It allows proxying **only** to services/paths stored in PostgreSQL
- It currently proxies **HTTP GET** with timeout + retries; other HTTP methods are scaffolded and return `not_implemented_error`

For local testing, the repo includes `mock-api-server`.

It also includes an `admin-api-server` for managing allowed services/targets in PostgreSQL.

## API
### Proxy server
Base: `http://localhost:8080`

- `GET /ping` - health check
- `GET /api/proxy/v1/:service/*path`
  - Forwards the request to `service.scheme://service.host/:path` (query string is preserved)
  - Response status code, headers (`Content-Type`) and body are passed through
  - Only allowed if both `:service` and the pair `(method, path)` exist in PostgreSQL (`services` + `targets`)
- `POST|PUT|DELETE /api/proxy/v1/:service/*path`
  - Currently returns `not_implemented_error`

### Admin server
Base: `http://localhost:8082`

- `GET /ping` - health check
- `GET /api/admin/v1/service`
- `GET /api/admin/v1/service/:name`
- `POST /api/admin/v1/service`
- `PUT /api/admin/v1/service/:name`
- `DELETE /api/admin/v1/service/:name`

Admin handlers are connected to PostgreSQL and support CRUD for services/targets.

### Mock server
Base: `http://localhost:8081`

- `GET /ping` - health check
- `GET|POST|PUT|DELETE /mock`
  - Returns JSON with `{ method, query, body }`
  - Used as an upstream target for the proxy during development

## Configuration
### Main config (`configs/config.yaml`)

- `proxy_server.port` - proxy listen port
- `proxy_server.pg.*` - PostgreSQL connection settings for proxy
- `mock_server.port` - mock listen port
- `mock_server.response_status_code` - mock response status
- `admin_server.port` - admin listen port
- `admin_server.pg.*` - PostgreSQL connection settings for admin

Servers support overriding listen port from environment variable `PORT` (via `viper` binding).

PostgreSQL fields in config:

- `user`, `password`, `host`, `port`, `database`, `sslmode`

Example:
```yaml
proxy_server:
  port: 8080
  pg:
    user: proxy_service_user
    password: proxy_service_password
    host: postgres
    port: 5432
    database: proxy_service_db
    sslmode: disable
```

## Retries / Timeout behavior (GET)
The proxy uses an internal HTTP client for GET requests:

- Applies `timeout` via `http.Client{Timeout: ...}`
- Retries when:
  - request returns an error, or
  - upstream responds with `502`, `503`, or `504`
- Retries up to `retry_count` times with `retry_interval` between attempts

## Logging
The project uses `zerolog` + Gin middlewares:

- `RequestIDMiddleware` generates/propagates `X-Request-ID`
- `ZerologMiddleware` logs method, path, request-id, response status, latency, and client IP

In `debug` mode, logs are also printed to stderr via a console writer.

## PostgreSQL
`test/docker-compose.yaml` starts PostgreSQL alongside the services.

Schema is defined in `internal/database/migrations`

The proxy/admin services use PostgreSQL as the source of truth for service allowlists.

If you want to initialize the DB locally, the repo includes:
```bash
make migrate-local
```

Note: this requires the `golang-migrate` CLI to be installed.

## Admin API payloads
### Create service
`POST /api/admin/v1/service`

```json
{
  "name": "mock",
  "scheme": "http",
  "host": "mock:8081",
  "timeout": 10.0,
  "retry_count": 3,
  "retry_interval": 0.1,
  "targets": [
    { "path": "/mock", "method": "GET" }
  ]
}
```

### Update service
`PUT /api/admin/v1/service/:name`

Request body has the same shape as create, plus `version` for optimistic locking.

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

## Testing
Run unit tests:
```bash
make test
```
