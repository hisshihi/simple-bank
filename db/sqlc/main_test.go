package sqlc

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/hisshihi/simple-bank/internal/config"
	_ "github.com/lib/pq"
)

var (
	testQueries *Queries
	testDB      *sql.DB
)

func TestMain(m *testing.M) {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("не удалось подключится к базе данных")
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
