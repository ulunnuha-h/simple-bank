package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/ulunnuha-h/simple_bank/util"
)

var testQuery *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	err = util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	// conn, err := pgx.Connect(context.Background(), dbSource)
	testDB, err = sql.Open(viper.GetString("DB_DRIVER"), viper.GetString("DB_SOURCE"))

	if err != nil {
		log.Fatal((err))
	}

	testQuery = New(testDB)

	os.Exit(m.Run())
}
