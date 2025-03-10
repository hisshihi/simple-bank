package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
)

// Обработка всех http запросов
type Server struct {
	store  sqlc.Store
	router *gin.Engine
}

// NewServer создаёт новый HTTP сервер и настраивает маршрутиризатор
func NewServer(store sqlc.Store) *Server {
	server := &Server{store: store}

	// создаёт новый маршрутизатор Gin с настройками по умолчанию
	router := gin.Default()

	// Настраиваем доверенные прокси
	router.SetTrustedProxies([]string{
		"127.0.0.1",      // локальный прокси
		"10.0.0.0/8",     // внутренняя сеть
		"172.16.0.0/12",  // Docker сети
		"192.168.0.0/16", // локальные сети
	})

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccountByID)
	router.GET("/accounts", server.listAccount)
	router.PUT("/accounts/:id", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	// сохраняет этот маршрутизатор в структуре сервера
	/*
		Через server.router мы потом будем добавлять все HTTP-маршруты (endpoints)
		Все middleware и обработчики запросов будут привязаны к этому router
		Когда сервер запустится, именно этот router будет обрабатывать все входящие HTTP-запросы
	*/
	server.router = router
	return server
}

// Запускает HTTP сервер
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
