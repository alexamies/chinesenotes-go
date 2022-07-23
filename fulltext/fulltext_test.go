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

//
// Unit tests for the fulltext package
//
package fulltext

import (
	"testing"
)

func TestGetMatch(t *testing.T) {
	const txt = "厚人倫，美教化，移風俗。故詩有六義焉：一曰風，二曰賦，三曰比，四曰興，五曰雅，六曰頌。"
	tests := []struct {
		name        string
		txt         string
		queryTerms  []string
		wantLM      string
		wantEM      bool
		wantSnippet string
	}{
		{
			name:        "Happy path 1",
			txt:         txt,
			queryTerms:  []string{"曰", "風"},
			wantLM:      "曰風",
			wantEM:      true,
			wantSnippet: txt,
		},
		{
			name:        "Happy path 2",
			txt:         txt,
			queryTerms:  []string{"一", "曰風"},
			wantLM:      "一曰風",
			wantEM:      true,
			wantSnippet: txt,
		},
		{
			name:        "Happy path 3",
			txt:         txt,
			queryTerms:  []string{"故", "詩", "一"},
			wantLM:      "故詩",
			wantEM:      false,
			wantSnippet: txt,
		},
		{
			name:        "Happy path 4",
			txt:         txt,
			queryTerms:  []string{"一", "詩", "有"},
			wantLM:      "詩有",
			wantEM:      false,
			wantSnippet: txt,
		},
		{
			name:        "Snippet empty",
			txt:         txt,
			queryTerms:  []string{"美", "移", "故"},
			wantLM:      "故",
			wantEM:      false,
			wantSnippet: txt,
		},
	}
	for _, tc := range tests {
		mt := getMatch(tc.txt, tc.queryTerms)
		if mt.LongestMatch != tc.wantLM {
			t.Errorf("TestGetMatch.%s: got LM %s but want %s", tc.name, mt.LongestMatch, tc.wantLM)
		}
		if mt.ExactMatch != tc.wantEM {
			t.Errorf("TestgetMatch.%s: got ExactMatch %t but want %t", tc.name, mt.ExactMatch, tc.wantEM)
		}
		if mt.Snippet != tc.wantSnippet {
			t.Errorf("TestGetMatch.%s: got snippet %q but want %q", tc.name, mt.Snippet, tc.wantSnippet)
		}
	}
}

// Test to load a local file
func TestGetMatching1(t *testing.T) {
	loader := LocalTextLoader{"../corpus"}
	queryTerms := []string{"漢代"}
	mt, err := loader.GetMatching("example_collection/example_collection001.txt", queryTerms)
	if err != nil {
		t.Errorf("TestGetMatching1: got an error %v", err)
	}
	if mt.Snippet == "" {
		t.Errorf("TestGetMatching1: snippet empty")
	}
	t.Logf("fulltext.TestGetMatching1: match: %v", mt)
}

// Test to load a local file
func TestGetMatching2(t *testing.T) {
	t.Log("fulltext.TestGetMatching: Begin unit test")
	loader := LocalTextLoader{"../corpus"}
	queryTerms := []string{"曰風", "曰"}
	mt, err := loader.GetMatching("example_collection/example_collection002.txt", queryTerms)
	if err != nil {
		t.Errorf("TestGetMatching2: got an error %v", err)
	}
	if mt.Snippet == "" {
		t.Errorf("TestGetMatching2: snippet empty")
	}
	t.Logf("fulltext.TestGetMatching2: match: %v", mt)
}

// Test to load a local file
func TestGetMatching3(t *testing.T) {
	t.Log("fulltext.TestGetMatching: Begin unit test")
	loader := LocalTextLoader{"../corpus"}
	queryTerms := []string{"曰", "曰風"}
	mt, err := loader.GetMatching("example_collection/example_collection002.txt", queryTerms)
	if err != nil {
		t.Errorf("TestGetMatching3: got an error %v", err)
	}
	if mt.Snippet == "" {
		t.Errorf("TestGetMatching3: snippet empty")
	}
	t.Logf("fulltext.TestGetMatching3: match: %v", mt)
}
