package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
)

type Server struct {
	store  *sqlc.Store
	router *gin.Engine
}

func NewServer(store *sqlc.Store) *Server {
	server := &Server{
		store: store,
	}
	router := gin.Default()

	router.POST("/account", server.createAccount)
	router.GET("/account/:id", server.getAccount)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
