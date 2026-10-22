package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Register env
	"github.com/artrctx/templ/cmd/api/env"
	"github.com/artrctx/templ/internal/auth"
	"github.com/artrctx/templ/internal/lib/logger"
	"github.com/artrctx/templ/internal/server"
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
	logProv, err := logger.Initialize(context.Background(), logger.Config{Env: env.AppEnv, Name: env.AppName, PosthogToken: env.PostHogToken})
	if err != nil {
		log.Fatalln(err)
	}
	if err := auth.Initialize(env.AuthURL); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("Failed to initialize authentication conn within alotted time")
		}
		log.Fatal(err)
	}
	defer func() {
		if err := auth.Close(); err != nil {
			log.Printf("auth provider shutdown failed with error: %v", err)
		}
		if err := logProv.Shutdown(context.Background()); err != nil {
			log.Printf("logger provider shutdown failed with error: %v", err)
		}
	}()

	srv, err := server.NewServer(context.Background(), server.Config{
		DatabaseURL:    env.DatabaseURL,
		Port:           env.Port,
		Env:            env.AppEnv,
		R2AccountID:    env.R2AccountID,
		R2AccessKey:    env.R2AccessKey,
		R2AccessSecret: env.R2AccessSecretKey,
		R2Bucket:       env.R2BucketName,
	})
	if err != nil {
		slog.Error("Failed Initiazing Server", "error", err)
		os.Exit(1)
	}

	// app close channel
	done := make(chan bool, 1)

	// Handle graceful shutdown in a seperate goroutine
	go gracefulShutdown(srv, done)

	slog.Info(fmt.Sprintf("templ server starting at %s", srv.Addr))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	slog.Info("Gracefully shutdown server.")
}
