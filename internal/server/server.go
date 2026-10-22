package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/artrctx/templ/internal/database"
	"github.com/artrctx/templ/internal/storage"
)

type Server struct {
	port    int
	env     string
	bucket  string
	db      *database.Service
	storage *storage.Service
}

type Config struct {
	DatabaseURL    string
	Port           int
	Env            string
	R2AccountID    string
	R2AccessKey    string
	R2AccessSecret string
	R2Bucket       string
}

func NewServer(ctx context.Context, cfg Config) (*http.Server, error) {
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	storage, err := storage.New(ctx, cfg.R2AccountID, cfg.R2AccessKey, cfg.R2AccessSecret)
	if err != nil {
		return nil, err
	}
	srv := Server{port: cfg.Port, env: cfg.Env, bucket: cfg.R2Bucket, db: db, storage: storage}
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      srv.Register(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}, nil
}
