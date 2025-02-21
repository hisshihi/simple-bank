package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank-go/db/sqlc"
	"github.com/hisshihi/simple-bank-go/util"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store sqlc.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server ,err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	// Установка gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	// Запуск тестов
	os.Exit(m.Run())
}
