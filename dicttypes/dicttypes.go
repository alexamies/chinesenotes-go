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
//
package dicttypes

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

// IsProperNoun tests whether the word is a proper noun.
// If the majority of word senses are proper nouns, then the word is marked
// as a proper noun.
func IsProperNoun(w *Word) bool {
	count := 0
	for _, ws := range w.Senses {
		if ws.Grammar == "proper noun" {
			count++
		}
	}
	return float64(count) / float64(len(w.Senses)) > 0.5
}
