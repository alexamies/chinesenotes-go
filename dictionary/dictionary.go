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
// Chinese-English dictionary database search functions

package dictionary

import (
	"context"
	"strings"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Dictionary is a struct to hold word dictionary indexes
type Dictionary struct {
	// Forward dictionary, lookup by Chinese word
	Wdict       map[string]*dicttypes.Word
	HeadwordIds map[int]*dicttypes.Word
}

func NewDictionary(wdict map[string]*dicttypes.Word) *Dictionary {
	hwIdMap := make(map[int]*dicttypes.Word)
	for _, w := range wdict {
		hwIdMap[w.HeadwordId] = w
	}
	return &Dictionary{
		Wdict:       wdict,
		HeadwordIds: hwIdMap,
	}
}

// ReverseIndex searches the dictionary by reverse lookup, eg to Chinese
type ReverseIndex interface {
	// Find searches from English, pinyin, or multilingual equivalents contained in notes to Chinese
	Find(ctx context.Context, query string) ([]dicttypes.WordSense, error)
}

type SubstringIndex interface {
	LookupSubstr(ctx context.Context, query, topic_en, subtopic_en string) (*Results, error)
}

type reverseIndexMem struct {
	revIndex map[string][]dicttypes.WordSense
}

func NewReverseIndex(dict *Dictionary, nExtractor *NotesExtractor) ReverseIndex {
	revIndex := map[string][]dicttypes.WordSense{}
	for _, v := range dict.HeadwordIds {
		for _, s := range v.Senses {
			tokens := splitEnglish(s.English)
			for _, eng := range tokens {
				e := strings.ToLower(eng)
				add(revIndex, e, s)
			}
			if len(s.Pinyin) > 0 {
				p := dicttypes.NormalizePinyin(s.Pinyin)
				add(revIndex, p, s)

			}
			if len(s.Notes) > 0 {
				equivalents := nExtractor.Extract(s.Notes)
				for _, eq := range equivalents {
					e := strings.ToLower(eq)
					add(revIndex, e, s)
				}
			}
		}
	}
	return reverseIndexMem{
		revIndex: revIndex,
	}
}

// add add the word sense s to the reverse index revIndex with given key
func add(revIndex map[string][]dicttypes.WordSense, key string, s dicttypes.WordSense) {
	if s.Traditional == "\\N" {
		s.Traditional = ""
	}
	if senses, ok := revIndex[key]; ok {
		// Avoid
		found := false
		for _, ws := range senses {
			if s.HeadwordId == ws.HeadwordId {
				found = true
			}
		}
		if !found {
			senses = append(senses, s)
			revIndex[key] = senses
		}
	} else {
		revIndex[key] = []dicttypes.WordSense{s}
	}
}

func (r reverseIndexMem) Find(ctx context.Context, query string) ([]dicttypes.WordSense, error) {
	return r.revIndex[query], nil
}

func splitEnglish(eng string) []string {
	tokens := strings.Split(eng, "; ")
	results := []string{}
	for _, t := range tokens {
		s := stripParen(t)
		results = append(results, stripStopWords(s))
	}
	return results
}

func stripStopWords(t string) string {
	if strings.HasPrefix(t, "a ") {
		return strings.Replace(t, "a ", "", 1)
	} else if strings.HasPrefix(t, "an ") {
		return strings.Replace(t, "an ", "", 1)
	} else if strings.HasPrefix(t, "to ") {
		return strings.Replace(t, "to ", "", 1)
	}
	return t
}

func stripParen(t string) string {
	if i := strings.Index(t, "("); i >= 0 {
		return strings.Trim(t[0:i], " ")
	}
	return t
}

// Ngrams finds the set of all substrings in the array longer than minLen characters
func Ngrams(chars []string, minLen int) []string {
	// log.Printf("ngrams, chars %v", chars)
	if len(chars) < 2 {
		return []string{}
	}
	ss := []string{}
	for i := range chars {
		for j := len(chars); j > 1; j-- {
			if i < j {
				x := chars[i:j]
				w := strings.Join(x, "")
				if len(x) >= minLen {
					// log.Printf("ngrams i=%d: j=%d, w=%s\n", i, j, w)
					ss = append(ss, w)
				}
			}
		}
	}
	// log.Printf("ngrams, ss %v\n", ss)
	return ss
}
