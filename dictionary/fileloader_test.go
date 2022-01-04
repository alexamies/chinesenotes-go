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
	"strings"
	"reflect"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

const inputTrad2SimpleNot121 = `# comment
393	了	\N	le	completion of an action	particle	动态助词	Aspectual Particle	现代汉语	Modern Chinese	虚词	Function Words	\N	le.mp3	In this usage 了 is an aspectual particle	393
5630	了	瞭	liǎo	to understand; to know	verb	\N	\N	文言文	Literary Chinese	\N	\N	\N	liao3.mp3	Traditional: 瞭; in the sense of 明白 or 清楚; as in 了解 (Guoyu '瞭' v 1)	393
16959	了	\N	le	modal particle	particle	语气助词	Modal Particle	现代汉语	Modern Chinese	虚词	Function Words	\N	le.mp3	In this use 了 appears at the end of a sentence as a modal particle	393
`

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
	if len(dict.Wdict) != 0 {
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
	const inputMultipleSenses = `# comment
25172	平地	\N	píngdì	flat land	noun	\N	\N	现代汉语	Modern Chinese	地理学	Geography	\N	\N	(CC-CEDICT '平地'; Guoyu '平地' 1)	25172
31834	平地	\N	píngdì	a plain	noun	\N	\N	现代汉语	Modern Chinese	地理学	Geography	\N	\N	(CC-CEDICT '平地'; Guoyu '平地' 2)	25172
`
	const inputTradDifferent = `# comment
8422	汉语	漢語	hànyǔ	Chinese language	noun	\N	\N	现代汉语	Modern Chinese	\N	\N	\N	\N	\N	8422
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
			expectError: false,
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
		{
			name: "Multiple senses",
			input: inputMultipleSenses,
			expectError: false,
			expectSize: 1,
			exampleSimp: "平地",
			expectPinyin: "píngdì",
			expectNoSenses: 2,
			expectDomain: "Modern Chinese",
			expectConcept: "",
			expectSubdomain: "Geography",
			expectNotes: "(CC-CEDICT '平地'; Guoyu '平地' 1)",
		},
		{
			name: "Traditional different to simplified",
			input: inputTradDifferent,
			expectError: false,
			expectSize: 2,
			exampleSimp: "漢語",
			expectPinyin: "hànyǔ",
			expectNoSenses: 1,
			expectDomain: "Modern Chinese",
			expectConcept: "",
			expectSubdomain: "",
			expectNotes: "",
		},
		{
			name: "Traditional to simplified not 1:1",
			input: inputTrad2SimpleNot121,
			expectError: false,
			expectSize: 2,
			exampleSimp: "瞭",
			expectPinyin: "le",
			expectNoSenses: 3,
			expectDomain: "Modern Chinese",
			expectConcept: "Aspectual Particle",
			expectSubdomain: "Function Words",
			expectNotes: "In this usage 了 is an aspectual particle",
		},
   }
  for _, tc := range tests {
		wdict := make(map[string]*dicttypes.Word)
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
			t.Fatalf("%s: expectSize got %d, want %d", tc.name, gotSize, tc.expectSize)
		}
		if tc.expectSize == 0 {
			continue
		}
		w, ok := wdict[tc.exampleSimp]
		if !ok {
			t.Fatalf("%s: did not find expected term for '%s'", tc.name, tc.exampleSimp)
		}
		if tc.expectPinyin != w.Pinyin {
			t.Errorf("%s: expectPinyin %s != %s", tc.name, tc.expectPinyin, w.Pinyin)
		}
		if tc.expectNoSenses != len(w.Senses) {
			t.Fatalf("%s: expectNoSenses got %d, want %d", tc.name, len(w.Senses), tc.expectNoSenses)
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


// TestLoadDictReaderDomains tests that loadDictReader sets values correctly
func TestLoadDictReaderValues(t *testing.T) {
	t.Log("TestLoadDictReaderValues: Begin unit tests")
	avoidSub := make( map[string]bool)
	ws1 := dicttypes.WordSense{
		Id: 393,
		HeadwordId: 393,
		Simplified: "了",
		Traditional: "\\N",
		Pinyin: "le",
		English: "completion of an action",
		Grammar: "particle",
		ConceptCN: "动态助词",
		Concept: "Aspectual Particle",
		DomainCN: "现代汉语",
		Domain: "Modern Chinese",
		Subdomain: "Function Words",
		Image: "\\N",
		MP3: "le.mp3",
		Notes: "In this usage 了 is an aspectual particle",
	}
	ws2 := dicttypes.WordSense{
		Id: 5630,
		HeadwordId: 393,
		Simplified: "了",
		Traditional: "瞭",
		Pinyin: "liǎo",
		English: "to understand; to know",
		Grammar: "verb",
		ConceptCN: "",
		Concept: "",
		DomainCN: "文言文",
		Domain: "Literary Chinese",
		Subdomain: "",
		Image: "\\N",
		MP3: "liao3.mp3",
		Notes: "Traditional: 瞭; in the sense of 明白 or 清楚; as in 了解 (Guoyu '瞭' v 1)",
	}
	ws3 := dicttypes.WordSense{
		Id: 16959,
		HeadwordId: 393,
		Simplified: "了",
		Traditional: "\\N",
		Pinyin: "le",
		English: "modal particle",
		Grammar: "particle",
		ConceptCN: "语气助词",
		Concept: "Modal Particle",
		DomainCN: "现代汉语",
		Domain: "Modern Chinese",
		Subdomain: "Function Words",
		Image: "\\N",
		MP3: "le.mp3",
		Notes: "In this use 了 appears at the end of a sentence as a modal particle",
	}
	hw1 := dicttypes.Word{
		Simplified: "了", 
		Traditional: "\\N",
		Pinyin: "le",
		Senses: []dicttypes.WordSense{ws1, ws2, ws3},
	}

	type test struct {
		name string
		input string
		expectWord dicttypes.Word
  }
  tests := []test{
		{
			name: "Traditional to simplified not 1:1",
			input: inputTrad2SimpleNot121,
			expectWord: hw1,
		},
   }
  for _, tc := range tests {
		wdict := make(map[string]*dicttypes.Word)
		r := strings.NewReader(tc.input)
		err := loadDictReader(r, wdict, avoidSub)
		if err != nil {
			t.Fatalf("%s: unexpected an error %v", tc.name, err)
		}
		w, ok := wdict["了"]
		if !ok {
			t.Fatalf("%s: did not find expected term for '%s'", tc.name, "了")
		}
		if !reflect.DeepEqual(w.Senses, tc.expectWord.Senses) {
			t.Fatalf("%s: got senses\n%v\nwant\n%v", tc.name, w.Senses, tc.expectWord.Senses)
		}
	}
}