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
package fileloader

import (
	"log"
	"testing"
)

// With no files
func TestLoadDictFile0(t *testing.T) {
	log.Printf("TestLoadDictFile0: Begin unit tests\n")
	fnames := []string{}
	dict, err := LoadDictFile(fnames)
	if err != nil {
		t.Fatalf("TestLoadDictFile0: Got error %v", err)
	}
	if len(dict) != 0 {
		t.Error("TestLoadDictFile0: len(dict) != 0")
	}
}

// With one file, 3 entries, 3 simplified + 1 traditional
func TestLoadDictFile1(t *testing.T) {
	fnames := []string{"../testdata/testdict.tsv"}
	dict, err := LoadDictFile(fnames)
	if err != nil {
		t.Fatalf("TestLoadDictFile1: Got an error: %v", err)
	}
	if len(dict) < 4 {
		t.Fatalf("TestLoadDictFile1: excpected at least 4, got %d", len(dict))
	}
	chinese := "邃古"
	word, ok := dict[chinese]
	if !ok {
		t.Fatalf("TestLoadDictFile1: could not find word %s", chinese)
	}
	senses := word.Senses
	if len(senses) ==0 {
		t.Fatalf("TestLoadDictFile1: expected > 0 senses, got %d", len(senses))
	}
	expectedDom := "Modern Chinese"
	domain := senses[0].Domain
	if expectedDom != domain {
		t.Errorf("TestLoadDictFile1: expected domain %s, got %s", expectedDom,
			domain)
	}
}
