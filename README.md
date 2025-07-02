# Go Scraper Service

A scalable, production-ready Go web scraping service with asynchronous job API, persistent job state (Redis), containerization, and comprehensive testing.

## Features

- **Asynchronous scraping API**: Submit jobs and poll for results.
- **Persistent job state**: Uses Redis for resilience and multi-instance support.
- **Scalable architecture**: Ready for distributed job processing.
- **Comprehensive API tests**: Uses `httptest` for handler coverage.
- **Containerized**: Docker and Docker Compose for easy deployment.

## Quick Start

### Prerequisites
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

### Run with Docker Compose

```
docker-compose up --build
```

- API: http://localhost:8080
- Redis: localhost:6379
- Redis Commander UI: http://localhost:8081

### API Endpoints

- `POST /scrape` — Submit a new scraping job
- `GET /job/status?id=...` — Get job status/results
- `GET /health` — Health check
- `GET /metrics` — Metrics (if enabled)

### Example: Submit a Scrape Job

```
curl -X POST http://localhost:8080/scrape -d '{"urls": ["https://example.com"]}' -H 'Content-Type: application/json'
```

### Example: Check Job Status

```
curl http://localhost:8080/job/status?id=<job_id>
```

## Architecture Overview

- **API Layer**: Handles HTTP requests, creates jobs, persists them in Redis.
- **Persistent Storage**: Redis stores all job state, making the system resilient to restarts and ready for distributed scaling.
- **Worker Model**: (Future) You can run multiple scraper instances as workers, all pulling jobs from Redis.
- **Containerization**: Dockerfile and docker-compose.yml for reproducible, portable deployment.

## Configuration

Environment variables (see `docker-compose.yml`):
- `SCRAPER_REDIS_ADDR` — Redis host (e.g., `redis:6379`)
- `SCRAPER_REDIS_PASSWORD` — Redis password (optional)
- `SCRAPER_REDIS_DB` — Redis DB index (default: 0)
- Other config: see `config.go` for all options.

## Testing

Run all tests:

```
go test ./...
```

## Extending for Distributed Job Queue

- Swap the in-process goroutine for a message queue (RabbitMQ, Kafka, or Redis Streams/Lists).
- Run multiple scraper workers, each pulling jobs from the queue and updating job state in Redis.
- This enables true horizontal scaling and high availability.

## License

MIT 