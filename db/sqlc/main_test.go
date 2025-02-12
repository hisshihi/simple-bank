package sqlc

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries // глобальная переменная для тестирования sqlc queries
var testDB *sql.DB // глобальная переменная для тестирования базы данных

func TestMain(m *testing.M) {
	var err error

	// Подключение к базе данных
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// Создание нового экземпляра Queries
	testQueries = New(testDB)

	// Запуск тестов
	os.Exit(m.Run())
}