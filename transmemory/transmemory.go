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
// Package for translation memory search
//

package transmemory

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

const (
	MAX_UNIGRAM = 8
)

// Encapsulates search recults
type Results struct {
	Words []dicttypes.Word
}

// Encapsulates translation memory searcher
type Searcher struct {
	database *sql.DB
	searchUnigramStmt *sql.Stmt
}

// Initialize SQL statement
func NewSearcher(ctx context.Context, database *sql.DB) (*Searcher, error) {
	unigramStmt, err := initUnigramStmt(ctx, database)
	if err != nil {
		return nil, err
	}
	return &Searcher{
		database: database,
		searchUnigramStmt: unigramStmt,
	}, nil
}

func initUnigramStmt(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx,
`SELECT
  word,
  count(*) as count
FROM tmindex_unigram
WHERE
  (ch = ? OR
  ch = ? OR
  ch = ? OR
  ch = ? OR
  ch = ? OR
  ch = ? OR
  ch = ? OR
  ch = ?)
  AND
  domain LIKE ?
GROUP BY word
ORDER BY count DESC LIMIT 50`)
}

// Search the trans memory for words containing the given unigrams
func (searcher *Searcher) queryUnigram(ctx context.Context, chars []string) ([]string, error) {
	var results *sql.Rows
	var err error
	results, err = searcher.searchUnigramStmt.QueryContext(ctx, chars[0], chars[1],
			chars[2], chars[3], chars[4], chars[5], chars[6], chars[7])
	if err != nil {
		applog.Error("queryUnigram, Error for query: ", err)
		return nil, err
	}
	var resSlice []string
	for results.Next() {
		var word string
		var count int
		err = results.Scan(&word, &count)
		if err != nil {
			applog.Error("queryUnigram, Error for scanning results: ", err)
			return nil, err
		}
		resSlice = append(resSlice, word)
	}
	applog.Info("queryUnigram, num results: ", len(resSlice))
	return resSlice, nil
}

// Searches the translation memory for approximate matches.
// Parameters
//   ctx Request context
//   query The search query
//   domain The domain to restrict the query to (optional)
//   wdict The full dictionary
// Retuns
//   A slice of approximate results
func (searcher *Searcher) Search(ctx context.Context,
		query string,
		domain string,
		wdict map[string]dicttypes.Word) (*Results, error) {
	chars := getChars(query)
	matches, err := searcher.queryUnigram(ctx, chars)
	if err != nil {
		return nil, fmt.Errorf("Search query error: %v", err)
	}
	words := combineResults(matches, wdict)
	return &Results{words}, nil
}

// Combines matches with dictionary defintions to send back to client
func combineResults(matches []string, wdict map[string]dicttypes.Word) []dicttypes.Word {
	var words []dicttypes.Word
	for _, match := range matches {
		if word, ok := wdict[match]; ok {
			words = append(words, word)
		}
	}
	return words
}

// Get the characters in the search query, padding to MAX_UNIGRAM with the
// last character
func getChars(query string) []string {
	runes := []rune(query)
	var chars []string
	for i := 0; i < MAX_UNIGRAM; i++ {
		if i < len(runes) {
			chars= append(chars, string(runes[i]))
			continue
		}
		chars= append(chars, string(runes[len(runes) - 1]))
	}
	return chars
}
