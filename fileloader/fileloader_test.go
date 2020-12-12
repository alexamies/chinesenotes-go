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
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// TestLoadNoDictFile tests with no files
func TestLoadNoDictFile(t *testing.T) {
	t.Log("TestLoadNoDictFile: Begin unit tests")
	appConfig := config.AppConfig{
		LUFileNames: []string{},
	}
	dict, err := LoadDictFile(appConfig)
	if err != nil {
		t.Fatalf("TestLoadNoDictFile: Got error %v", err)
	}
	if len(dict) != 0 {
		t.Error("TestLoadNoDictFile: len(dict) != 0")
	}
}

// TestLoadDictReader tests loadDictReader
func TestLoadDictReader(t *testing.T) {
	t.Log("TestLoadDictReader: Begin unit tests")
	avoidSub := make( map[string]bool)
	const inputOneEntry = `# comment
2	邃古	\N	suìgǔ	remote antiquity	noun	\N	\N	现代汉语	Modern Chinese	\N	\N	\N	\N	(CC-CEDICT '邃古'; Guoyu '邃古')	2
`
	const inputEmptyNotes = `# comment
2	邃古	\N	suìgǔ	remote antiquity	noun	\N	\N	现代汉语	Modern Chinese	\N	\N	\N	\N	\N	2
`
	const inputTwoEntries = `# comment
2	邃古	\N	suìgǔ	remote antiquity	noun	\N	\N	现代汉语	Modern Chinese	\N	\N	\N	\N	\N	2
25172	平地	\N	píngdì	flat land	noun	\N	\N	现代汉语	Modern Chinese	地理学	Geography	\N	\N	(CC-CEDICT '平地'; Guoyu '平地' 1)	25172
`

	type test struct {
		name string
		input string
		expectError bool
		expectSize int
		exampleSimp string
		expectPinyin string
		expectNoSenses int
		expectDomain string
		expectConcept string
		expectSubdomain string		
		expectNotes string		
  }
  tests := []test{
		{
			name: "Invalid entry",
			input: "Hello, Dictionary!",
			expectError: true,
			expectSize: 0,
			exampleSimp: "",
			expectPinyin: "",
			expectNoSenses: 0,
			expectDomain: "",
			expectConcept: "",
			expectSubdomain: "",
			expectNotes: "",
		},
		{
			name: "One entry",
			input: inputOneEntry,
			expectError: false,
			expectSize: 1,
			exampleSimp: "邃古",
			expectPinyin: "suìgǔ",
			expectNoSenses: 1,
			expectDomain: "Modern Chinese",
			expectConcept: "",
			expectSubdomain: "",
			expectNotes: "(CC-CEDICT '邃古'; Guoyu '邃古')",
		},
		{
			name: "Empty notes",
			input: inputEmptyNotes,
			expectError: false,
			expectSize: 1,
			exampleSimp: "邃古",
			expectPinyin: "suìgǔ",
			expectNoSenses: 1,
			expectDomain: "Modern Chinese",
			expectConcept: "",
			expectSubdomain: "",
			expectNotes: "",
		},
		{
			name: "Subdomain not empty",
			input: inputTwoEntries,
			expectError: false,
			expectSize: 2,
			exampleSimp: "平地",
			expectPinyin: "píngdì",
			expectNoSenses: 1,
			expectDomain: "Modern Chinese",
			expectConcept: "",
			expectSubdomain: "Geography",
			expectNotes: "(CC-CEDICT '平地'; Guoyu '平地' 1)",
		},
   }
  for _, tc := range tests {
		wdict := make(map[string]dicttypes.Word)
		r := strings.NewReader(tc.input)
		err := loadDictReader(r, wdict, avoidSub)
		if tc.expectError && (err == nil) {
			t.Fatalf("%s: expected an error but got none", tc.name)
		}
		if tc.expectError {
			continue
		}
		if !tc.expectError && (err != nil) {
			t.Fatalf("%s: did not expect an error but got %v", tc.name, err)
		}
		gotSize := len(wdict)
		if tc.expectSize != gotSize {
			t.Fatalf("%s: expectSize %d != %d", tc.name, tc.expectSize, gotSize)
		}
		w, ok := wdict[tc.exampleSimp]
		if !ok {
			t.Fatalf("%s: did not find expected term for '%s'", tc.name, tc.exampleSimp)
		}
		if tc.expectPinyin != w.Pinyin {
			t.Errorf("%s: expectPinyin %s != %s", tc.name, tc.expectPinyin, w.Pinyin)
		}
		if tc.expectNoSenses != len(w.Senses) {
			t.Fatalf("%s: expectNoSenses %d != %d", tc.name, tc.expectNoSenses, len(w.Senses))
		}
		s := w.Senses[0]
		if tc.expectDomain != s.Domain {
			t.Errorf("%s: expectDomain %s != %s", tc.name, tc.expectDomain, s.Domain)
		}
		if tc.expectConcept != s.Concept {
			t.Errorf("%s: Concept '%s' != %s", tc.name, tc.expectConcept, s.Concept)
		}
		if tc.expectSubdomain != s.Subdomain {
			t.Errorf("%s: Subdomain '%s' != %s", tc.name, tc.expectSubdomain, s.Subdomain)
		}
		if tc.expectNotes != s.Notes {
			t.Errorf("%s: Notes '%s' != %s", tc.name, tc.expectNotes, s.Notes)
		}
	}
}
