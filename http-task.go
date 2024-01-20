package service

import (
	"context"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/acsl-go/logger"
)

func HttpServer(name, addr string, router http.Handler) ServiceTask {
	return func(wg *sync.WaitGroup, quit_signal chan os.Signal) {
		defer wg.Done()

		server := &http.Server{
			Addr:    addr,
			Handler: router,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("%s", err)
			}
		}()

		logger.Info("HTTP server %s started on %s\n", name, addr)

		<-quit_signal
		quit_signal <- syscall.SIGTERM

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error:  %+v\n", err)
		}

		logger.Info("HTTP server %s on %s stopped gracefully\n", name, addr)
	}
}
