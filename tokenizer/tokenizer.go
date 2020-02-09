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
// Package for tokenizer of Chinese text
//

package tokenizer


import (
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"log"
	"strings"
	"unicode/utf8"
)

// Tokenizes Chinese text
type Tokenizer interface {
	Tokenize(fragment string) []TextToken
}

// Tokenizes Chinese text using a dictionary
type DictTokenizer struct{WDict map[string]dicttypes.Word}

// A text token contains the results of tokenizing a string
type TextToken struct{
	Token string
	DictEntry dicttypes.Word
	Senses []dicttypes.WordSense
}

// Tokenizes a Chinese text string into words and other terms in the dictionary.
// If the terms are not found in the dictionary then individual characters will
// be returned. Compares left to right and right to left greedy methods, taking
// the one with the least tokens.
func (tokenizer DictTokenizer) Tokenize(fragment string) []TextToken {
	tokens1 := tokenizer.greedyLtoR(fragment)
	tokens2 := tokenizer.greedyRtoL(fragment)
	if len(tokens1) < len(tokens2) {
		return tokens1
	}
	return tokens2
}

// Tokenizes text with a greedy knapsack-like algorithm, scanning left to
// right.
func (tokenizer DictTokenizer) greedyLtoR(fragment string) []TextToken {
	tokens := []TextToken{}
	if len(fragment) == 0 {
		return tokens
	}
	characters := strings.Split(fragment, "")
	for i := 0; i < len(characters); i++ {
		for j := len(characters); j > 0; j-- {
			w := strings.Join(characters[i:j], "")
			//log.Printf("greedyLtoR: w = %s\n", w)
			if entry, ok := tokenizer.WDict[w]; ok {
				token := TextToken{w, entry, []dicttypes.WordSense{}}
				tokens = append(tokens, token)
				i = j - 1
				j = 0
			} else if utf8.RuneCountInString(w) == 1 {
				log.Printf("greedyLtoR: found unknown character %s\n", w)
				token := TextToken{}
				token.Token = w
				tokens = append(tokens, token)
				break
			}
		}
	}
	return tokens
}

// Tokenizes text with a greedy knapsack-like algorithm, scanning right to
// left.
func (tokenizer DictTokenizer) greedyRtoL(fragment string) []TextToken {
	tokens := []TextToken{}
	if len(fragment) == 0 {
		return tokens
	}
	characters := strings.Split(fragment, "")
	for i := len(characters); i > 0; i-- {
		for j := 0; j < len(characters); j++ {
			if i < j {
				break
			}
			w := strings.Join(characters[j:i], "")
			//log.Printf("greedyRtoL: i, j, w = %d, %d, %s\n", i, j, w)
			if entry, ok := tokenizer.WDict[w]; ok {
				token := TextToken{w, entry, []dicttypes.WordSense{}}
				tokens = append([]TextToken{token}, tokens...)
				i = j
				j = -1
			} else if utf8.RuneCountInString(w) == 1 {
				//log.Printf("greedyRtoL: found unknown character %s\n", w)
				token := TextToken{}
				token.Token = w
				tokens = append([]TextToken{token}, tokens...)
				break
			}
		}
	}
	return tokens
}
