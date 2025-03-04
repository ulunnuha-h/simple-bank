package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/ulunnuha-h/simple_bank/api"
	db "github.com/ulunnuha-h/simple_bank/db/sqlc"
	"github.com/ulunnuha-h/simple_bank/util"
)

func main(){
	err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Failed to load .env file")
	}

	conn, err := sql.Open(viper.GetString("DB_DRIVER"), viper.GetString("DB_SOURCE"))

	if err != nil {
		log.Fatal((err))
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(viper.GetString("SERVER_ADDRESS"))
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
	
}