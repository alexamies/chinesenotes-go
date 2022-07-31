// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// termfreq provides IO functions to read term frequency information from
// Firestore
package termfreq

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/alexamies/chinesenotes-go/find"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"
)

const (
	k          float64 = 1.5
	b          float64 = 0.65
	avDocLen           = 4497
	queryLimit         = 300
)

// fsClient defines Firestore interfaces needed
type fsClient interface {
	Collection(path string) *firestore.CollectionRef
}

type TermFreqDoc struct {
	Term       string  `firestore:"term"`
	Freq       int64   `firestore:"freq"`
	Collection string  `firestore:"collection"`
	Document   string  `firestore:"document"`
	IDF        float64 `firestore:"idf"`
	DocLen     int64   `firestore:"doclen"`
}

type fsDocFinder struct {
	client       fsClient
	corpus       string
	generation   int
	addDirectory bool
}

// NewFirestoreDocFinder creates a TermFreqDocFinder implemented with a Firestore client
// Set addDirectory if you need to add a directory prefix to the collection names
func NewFirestoreDocFinder(client fsClient, corpus string, generation int, addDirectory bool) find.TermFreqDocFinder {
	log.Printf("NewFirestoreDocFinder: instantiating new instance with corpus %s, generation %d", corpus, generation)
	return fsDocFinder{
		client:       client,
		corpus:       corpus,
		generation:   generation,
		addDirectory: addDirectory,
	}
}

// bm25 computes the BM25 score using the formula (Zhai and Massung 2016, loc. 2423):
//
// f(q, d) = Σ c(w, q) {[k + 1] c(w, d) / [c(w, d) + k(1 - b + |d| / avdl)]} idf(w)
//
// where
// f is the BM25 value, which is summed over all words w in both documents d and query q
// c(w, q) is the count of words w in the query (assume 1, terms with 0 will not be selected)
// c(w, d) is the count of words w in document d
// k is a parameter. Robertson and Zaragoza recommend 1.2 < k < 2 ( 2009, p. 360)
// b is a parameter. Robertson and Zaragoza recommend 0.5 < b < 0.8 ( 2009, p. 360)
// |d| is the length of document d
// avdl is the average document length (depends on corpus)
// idf(w) is the inverse document frequency of word w (precomputed)
//
// This software uses k = 1.5, and b = 0.65.
func bm25(entries []*TermFreqDoc) float64 {
	score := 0.0
	for _, w := range entries {
		score += (k + 1.0) * float64(w.Freq) / (float64(w.Freq) + k*(1.0-b+float64(w.DocLen)/float64(avDocLen))) * w.IDF
	}
	return score
}

// bitvector computes the bit vector product; that is, how many terms in the query are present in the document.
// The value is normalized by the number of terms in the query.
func bitvector(terms []string, entries []*TermFreqDoc) float64 {
	if len(terms) > 0 {
		return float64(len(entries)) / float64(len(terms))
	}
	return float64(0.0)
}

// FindDocsBigramFreq finds documents with occurences of any of the bigram given in the corpus ordered by BM25 score
func (f fsDocFinder) FindDocsBigramFreq(ctx context.Context, bigrams []string) ([]find.BM25Score, error) {
	fbCol := fmt.Sprintf("%s_bigram_doc_freq%d", f.corpus, f.generation)
	return findDocsTermFreq(ctx, f.client, fbCol, bigrams, f.addDirectory)
}

// FindDocsTermFreq finds documents with occurences of any of the terms given in the corpus ordered by BM25 score
func (f fsDocFinder) FindDocsTermFreq(ctx context.Context, terms []string) ([]find.BM25Score, error) {
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", f.corpus, f.generation)
	return findDocsTermFreq(ctx, f.client, fbCol, terms, f.addDirectory)
}

