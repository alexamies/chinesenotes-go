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
	"database/sql"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

func initDBCon() (*sql.DB, error) {
	conString := webconfig.DBConfig()
	return sql.Open("mysql", conString)
}

func mockDict() map[string]dicttypes.Word {
	w1 := dicttypes.Word{
		Simplified: "结实",
		Traditional: "結實",
		Pinyin: "jiēshi",
		HeadwordId: 10778,
		Senses: []dicttypes.WordSense{},
	}
	w2 := dicttypes.Word{
		Simplified: "结",
		Traditional: "結",
		Pinyin: "jiē",
		HeadwordId: 42,
		Senses: []dicttypes.WordSense{},
	}
	w3 := dicttypes.Word{
		Simplified: "实",
		Traditional: "實",
		Pinyin: "shí",
		HeadwordId: 43,
		Senses: []dicttypes.WordSense{},
	}
	w4 := dicttypes.Word{
		Simplified: "开花结实",
		Traditional: "開花結實",
		Pinyin: "kāi huā jiē shi",
		HeadwordId: 100973,
		Senses: []dicttypes.WordSense{},
	}
	w5 := dicttypes.Word{
		Simplified: "大方广入如来智德不思议经",
		Traditional: "大方廣入如來智德不思議經",
		Pinyin: "Dàfāngguǎng Rù Rúlái Zhì Dé Bù Sīyì Jīng",
		HeadwordId: 1234,
		Senses: []dicttypes.WordSense{},
	}
	w6 := dicttypes.Word{
		Simplified: "从门入者不是家珍",
		Traditional: "從門入者不是家珍",
		Pinyin: "cóng mén rù zhě bùshì jiāzhēn",
		HeadwordId: 1235,
		Senses: []dicttypes.WordSense{},
	}
	wdict := make(map[string]dicttypes.Word)
	wdict[w1.Traditional] = w1
	wdict[w2.Traditional] = w2
	wdict[w3.Traditional] = w3
	wdict[w4.Traditional] = w4
	wdict[w5.Traditional] = w5
	wdict[w6.Traditional] = w6
	return wdict
}

func TestAddMatches(t *testing.T) {
	wdict := mockDict()
	type test struct {
		name string
		matches1 []*tmResult
		matches2 []*tmResult
		expect []*tmResult
  }
	var matches1, matches2, expect []*tmResult
  tests := []test{
		{
			name: "Empty set",
			matches1: matches1,
			matches2: matches2,
			expect: expect,
		},
  }
  for _, tc := range tests {
		result := addMatches(tc.matches1, tc.matches2, wdict)
		if len(tc.expect) != len(result) {
			t.Fatalf("%s: expected len %d, got %d", tc.name, len(tc.expect),
				len(result))
		}
	}
}

// Test getChars function
func TestCombineResults(t *testing.T) {
	type test struct {
		name string
		query string
		matches []*tmResult
		expect []string
  }
  // for query 結實
  mPartial := tmResult{
		term: "結",
		unigramCount: 1,
  }
  mExact := tmResult{
		term: "結實",
		unigramCount: 2,
  }
  mPoor := tmResult{
		term: "實",
		unigramCount: 1,
  }
  mLong := tmResult{
		term: "開花結實",
		unigramCount: 2,
  }
  matches := []*tmResult{&mPartial, &mExact, &mPoor, &mLong}
  expect := []string{"結實", "結",  "實", "開花結實"}
  // For query 把手拽不入
  mLong1 := tmResult{
		term: "大方廣入如來智德不思議經",
		unigramCount: 2,
  }
  mLong2 := tmResult{
		term: "從門入者不是家珍",
		unigramCount: 2,
  }
  lMatches := []*tmResult{&mLong1, &mLong2}
  lExpect := []string{"從門入者不是家珍", "大方廣入如來智德不思議經"}
  tests := []test{
		{
			name: "happy path",
			query: "結實",
			matches: matches,
			expect: expect,
		},
		{
			name: "long strings",
			query: "把手拽不入",
			matches: lMatches,
			expect: lExpect,
		},
  }
  var pinyinMatches []*tmResult
  wdict := mockDict()
  for _, tc := range tests {
		result := combineResults(tc.query, tc.matches, pinyinMatches, wdict)
		if len(tc.expect) != len(result) {
			t.Errorf("%s: expected len %d, got %d", tc.name, len(tc.expect),
				len(result))
			continue
		}
		for i, term := range tc.expect {
			if term != result[i].Traditional {
				t.Errorf("%s, query: %s, %d: expected %s, got %s",
						tc.name, tc.query, i, term, result[i].Traditional)
			}
		}
	}
}

