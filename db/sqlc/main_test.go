package sqlc

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/hisshihi/simple-bank-go/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries // глобальная переменная для тестирования sqlc queries
var testDB *sql.DB // глобальная переменная для тестирования базы данных

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Подключение к базе данных
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// Создание нового экземпляра Queries
	testQueries = New(testDB)

	// Запуск тестов
	os.Exit(m.Run())
}