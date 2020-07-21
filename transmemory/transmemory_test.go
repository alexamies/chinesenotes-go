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
func TestCombineResults(t *testing.T) {
	type test struct {
		name string
		query string
		matches []tmResult
		expect []string
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
  mPoor := tmResult{
		term: "實",
		unigramCount: 1,
		hamming: 2,
  }
  matches := []tmResult{mPartial, mExact, mPoor}
  expect := []string{"結實", "結", "實"}
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
	wdict := make(map[string]dicttypes.Word)
	wdict[w1.Traditional] = w1
	wdict[w2.Traditional] = w2
	wdict[w3.Traditional] = w3
  tests := []test{
		{
			name: "Happy path",
			query: "結實",
			matches: matches,
			expect: expect,
		},
   }
  for _, tc := range tests {
		result := combineResults(tc.query, tc.matches, wdict)
		if len(tc.expect) != len(result) {
			t.Errorf("%s: expected len %d, got %d", tc.name, len(tc.expect),
				len(result))
			continue
		}
		for i, term := range tc.expect {
			if term != result[i].Traditional {
				t.Errorf("%s: expected %s, got %s", tc.name, term, result[i].Traditional)
			}
		}
	}
}

// Test getChars function
func TestCombineScores(t *testing.T) {
	type test struct {
		name string
		query string
		match tmResult
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
  tests := []test{
		{
			name: "Happy path",
			query: "結實",
			match: mPartial,
			expect: 0.0,
		},
		{
			name: "Same",
			query: "結實",
			match: mExact,
			expect: 1.0,
		},
		{
			name: "differenter and differenter",
			query: "結實",
			match: mNoOverlap,
			expect: -1.0,
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