// Test getChars function
func TestCombineScores(t *testing.T) {
	type test struct {
		name string
		query string
		match *tmResult
		expect float64
  }
  mPartial := tmResult{
		term: "結",
		unigramCount: 1,
		hamming: 1,
  }
  mExact := tmResult{
		term: "結實",
		unigramCount: 2,
		hamming: 0,
  }
  mNoOverlap := tmResult{
		term: "",
		unigramCount: 0,
		hamming: 2,
  }
  const normalUni = 2.0 / 5.0
  const normalHamming = 12 / 5.0
  const expectLong = normalUni * uniCountWeight +
  		normalHamming * hammingWeight
  mLong := tmResult{
		term: "大方廣入如來智德不思議經)",
		unigramCount: 2,
		hamming: 12,
  }
  tests := []test{
		{
			name: "Happy path",
			query: "結實",
			match: &mPartial,
			expect: 0.0,
		},
		{
			name: "Same",
			query: "結實",
			match: &mExact,
			expect: 1.0,
		},
		{
			name: "differenter and differenter",
			query: "結實",
			match: &mNoOverlap,
			expect: -1.0,
		},
		{
			name: "long example",
			query: "把手拽不入",
			match: &mLong,
			expect: expectLong,
		},
   }
  for _, tc := range tests {
		result := combineScores(tc.query, tc.match)
		if tc.expect != result {
			t.Errorf("%s: expected %f, got %f", tc.name, tc.expect,
					result)
		}
	}
}

// Test getChars function
func TestGetChars(t *testing.T) {
	t.Log("TestGetChars: Begin unit tests")
	type test struct {
		query string
		expect []string
  }
	trad := []string{"結", "實"}
	t.Logf("TestGetChars: trad: %s", "結實")
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
func TestHamming(t *testing.T) {
	type test struct {
		name string
		query string
		term string
		expect int
  }
  tests := []test{
		{
			name: "Same string",
			query: "結實",
			term: "結實",
			expect: 0,
		},
		{
			name: "Zero length term",
			query: "結實",
			term: "",
			expect: 2,
		},
		{
			name: "One char same",
			query: "結實",
			term: "結束",
			expect: 1,
		},
		{
			name: "Query is a substring",
			query: "一玉",
			term: "越海一玉",
			expect: 4,
		},
		{
			name: "Term is longer than query",
			query: "一玉",
			term: "一玉一",
			expect: 1,
		},
		{
			name: "Long example with no overlap",
			query: "把手拽不入",
			term: "大方廣入如來智德不思議經",
			expect: 12,
		},
  }
  for _, tc := range tests {
		result := hammingDist(tc.query, tc.term)
		if tc.expect != result {
			t.Errorf("%s: expected %d, got %d", tc.name, tc.expect,
					result)
		}
	}
}

// Test getChars function
func TestQueryPinyin(t *testing.T) {
	ctx := context.Background()
	database, err := initDBCon()
	if err != nil {
		t.Skipf("cannot connect to database: %v", err)
	}
	searcher, err := NewSearcher(ctx, database)
	if err != nil {
		t.Skipf("cannot create a searcher: %v", err)
	}
	wdict := mockDict()
	matches, err := searcher.queryPinyin(ctx, "結實", "", wdict)
	if err != nil {
		t.Fatalf("error calling queryPinyin: %v", err)
	}
	if len(matches) == 0 {
		t.Errorf("no results")
	}
}

// Test getChars function
func TestSearch(t *testing.T) {
	ctx := context.Background()
	database, err := initDBCon()
	if err != nil {
		t.Skipf("cannot connect to database: %v", err)
	}
	searcher, err := NewSearcher(ctx, database)
	if err != nil {
		t.Skipf("cannot create a searcher: %v", err)
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
		expectTop string
  }
  tests := []test{
		{
			name: "Happy path",
			query: "結實", 
			domain: "",
			expectNo: 1,
			expectTop: "結實",
		},
		{
			name: "With domain",
			query: "結實", 
			domain: "Buddhism",
			expectNo: 0,
			expectTop: "結實",
		},
  }
  for _, tc := range tests {
		results, err := searcher.Search(ctx, tc.query, tc.domain, wdict)
		if err != nil {
			t.Fatalf("%s: error calling search: %v", tc.name, err)
		}
		numRes := len(results.Words)
		if tc.expectNo != numRes {
			t.Errorf("%s: expect no results: %d, got: %d",
				tc.name, tc.expectNo, numRes)
		}
		if numRes > 0 {
			top := results.Words[0].Traditional
			if tc.expectTop != top {
				t.Errorf("%s: expect top result: %s, got: %s",
					tc.name, tc.expectTop, top)
			}
		}
	}
}
