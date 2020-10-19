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
// Chinese-English dictionary type definitions
package dicttypes

import (
	"strings"
)

// A top level word structure that may include multiple word senses
type Word struct {
	Simplified, Traditional, Pinyin string
	HeadwordId int
	Senses []WordSense
}

// Defines a single sense of a Chinese word
type WordSense struct {
	Id, HeadwordId int
	Simplified, Traditional, Pinyin, English, Grammar, Concept, ConceptCN, Domain,
			DomainCN, Subdomain, SubdomainCN, Image, MP3, Notes string
}

// May be sorted into descending order with most frequent bigram first
type Words []Word

// May be sorted into descending order with most frequent bigram first
type WordSenses []WordSense

// DictionaryConfig encapsulates parameters for dictionary configuration
type DictionaryConfig struct {
	AvoidSubDomains map[string]bool
	DictionaryDir string
}

// Clones the headword definition without the attached array of word senses
func CloneWord(w Word) Word {
	return Word{
		HeadwordId: w.HeadwordId,
		Simplified: w.Simplified,
		Traditional: w.Traditional,
		Pinyin: w.Pinyin,
		Senses: w.Senses,
	}
}

// Tests whether the word is a function word
func (ws *WordSense) IsFunctionWord() bool {
	functionPOS := map[string]bool{
		"adverb": true,
		"conjunction": true,
		"interjection": true,
		"interrogative pronoun": true,
		"measure word": true,
		"particle": true,
		"prefix": true,
		"preposition": true,
		"pronoun": true,
		"suffix": true,
	}	
	functionalWords := map[string]bool{
		"有": true,
		"是": true,
		"诸": true,
		"故": true,
		"出": true,
		"当": true,
		"若": true,
		"如": true,
	}
	return functionalWords[ws.Simplified] || functionPOS[ws.Grammar]
}

// IsProperNoun tests whether the word is a proper noun.
// If the majority of word senses are proper nouns, then the word is marked
// as a proper noun.
func (w Word) IsProperNoun() bool {
	count := 0
	for _, ws := range w.Senses {
		if ws.Grammar == "proper noun" {
			count++
		}
	}
	return float64(count) / float64(len(w.Senses)) > 0.5
}

func (hwArr Words) Len() int {
	return len(hwArr)
}

func (hwArr Words) Swap(i, j int) {
	hwArr[i], hwArr[j] = hwArr[j], hwArr[i]
}

func (hwArr Words) Less(i, j int) bool {
	noTones1 := normalizePinyin(hwArr[i].Pinyin)
	noTones2 := normalizePinyin(hwArr[j].Pinyin)
	return noTones1 < noTones2
}

func (senses WordSenses) Len() int {
	return len(senses)
}

func (senses WordSenses) Swap(i, j int) {
	senses[i], senses[j] = senses[j], senses[i]
}

func (senses WordSenses) Less(i, j int) bool {
	noTones1 := normalizePinyin(senses[i].Pinyin)
	noTones2 := normalizePinyin(senses[j].Pinyin)
	return noTones1 < noTones2
}

// Removes the tone diacritics from a Pinyin string
func normalizePinyin(pinyin string) string {
	runes := []rune{}
	for _, r := range pinyin {
		n, ok := NORMAL[r]
		if ok {
			runes = append(runes, n)
		} else {
			runes = append(runes, r)
		}
	}
	return strings.ToLower(string(runes))
}

var NORMAL = map[rune]rune{
	'ā': 'a',
	'á': 'a',
	'ǎ': 'a',
	'à': 'a',
	'ē': 'e',
	'é': 'e',
	'ě': 'e',
	'è': 'e',
	'ī': 'i',
	'í': 'i',
	'ǐ': 'i',
	'ì': 'i',
	'ō': 'o',
	'ó': 'o',
	'ǒ': 'o',
	'ò': 'o',
	'ū': 'u',
	'ú': 'u',
	'ǔ': 'u',
	'ù': 'u',
	'Ā': 'a',
	'Á': 'a',
	'Ǎ': 'a',
	'À': 'a',
	'Ē': 'e',
	'É': 'e',
	'Ě': 'e',
	'È': 'e',
	'Ī': 'i',
	'Í': 'i',
	'Ǐ': 'i',
	'Ì': 'i',
	'Ō': 'o',
	'Ó': 'o',
	'Ǒ': 'o',
	'Ò': 'o',
	'Ū': 'u',
	'Ú': 'u',
	'Ǔ': 'u',
	'Ù': 'u',
}
