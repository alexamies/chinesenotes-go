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

// Trival test
func TestGetMatches0(t *testing.T) {
	t.Log("fulltext.TestGetMatches0: Begin unit test")
	queryTerms := []string{}
	fileNames := []string{}
	dm := GetMatches(fileNames, queryTerms)
	t.Logf("fulltext.TestGetMatches0: match: %v", dm)
}

// Trival test
func TestGetMatches1(t *testing.T) {
	t.Log("fulltext.TestGetMatches1: Begin unit test")
	queryTerms := []string{"曰風"}
	fn := "example_collection/example_collection002.txt"
	fileNames := []string{fn}
	docMatches := GetMatches(fileNames, queryTerms)
	if len(docMatches) == 0 {
		t.Error("docMatches empty")
		return
	}
	snippet := docMatches[fn].MT.Snippet
	if len(snippet) == 0 {
		t.Error("Tsnippet empty")
		return
	}
	t.Logf("fulltext.TestGetMatches1: match: %v", docMatches)
}

// Trival test
func TestGetMatches2(t *testing.T) {
	queryTerms := []string{"曰風"}
	fn0 := "example_collection/example_collection001.txt"
	fn1 := "example_collection/example_collection002.txt"
	fileNames := []string{fn0, fn1}
	docMatches := GetMatches(fileNames, queryTerms)
	if len(docMatches) == 0 {
		t.Error("docMatches empty")
		return
	}
	snippet := docMatches[fn1].MT.Snippet
	if len(snippet) == 0 {
		t.Error("snippet empty")
		return
	}
}
