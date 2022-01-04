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
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Dictionary is a struct to hold word dictionary indexes
type Dictionary struct {
	// Forward dictionary, lookup by Chinese word
	Wdict map[string]*dicttypes.Word
	HeadwordIds map[int]*dicttypes.Word
}

func NewDictionary(wdict map[string]*dicttypes.Word) *Dictionary {
	hwIdMap := make(map[int]*dicttypes.Word)
	for _, w := range wdict {
		hwIdMap[w.HeadwordId] = w
	}
	return &Dictionary{
		Wdict: wdict,
		HeadwordIds: hwIdMap,
	}
}

// Searcher looks up Chinese words by either Chinese or English.
// 
// If the dictionary searcher cannot connect to the database then
// it will run in degraded mode by looking up Chinese words form dictionary
// file.
type Searcher struct {
	database *sql.DB
	findEnglishStmt *sql.Stmt
	findSubstrStmt *sql.Stmt
	initialized bool
}

// NewSearcher initialize SQL statements
func NewSearcher(ctx context.Context, database *sql.DB) *Searcher {
	s := Searcher{}
	if database != nil {
		var err error
		s.findEnglishStmt, err = initEnglishQuery(ctx, database)
		if err != nil {
			log.Printf("NewSearcher, database statement initializaton error %v", err)
			return &s
		}
		s.findSubstrStmt, err = initSubtrQuery(ctx, database)
		if err != nil {
			log.Printf("NewSearcher, substr query initializaton error \n%v\n", err)
			return &s
		}
	}
	s.initialized = true
	return &s
}

func initEnglishQuery(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT simplified, traditional, pinyin, english, notes, headword
FROM words
WHERE pinyin = ? OR english LIKE ?
LIMIT 20`)
}

// Initialized returns true if there were no error in initialization.
func (s *Searcher) Initialized() bool {
	return s.initialized
}

// FindWordsByEnglish returns the word senses with English approximate or Pinyin exact match
func (searcher *Searcher) FindWordsByEnglish(ctx context.Context,
		query string) ([]dicttypes.WordSense, error) {
	log.Printf("findWordsByEnglish, query = %s", query)
	likeEnglish := "%" + query + "%"
	if searcher.findEnglishStmt == nil {
		return nil, fmt.Errorf("FindWordsByEnglish,findEnglishStmt is nil query = %s",
			query)
	}
	results, err := searcher.findEnglishStmt.QueryContext(ctx, query, likeEnglish)
	if err != nil {
		log.Printf("FindWordsByEnglish, Error for query: %s, error %v", query, err)
		// Retry
		results, err = searcher.findEnglishStmt.QueryContext(ctx, query, query)
		if err != nil {
			log.Printf("FindWordsByEnglish, Give up after retry: %s, error: %v", query, err)
			return nil, err
		}
	}
	senses := []dicttypes.WordSense{}
	for results.Next() {
		ws := dicttypes.WordSense{}
		var hw sql.NullInt64
		var trad, pinyin, english, notes sql.NullString
		results.Scan(&ws.Simplified, &trad, &pinyin, &english, &notes, &hw)
		log.Printf("FindWordsByEnglish, simplified, headword = %s, %v",
			ws.Simplified, hw)
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
		senses = append(senses, ws)
	}
	log.Printf("FindWordsByEnglish, len(senses): %d", len(senses))
	return senses, nil
}
