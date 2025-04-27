package gapi

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
	"github.com/hisshihi/simple-bank/internal/config"
	"github.com/hisshihi/simple-bank/pb"
	"github.com/hisshihi/simple-bank/pkg/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config     config.Config
	store      sqlc.Store
	tokenMaker util.Maker
	router     *gin.Engine
}

// NewServer creates a new gRPC server
func NewServer(config config.Config, store sqlc.Store) (*Server, error) {
	tokenMaker, err := util.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
