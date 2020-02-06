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
	Tokenize(fragment string) []string
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
// be returned.
func (tokenizer DictTokenizer) Tokenize(fragment string) []TextToken {
	tokens := []TextToken{}
	if len(fragment) == 0 {
		return tokens
	}
	characters := strings.Split(fragment, "")
	for i := 0; i < len(characters); i++ {
		for j := len(characters); j > 0; j-- {
			w := strings.Join(characters[i:j], "")
			if entry, ok := tokenizer.WDict[w]; ok {
				token := TextToken{w, entry, []dicttypes.WordSense{}}
				tokens = append(tokens, token)
				i = j - 1
				j = 0
			} else if utf8.RuneCountInString(w) == 1 {
				log.Printf("Tokenize: found unknown character %s\n", w)
				token := TextToken{}
				token.Token = w
				tokens = append(tokens, token)
				break
			}
		}
	}

	return tokens
}
