package sqlc

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	testDB, err := NewTestDB()
	if err != nil {
		fmt.Errorf("ошибка в файле main_test.go", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
