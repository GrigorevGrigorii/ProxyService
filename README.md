# ProxyService

This project was created **for educational purposes only**, as a small playground to learn **Go** (HTTP routing, YAML configuration, and basic proxying with retries/timeouts).

## What it is
`ProxyService` is a lightweight API proxy/gateway:

- It exposes a single proxy API under `/api/v1/...`
- It allows proxying **only** to services/paths explicitly configured in `configs/services.yaml`
- For now it implements proxying for **HTTP GET** (with timeout + retries); other HTTP methods are scaffolded and return `not_implemented_error`

For local testing the repo also includes a tiny `mock-api-server`.

## API
### Proxy server
Base: `http://localhost:8080`

- `GET /ping` - health check
- `GET /api/v1/:service/*path`
  - Forwards the request to `service.scheme://service.host/:path` (query string is preserved)
  - Response status code, headers (`Content-Type`) and body are passed through
  - Only allowed if both `:service` and the pair `(method, path)` exist in `configs/services.yaml`
- `POST|PUT|DELETE /api/v1/:service/*path`
  - Currently returns `not_implemented_error`

### Mock server
Base: `http://localhost:8081`

- `GET /ping` - health check
- `GET|POST|PUT|DELETE /mock`
  - Returns JSON with `{ method, query, body }`
  - Used as an upstream target for the proxy during development

## Configuration
### Main config (`configs/config.yaml`)
Contains:

- `proxy_server.port` - proxy listen port
- `proxy_server.services_path` - path to `configs/services.yaml`
- `mock_server.port` - mock listen port
- `mock_server.response_status_code` - mock response status

Both servers also support overriding the listen port from environment variable `PORT` (via `viper` binding).

### Services allowlist (`configs/services.yaml`)
This file defines a YAML array of services. Each service has:

- `name` - service key used as `:service` in the proxy URL
- `scheme` / `host` - where the proxy forwards requests
- `targets` - list of allowed routes:
  - `path` - upstream path (e.g. `/mock`)
  - `method` - HTTP method (e.g. `GET`)
- `timeout` - upstream timeout (seconds)
- `retry_count` / `retry_interval` - retry behavior for GET requests

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

Test with curl:
```bash
curl "http://localhost:8080/api/v1/mock/mock?x=1"
```

### Local run (Docker Compose)
```bash
make start-containers-with-build
```

Proxy is available on `127.0.0.1:8080`, mock on `127.0.0.1:8081`.

## Retries / Timeout behavior (GET)
The proxy uses an internal HTTP client for GET requests:

- Applies `timeout` via `http.Client{Timeout: ...}`
- Retries when:
  - request returns an error, or
  - upstream responds with `502`, `503`, or `504`
- Retries up to `retry_count` times with `retry_interval` between attempts

## Testing
Run unit tests:
```bash
make test
```

There are tests for:
- YAML services loading (`configs/services.yaml` -> in-memory structures)
- Proxy routing allowlist logic
- GET proxy forwarding behavior (including context cancellation and error propagation)
