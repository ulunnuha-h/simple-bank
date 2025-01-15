package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:admin123@localhost:5432/simple_bank?sslmode=disable"
)

var testQuery *Queries

func TestMain(m *testing.M) {
	conn, err := pgx.Connect(context.Background(), dbSource)

	if err != nil {
		log.Fatal((err))
	}

	testQuery = New(conn)

	os.Exit(m.Run())
}
