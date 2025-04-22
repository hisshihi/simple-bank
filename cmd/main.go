package main

import (
	"database/sql"
	"log"

	"github.com/hisshihi/simple-bank/db/sqlc"
	"github.com/hisshihi/simple-bank/internal/config"
	"github.com/hisshihi/simple-bank/internal/service/api"
	_ "github.com/lib/pq"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config", err)
	}

	if config.ENV == "production" {
		conn, err := sql.Open(config.DBDriver, config.DBSouceProd)
		if err != nil {
			log.Fatal("не удалось подключится к базе данных")
		}

		store := sqlc.NewStore(conn)
		server, err := api.NewServer(config, store)
		if err != nil {
			log.Fatal("cannot create server", err)
		}

		err = server.Start(config.ServerAddress)
		if err != nil {
			log.Fatal("cannot create server", err)
		}
	} else {
		conn, err := sql.Open(config.DBDriver, config.DBSource)
		if err != nil {
			log.Fatal("не удалось подключится к базе данных")
		}

		store := sqlc.NewStore(conn)
		server, err := api.NewServer(config, store)
		if err != nil {
			log.Fatal("cannot create server", err)
		}

		err = server.Start(config.ServerAddress)
		if err != nil {
			log.Fatal("cannot create server", err)
		}
	}
}
