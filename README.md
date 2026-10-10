# Distributed Task Queue

A Go-based asynchronous task processing service built with Gin, goroutines, and channels. It accepts tasks through a REST API, processes them using concurrent workers, and tracks their execution status.

## Features

* **REST API:** Submit tasks and retrieve their current status.
* **Concurrent workers:** Process tasks using goroutines and a channel-based queue.
* **Task lifecycle tracking:** Track tasks through `pending`, `processing`, `completed`, and `failed` states.
* **Failure handling:** Record processing errors when a task fails.
* **Manual retries:** Retry eligible failed tasks through an API endpoint.
* **Layered architecture:** Separate handlers, services, queue management, workers, and repository storage.

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
                Service
                 /   \
                v     v
          Repository  Queue
                ^       |
                |       v
                +---- Workers
```

The repository currently uses in-memory storage. Tasks and queue contents are not durable across server restarts.

## Tech Stack

* **Language:** Go
* **HTTP framework:** Gin
* **Concurrency:** Goroutines and channels
* **Storage:** In-memory repository
* **API format:** JSON over HTTP

## Getting Started

### Prerequisites

* Go installed
* Git installed

### Clone the repository

```bash
  git clone https://github.com/riteshdhurwey/distributed-task-queue.git
cd distributed-task-queue
```

### Install dependencies

```bash
go mod tidy
```

### Run the server

```bash
go run ./cmd/server
```

The server should start on the address configured in your application, for example `localhost:8080`.

## API Endpoints

### 1. Create a task

`POST /tasks`

Request:

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"name":"Send Email"}'
```

A successful request returns the created task and its initial or current status, depending on how quickly a worker begins processing it.

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
  "created_at": "2026-10-08T19:28:10+05:30",
  "updated_at": "2026-10-08T19:28:15+05:30"
}
```

The response above is illustrative; timestamps and task status depend on actual execution.

### 3. Retry a failed task

`POST /tasks/:id/retry`

```bash
curl -X POST http://localhost:8080/tasks/1/retry
```

Only tasks meeting the retry eligibility rules can be retried. The endpoint returns an appropriate response when the request is invalid or the task cannot be retried.

## Task Lifecycle

```text
pending → processing → completed
                    ↘ failed
```

A failed task can be manually retried, returning it to the pending state for another processing attempt.

## Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   ├── model/
│   ├── queue/
│   ├── repository/
│   ├── service/
│   └── worker/
├── config/
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

## Current Limitations

* Tasks and queued work are held in memory.
* Tasks may be lost if the process exits or crashes.
* Automatic retry scheduling and exponential backoff are not implemented.
* Production-grade shutdown, recovery, and observability are still areas for improvement.

## Roadmap

* [ ] PostgreSQL persistence
* [ ] Automatic retries with exponential backoff
* [ ] Graceful worker shutdown
* [ ] Unit and integration tests
* [ ] Structured logging and metrics
* [ ] Docker support
* [ ] Durable message broker integration

## License

Choose a license before publishing the project for reuse, such as the MIT License.
