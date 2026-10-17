# Distributed Task Queue

A Go-based asynchronous task processing service built with Gin, PostgreSQL, and Redis. It accepts tasks through a REST API, queues task IDs in Redis, processes tasks concurrently using workers, and persists task execution status in PostgreSQL.

## Features

- **REST API:** Submit tasks and retrieve their status.
- **Concurrent workers:** Process tasks using goroutines.
- **Redis queue:** Use a Redis Sorted Set (ZSET) to order queued task IDs by score.
- **PostgreSQL persistence:** Store task details, status, attempts, and timestamps.
- **Task lifecycle tracking:** Track tasks through `pending`, `processing`, `completed`, and `failed` states.
- **Failure handling:** Record task processing errors.
- **Startup recovery:** Reload pending tasks from PostgreSQL and re-enqueue them after a server restart.
- **Layered architecture:** Separate handlers, services, repositories, queue operations, and workers.

## Architecture

```text
                   Client
                     |
                     v
                 Gin REST API
                     |
                     v
                   Handler
                     |
                     v
                 Task Service
                   /      \
                  v        v
          PostgreSQL      Redis
          Repository      ZSET
              ^              |
              |              v
              +----------- Worker
```

PostgreSQL stores the task records, while Redis maintains the queue of task IDs. Workers retrieve task details from PostgreSQL and update their execution status after processing.

At startup, the application loads eligible pending tasks from PostgreSQL and re-enqueues them in Redis.

## Tech Stack

- **Language:** Go
- **HTTP framework:** Gin
- **Database:** PostgreSQL
- **Queue:** Redis Sorted Set (ZSET)
- **Concurrency:** Goroutines
- **API format:** JSON over HTTP

## Getting Started

### Prerequisites

- Go installed
- PostgreSQL installed and running
- Redis installed and running
- Git installed

### Clone the repository

```bash
git clone https://github.com/riteshdhurwey/distributed-task-queue.git
cd distributed-task-queue
```

### Configure environment variables

Create a `.env` file in the project root with the database, Redis, and server settings expected by your application.

Do not commit credentials or secrets.

### Set up the database

Create your PostgreSQL database and apply the task table schema used by the application.

Ensure PostgreSQL and Redis are running before starting the server.

### Install dependencies

```bash
go mod tidy
```

### Run the server

```bash
go run ./cmd/server
```

The server listens on the address configured in your application, for example `localhost:8080`.

## API Endpoints

### 1. Create a task

`POST /tasks`

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"name":"Send Email"}'
```

The service persists the task in PostgreSQL and enqueues its ID in Redis.

### 2. Get task status

`GET /tasks/:id`

```bash
curl http://localhost:8080/tasks/1
```

Example response:

```json
{
  "id": 1,
  "name": "Send Email",
  "status": "completed",
  "attempts": 1,
  "created_at": "2026-10-14T22:42:55+05:30",
  "updated_at": "2026-10-15T00:57:51+05:30"
}
```

The response is illustrative; actual timestamps and status depend on execution.

### 3. Retry a failed task

`POST /tasks/:id/retry`

```bash
curl -X POST http://localhost:8080/tasks/1/retry
```

This endpoint is available if manual retry functionality is implemented and registered in the application.

## Task Lifecycle

```text
pending → processing → completed
                 |
                 v
               failed
                 |
                 v
          manual retry
                 |
                 v
              pending
```

Tasks are persisted in PostgreSQL, and pending task IDs are queued in Redis.

## Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── config/
├── docs/
├── internal/
│   ├── database/
│   ├── handler/
│   ├── model/
│   ├── queue/
│   ├── repository/
│   ├── service/
│   └── worker/
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

## Current Limitations

- Reliable recovery of tasks claimed by workers that crash still needs to be addressed.
- A failure between database insertion and Redis enqueueing can leave a task unqueued.
- Automatic retries with exponential backoff are not yet implemented.
- Graceful shutdown, structured logging, metrics, and production-grade reliability need further work.

## Roadmap

- [x] PostgreSQL task persistence
- [x] Redis ZSET queue integration
- [x] Startup recovery of pending tasks
- [x] Reliable task claiming and crash recovery
- [ ] Automatic retries with exponential backoff
- [ ] Graceful worker shutdown
- [ ] Unit and integration tests
- [ ] Structured logging and metrics
- [ ] Docker support
- [ ] Transactional outbox for reliable database-to-queue delivery


