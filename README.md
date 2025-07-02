<p align="center">
  <img src="./assets/arachne-logo-transparent.png" alt="Arachne Logo" width="400">
</p>

<h1 align="center">Arachne</h1>

<p align="center">
  A scalable, production-ready Go web scraping service with an asynchronous job API, persistent job state, and containerization.
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/kareemsasa3/arachne"><img src="https://goreportcard.com/badge/github.com/kareemsasa3/arachne" /></a>
  <a href="https://github.com/kareemsasa3/arachne/blob/main/LICENSE"><img src="https://img.shields.io/github/license/kareemsasa3/arachne?style=flat-square&color=blue" /></a>
  <a href="https://hub.docker.com/r/kareemsasa3/arachne"><img src="https://img.shields.io/docker/pulls/kareemsasa3/arachne.svg" /></a>
  <a href="https://github.com/kareemsasa3/arachne/releases/latest"><img src="https://img.shields.io/github/v/release/kareemsasa3/arachne" /></a>
</p>

## ✨ Features

- **🚀 Asynchronous scraping API**: Submit jobs and poll for results
- **💾 Persistent job state**: Uses Redis for resilience and multi-instance support
- **📈 Scalable architecture**: Ready for distributed job processing
- **🛡️ Circuit breaker pattern**: Protects against cascading failures
- **🔄 Retry logic**: Handles transient failures with exponential backoff
- **📊 Comprehensive metrics**: Real-time performance monitoring
- **🧪 Comprehensive API tests**: Uses `httptest` for handler coverage
- **🐳 Containerized**: Docker and Docker Compose for easy deployment
- **🏥 Health checks**: System monitoring and observability

## 🚀 Quick Start

### Prerequisites
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

### Run with Docker Compose

```bash
# Clone the repository
git clone <your-repo-url>
cd go-practice

# Start the service
docker-compose up --build
```

**Services Available:**
- **API Server**: http://localhost:8080
- **Redis**: localhost:6379
- **Redis Commander UI**: http://localhost:8081

- API: http://localhost:8080
- Redis: localhost:6379
- Redis Commander UI: http://localhost:8081

### 📡 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/scrape` | Submit a new scraping job |
| `GET` | `/scrape/status?id=<job_id>` | Get job status and results |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | System metrics (if enabled) |

### 📝 Example: Submit a Scrape Job

```bash
curl -X POST http://localhost:8080/scrape \
  -H 'Content-Type: application/json' \
  -d '{
    "urls": ["https://golang.org", "https://httpbin.org/get"]
  }'
```

### 📊 Example: Check Job Status

```bash
curl "http://localhost:8080/scrape/status?id=<job_id>"
```

### 🌐 Example: Scrape a Single Site

```bash
curl -X POST http://localhost:8080/scrape \
  -H 'Content-Type: application/json' \
  -d '{
    "site_url": "https://jsonplaceholder.typicode.com/posts"
  }'
```

## 🏗️ Architecture Overview

- **🌐 API Layer**: Handles HTTP requests, creates jobs, persists them in Redis
- **💾 Persistent Storage**: Redis stores all job state, making the system resilient to restarts and ready for distributed scaling
- **🛡️ Circuit Breaker**: Protects against cascading failures with automatic recovery
- **🔄 Retry Logic**: Handles transient failures with exponential backoff
- **📊 Metrics & Monitoring**: Real-time performance tracking and health checks
- **🐳 Containerization**: Dockerfile and docker-compose.yml for reproducible, portable deployment
- **🔧 Worker Model**: Ready for multiple scraper instances as workers, all pulling jobs from Redis

## ⚙️ Configuration

Environment variables (see `.env.example` and `docker-compose.yml`):

| Variable | Description | Default |
|----------|-------------|---------|
| `SCRAPER_API_PORT` | API server port | `8080` |
| `SCRAPER_REDIS_ADDR` | Redis host | `redis:6379` |
| `SCRAPER_REDIS_PASSWORD` | Redis password | (empty) |
| `SCRAPER_REDIS_DB` | Redis DB index | `0` |
| `SCRAPER_MAX_CONCURRENT` | Max concurrent requests | `5` |
| `SCRAPER_REQUEST_TIMEOUT` | Request timeout | `10s` |
| `SCRAPER_CIRCUIT_BREAKER_THRESHOLD` | Circuit breaker threshold | `3` |
| `SCRAPER_ENABLE_METRICS` | Enable metrics collection | `true` |

See `config.go` for all available options.

## 🧪 Testing

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -v -cover ./...
```

## 🚀 Extending for Distributed Job Queue

- **Message Queue**: Swap the in-process goroutine for a message queue (RabbitMQ, Kafka, or Redis Streams/Lists)
- **Worker Scaling**: Run multiple scraper workers, each pulling jobs from the queue and updating job state in Redis
- **Load Balancing**: Distribute scraping load across multiple instances
- **High Availability**: Enable true horizontal scaling and fault tolerance

## 📈 Performance Features

- **Circuit Breaker**: Automatic failure detection and recovery
- **Rate Limiting**: Domain-specific and global rate limiting
- **Retry Logic**: Exponential backoff for transient failures
- **Metrics Collection**: Real-time performance monitoring
- **Health Checks**: System health monitoring

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details. 