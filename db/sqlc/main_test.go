package sqlc

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/hisshihi/simple-bank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Открываем соединение с базой данных
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// Создаем тестовый запрос
	testQueries = New(testDB)

	// Запускаем тесты
	os.Exit(m.Run())
}