# ProxyService

This project was created **for educational purposes only**, as a small playground to learn **Go** (HTTP routing, YAML configuration, and basic proxying with retries/timeouts).

## What it is
`ProxyService` is a lightweight API gateway:

- It exposes a proxy API under `/api/proxy/v1/...`
- It allows proxying **only** to services/paths explicitly configured in `configs/services.yaml`
- It currently proxies **HTTP GET** with timeout + retries; other HTTP methods are scaffolded and return `not_implemented_error`

For local testing, the repo includes `mock-api-server`.

It also includes an `admin-api-server` (scaffolded) and PostgreSQL migrations as groundwork for loading/managing services from a database.

## API
### Proxy server
Base: `http://localhost:8080`

- `GET /ping` - health check
- `GET /api/proxy/v1/:service/*path`
  - Forwards the request to `service.scheme://service.host/:path` (query string is preserved)
  - Response status code, headers (`Content-Type`) and body are passed through
  - Only allowed if both `:service` and the pair `(method, path)` exist in `configs/services.yaml`
- `POST|PUT|DELETE /api/proxy/v1/:service/*path`
  - Currently returns `not_implemented_error`

### Admin server (scaffold)
Base: `http://localhost:8082`

- `GET /ping` - health check
- `GET /api/admin/v1/service`
- `GET /api/admin/v1/service/:name`
- `POST /api/admin/v1/service/:name`
- `PUT /api/admin/v1/service/:service`
- `DELETE /api/admin/v1/service/:service`

All admin handlers currently return `not_implemented_error` and are intended for the future milestone: storing services/targets in PostgreSQL.

### Mock server
Base: `http://localhost:8081`

- `GET /ping` - health check
- `GET|POST|PUT|DELETE /mock`
  - Returns JSON with `{ method, query, body }`
  - Used as an upstream target for the proxy during development

## Configuration
### Main config (`configs/config.yaml`)

- `proxy_server.port` - proxy listen port
- `proxy_server.services_path` - path to `configs/services.yaml`
- `mock_server.port` - mock listen port
- `mock_server.response_status_code` - mock response status
- `admin_server.port` - admin listen port

All servers also support overriding the listen port from environment variable `PORT` (via `viper` binding).

### Services allowlist (`configs/services.yaml`)

This file defines a YAML array of services. Each service has:

- `name` - service key used as `:service` in the proxy URL
- `scheme` / `host` - where the proxy forwards requests
- `targets` - list of allowed routes:
  - `path` - upstream path (e.g. `/mock`)
  - `method` - HTTP method (e.g. `GET`)
- `timeout` - upstream timeout (seconds)
- `retry_count` / `retry_interval` - retry behavior for GET requests

The provided sample file includes:

- `mock` - upstream host `localhost:8081` (useful for local, non-Docker runs)
- `mock-docker` - upstream host `mock:8081` (useful inside Docker networks)

Example (trimmed):
```yaml
- name: mock
  host: localhost:8081
  scheme: http
  targets:
    - path: /mock
      method: GET
  timeout: 10.0
  retry_count: 3
  retry_interval: 0.1
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

## PostgreSQL (roadmap groundwork)
`test/docker-compose.yaml` starts PostgreSQL alongside the services.

Schema is defined in `internal/database/migrations` and currently creates:

- `services` table (service name, host, scheme, timeout, retry settings)
- `targets` table (allowed `(service_name, path, method)` routes)

The proxy/admin services still read allowlists from `configs/services.yaml` for now (the migration to load services from PG is still in the roadmap).

If you want to initialize the DB locally, the repo includes:
```bash
make migrate-local
```

Note: this requires the `migrate` CLI to be installed.

## How to run
### Local run (Go)

Terminal 1:
```bash
make start-mock
```

Terminal 2:
```bash
make start-proxy
```

Terminal 3 (optional, scaffold):
```bash
make start-admin
```

Test with curl:
```bash
curl "http://localhost:8080/api/proxy/v1/mock/mock?x=1"
```

Admin (currently scaffolded):
```bash
curl "http://localhost:8082/api/admin/v1/service"
```

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

There are tests for:

- YAML services loading (`configs/services.yaml` -> in-memory structures)
- Proxy routing allowlist logic
- GET proxy forwarding behavior (including context cancellation and error propagation)
