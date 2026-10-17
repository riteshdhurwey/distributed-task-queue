# Worker Concurrency Test Documentation

## 1. Overview

The `TestWorkersProcessTasksOnce` test verifies that multiple concurrent workers can process queued tasks and that each task is processed exactly once during normal execution.

The test uses:
- **Go goroutines** to run multiple workers concurrently.
- **PostgreSQL** to persist task details and processing status.
- **Redis ZSET** to queue and order task IDs.
- **Mutex (`sync.Mutex`)** to protect a shared task execution counter.
- **WaitGroup (`sync.WaitGroup`)** to wait for worker goroutines to stop.
- **Race detector** to identify unsynchronized memory access.

**Important:** This test validates normal concurrent execution. It does not prove exactly-once processing under worker crashes or external side effects.

## 2. Test Location

File:

`internal/worker/worker_test.go`

Test function:

`TestWorkersProcessTasksOnce`

## 3. How the Test Works

### Step 1: Track task executions

The test creates a map called `counts` that stores how many times each task is processed.

A custom processor, `testProcess`, increments the count for the corresponding task ID. A mutex protects the map because multiple workers may access it simultaneously.

The custom processor returns immediately instead of sleeping, making the test faster and easier to repeat.

### Step 2: Connect to PostgreSQL and Redis

The test connects to the local PostgreSQL database and Redis server using the project's existing configuration.

It initializes:
- A PostgreSQL repository for task persistence.
- A Redis repository for queue operations.

Both connections are closed when the test finishes.

**Prerequisites:** PostgreSQL and Redis must be running, and the project's environment configuration must be available.

### Step 3: Create and enqueue 100 tasks

The test creates 100 tasks with names such as `test-task-0` and `test-task-1`.

Each task initially has the status `pending`.

For every task:
1. Save the task in PostgreSQL.
2. Add its task ID to the Redis ZSET.
3. Record its ID in `taskIDs` for later verification and cleanup.

### Step 4: Start five workers

The test creates five worker instances and runs each worker's `Start(ctx)` method in a separate goroutine.

All five workers consume tasks from the same Redis queue and use the same PostgreSQL repository.

This simulates concurrent task processing.

### Step 5: Wait for completion

The test checks PostgreSQL repeatedly to determine whether all 100 tasks have reached the `completed` status.

It checks every 100 milliseconds and allows up to 30 seconds for completion.

If all tasks complete before the deadline, the test proceeds. Otherwise, it reports a timeout failure.

### Step 6: Stop the workers

The test cancels the shared context. Workers should stop when they observe cancellation.

The `WaitGroup` waits for all worker goroutines to exit before the test continues.

### Step 7: Verify execution counts

The test checks the execution count for every task.

Each task must have a count of exactly `1`. A count of `0` indicates that the processor was never called; a count greater than `1` indicates duplicate processor calls.

Any mismatch causes the test to fail.

### Step 8: Clean up test data

The deferred cleanup removes the test task IDs from the Redis ZSET and deletes their corresponding PostgreSQL records.

Cleanup should remove only records created by this test, not unrelated application data.

## 4. Commands to Run the Test

Run only this test:

```bash
go test -v ./internal/worker -run TestWorkersProcessTasksOnce
```

Run all tests in the worker package:

```bash
go test -v ./internal/worker
```

Run the test with Go's race detector:

```bash
go test -race -v ./internal/worker -run TestWorkersProcessTasksOnce
```

Run the test 10 times to check for intermittent concurrency problems:

```bash
go test -count=10 ./internal/worker -run TestWorkersProcessTasksOnce
```

Run all project tests:

```bash
go test ./...
```

Run all project tests with the race detector:

```bash
go test -race ./...
```

## 5. What Each Command Checks

| Command | Purpose |
|---|---|
| `-v` | Displays detailed test output. |
| `-run TestWorkersProcessTasksOnce` | Runs only the named test. |
| `-race` | Detects data races during execution. |
| `-count=10` | Executes the test 10 times. |
| `./internal/worker` | Tests the worker package. |
| `./...` | Tests all packages in the module. |

## 6. Expected Results

A successful test should produce output similar to:

```text
=== RUN   TestWorkersProcessTasksOnce
--- PASS: TestWorkersProcessTasksOnce
PASS
ok      taskqueue/internal/worker
```

The exact duration may differ between runs.

With the race detector enabled, the test should pass without any reported data races.

## 7. What This Test Proves

When it passes, the test provides evidence that:

- Five workers can run concurrently.
- The submitted tasks can be processed to completion.
- The custom processor was called once per task during this test run.
- The execution counter is protected against data races.
- Worker goroutines can be stopped through context cancellation.

## 8. What This Test Does Not Prove

The test does not establish that:

- Tasks can always recover after a worker crashes.
- A task cannot be lost between Redis dequeue and PostgreSQL claiming.
- External side effects, such as sending an email, happen exactly once.
- The system guarantees exactly-once delivery.
- The system remains correct under every possible timing or failure scenario.

The race detector checks memory synchronization; it does not detect every logical concurrency bug.

## 9. Future Testing Improvements

Potential next tests include:

1. **Worker crash recovery:** Verify that a task left in `processing` can be recovered after its lease expires.
2. **Redis dequeue failure:** Verify that queue errors do not silently lose tasks.
3. **Duplicate claiming:** Verify that two workers cannot claim the same task concurrently.
4. **Failure and retry:** Verify that failed tasks follow the configured retry policy.
5. **Test isolation:** Use a dedicated test database and isolated Redis queue so repeated test runs cannot interfere with application data.

## 10. Summary

This test is an integration test for normal concurrent task processing. It exercises the worker, PostgreSQL repository, and Redis queue together, rather than testing only an individual function.

Passing it is a useful baseline before adding crash-recovery and retry tests.

