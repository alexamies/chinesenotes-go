package dictionary

//
// Package for looking up words and multiword expressions.
//

import (
	"context"
	"fmt"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)


// Encapsulates term lookup recults
type Results struct {
	Words []dicttypes.Word
}


// SubstringIndexMem looks up substrings from a map loaded from a file.
type SubstringIndexMem struct {
}

// NewSubstringIndexMem initialize a SubstringIndexMem
func NewSubstringIndexMem(ctx context.Context) (SubstringIndex, error) {
	return &SubstringIndexMem{}, nil
}

// Lookup a term based on a substring and a topic
func (searcher SubstringIndexMem) LookupSubstr(ctx context.Context, query, topic_en, subtopic_en string) (*Results, error) {
	if query == "" {
		return nil, fmt.Errorf("query string is empty")
	}
	return &Results{}, nil
}

// Used for grouping word senses by similar headwords in result sets
func addWordSense2Map(wmap map[string]dicttypes.Word, ws dicttypes.WordSense) {
	word, ok := wmap[ws.Simplified]
	if ok {
		word.Senses = append(word.Senses, ws)
		wmap[word.Simplified] = word
	} else {
		word = dicttypes.Word{}
		word.Simplified = ws.Simplified
		word.Traditional = ws.Traditional
		word.Pinyin = ws.Pinyin
		word.HeadwordId = ws.HeadwordId
		word.Senses = []dicttypes.WordSense{ws}
		wmap[word.Simplified] = word
	}
}

func wordMap2Array(wmap map[string]dicttypes.Word) []dicttypes.Word {
	words := []dicttypes.Word{}
	for _, w := range wmap {
		words = append(words, w)
	}
	return words
}
