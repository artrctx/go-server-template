package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Register env
	"github.com/artrctx/shuffle-core/internal/auth"
	_ "github.com/artrctx/shuffle-core/internal/env"
	"github.com/artrctx/shuffle-core/internal/lib/logger"
	"github.com/artrctx/shuffle-core/internal/server"
)

func gracefulShutdown(srv *http.Server, done chan bool) {
	// listen to interrupt signal from os
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal
	<-ctx.Done()

	slog.InfoContext(ctx, "shutting down. press Ctrl+C again to force quit")
	stop() // Allow Ctrl+C to force shutdown

	// 5 second max request handling time
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "Server forced to shutdown with error", slog.Any("error", err))
	}

	slog.Info("Server exiting")

	done <- true
}

func main() {
	logProv, err := logger.Initialize(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := auth.Close(); err != nil {
			log.Printf("auth provider shutdown failed with error: %v", err)
		}
		if err := logProv.Shutdown(context.Background()); err != nil {
			log.Printf("logger provider shutdown failed with error: %v", err)
		}
	}()

	srv, err := server.NewServer()
	if err != nil {
		slog.Error("Failed Initiazing Server", "error", err)
		os.Exit(1)
	}

	// app close channel
	done := make(chan bool, 1)

	// Handle graceful shutdown in a seperate goroutine
	go gracefulShutdown(srv, done)

	slog.Info(fmt.Sprintf("shuffle server starting at %s", srv.Addr))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	slog.Info("Gracefully shutdown server.")
}
