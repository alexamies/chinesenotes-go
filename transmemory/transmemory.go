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
	"github.com/alexamies/chinesenotes-go/webconfig"
)

const (
	MAX_UNIGRAM = 8
)

var (
	database *sql.DB
	searchUnigramStmt *sql.Stmt
)

// Encapsulates search recults
type Results struct {
	Words []dicttypes.Word
}

// Initialize SQL statement
func init() {
	ctx := context.Background()
	err := initSQL(ctx)
	if err != nil {
		applog.Error("init, unable to intiialize SQL statement: ", err)
	}
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

func initSQL(ctx context.Context) error {
	conString := webconfig.DBConfig()
	db, err := sql.Open("mysql", conString)
	if err != nil {
		return err
	}
	database = db

	searchUnigramStmt, err = database.PrepareContext(ctx,
`SELECT
  word,
  count(*) as count
FROM tmindex_unigram
WHERE 
  ch = '?' OR
  ch = '?' OR
  ch = '?' OR
  ch = '?' OR
  ch = '?' OR
  ch = '?' OR
  ch = '?' OR
  ch = '?'
GROUP BY word
ORDER BY count DESC LIMIT 50`)
  if err != nil {
  	applog.Error("transmemory.initStatements() searchUnigramStmt err: ", err)
    applog.Info("transmemory.initStatements() conString: ", conString)
    return err
  }
  return nil
}

// Search the trans memory for words containing the given unigrams
func queryUnigram(ctx context.Context, chars []string) ([]string, error) {
	applog.Info("queryUnigram, terms = ", chars)
	if searchUnigramStmt == nil {
		applog.Error("queryUnigram, searchUnigramStmt == nil")
		// Re-initialize
		err := initSQL(ctx)
		if err != nil {
			applog.Error("queryUnigram, unable to intiialize SQL statement: ", err)
		  return []string{}, err
		}
	}
	var results *sql.Rows
	var err error
	results, err = searchUnigramStmt.QueryContext(ctx, chars[0], chars[1],
			chars[2], chars[3], chars[4], chars[5], chars[6], chars[7])
	if err != nil {
		applog.Error("queryUnigram, Error for query: ", chars, err)
		return []string{}, err
	}
	resSlice := []string{}
	for results.Next() {
		var word string
		results.Scan(&word)
		resSlice = append(resSlice, word)
	}
	return resSlice, nil
}

// Get the characters in the search query, padding to MAX_UNIGRAM with the
// last character
func getChars(query string) []string {
	var chars []string
	for i := 0; i < MAX_UNIGRAM; i++ {
		if i < len(query) {
			chars= append(chars, string(query[i]))
			continue
		}
		chars= append(chars, string(query[len(query) - 1]))
	}
	return chars
}

// Searches the translation memory for approximate matches.
// Parameters
//   ctx Request context
//   query The search query
//   wdict The full dictionary
// Retuns
//   A slice of approximate results
func Search(ctx context.Context,
		query string,
		wdict map[string]dicttypes.Word) (*Results, error) {
	chars := getChars(query)
	matches, err := queryUnigram(ctx, chars)
	if err != nil {
		return nil, fmt.Errorf("Search query error: %v", err)
	}
	words := combineResults(matches, wdict)
	return &Results{words}, nil
}
