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


// Unit tests for config functions
package config

import (
	"log"
	"reflect"
	"testing"
)

// Test trivial query with empty chunk
func TestCorpusDataDir(t *testing.T) {
	log.Printf("TestCorpusDataDir: Begin unit tests\n")
	result := CorpusDataDir()
	expect := "../data/corpus"
	if expect != result {
		t.Errorf("expected: %s, got: %s", expect, result)
	}
}

// Test AvoidSubDomains
func TestAvoidSubDomains(t *testing.T) {
	result := AvoidSubDomains()
	expect := make(map[string]bool)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected: %v, got: %v", expect, result)
	}
}
