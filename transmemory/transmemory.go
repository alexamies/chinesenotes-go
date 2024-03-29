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
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

const (
	maxUnigram             = 8
	maxResultsSubstrings   = 10
	maxResultsNoSubstrings = 3
	// Decision point from decision tree classification training for Unicode count / query len
	// divided by query length
	uniCountMin float64 = 0.37
	// Decision point from training for Hamming distance divided by query length / query len
	hammingDistMax float64 = 0.59
	// Decision point from training with substrings excluded for Unicode count (greater than or equal)
	uniCountNoSubMin int = 3
	// Decision point from training with substrings excluded for Hamming distance (less than or equal)
	hammingNoSubMax int = 8
)

// Encapsulates search recults
type Results struct {
	Words []dicttypes.Word
}

// Encapsulates search recults
type tmResult struct {
	term         string
	unigramCount int
	hamming      int
	hasPinyin    int
	inNotes      int
	isSubstring  int
	relevant     int
}

// Searcher finds similar phrases
type Searcher interface {

	// Search for phrases similar to the given query, optionally with a particular domain
	// Parameters
	//   ctx Request context
	//   query The search query
	//   domain The domain to restrict the query to (optional)
	//   wdict The full dictionary
	// Retuns
	//   A slice of approximate results
	Search(ctx context.Context, query string, domain string, includeSubstrings bool, wdict map[string]*dicttypes.Word) (*Results, error)
}

// pinyinSearcher finds similar phrases with matching Pinyin
type pinyinSearcher interface {

	// Search for phrases similar to the given query, optionally with a particular domain
	// Parameters
	//   ctx Request context
	//   query The search query
	//   domain The domain to restrict the query to (optional)
	//   wdict The full dictionary
	// Retuns
	//   A slice of approximate results
	queryPinyin(ctx context.Context, query, domain string, wdict map[string]*dicttypes.Word) ([]tmResult, error)
}

// unigramSearcher finds similar phrases with several matching characters
type unigramSearcher interface {

	// queryUnigram searches for phrases similar to the given query, optionally with a particular domain
	// Parameters
	//   ctx Request context
	//   query The search query
	//   domain The domain to restrict the query to (optional)
	//   wdict The full dictionary
	// Retuns
	//   A slice of approximate results
	queryUnigram(ctx context.Context, chars []string, domain string) ([]tmResult, error)
}

// searcher implements the Searcher interface with pinyin and unigram translation memory searchers
type searcher struct {
	ps pinyinSearcher
	us unigramSearcher
}

// newSearcher initializes an implementation of the Searcher interface
func newSearcher(ps pinyinSearcher, us unigramSearcher) (Searcher, error) {
	return searcher{
		ps: ps,
		us: us,
	}, nil
}

// memPinyinSearcher is a translation memory Searcher implementation based on in memory queries for similar pinyin
type memPinyinSearcher struct {
	revIndex dictionary.ReverseIndex
}

// newMemPinyinSearcher initializes a pinyinSearcher implementation based on in-memory queries
func newMemPinyinSearcher(revIndex dictionary.ReverseIndex) (pinyinSearcher, error) {
	return memPinyinSearcher{
		revIndex: revIndex,
	}, nil
}

// queryPinyin searches for phrases with matching pinyin
func (s memPinyinSearcher) queryPinyin(ctx context.Context, query, domain string, wdict map[string]*dicttypes.Word) ([]tmResult, error) {
	pinyin := findPinyin(query, wdict)
	if len(pinyin) == 0 {
		return nil, fmt.Errorf("fsPinyinSearcher.queryPinyin, No pinyin for query,%s", query)
	}
	results := []tmResult{}
	revResults, err := s.revIndex.Find(ctx, pinyin)
	if err != nil {
		return nil, fmt.Errorf("memPinyinSearcher.queryPinyin error from revIndex: %v", err)
	}
	revMap := map[string]bool{}
	for _, ws := range revResults {
		if len(domain) > 0 && domain != ws.Domain {
			continue
		}
		term := ws.Simplified
		_, ok := revMap[term]
		if !ok {
			tmr := tmResult{
				term:         term,
				hasPinyin:    1,
				unigramCount: charsContained(query, ws.Simplified, ws.Traditional),
			}
			results = append(results, tmr)
			revMap[term] = true
		}
	}
	// log.Printf("memPinyinSearcher.queryPinyin, %d results found, query: %s pinyin: %s", len(results), query, pinyin)
	return results, nil
}

