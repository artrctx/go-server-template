package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/artrctx/shuffle-core/internal/database/repository/deck"
	"github.com/artrctx/shuffle-core/internal/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Service struct {
	pool *pgxpool.Pool
	Deck *deck.Queries
}

var dbInst *Service

func New(connStr string) (*Service, error) {
	pool, err := pgxpool.New(context.Background(), connStr)

	if err != nil {
		return nil, err
	}

	//? Commenting out to be lazy
	// if err := dbConn.Ping(); err != nil {
	// 	return nil, err
	// }

	dbInst = &Service{
		pool: pool,
		Deck: deck.New(pool),
	}

	return dbInst, nil
}

func Get() (*Service, error) {
	if dbInst != nil {
		return dbInst, nil
	}
	return New(env.DatabaseUrl)
}

func (s *Service) Conn() *pgxpool.Pool {
	return s.pool
}

func (s *Service) DeckTx(ctx context.Context) (pgx.Tx, *deck.Queries, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	return tx, deck.New(tx), nil
}

func (s *Service) Health(ctx context.Context) map[string]string {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.pool.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "The database is healthy."

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.pool.Stat()
	stats["total_connections"] = strconv.FormatInt(int64(dbStats.TotalConns()), 10)
	stats["open_connections"] = strconv.FormatInt(int64(dbStats.TotalConns()), 10)
	stats["in_use"] = strconv.Itoa(int(dbStats.AcquiredConns()))
	stats["idle"] = strconv.Itoa(int(dbStats.IdleConns()))
	stats["wait_duration"] = dbStats.AcquireDuration().String()

	if dbStats.TotalConns() > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.AcquireCount() > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleDestroyCount() > int64(dbStats.TotalConns())/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeDestroyCount() > int64(dbStats.TotalConns())/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}
	return stats
}

func (s *Service) Close() {
	s.pool.Close()
}
