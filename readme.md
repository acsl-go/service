# ACSL-GO Service Framework

See API reference [Here](./reference.md).

## Introduction
ACSL-GO Service Framework is a tiny framework for services. It focuses on the following aspects:
- Sub-service scheduling, in a common scenario, we may need to run an HTTP server and schedule some timer tasks. ACSL-GO Service Framework provides a simple way to do this.
- Graceful-shutdown HTTP server, a simple implementation of graceful-shutdown HTTP server as a sub-service (task).
- Timer task, a simple implementation of timer task as a sub-service (task).

## Usage
### Integration
```go
import (
    "github.com/acsl-go/service"
)
```

### Create an HTTP server
```go
var httpHandler http.Handler
httpServer := service.HttpServer("http-server", ":8080", httpHandler)
```
Or, if you work with [Gin](https://github.com/gin-gonic/gin):
```go
router := gin.New()
httpServer := service.HttpServer("http-server", ":8080", router)
```

### Create a timer task
```go
timerTask := service.Timer(5 * time.Second, func() {
    // do something
})
```

### Run the HTTP server and timer task
```go
service.Run(httpServer, timerTask)
```

### Sample code
```go
package main

import (
    "fmt"
    "time"
	"github.com/acsl-go/service"
	"github.com/gin-gonic/gin"
)

func apiInit(router *gin.Engine) {
    router.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
}

func ginTask(name, addr string, initializer func(*gin.Engine)) service.ServiceTask {
	router := gin.New()
	initializer(router)
	return service.HttpServer(name, addr, router)
}

func main() {
	service.Run(
		ginTask("API", ":8080", apiInit),
		service.Timer(5*time.Second, func() {
            fmt.Println("timer task")
        }),
	)
}
```