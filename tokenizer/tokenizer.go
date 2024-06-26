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

// Package for tokenizing of Chinese text into multi-character terms and
// corresponding English equivalents.
package tokenizer

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Tokenizes Chinese text
type Tokenizer interface {
	Tokenize(fragment string) []TextToken
}

// Tokenizes Chinese text using a dictionary
type DictTokenizer[V any] struct{
	wDict map[string]V
}

func NewDictTokenizer[V any](wDict map[string]V) *DictTokenizer[V] {
	tokenizer := DictTokenizer[V]{
		wDict: wDict,
	}
	return &tokenizer
}

// A text token contains the results of tokenizing a string
type TextToken struct{
	Token string
	DictEntry dicttypes.Word
	Senses []dicttypes.WordSense
}

func newTextToken(token string, v interface{}) TextToken {
	if s, ok := v.(*dicttypes.Word); ok {
		return TextToken{
			Token: token,
			DictEntry: *s,
		}
	}
	return TextToken{
		Token: token,
	}
}

// Tokenizes a Chinese text string into words and other terms in the dictionary.
// If the terms are not found in the dictionary then individual characters will
// be returned. Compares left to right and right to left greedy methods, taking
// the one with the least tokens.
// Long text is handled by breaking the string into segments delimited by
// punctuation or non-Chinese characters.
func (tokenizer DictTokenizer[V]) Tokenize(text string) []TextToken {
	tokens := []TextToken{}
	segments := Segment(text)
	for _, segment := range segments {
		if segment.Chinese {
			tokens1 := tokenizer.greedyLtoR(segment.Text)
			tokens2 := tokenizer.greedyRtoL(segment.Text)
			if len(tokens2) < len(tokens1) {
				tokens = append(tokens, tokens2...)
			} else {
				tokens = append(tokens, tokens1...)
			}
		} else {
			token := TextToken{
				Token: segment.Text,
			}
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// term looks up either a simple string of the full dictionary term
func term[V any](dict map[string]V, w string) (V, bool) {
	v, ok := dict[w]
	return v, ok
}

// Tokenizes text with a greedy knapsack-like algorithm, scanning left to
// right.
func (tokenizer DictTokenizer[V]) greedyLtoR(fragment string) []TextToken {
	tokens := []TextToken{}
	if len(fragment) == 0 {
		return tokens
	}
	characters := strings.Split(fragment, "")
	for i := 0; i < len(characters); i++ {
		for j := len(characters); j > 0; j-- {
			w := strings.Join(characters[i:j], "")
			//log.Printf("greedyLtoR: w = %s\n", w)
			if entry, ok := term(tokenizer.wDict, w); ok {
				token := newTextToken(w, entry)
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
func (tokenizer DictTokenizer[V]) greedyRtoL(fragment string) []TextToken {
	tokens := []TextToken{}
	if len(fragment) == 0 {
		return tokens
	}
	characters := strings.Split(fragment, "")
	for i := len(characters); i > 0; i-- {
		for j := 0; j < i; j++ {
			w := strings.Join(characters[j:i], "")
			//log.Printf("greedyRtoL: i, j, w = %d, %d, %s\n", i, j, w)
			if entry, ok := tokenizer.wDict[w]; ok {
				token := newTextToken(w, entry)
				tokens = append([]TextToken{token}, tokens...)
				i = j + 1
				break
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
