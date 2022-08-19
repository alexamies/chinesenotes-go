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

package transmemory

import (
	"context"
	"errors"
	"fmt"
	"log"

	"google.golang.org/api/iterator"

	"github.com/alexamies/chinesenotes-go/dictionary"	
)

// TMIndexUnigram holds characters for a term, used as an index for translation memory
type TMIndexUnigram struct {
	Ch  string    `firestore:"ch"`
	Word  string   `firestore:"word"`
}

// TMIndexDomain holds characters for a term with domain
type TMIndexDomain struct {
	Ch  string    `firestore:"ch"`
	Word  string   `firestore:"word"`
	Domain  string   `firestore:"domain"`
}

// fsSearcher is a translation memory Searcher implementation based on Firestore queries for matching unigrams
type fsSearcher struct {
	client 			dictionary.FsClient
	corpus     	string
	generation 	int
}

// NewFSSearcher initializes a Searcher implementation based on Firestore queries
func NewFSSearcher(client dictionary.FsClient, corpus string, generation int, revIndex dictionary.ReverseIndex) (Searcher, error) {
	us, err := newFSUniSearcher(client, corpus, generation)
	if err != nil {
		return nil, fmt.Errorf("error from newFSUniSearcher: %v", err)
	}
	ps, err := newMemPinyinSearcher(revIndex)
	if err != nil {
		return nil, fmt.Errorf("error from newMemPinyinSearcher: %v", err)
	}
	return newSearcher(ps, us)
}

// newFSUniSearcher initializes a Searcher implementation based on Firestore queries
func newFSUniSearcher(client dictionary.FsClient, corpus string, generation int) (unigramSearcher, error) {
	if client == nil {
		return nil, fmt.Errorf("Firestore client is nil")
	}
	return fsSearcher{
		client: client,
		corpus: corpus,
		generation: generation,
	}, nil
}

// queryUnigram searches for phrases with matching characters
func (s fsSearcher) queryUnigram(ctx context.Context, chars []string, domain string) ([]tmResult, error) {
	if len(domain) > 0 {
		return s.queryUnigramDom(ctx, chars, domain)
	}
	fsCol := fmt.Sprintf("%s_transmemory_uni_%d", s.corpus, s.generation)
	// log.Printf("fsSearcher.queryUnigram, fsCol: %s", fsCol)
	col := s.client.Collection(fsCol)
	if col == nil {
		return nil, errors.New("fsSearcher.queryUnigram collection is empty")
	}
	q := col.Where("ch", "in", chars).Limit(2000)
	iter := q.Documents(ctx)
	defer iter.Stop()
	rMap := map[string]tmResult{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fsSearcher.queryUnigram iteration error: %v", err)
		}
		var d TMIndexUnigram
		err = ds.DataTo(&d)
		if err != nil {
			return nil, fmt.Errorf("fsSearcher.queryUnigram type conversion error: %v", err)
		}
		tmr, ok := rMap[d.Word]
		if !ok {
			r := tmResult{
				term: d.Word,
				unigramCount: 1,
			}
			rMap[d.Word] = r
		} else {
			r := tmResult{
				term: d.Word,
				unigramCount: tmr.unigramCount + 1,
			}
			rMap[d.Word] = r
		}
	}
	results := []tmResult{}
	for _, v := range rMap {
		results = append(results, v)
	}
	// log.Printf("fsSearcher.queryUnigram, %d results with query %v", len(results), chars)
	return results, nil
}

// queryUnigramDom searches for phrases with matching characters with a similar domain
func (s fsSearcher) queryUnigramDom(ctx context.Context, chars []string, domain string) ([]tmResult, error) {
	fsCol := fmt.Sprintf("%s_transmemory_dom_%d", s.corpus, s.generation)
	col := s.client.Collection(fsCol)
	if col == nil {
		return nil, errors.New("fsSearcher.queryUnigramDom collection is empty")
	}
	q := col.Where("ch", "in", chars).Where("domain", "==", domain).Limit(2000)
	iter := q.Documents(ctx)
	defer iter.Stop()
	rMap := map[string]tmResult{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fsSearcher.queryUnigramDom iteration error: %v", err)
		}
		var d TMIndexDomain
		err = ds.DataTo(&d)
		if err != nil {
			return nil, fmt.Errorf("fsSearcher.queryUnigramDom type conversion error: %v", err)
		}
		tmr, ok := rMap[d.Word]
		if !ok {
			r := tmResult{
				term: d.Word,
				unigramCount: 1,
			}
			rMap[d.Word] = r
		} else {
			r := tmResult{
				term: d.Word,
				unigramCount: tmr.unigramCount + 1,
			}
			rMap[d.Word] = r
		}
	}
	results := []tmResult{}
	for _, v := range rMap {
		results = append(results, v)
	}
	log.Printf("fsSearcher.queryUnigramDom, %d results found", len(results))
	return results, nil
}