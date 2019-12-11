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
 	"fmt"
	"testing"
)

// Trival test
func TestGetMatches0(t *testing.T) {
	fmt.Printf("fulltext.TestGetMatches0: Begin unit test\n")
	queryTerms := []string{}
	fileNames := []string{}
	dm := GetMatches(fileNames, queryTerms)
	fmt.Printf("fulltext.TestGetMatches0: match: %v\n", dm)
}

// Trival test
func TestGetMatches1(t *testing.T) {
	fmt.Printf("fulltext.TestGetMatches1: Begin unit test\n")
	queryTerms := []string{"曰風"}
	fn := "shijing/shijing001.txt"
	fileNames := []string{fn}
	docMatches := GetMatches(fileNames, queryTerms)
	if len(docMatches) == 0 {
		t.Errorf("TestGetMatches1: docMatches empty\n")
		return
	}
	snippet := docMatches[fn].MT.Snippet
	if len(snippet) == 0 {
		t.Errorf("TestGetMatches1: snippet empty\n")
		return
	}
	fmt.Printf("fulltext.TestGetMatches1: match: %v\n", docMatches)
}

// Trival test
func TestGetMatches2(t *testing.T) {
	fmt.Printf("fulltext.TestGetMatches1: Begin unit test\n")
	queryTerms := []string{"曰風"}
	fn0 := "shijing/shijing001.txt"
	fn1 := "shijing/shijing002.txt"
	fileNames := []string{fn0, fn1}
	docMatches := GetMatches(fileNames, queryTerms)
	if len(docMatches) == 0 {
		t.Errorf("TestGetMatches2: docMatches empty\n")
		return
	}
	snippet := docMatches[fn0].MT.Snippet
	if len(snippet) == 0 {
		t.Errorf("TestGetMatches2: snippet empty\n")
		return
	}
	fmt.Printf("fulltext.TestGetMatches2: match: %v\n", docMatches)
}
