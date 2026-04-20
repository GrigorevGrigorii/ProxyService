# Proxy Service

**This project was made for educational purposes.**  
It demonstrates a production-grade microservices architecture in Go, including a configurable reverse proxy, background caching, RBAC authorization, database migrations, Redis queuing, and AWS deployment (AWS deployment is not working now because I've reached my free tier quota).

---

## 📋 Overview

A lightweight, configurable **reverse proxy service** with the following capabilities:

- **Admin API** – CRUD management of allowed upstream services and their proxy targets.
- **Proxy API** – Forwards allowed HTTP requests (GET fully supported, others stubbed) to upstream services.
- **Intelligent caching** – Automatic background caching of GET responses using Redis + periodic tasks (Asynq).
- **Authentication & Authorization** – AWS Cognito (via ALB) + Casbin RBAC.
- **Retry + timeout** logic per service.
- **Fully observable** – Structured logging, request IDs, Swagger docs.

Designed to run on **AWS ECS Fargate** with **Postgres** and **Redis Sentinel**, but easy to run locally via Docker Compose.
A Kubernetes-based cloud deployment alternative is also available in `deployments/k8s/` (not fully validated in cloud yet because free tier quota is over).

## ✨ Features

- **Dynamic configuration** of services/targets via Admin API
- **Per-target caching** with configurable TTL (only for GET + specific query)
- **Background worker** that pre-fills cache using Asynq periodic tasks + leader election
- **Optimistic locking** on service updates (version field)
- **Query parameter normalization** for cache keys
- **Full Swagger/OpenAPI** documentation
- **Health-check ready** (`/ping`)
- **Production-ready** observability (Zerolog + request tracing)
- **Local testing** with mock upstream + full Docker Compose stack
- **CI/CD** ready (GitHub Actions for build/test + AWS deployment)

## 🛠️ Tech Stack

| Layer              | Technology                              |
|--------------------|-----------------------------------------|
| Language           | Go 1.25                                 |
| Web Framework      | Gin                                     |
| ORM                | GORM v2                                 |
| Database           | PostgreSQL + custom types/enums         |
| Cache & Queue      | Redis + Asynq (periodic tasks)          |
| Auth               | AWS Cognito + Casbin RBAC               |
| API Docs           | Swaggo (Swagger 2.0)                    |
| Logging            | Zerolog                                 |
| Config             | Viper                                   |
| Migrations         | golang-migrate                          |
| Deployment         | AWS ECS Fargate + Lambda (migrations)   |
| CI/CD              | GitHub Actions                          |
| Local Dev          | Docker Compose                          |

## 🚀 Quick Start (Local)

### 1. Clone & prepare

```bash
git clone <your-repo>
cd proxy-service
```

### 2. Start all services
```bash
make start-containers-with-build          # or make start-containers
```
This spins up:
- Postgres
- Redis (master + 2 replicas + 3 sentinels)
- Proxy API (:8080)
- Admin API (:8082)
- Mock upstream (:8081)
- Background worker & scheduler

### 3. Run migrations
Install `migrate` tool using [instruction](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md#installation)
```bash
make migrate-local
```

### 4. Fill the database with test data
```bash
make init-pg-data
```

## 📡 API Documentation

### Admin API (port 8082)
- **Base**: `/api/admin`
- **Swagger**: http://localhost:8082/api/admin/swagger/index.html

Endpoints:
- `GET /v1/service` – List all services
- `GET /v1/service/{name}` – Get service + targets
- `POST /v1/service` – Create service
- `PUT /v1/service/{name}` – Update service (with version)
- `DELETE /v1/service/{name}` – Delete service

### Proxy API (port 8080)
- **Base**: `/api/proxy`
- **Swagger**: http://localhost:8080/api/proxy/swagger/index.html

Example usage:
```http
GET /api/proxy/v1/my-service/users?active=true
```

## 🧪 Testing

```bash
make test
```

Includes:
- Unit tests for handlers, cache task, repository, validators, etc.
- SQL mocking with go-sqlmock

## 📦 Deployment

The project includes ready-to-use AWS ECS task definitions and GitHub Actions workflows:

- `.github/workflows/deploy_to_aws.yml` – Builds & deploys all microservices to ECS Fargate
- `.github/workflows/migrate_db.yml` – Builds Lambda and runs DB migrations on every migration change

See `deployments/aws-ecs-task-definitions/` and `build/package/Dockerfile`.

As an alternative to ECS, there is also a Kubernetes deployment setup in `deployments/k8s/` with `cloud` and `local` overlays.
Cloud correctness of the Kubernetes path has not been fully verified yet because the AWS free tier quota is over.

## 📁 Project Structure (key folders)

```
proxy-service/
├── cmd/                    # Entry points (proxy, admin, background, mock)
├── internal/
│   ├── auth/               # Auth checkers for AuthMiddleware (contains only AWS Cognito as example)
│   ├── database/           # GORM models, repository, migrations
│   ├── handlers/           # Gin handlers + validation
│   ├── cache/              # Redis cache layer
│   ├── background/         # Asynq tasks & leader election
│   ├── httpclient/         # Retry-enabled HTTP client
│   ├── middlewares/        # Auth, logging, request ID
│   ├── models/             # DTOs & converters
│   └── config/             # Viper config loading
├── configs/                # YAML configs + Casbin policy
├── api/*/docs/             # Generated Swagger
├── deployments/            # AWS ECS definitions + Kubernetes manifests
├── test/                   # docker-compose.yaml
├── Makefile
└── go.mod
```

## 🔧 Configuration

All services read from `configs/*.yaml` (overridable via environment variables).  
Example environment variables are already set in Docker Compose, ECS task definitions, and Kubernetes overlays.

## Contributing

This is an **educational project**. Feel free to fork, experiment, and improve it!
