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
// Package for looking up words and multiword expressions.
//
package dictionary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)


// Encapsulates term lookup recults
type Results struct {
	Words []dicttypes.Word
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

func initSubtrQuery(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT simplified, traditional, pinyin, english, notes, headword 
FROM words 
WHERE
  (simplified LIKE ? OR traditional LIKE ?)
  AND 
  (topic_en = ? OR parent_en = ?)
LIMIT 100`)
}

// Lookup a term based on a substring and a topic
func (searcher *Searcher) LookupSubstr(ctx context.Context,
		query, topic_en, subtopic_en string) (*Results, error) {
	if query == "" {
		return nil, errors.New("Query string is empty")
	}
	log.Printf("LookupSubstr, query %s, topic = %s", query, topic_en)
	likeTerm := "%" + query + "%"
	results, err := searcher.findSubstrStmt.QueryContext(ctx, likeTerm, likeTerm,
		topic_en, subtopic_en)
	if err != nil {
		return nil, fmt.Errorf("LookupSubstr, Error for query %s: %v", query, err)
	}
	wmap := map[string]dicttypes.Word{}
	for results.Next() {
		ws := dicttypes.WordSense{}
		var hw sql.NullInt64
		var trad, pinyin, english, notes sql.NullString
		results.Scan(&ws.Simplified, &trad, &pinyin, &english, &notes, &hw)
		if trad.Valid {
			ws.Traditional = trad.String
		}
		if pinyin.Valid {
			ws.Pinyin = pinyin.String
		}
		if english.Valid {
			ws.English = english.String
		}
		if notes.Valid {
			ws.Notes = notes.String
		}
		if hw.Valid {
			ws.HeadwordId = int(hw.Int64)
		}
		addWordSense2Map(wmap, ws)
	}
	log.Printf("LookupSubstr, len(wmap): %d", len(wmap))
	words := wordMap2Array(wmap)
	return &Results{words}, nil
}

func wordMap2Array(wmap map[string]dicttypes.Word) []dicttypes.Word {
	words := []dicttypes.Word{}
	for _, w := range wmap {
		words = append(words, w)
	}
	return words
}
