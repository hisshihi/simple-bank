package main

import (
	"database/sql"
	"log"

	"github.com/hisshihi/simple-bank-go/api"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
	"github.com/hisshihi/simple-bank-go/util"

	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	
	// Подключение к базе данных
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := sqlc.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
