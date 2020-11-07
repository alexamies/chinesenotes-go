// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


// Unit tests for query parsing functions
package dictionary

import (
	"context"
	"database/sql"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
)

func initDBCon() (*sql.DB, error) {
	conString := config.DBConfig()
	return sql.Open("mysql", conString)
}

// Test trivial query with empty dictionary
func TestFindWordsByEnglish(t *testing.T) {
	t.Log("TestFindWordsByEnglish1: Begin unit tests")
	ctx := context.Background()
	database, err := initDBCon()
	if err != nil {
		t.Skipf("TestFindWordsByEnglish: cannot connect to database: %v", err)
	}
	dictSearcher := NewSearcher(ctx, database)
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
