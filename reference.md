# API Reference of ACSL-GO Service Framework

## Run
Run tasks.
```go
func Run(tasks ...Task)
```

### Parameters
- `tasks ...Task`: Tasks to run.

### Remarks
This function runs all the tasks in different goroutines and waits for the termination of all the tasks.

This function will observe the `SIGTERM` signal and pass it to all tasks via the `quit_signal` parameter of Task.

This function will block until all tasks are terminated.


## Task

A task is a function that runs in a goroutine.

Applications should implement their task functions to run their business logic.

```go
type ServiceTask func(*sync.WaitGroup, chan os.Signal)
```

### Parameters
- `wg`: A wait group to wait for the termination of the task. A best practice is to call `defer wg.Done()` at the beginning of the task function.
- `quit_signal`: A channel to receive the `SIGTERM` signal.

### Remarks
A task function represents a task. It should be passed to the `Run` function to run.

When the function returns, the task is considered terminated.

If the task is a long-running task, it should observe the `quit_signal` parameter and terminate gracefully when receiving the `SIGTERM` signal.

## HttpServer

Create an HTTP server task.

```go
func HttpServer(name, addr string, handler http.Handler) ServiceTask
```

### Parameters
- `name`: The name of the task, shown in the log output.
- `addr`: The address to listen on.
- `handler`: The HTTP handler.

### Return
A task function that represents an HTTP server.

## Timer

Create a timer task.

```go
func Timer(interval time.Duration, task func()) ServiceTask
```

### Parameters
- `interval`: The interval of the timer.
- `task`: The task to run.

### Return
A task function that represents a timer task.

### Remarks
The timer task will run the task function periodically.