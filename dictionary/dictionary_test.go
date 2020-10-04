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

	"github.com/alexamies/chinesenotes-go/webconfig"
)

func initDBCon() (*sql.DB, error) {
	conString := webconfig.DBConfig()
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

// Test trivial query with empty dictionary
func TestLoadDict(t *testing.T) {
	ctx := context.Background()
	database, err := initDBCon()
	if err != nil {
		t.Skipf("TestLoadDict: cannot connect to database: %v", err)
	}
	wdict, err := LoadDict(ctx, database)
	if err != nil {
		t.Fatalf("TestLoadDict: not able to load dictionary, skipping tests: %v\n", err)
	}
	if len(wdict) == 0 {
		t.Fatalf("TestLoadDict: len(wdict) = %d", len(wdict))
	}
	t.Logf("TestLoadDict: len(wdict): %d", len(wdict))
	trad := "煸"
	w1, ok := wdict[trad]
	if !ok {
		t.Fatalf("TestLoadDict: !ok: %s", trad)
	}
	if w1.HeadwordId == 0 {
		t.Error("TestLoadDict: w.HeadwordId == 0")
	}
	expectPinyin := "biān"
	if expectPinyin != w1.Pinyin {
		t.Errorf("TestLoadDict: expected pinyin: %s, got: %s", expectPinyin,
			w1.Pinyin)
	}
	w2 := wdict["與"]
	if w2.HeadwordId == 0 {
		t.Error("TestLoadDict: w.HeadwordId == 0")
	}
	if w2.Pinyin == "" {
		t.Error("TestLoadDict: w2.Pinyin == ''")
	}
	w3 := wdict["來"]
	if len(w3.Senses) < 2 {
		t.Error("len(w3.Senses) < 2, ", len(w3.Senses))
	}
	w4 := wdict["发"]
	if len(w4.Senses) < 2 {
		t.Error("len(w4.Senses) < 2, ", len(w4.Senses))
	}
}