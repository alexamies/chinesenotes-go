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
	pinyinWeight float64 = 0.5
	notesWeight float64 = 1.0
	uniCountWeight float64 = 1.0
	hammingWeight float64 = -1.0
	substringWeight float64 = 5.0
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
	hasPinyin int
	inNotes int
	isSubstring int
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
		domain string, wdict map[string]dicttypes.Word) ([]*tmResult, error) {
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
	var resSlice []*tmResult
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
		resSlice = append(resSlice, &result)
	}
	log.Printf("queryPinyin, num results: %d\n", len(resSlice))
	return resSlice, nil
}

// Search the trans memory for words containing the given unigrams
func (searcher *Searcher) queryUnigram(ctx context.Context, chars []string,
		domain string) ([]*tmResult, error) {
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
	var resSlice []*tmResult
	for results.Next() {
		var result tmResult
		err = results.Scan(&result.term, &result.unigramCount)
		if err != nil {
			return nil, fmt.Errorf("queryUnigram, Error for scanning results:\n%v", err)
		}
		resSlice = append(resSlice, &result)
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

// Adds two sets of matches with no dups, including simplified vs trad dups
func addMatches(matches1, matches2 []*tmResult,
		wdict map[string]dicttypes.Word) []*tmResult {
	matchMap := make(map[int]*tmResult)
	var mDups []*tmResult
	for _, m := range matches1 {
		mDups = append(mDups, m)
	}
	for _, m := range matches2 {
		mDups = append(mDups, m)
	}
	for _, m := range mDups {
		if w, ok := wdict[m.term]; ok {
			hwId := w.HeadwordId
			if m2, ok := matchMap[hwId]; ok {
				if m.combinedScore > m2.combinedScore {
					matchMap[hwId] = m
				}
				continue
			}
			matchMap[hwId] = m
		}
	}
	var matches []*tmResult
	for _, m := range matchMap {
		matches = append(matches, m)
	}
	return matches
}

// Combines matches with dictionary defintions to send back to client
func combineResults(query string,
		matches, pinyinMatches []*tmResult,
		wdict map[string]dicttypes.Word) []dicttypes.Word {
	fillHamming(query, matches)
	fillHamming(query, pinyinMatches)
	fillSubstring(query, matches)
	fillSubstring(query, pinyinMatches)
	for i := range matches {
		matches[i].combinedScore = combineScores(query, matches[i])
	}
	for i := range pinyinMatches {
		pinyinMatches[i].combinedScore = combineScores(query, pinyinMatches[i])
	}
	matches = addMatches(matches, pinyinMatches, wdict)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].combinedScore > matches[j].combinedScore
	})
	var words []dicttypes.Word
	for _, match := range matches {
		if word, ok := wdict[match.term]; ok {
			words = append(words, word)
		}
	}
	printTopResults(query, matches)
	return words
}

// Compute combined score for result
func combineScores(query string, match *tmResult) float64 {
	l := len([]rune(query))
	if l == 0 {
		return float64(100)
	}
	normalUni := float64(match.unigramCount) / float64(l)
	normalHamming := float64(match.hamming) / float64(l)
	return normalUni * uniCountWeight +
			normalHamming * hammingWeight +
			float64(match.hasPinyin) * pinyinWeight +
			float64(match.isSubstring) * substringWeight +
			float64(match.inNotes) * notesWeight
}

// Fill in hamming distance for match results
func fillHamming(query string, matches []*tmResult) {
	for _, match := range matches {
		match.hamming = hammingDist(query, match.term)
	}
}

// Fill in hamming distance for match results
func fillSubstring(query string, matches []*tmResult) {
	for _, match := range matches {
		match.isSubstring = eitherSubstring(query, match.term)
	}
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
func printTopResults(query string, matches []*tmResult) {
	log.Printf("transmemory.printTopResults, query: %s" +
			", top results", query)
	if len(matches) == 0 {
		log.Printf("transmemory.Search no results")
		return
	}
	var sb strings.Builder
	sb.WriteString("\nTerm, Has Pinyin, In Notes, Unigram count, Hamming, " +
			"Substring, Combined\n")
	for i := 0; i < 10; i++ {
		if i == len(matches) {
			break
		}
		m := fmt.Sprintf("%d: %s, %d, %d, %d, %d, %d, %f\n", i, matches[i].term,
			matches[i].hasPinyin, matches[i].inNotes, 
			matches[i].unigramCount, matches[i].hamming, matches[i].isSubstring,
			matches[i].combinedScore)
		sb.WriteString(m)
	}
	log.Printf("transmemory.printTopResults, matchs:\n%s\n", sb.String())
}
