package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:admin123@localhost:5432/simple_bank?sslmode=disable"
)

var testQuery *Queries

func TestMain(m *testing.M) {
	// conn, err := pgx.Connect(context.Background(), dbSource)
	conn, err := sql.Open(dbDriver, dbSource)

	if err != nil {
		log.Fatal((err))
	}

	testQuery = New(conn)

	os.Exit(m.Run())
}
