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


// Unit tests for lookup package
package dictionary

import (
	"context"
	"log"
	"testing"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Query expecting empty list
func TestAddWordSense2Map(t *testing.T) {
	wmap := map[string]dicttypes.Word{}
	ws := dicttypes.WordSense{
		Id: 1,
		HeadwordId: 1,
		Simplified: "我",
		Traditional: "",
		Pinyin: "wǒ",
		English: "me",
		Notes: "No notes",
	}
	addWordSense2Map(wmap, ws)
	if len(wmap) != 1 {
		t.Error("TestAddWordSense2Map: unexpected length, ", len(wmap))
	}
}

// Test trivial query with empty query, expect error
func TestLookupSubstr(t *testing.T) {
	log.Printf("TestLookupSubstr: Begin unit tests\n")
	ctx := context.Background()
	database, err := InitDBCon()
	if err != nil {
		t.Fatalf("TestFindWordsByEnglish: cannot connect to database: %v", err)
	}
	dictSearcher, err := NewSearcher(ctx, database)
	if err != nil {
		t.Fatalf("TestFindWordsByEnglish: cannot create dictSearcher: %v", err)
	}
	type test struct {
		name string
		query string
		domain string
		expectErr bool
		expectNum int
  }
  tests := []test{
		{	name: "expect error",
			query: "",
			domain: "",
			expectErr: true,
		 	expectNum: 0,
		 },
		{	name: "expect empty",
			query: "我還不知道",
			domain: "",
			expectErr: false,
		 	expectNum: 0,
		 },
		{	name: "invalid domain",
			query: "置",
			domain: "invalid",
			expectErr: false,
		 	expectNum: 0,
		 },
  }
  for _, tc := range tests {
		results, err := dictSearcher.LookupSubstr(ctx, tc.query, "", "")
		if tc.expectErr && err == nil {
			t.Errorf("TestLookupSubstr: %s, expect an error, got none", tc.name)
			continue
		}
		if tc.expectErr {
			continue
		}
		if !tc.expectErr && err != nil {
			t.Errorf("TestLookupSubstr: %s, expect no error, got: %v", tc.name, err)
			continue
		}
		if results == nil {
			t.Errorf("TestLookupSubstr: %s, results nil", tc.name)
			continue
		}
		resNum := len(results.Words)
		if tc.expectNum != resNum {
			t.Errorf("TestLookupSubstr: %s, expected %d results, got: %v", tc.name,
					tc.expectNum, resNum)
		}
	}
}
