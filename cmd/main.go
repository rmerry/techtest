package main

import (
	"asostechtest/internal/api"
	"asostechtest/internal/datastore"
	"asostechtest/internal/sessionstore"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	port                    = 8081
	gracefulShutdownTimeout = 20 * time.Second
	maxSessionAge           = time.Minute * 10
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	logger = logger.With("component", "main")

	db := datastore.NewInMemory(maxSessionAge)
	sessionStore := sessionstore.New(db, maxSessionAge)
	handlers := api.NewHTTPHandlers(sessionStore)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handlers.Router,
	}
	go func() {
		logger.Info("starting server", "port", port)
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			logger.Info("server stopped")
			return
		}

		// Log error cases and cancel context to prevent application running
		// without a server.
		if err == syscall.EADDRINUSE {
			logger.Error("starting server", "err", err)
		} else {
			logger.Error("shutting down server", "err", err)
		}
		cancel()
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")
	logger.Info("stopping server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	srv.Shutdown(shutdownCtx)
	defer cancel()

	// Called after server gracefully shutdown to allow for inflight requests
	// to be successfully handled.
	db.Close()
}
