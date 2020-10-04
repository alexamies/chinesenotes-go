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

func TestgetMatch0(t *testing.T) {
	t.Log("fulltext.TestgetMatch0: Begin unit test")
	txt := "厚人倫，美教化，移風俗。故詩有六義焉：一曰風，二曰賦，三曰比，四曰興，五曰雅，六曰頌。"
	queryTerms := []string{"曰", "風"}
	mt := getMatch(txt, queryTerms)
	if mt.Snippet == "" {
		t.Errorf("TestgetMatch0: snippet empty")
	}
	expectLM := "曰風"
	if mt.LongestMatch != expectLM {
		t.Errorf("TestgetMatch0: expect %s. got %s", expectLM, mt.LongestMatch)
	}
	expectEM := true
	if mt.ExactMatch != expectEM {
		t.Errorf("TestgetMatch0: expect %v. got %v", expectEM, mt.ExactMatch)
	}
}

func TestgetMatch1(t *testing.T) {
	t.Log("fulltext.TestgetMatch0: Begin unit test")
	txt := "厚人倫，美教化，移風俗。故詩有六義焉：一曰風，二曰賦，三曰比，四曰興，五曰雅，六曰頌。"
	queryTerms := []string{"一", "曰風"}
	mt := getMatch(txt, queryTerms)
	if mt.Snippet == "" {
		t.Errorf("TestgetMatch1: snippet empty")
	}
	expectLM := "一曰風"
	if mt.LongestMatch != expectLM {
		t.Errorf("TestgetMatch1: expect %s. got %s", expectLM, mt.LongestMatch)
	}
	expectEM := true
	if mt.ExactMatch != expectEM {
		t.Errorf("TestgetMatch1: expect %v. got %v", expectEM, mt.ExactMatch)
	}
}

func TestgetMatch2(t *testing.T) {
	t.Log("fulltext.TestgetMatch0: Begin unit test")
	txt := "厚人倫，美教化，移風俗。故詩有六義焉：一曰風，二曰賦，三曰比，四曰興，五曰雅，六曰頌。"
	queryTerms := []string{"故", "詩", "一"}
	mt := getMatch(txt, queryTerms)
	if mt.Snippet == "" {
		t.Errorf("TestgetMatch2: snippet empty")
	}
	expectLM := "故詩"
	if mt.LongestMatch != expectLM {
		t.Errorf("TestgetMatch2: expect %s. got %s", expectLM, mt.LongestMatch)
	}
	expectEM := false
	if mt.ExactMatch != expectEM {
		t.Errorf("TestgetMatch2: expect %v. got %v", expectEM, mt.ExactMatch)
	}
}

func TestgetMatch3(t *testing.T) {
	t.Log("fulltext.TestgetMatch0: Begin unit test")
	txt := "厚人倫，美教化，移風俗。故詩有六義焉：一曰風，二曰賦，三曰比，四曰興，五曰雅，六曰頌。"
	queryTerms := []string{"一", "詩", "有"}
	mt := getMatch(txt, queryTerms)
	if mt.Snippet == "" {
		t.Errorf("TestgetMatch3: snippet empty")
	}
	expectLM := "詩有"
	if mt.LongestMatch != expectLM {
		t.Errorf("TestgetMatch3: expect %s. got %s", expectLM, mt.LongestMatch)
	}
	expectEM := false
	if mt.ExactMatch != expectEM {
		t.Errorf("TestgetMatch3: expect %v. got %v", expectEM, mt.ExactMatch)
	}
}

func TestgetMatch4(t *testing.T) {
	t.Log("fulltext.TestgetMatch0: Begin unit test")
	txt := "厚人倫，美教化，移風俗。故詩有六義焉：一曰風，二曰賦，三曰比，四曰興，五曰雅，六曰頌。"
	queryTerms := []string{"美", "移", "故"}
	mt := getMatch(txt, queryTerms)
	if mt.Snippet == "" {
		t.Errorf("TestgetMatch4: snippet empty")
	}
	if mt.LongestMatch == "" {
		t.Errorf("TestgetMatch4: LongestMatch empty")
	}
	expectEM := false
	if mt.ExactMatch != expectEM {
		t.Errorf("TestgetMatch4: expect %v. got %v", expectEM, mt.ExactMatch)
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