// Searches the translation memory for approximate matches.
// Parameters
//   ctx Request context
//   query The search query
//   domain The domain to restrict the query to (optional)
//   wdict The full dictionary
// Retuns
//   A slice of approximate results
func (s searcher) Search(ctx context.Context, query, domain string, includeSubstrings bool, wdict map[string]*dicttypes.Word) (*Results, error) {
	// log.Printf("searcher.Search, query: %s domain: %s", query, domain)
	if s.ps == nil {
		return nil, fmt.Errorf("searcher: ps is nil")
	}
	chars := strings.Split(query, "")
	var matches []tmResult
	var err error
	if s.us != nil {
		matches, err = s.us.queryUnigram(ctx, chars, domain)
		if err != nil {
			return nil, fmt.Errorf("Search query error:\n%v", err)
		}
	}
	// log.Printf("searcher.Search, %d results found from queryUnigram", len(matches))
	pinyinMatches, err := s.ps.queryPinyin(ctx, query, domain, wdict)
	// log.Printf("searcher.Search, %d results found from pinyinMatches", len(pinyinMatches))
	if includeSubstrings {
		words := combineResults(query, matches, pinyinMatches, wdict)
		return &Results{
			Words: words,
		}, nil
	}
	words := combineResultsNoSubstrings(query, matches, pinyinMatches, wdict)
	log.Printf("searcher.Search, %d results found for query: %s", len(words), query)
	return &Results{
		Words: words,
	}, nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// combineResults combines matches with dictionary defintions to send back to client
func combineResults(query string, matches, pinyinMatches []tmResult, wdict map[string]*dicttypes.Word) []dicttypes.Word {
	// log.Printf("combineResults query: %s, uni matches: %d, pinyin matches: %d", query, len(matches), len(pinyinMatches))
	relevantMap := map[string]tmResult{}
	for _, m := range matches {
		m.hamming = hammingDist(query, m.term)
		m.isSubstring = eitherSubstring(query, m.term)
		m.relevant = predictRelevanceNorm(query, m)
		relevantMap[m.term] = m
	}
	for _, m := range pinyinMatches {
		m.hamming = hammingDist(query, m.term)
		m.isSubstring = eitherSubstring(query, m.term)
		m.relevant = predictRelevanceNorm(query, m)
		relevantMap[m.term] = m
	}
	allMatches := []tmResult{}
	for _, v := range relevantMap {
		allMatches = append(allMatches, v)
	}
	printResults(query, allMatches, "with substrings")

	// Eliminate dups with a map since simplified and traditional may both match
	uMap := map[int]tmResult{}
	for _, m := range allMatches {
		// log.Printf("combineResults query: %s, m.term = %s, m.relevant = %d", query, m.term, m.relevant)
		if m.relevant == 1 && len(uMap) < maxResultsSubstrings {
			if w, ok := wdict[m.term]; ok {
				uMap[w.HeadwordId] = m
			}
		}
	}

	// Sort for most relevant based on unigram count
	relevantMatches := []tmResult{}
	for _, m := range uMap {
		relevantMatches = append(relevantMatches, m)
	}
	// log.Printf("combineResults query: %s, len(uMap) = %d, len(relevantMatches) = %d", query, len(uMap), len(relevantMatches))
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].unigramCount > matches[j].unigramCount
	})

	// Transform into a slice of Words
	var words []dicttypes.Word
	for _, match := range relevantMatches {
		if word, ok := wdict[match.term]; ok {
			words = append(words, *word)
		}
	}
	// log.Printf("transmemory.combineResults, query: %s, matchs (%d): ", query, len(words))
	return words
}

