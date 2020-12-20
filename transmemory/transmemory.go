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

// Package for translation memory search
package transmemory

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

const (
	maxUnigram = 8
	// Decision point from decision tree training for Unicode count divided by query length
	uniCountDP float64 = 0.37
	// Decision point from decision tree training for Hamming distance divided by query length
	hammingDP float64 = 0.59
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
	hasPinyin int
	inNotes int
	isSubstring int
	relevant int
}

// Encapsulates translation memory searcher
type Searcher struct {
	database *sql.DB
	databaseInitialized bool
	pinyinStmt *sql.Stmt
	pinyinDomainStmt *sql.Stmt
	unigramStmt *sql.Stmt
	uniDomainStmt *sql.Stmt
}

// Initialize SQL statement
func NewSearcher(ctx context.Context, database *sql.DB) (*Searcher, error) {
	pinyinStmt, err := initPinyinStmt(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("NewSearcher: unable to prepare pinyinStmt:\n%v", err)
	}
	pinyinDomainStmt, err := initPinyinDomainStmt(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("NewSearcher: unable to prepare pinyinDomainStmt:\n%v", err)
	}
	unigramStmt, err := initUnigramStmt(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("NewSearcher: unable to prepare unigramStmt:\n%v", err)
	}
	uniDomainStmt, err := initUniDomainStmt(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("NewSearcher: unable to prepare uniDomainStmt:\n%v", err)
	}
	return &Searcher{
		database: database,
		databaseInitialized: true,
		pinyinStmt: pinyinStmt,
		pinyinDomainStmt: pinyinDomainStmt,
		unigramStmt: unigramStmt,
		uniDomainStmt: uniDomainStmt,
	}, nil
}

// Returns the word senses with English approximate or Pinyin exact match
func (searcher *Searcher) DatabaseInitialized() bool {
	return searcher.databaseInitialized
}

// Find words with similar pinyin or with notes conaining the query
func initPinyinStmt(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT DISTINCT simplified
FROM words
WHERE
  pinyin LIKE ? OR notes LIKE ?
LIMIT 20`)
}

// Find words with similar pinyin or with notes conaining the query
// for a given domain
func initPinyinDomainStmt(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT DISTINCT simplified
FROM words
WHERE
  (pinyin LIKE ? OR notes LIKE ?)
  AND
  (topic_en = ?)
LIMIT 20`)
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
FROM tmindex_uni_domain
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
  domain = ?
