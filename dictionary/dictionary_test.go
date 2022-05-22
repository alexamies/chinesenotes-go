package dictionary

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
)

func initDBCon() (*sql.DB, error) {
	conString := config.DBConfig()
	return sql.Open("mysql", conString)
}

// Test trivial query with empty dictionary
func TestFindWordsByEnglish(t *testing.T) {
	if dbHost := os.Getenv("DATABASE"); len(dbHost) == 0 {
		t.Skip("TestFindWordsByEnglish: skipping")
	}
	t.Log("TestFindWordsByEnglish: Begin unit tests")
	ctx := context.Background()
	database, err := initDBCon()
	if err != nil {
		t.Skipf("TestFindWordsByEnglish: cannot connect to database: %v", err)
	}
	dictSearcher := NewDBSearcher(ctx, database)
	if !dictSearcher.Initialized() {
		t.Skip("TestFindWordsByEnglish: cannot iniitalize DB")
	}
	senses, err := dictSearcher.FindWordsByEnglish(ctx, "hello")
	if err != nil {
		t.Errorf("TestFindWordsByEnglish: got error: %v", err)
	}
	if len(senses) == 0 {
		t.Errorf("TestFindWordsByEnglish: got no results: %d", len(senses))
	}
}