// Combines matches with dictionary defintions, excluding substrings.
// It is ok for the query to be a substring of a similar term but not the other
// way around.
func combineResultsNoSubstrings(query string,
	matches, pinyinMatches []tmResult,
	wdict map[string]*dicttypes.Word) []dicttypes.Word {
	relevantMap := map[string]tmResult{}
	for _, m := range matches {
		m.hamming = hammingDist(query, m.term)
		if strings.Contains(m.term, query) {
			m.isSubstring = 1
		}
		m.relevant = predictRelevance(query, m)
		relevantMap[m.term] = m
	}
	for _, m := range pinyinMatches {
		m.hamming = hammingDist(query, m.term)
		if strings.Contains(m.term, query) {
			m.isSubstring = 1
		}
		m.relevant = predictRelevance(query, m)
		relevantMap[m.term] = m
	}
	allMatches := []tmResult{}
	for _, v := range relevantMap {
		if !strings.Contains(query, v.term) {
			allMatches = append(allMatches, v)
		}
	}
	printResults(query, allMatches, "substrings excluded")

	// Eliminate dups with a map since simplified and traditional may both match
	uMap := map[int]tmResult{}
	for _, m := range allMatches {
		if m.relevant == 1 && len(uMap) < maxResultsNoSubstrings {
			if w, ok := wdict[m.term]; ok {
				uMap[w.HeadwordId] = m
			}
		}
	}

	// Sort for most relevant based on unigram count
	relevantMatches := []tmResult{}
	for _, m := range uMap {
		relevantMatches = append(relevantMatches, m)
	}
	sort.Slice(relevantMatches, func(i, j int) bool {
		return relevantMatches[i].unigramCount > relevantMatches[j].unigramCount
	})
	var words []dicttypes.Word
	for _, match := range relevantMatches {
		// log.Printf("transmemory.combineResultsNoSubstrings, words: %s",match.term)
		if word, ok := wdict[match.term]; ok {
			words = append(words, *word)
		}
	}
	//log.Printf("transmemory.combineResultsNoSubstrings, query: %s, matchs (%d): ", query, len(words))
	return words
}

// Predict relevance based on parameters (not normaized by query length)
// Returns
//   1 if relevant, 0 if not relevant
func predictRelevance(query string, m tmResult) int {
	l := len([]rune(query))
	if l == 0 {
		return 0
	}
	if m.isSubstring == 1 || m.inNotes == 1 {
		return 1
	}
	if m.unigramCount >= 1 && m.hasPinyin == 1 {
		return 1
	}
	if m.unigramCount >= uniCountNoSubMin && m.hamming <= hammingNoSubMax {
		return 1
	}
	return 0
}

// Predict relevance based on parameters normaized by query length
// Returns
//   1 if relevant, 0 if not relevant
func predictRelevanceNorm(query string, m tmResult) int {
	l := len([]rune(query))
	if l == 0 {
		return 0
	}
	if m.isSubstring == 1 || m.inNotes == 1 {
		return 1
	}
	// log.Printf("predictRelevanceNorm, query: %s, term: %s, unigramCount: %d, hasPinyin: %d", query, m.term, m.unigramCount, m.hasPinyin)
	if m.unigramCount >= 1 && m.hasPinyin == 1 {
		return 1
	}
	normalUni := float64(m.unigramCount) / float64(l)
	normalHamming := float64(m.hamming) / float64(l)
	if normalUni >= uniCountMin && normalHamming <= hammingDistMax {
		return 1
	}
	return 0
}

// charsContained computes the number of overlapping chars contained in the query and the term
func charsContained(query, simplified, traditional string) int {
	counted := map[string]bool{}
	nChars := 0
	sChars := strings.Split(simplified, "")
	for _, c := range sChars {
		if strings.Contains(query, c) && !counted[c] {
			nChars++
			counted[c] = true
		}
	}
	tChars := strings.Split(traditional, "")
	for _, c := range tChars {
		if strings.Contains(query, c) && !counted[c] {
			nChars++
			counted[c] = true
		}
	}
	return nChars
}

// Finds the pinyin for a given Chinese string
func findPinyin(query string, wdict map[string]*dicttypes.Word) string {
	pinyin := ""
	chars := strings.Split(query, "")
	for _, ch := range chars {
		word, ok := wdict[ch]
		if ok {
			pinyin += dicttypes.NormalizePinyin(word.Pinyin)
		} else {
			log.Printf("findPinyin: query %s, char %s not found", query, ch)
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
			chars = append(chars, string(runes[i]))
			continue
		}
		chars = append(chars, string(runes[len(runes)-1]))
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
func printResults(query string, matches []tmResult, description string) {
	if len(matches) == 0 {
		log.Printf("transmemory.printResults no results")
		return
	}
	log.Printf("\nQuery, rank, Term, Has Pinyin, In Notes, Unigram count, " +
		"Hamming, Substring, Relevant\n")
	for i, m := range matches {
		if i == 10 {
			break
		}
		log.Printf("transmemory.printTopResults result: %s, %d, %s, %d, %d, %d, "+
			"%d, %d, %d\n", query, i, m.term, m.hasPinyin, m.inNotes,
			m.unigramCount, m.hamming, m.isSubstring, m.relevant)
	}
	log.Printf("transmemory.printResults %s, query: %s, matchs (%d): ",
		description, query, len(matches))
}
