package database

import (
	"context"
	"testing"

	"github.com/artrctx/shuffle-core/tests/dbtest"
)

var dbConnStr string

func TestMain(m *testing.M) {
	dbtest.RunWithPostgres(m, &dbConnStr)
}

func TestServiceNew(t *testing.T) {
	srv, err := New(dbConnStr)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestServiceHealth(t *testing.T) {
	srv, err := New(dbConnStr)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	stats := srv.Health(context.Background())

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "The database is healthy." {
		t.Fatalf("expected message to be 'The database is healthy.', got %s", stats["message"])
	}
}

func TestServiceClose(t *testing.T) {
	srv, err := New(dbConnStr)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	srv.Close()
}