// findDocsTermFreq finds documents with occurences of any of the terms or bigrams
func findDocsTermFreq(ctx context.Context, client fsClient, fbCol string, terms []string, addDirectory bool) ([]find.BM25Score, error) {
	col := client.Collection(fbCol)
	if col == nil {
		return nil, fmt.Errorf("findDocsTermFreq collection is empty")
	}
	q := col.Where("term", "in", terms).OrderBy("freq", firestore.Desc).Limit(queryLimit)
	iter := q.Documents(ctx)
	defer iter.Stop()
	docs := map[string][]*TermFreqDoc{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("findDocsTermFreq iteration error: %v", err)
		}
		var tf TermFreqDoc
		err = ds.DataTo(&tf)
		if err != nil {
			return nil, fmt.Errorf("findDocsTermFreq type conversion error: %v", err)
		}
		log.Printf("findDocsTermFreq %s: freq: %d, idf: %0.3f, DocLen: %d in doc:%s, col:%s", tf.Term, tf.Freq, tf.IDF, tf.DocLen, tf.Document, tf.Collection)
		d, ok := docs[tf.Document]
		if ok {
			d = append(d, &tf)
			docs[tf.Document] = d
		} else {
			docs[tf.Document] = []*TermFreqDoc{&tf}
		}
	}
	scores := findScores(docs, terms, addDirectory)
	log.Printf("findDocsTermFreq: for terms %v, found %d matching docs", terms, len(scores))
	return scores, nil
}

// FindDocsTermCo finds documents within the scope of a corpus collection
func (f fsDocFinder) FindDocsBigramCo(ctx context.Context, bigrams []string, col string) ([]find.BM25Score, error) {
	fbCol := fmt.Sprintf("%s_bigram_doc_freq%d", f.corpus, f.generation)
	return findDocsCol(ctx, f.client, fbCol, bigrams, col, f.addDirectory)
}

// FindDocsTermCo finds documents within the scope of a corpus collection
func (f fsDocFinder) FindDocsTermCo(ctx context.Context, terms []string, col string) ([]find.BM25Score, error) {
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", f.corpus, f.generation)
	return findDocsCol(ctx, f.client, fbCol, terms, col, f.addDirectory)
}

// findDocsCol finds documents within the scope of a corpus collection
func findDocsCol(ctx context.Context, client fsClient, fbCol string, terms []string, colName string, addDirectory bool) ([]find.BM25Score, error) {
	col := client.Collection(fbCol)
	if col == nil {
		return nil, fmt.Errorf("findDocsCol collection is empty")
	}
	q := col.Where("term", "in", terms).Where("collection", "==", colName).Limit(queryLimit)
	iter := q.Documents(ctx)
	defer iter.Stop()
	docs := map[string][]*TermFreqDoc{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("findDocsCol iteration error: %v", err)
		}
		var tf TermFreqDoc
		err = ds.DataTo(&tf)
		if err != nil {
			return nil, fmt.Errorf("findDocsCol type conversion error: %v", err)
		}
		d, ok := docs[tf.Document]
		if ok {
			d = append(d, &tf)
			docs[tf.Document] = d
		} else {
			docs[tf.Document] = []*TermFreqDoc{&tf}
		}
	}
	scores := findScores(docs, terms, addDirectory)
	log.Printf("findDocsCol: for terms %v, found %d matching docs", terms, len(scores))
	return scores, nil
}

// addDirectoryToCol adds a directory prefix matching the doc to col
func addDirectoryToCol(col, doc string) string {
	i := strings.Index(doc, "/")
	if i > 0 {
		dir := doc[:i]
		return dir + "/" + col
	}
	return col
}

func findScores(docs map[string][]*TermFreqDoc, terms []string, addDirectory bool) []find.BM25Score {
	scores := []find.BM25Score{}
	for k, v := range docs {
		col := ""
		if len(v) > 0 {
			col = v[0].Collection
		}
		containsTerms := ""
		for _, tf := range v {
			containsTerms = containsTerms + tf.Term
		}
		if addDirectory {
			col = addDirectoryToCol(col, k)
		}
		d := find.BM25Score{
			Document:      k,
			Collection:    col,
			Score:         bm25(v),
			BitVector:     bitvector(terms, v),
			ContainsTerms: containsTerms,
		}
		scores = append(scores, d)
	}
	return scores
}
