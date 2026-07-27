package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/artrctx/shuffle-core/internal/database"
	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/artrctx/shuffle-core/internal/storage"
)

type Server struct {
	port    int
	db      *database.Service
	storage *storage.Service
}

func NewServer() (*http.Server, error) {
	db, err := database.Get()
	if err != nil {
		return nil, err
	}
	storage, err := storage.Get(context.Background())
	if err != nil {
		return nil, err
	}
	srv := Server{env.Port, db, storage}
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.Register(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}, nil
}
