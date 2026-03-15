package service

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/acsl-go/logger"
)

func HttpServer(name, addr string, router http.Handler) ServiceTask {
	return func(ctx context.Context) {

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

		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error:  %+v\n", err)
		}

		logger.Info("HTTP server %s on %s stopped gracefully\n", name, addr)
	}
}

func HttpsServer(name, addr, certFile, keyFile string, router http.Handler) ServiceTask {
	return func(ctx context.Context) {

		server := &http.Server{
			Addr:    addr,
			Handler: router,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}

		go func() {
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatal("%s", err)
			}
		}()

		logger.Info("HTTPS server %s started on %s\n", name, addr)

		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error:  %+v\n", err)
		}

		logger.Info("HTTPS server %s on %s stopped gracefully\n", name, addr)
	}
}
