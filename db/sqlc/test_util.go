package sqlc

import (
	"database/sql"
	"log"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

func NewTestDB() (*sql.DB, error) {
	var err error
	testDB, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("не удалось подключится к базе данных")
	}
	return testDB, nil
}