GROUP BY word
ORDER BY count DESC LIMIT 50`)
}

// Search the trans memory for words containing the given unigrams
func (searcher *Searcher) queryPinyin(ctx context.Context, query,
		domain string, wdict map[string]dicttypes.Word) ([]tmResult, error) {
	pinyin := findPinyin(query, wdict)
	if len(pinyin) == 0 {
		return nil, fmt.Errorf("queryPinyin, No pinyin for query,\n%s", query)
	}
	var results *sql.Rows
	var err error
	if len(domain) == 0 {
		results, err = searcher.pinyinStmt.QueryContext(ctx, pinyin, query)
	} else {
		results, err = searcher.pinyinDomainStmt.QueryContext(ctx, pinyin, query,
				domain)
	}
	if err != nil {
		return nil, fmt.Errorf("queryPinyin, Error for query, %s:\n%v", query, err)
	}
	var resSlice []tmResult
	for results.Next() {
		var result tmResult
		err = results.Scan(&result.term)
		if err != nil {
			return nil, fmt.Errorf("queryPinyin, Error for scanning results, %s:\n%v",
					query, err)
		}
		if strings.Contains(result.term, query) {
			result.inNotes = 1
		} else {
			result.hasPinyin = 1
		}
		resSlice = append(resSlice, result)
	}
	log.Printf("queryPinyin, num results: %d\n", len(resSlice))
	return resSlice, nil
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
		return nil, fmt.Errorf("queryUnigram, Error for query:\n%v", err)
	}
	var resSlice []tmResult
	for results.Next() {
		var result tmResult
		err = results.Scan(&result.term, &result.unigramCount)
		if err != nil {
			return nil, fmt.Errorf("queryUnigram, Error for scanning results:\n%v", err)
		}
		resSlice = append(resSlice, result)
	}
	log.Printf("queryUnigram, num results: %d\n", len(resSlice))
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
		includeSubstrings bool,
		wdict map[string]dicttypes.Word) (*Results, error) {
	chars := getChars(query)
	matches, err := searcher.queryUnigram(ctx, chars, domain)
	if err != nil {
		return nil, fmt.Errorf("Search query error:\n%v", err)
	}
	pinyinMatches, err := searcher.queryPinyin(ctx, query, domain, wdict)
	words := combineResults(query, matches, pinyinMatches, wdict)
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
		matches, pinyinMatches []tmResult,
		wdict map[string]dicttypes.Word) []dicttypes.Word {
	relevantMap := map[string]tmResult{}
	for _, m := range matches {
		m.hamming = hammingDist(query, m.term)
		m.isSubstring = eitherSubstring(query, m.term)
		if predictRelevance(query, m) {
			m.relevant = 1
		} else {
			m.relevant = 0
		}
		relevantMap[m.term] = m
	}
	for _, m := range pinyinMatches {
		m.hamming = hammingDist(query, m.term)
		m.isSubstring = eitherSubstring(query, m.term)
		if predictRelevance(query, m) {
			m.relevant = 1
		} else {
			m.relevant = 0
		}
		relevantMap[m.term] = m
	}
	allMatches := []tmResult{}
	for _, v := range relevantMap {
		allMatches = append(allMatches, v)
	}
	printTopResults(query, allMatches)
	relevantMatches := []tmResult{}
	for _, m := range allMatches {
		if m.relevant == 1 {
			relevantMatches = append(relevantMatches, m)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].unigramCount > matches[j].unigramCount
	})
	var words []dicttypes.Word
	for _, match := range relevantMatches {
		if word, ok := wdict[match.term]; ok {
			words = append(words, word)
		}
	}
	return words
}

// Predict relevance based on decision tree analysis
// Returns
//   bool - true if relevant, false if not relevant
func predictRelevance(query string, m tmResult) bool {
	l := len([]rune(query))
	if l == 0 {
		return false
	}
	if m.isSubstring == 1 || m.hasPinyin == 1 || m.inNotes == 1 {
		return true
	}
	normalUni := float64(m.unigramCount) / float64(l)
	normalHamming := float64(m.hamming) / float64(l)
	//log.Printf("transmemory.predictRelevance, query: %s, term: %s, " +
	//		"normalUni: %f, normalHamming: %f: ", query, m.term, normalUni,
	//		normalHamming)
  if normalUni >= uniCountDP && normalHamming <= hammingDP {
  	return true
  }
  return false
}

// Finds the pinyin for a given Chinese string
func findPinyin(query string, wdict map[string]dicttypes.Word) string {
	pinyin := ""
	for _, ch := range query {
		if word, ok := wdict[string(ch)]; ok {
			pinyin += word.Pinyin
		}
	}
	return pinyin
}

// Get the characters in the search query, padding to maxUnigram with the
// last character
func getChars(query string) []string {
	runes := []rune(query)
	var chars []string
	for i := 0; i < maxUnigram; i++ {
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

// Find whether eight string is a substring of the other, unless only a single
// character
// Return
//   1 if either is a substring
//   0 otherwise
func eitherSubstring(s1, s2 string) int {
	if len(s1) == 1 || len(s1) == 1 {
		return 0
	}
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		return 1
	}
	return 0
}

// Prints top search results
func printTopResults(query string, matches []tmResult) {
	if len(matches) == 0 {
		log.Printf("transmemory.Search no results")
		return
	}
	log.Printf("\nQuery, rank, Term, Has Pinyin, In Notes, Unigram count, " +
			"Hamming, Substring, Relevant\n")
	for i, m := range matches {
		if i == 10 {
			break
		}
		log.Printf("transmemory.printTopResults result: %s, %d, %s, %d, %d, %d, " +
				"%d, %d, %d\n", query, i, m.term, m.hasPinyin, m.inNotes,
				m.unigramCount, m.hamming, m.isSubstring, m.relevant)
	}
	log.Printf("transmemory.printTopResults, query: %s, matchs (%d): ",
			query, len(matches))
}
