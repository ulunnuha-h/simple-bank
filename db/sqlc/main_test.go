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
var testDB *sql.DB

func TestMain(m *testing.M) {
	// conn, err := pgx.Connect(context.Background(), dbSource)
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)

	if err != nil {
		log.Fatal((err))
	}

	testQuery = New(testDB)

	os.Exit(m.Run())
}
