package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/hisshihi/simple-bank/db/sqlc"
	"github.com/hisshihi/simple-bank/internal/config"
	"github.com/hisshihi/simple-bank/internal/service/api"
	"github.com/hisshihi/simple-bank/internal/service/gapi"
	"github.com/hisshihi/simple-bank/pb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
		runGrpcServer(config, store)
	} else {
		conn, err := sql.Open(config.DBDriver, config.DBSource)
		if err != nil {
			log.Fatal("не удалось подключится к базе данных")
		}

		store := sqlc.NewStore(conn)
		runGrpcServer(config, store)
	}
}

func runGrpcServer(config config.Config, store sqlc.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server", err)
	}
}

func runGinServer(config config.Config, store sqlc.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
