package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// Установка gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	// Запуск тестов
	os.Exit(m.Run())
}
