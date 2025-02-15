package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
)

// Server обрабатывает все HTTP-запросы к банковскому сервису
type Server struct {
	store sqlc.Store
	router *gin.Engine
}

// NewServer создаёт новый HTTP-сервер и настраивает маршруты
func NewServer(store sqlc.Store) *Server {
	server := &Server{store: store} // создаём сервер
	router := gin.Default() // создаём маршрутизатор

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	router.PUT("/accounts/:id", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	server.router = router // присваиваем маршрутизатор серверу
	return server // возвращаем сервер
}

// Запускает HTTP сервер по указанному адресу
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}