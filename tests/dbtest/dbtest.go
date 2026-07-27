package dbtest

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartPostgresContainer(connStr *string) (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	dbName, username, password := "testdb", "testusr", "password"

	ctx := context.Background()
	dbContainer, err := postgres.Run(
		ctx,
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		// mute test container log
		testcontainers.WithLogger(log.New(io.Discard, "", 0)),
		testcontainers.WithAdditionalWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)

	if err != nil {
		return nil, err
	}

	dbHost, err := dbContainer.Host(ctx)
	if err != nil {
		return dbContainer.Terminate, err
	}
	host := dbHost

	dbPort, err := dbContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}
	port, sslmode, schema := dbPort.Port(), "disable", "public"

	*connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s", username, password, host, port, dbName, sslmode, schema)

	return dbContainer.Terminate, nil
}

func RunWithPostgres(m *testing.M, connStr *string) {
	teardown, err := StartPostgresContainer(connStr)
	if err != nil {
		log.Fatalf("could not start postgres test container: %v", err)
	}

	m.Run()

	if err := teardown(context.Background()); err != nil {
		log.Fatalf("could not teardown postgres test container: %v", err)
	}
}
