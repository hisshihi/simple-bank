package api

import (
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
	"github.com/hisshihi/simple-bank/db/sqlc"
	"github.com/hisshihi/simple-bank/internal/config"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store sqlc.Store) *Server {
	config := config.Config{
		TokenSymmetricKey:  gofakeit.Password(true, true, true, true, false, 32),
		AccesTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
