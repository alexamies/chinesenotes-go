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

// Functions for parsing a search query
package find

import (
	"unicode"

	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
)

// Parses input queries into a slice of text segments
type QueryParser interface {
	ParseQuery(query string) []TextSegment
}

type DictQueryParser struct{
	Tokenizer *tokenizer.DictTokenizer[*dicttypes.Word]
}

// A text segment contains the QueryText searched for and possibly a matching
// dictionary entry. There will only be matching dictionary entries for
// Chinese words in the dictionary. Non-Chinese text, punctuation, and unknown
// Chinese words will have nil DictEntry values and matching values will be
// included in the Senses field.
type TextSegment struct {
	QueryText string
	DictEntry dicttypes.Word
	Senses    []dicttypes.WordSense
}

// Creates a QueryParser
func NewQueryParser(dict map[string]*dicttypes.Word) QueryParser {
	tokenizer := tokenizer.NewDictTokenizer(dict)
	return DictQueryParser{tokenizer}
}

// The method for parsing the query text in this function is based on dictionary
// lookups
func (parser DictQueryParser) ParseQuery(query string) []TextSegment {
	return parser.get_chunks(query)
}

// Segments the text string into chunks that are CJK and non-CJK or puncuation
func (parser DictQueryParser) get_chunks(text string) []TextSegment {
	chunks := []TextSegment{}
	cjk := ""
	noncjk := ""
	for _, character := range text {
		if is_cjk(character) {
			if noncjk != "" {
				seg := TextSegment{}
				seg.QueryText = noncjk
				chunks = append(chunks, seg)
				noncjk = ""
			}
			cjk += string(character)
		} else if cjk != "" {
			segments := parser.parse_chinese(cjk)
			chunks = append(chunks, segments...)
			cjk = ""
			noncjk += string(character)
		} else {
			noncjk += string(character)
		}
	}
	if cjk != "" {
		segments := parser.parse_chinese(cjk)
		chunks = append(chunks, segments...)
	}
	if noncjk != "" {
		seg := TextSegment{}
		seg.QueryText = noncjk
		chunks = append(chunks, seg)
	}
	return chunks
}

// Tests whether the symbol is a CJK character, excluding punctuation
// Only looks at the first charater in the string
func is_cjk(r rune) bool {
	return unicode.Is(unicode.Han, r) && !unicode.IsPunct(r)
}

// Segments Chinese text based on dictionary entries
func (parser DictQueryParser) parse_chinese(text string) []TextSegment {
	tokens := parser.Tokenizer.Tokenize(text)
	terms := []TextSegment{}
	for _, token := range tokens {
		seg := TextSegment{token.Token, token.DictEntry, token.Senses}
		terms = append(terms, seg)
	}
	return terms
}

func toQueryTerms(terms []TextSegment) []string {
	queryTerms := []string{}
	for _, term := range terms {
		queryTerms = append(queryTerms, term.QueryText)
	}
	return queryTerms
}
