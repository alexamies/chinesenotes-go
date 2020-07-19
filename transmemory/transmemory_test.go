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


// Unit tests for transmemory functions
package transmemory

import (
	"context"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"log"
	"testing"
)

// Test getChars function
func TestGetChars(t *testing.T) {
	log.Printf("TestGetChars: Begin unit tests\n")
	type test struct {
		query string
		expect []string
  }
	trad := []string{"結", "實"}
	log.Printf("TestGetChars: trad: %s", "結實")
  tests := []test{
		{query: "結實", expect: trad},
  }
  for _, tc := range tests {
		result := getChars(tc.query)
		for i, ch := range tc.query {
			if string(ch) != result[i] {
				t.Errorf("TestGetChars: expected %s, got %s", string(ch), result[i])
			}
		}
	}
}

// Test getChars function
func TestSearch(t *testing.T) {
	ctx := context.Background()
	database, err := dictionary.InitDBCon()
	if err != nil {
		t.Fatalf("TestSearch: cannot connect to database: %v", err)
	}
	searcher, err := NewSearcher(ctx, database)
	if err != nil {
		t.Fatalf("TestSearch: cannot create a searcher: %v", err)
	}
	w1 := dicttypes.Word{
		Simplified: "结实",
		Traditional: "結實",
		Pinyin: "jiēshi",
		HeadwordId: 10778,
		Senses: []dicttypes.WordSense{},
	}
	wdict := make(map[string]dicttypes.Word)
	wdict[w1.Traditional] = w1
	type test struct {
		name string
		query string
		domain string
		expectNo int
  }
  tests := []test{
		{
			name: "Happy path",
			query: "結實", 
			domain: "",
			expectNo: 1,
		},
		{
			name: "With domain",
			query: "結實", 
			domain: "Buddhism",
			expectNo: 0,
		},
  }
  for _, tc := range tests {
		results, err := searcher.Search(ctx, tc.query, tc.domain, wdict)
		if err != nil {
			t.Fatalf("TestSearch %s: error calling search: %v", tc.name, err)
		}
		numRes := len(results.Words)
		if tc.expectNo != numRes {
			t.Errorf("TestSearch %s: expect no results: %d, got: %d",
				tc.name, tc.expectNo, numRes)
		}
	}
}
