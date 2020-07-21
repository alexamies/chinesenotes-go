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
	"sort"
)

const (
	MAX_UNIGRAM = 8
	UNI_COUNT_WEIGHT float64 = 1.0
	HAMMING_WEIGHT float64 = -1.0
)

// Encapsulates search recults
type Results struct {
	Words []dicttypes.Word
}

// Encapsulates search recults
type tmResult struct {
	term string
	unigramCount int
	hamming int
	combinedScore float64
}

// Encapsulates translation memory searcher
type Searcher struct {
	database *sql.DB
	unigramStmt *sql.Stmt
	uniDomainStmt *sql.Stmt
}

// Initialize SQL statement
func NewSearcher(ctx context.Context, database *sql.DB) (*Searcher, error) {
	unigramStmt, err := initUnigramStmt(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("NewSearcher: unable to prepare unigramStmt: %v", err)
	}
	uniDomainStmt, err := initUniDomainStmt(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("NewSearcher: unable to prepare uniDomainStmt: %v", err)
	}
	return &Searcher{
		database: database,
		unigramStmt: unigramStmt,
		uniDomainStmt: uniDomainStmt,
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
GROUP BY word
ORDER BY count DESC LIMIT 50`)
}

func initUniDomainStmt(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
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
func (searcher *Searcher) queryUnigram(ctx context.Context, chars []string,
		domain string) ([]tmResult, error) {
	var results *sql.Rows
	var err error
	if len(domain) == 0 {
		results, err = searcher.unigramStmt.QueryContext(ctx, chars[0], chars[1],
				chars[2], chars[3], chars[4], chars[5], chars[6], chars[7])
	} else {
		results, err = searcher.uniDomainStmt.QueryContext(ctx, chars[0], chars[1],
				chars[2], chars[3], chars[4], chars[5], chars[6], chars[7], domain)
	}
	if err != nil {
		return nil, fmt.Errorf("queryUnigram, Error for query: %v", err)
	}
	var resSlice []tmResult
	for results.Next() {
		var result tmResult
		err = results.Scan(&result.term, &result.unigramCount)
		if err != nil {
			return nil, fmt.Errorf("queryUnigram, Error for scanning results: %v", err)
		}
		resSlice = append(resSlice, result)
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
	matches, err := searcher.queryUnigram(ctx, chars, domain)
	if err != nil {
		return nil, fmt.Errorf("Search query error: %v", err)
	}
	words := combineResults(query, matches, wdict)
	return &Results{words}, nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Combines matches with dictionary defintions to send back to client
func combineResults(query string,
		matches []tmResult,
		wdict map[string]dicttypes.Word) []dicttypes.Word {
	for i := range matches {
		matches[i].combinedScore = combineScores(query, matches[i])
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].combinedScore > matches[j].combinedScore
	})
	var words []dicttypes.Word
	for _, match := range matches {
		if word, ok := wdict[match.term]; ok {
			words = append(words, word)
		}
	}
	return words
}

// Compute combined score for result
func combineScores(query string, match tmResult) float64 {
	l := len([]rune(query))
	if l == 0 {
		return float64(100)
	}
	normalUni := float64(match.unigramCount) / float64(l)
	normalHamming := float64(match.hamming) / float64(l)
	return normalUni * UNI_COUNT_WEIGHT + normalHamming * HAMMING_WEIGHT
}

// Fill in hamming distance for match results
func fillHamming(query string, matches []tmResult) {
	for _, match := range matches {
		match.hamming = hammingDist(query, match.term)
	}
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

// Compute hamming distance based on similar characters
func hammingDist(query, term string) int {
	hamming := 0
	rQuery := []rune(query)
	rTerm := []rune(term)
	for i := 0; i < len(rQuery); i++ {
		if i < len(rTerm) {
			if rQuery[i] != rTerm[i] {
				hamming += 1
			}
			continue
		}
		break
	}
	hamming += absInt(len(rQuery) - len(rTerm))
	return hamming
}
