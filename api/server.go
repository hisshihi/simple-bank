package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
	"github.com/hisshihi/simple-bank-go/token"
	"github.com/hisshihi/simple-bank-go/util"
)

// Server обрабатывает все HTTP-запросы к банковскому сервису
type Server struct {
	config     util.Config
	store      sqlc.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer создаёт новый HTTP-сервер и настраивает маршруты
func NewServer(config util.Config, store sqlc.Store) (*Server, error) {
	tokenMaker, err := token.NewPASETOMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	} // создаём сервер

	// добавляем валидатор для проверки валюты
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter() // настраиваем маршруты
	return server, nil   // возвращаем сервер
}

// Настраивает маршруты
func (server *Server) setupRouter() {
	router := gin.Default() // создаём маршрутизатор

	// endpoints пользователей
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	router.PUT("/accounts/:id", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	// endpoints переводов
	router.POST("/transfers", server.createTransfer)

	// endpoints пользователей
	router.POST("/users", server.createUser)
	router.GET("/users/:username", server.getUser)
	router.POST("/users/login", server.loginUser)

	server.router = router // присваиваем маршрутизатор серверу
}

// Запускает HTTP сервер по указанному адресу
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
