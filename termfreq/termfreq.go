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

	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"
)

const (
	k        float64 = 1.5
	b        float64 = 0.65
	avDocLen         = 4497
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

type BM25Score struct {
	Document   string
	Collection string
	Score      float64
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

// FindDocsBigramFreq finds documents with occurences of any of the bigram given in the corpus ordered by BM25 score
func FindDocsBigramFreq(ctx context.Context, client fsClient, corpus string, generation int, bigrams []string) ([]BM25Score, error) {
	fbCol := fmt.Sprintf("%s_bigram_doc_freq%d", corpus, generation)
	return findDocsTermFreq(ctx, client, fbCol, bigrams)
}

// FindDocsTermFreq finds documents with occurences of any of the terms given in the corpus ordered by BM25 score
func FindDocsTermFreq(ctx context.Context, client fsClient, corpus string, generation int, terms []string) ([]BM25Score, error) {
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", corpus, generation)
	return findDocsTermFreq(ctx, client, fbCol, terms)
}

// findDocsTermFreq finds documents with occurences of any of the terms or bigrams
func findDocsTermFreq(ctx context.Context, client fsClient, fbCol string, terms []string) ([]BM25Score, error) {
	col := client.Collection(fbCol)
	if col == nil {
		return nil, fmt.Errorf("findDocsTermFreq collection is empty")
	}
	q := col.Where("term", "in", terms).OrderBy("freq", firestore.Desc).Limit(100)
	iter := q.Documents(ctx)
	defer iter.Stop()
	docs := map[string][]*TermFreqDoc{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("FindDocsTermFreq iteration error: %v", err)
		}
		var tf TermFreqDoc
		err = ds.DataTo(&tf)
		if err != nil {
			return nil, fmt.Errorf("FindDocsTermFreq type conversion error: %v", err)
		}
		log.Printf("%s: %d in doc %s, col %s", tf.Term, tf.Freq, tf.Document, tf.Collection)
		d, ok := docs[tf.Document]
		if ok {
			d = append(d, &tf)
			docs[tf.Document] = d
		} else {
			docs[tf.Document] = []*TermFreqDoc{&tf}
		}
	}
	scores := []BM25Score{}
	for k, v := range docs {
		col := ""
		if len(v) > 0 {
			col = v[0].Collection
		}
		d := BM25Score{
			Document:   k,
			Collection: col,
			Score:      bm25(v),
		}
		scores = append(scores, d)
	}
	return scores, nil
}

// FindDocsTermCo finds documents within the scope of a corpus collection
func FindDocsBigramCo(ctx context.Context, client fsClient, corpus string, generation int, bigrams []string, col string) ([]BM25Score, error) {
	fbCol := fmt.Sprintf("%s_bigram_doc_freq%d", corpus, generation)
	return findDocsCol(ctx, client, fbCol, bigrams, col)
}

// FindDocsTermCo finds documents within the scope of a corpus collection
func FindDocsTermCo(ctx context.Context, client fsClient, corpus string, generation int, terms []string, col string) ([]BM25Score, error) {
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", corpus, generation)
	return findDocsCol(ctx, client, fbCol, terms, col)
}

// findDocsCol finds documents within the scope of a corpus collection
func findDocsCol(ctx context.Context, client fsClient, fbCol string, terms []string, colName string) ([]BM25Score, error) {
	col := client.Collection(fbCol)
	if col == nil {
		return nil, fmt.Errorf("findDocsCol collection is empty")
	}
	q := col.Where("term", "in", terms).Where("collection", "==", colName).Limit(100)
	iter := q.Documents(ctx)
	defer iter.Stop()
	docs := map[string][]*TermFreqDoc{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("FindDocsTermFreq iteration error: %v", err)
		}
		var tf TermFreqDoc
		err = ds.DataTo(&tf)
		if err != nil {
			return nil, fmt.Errorf("FindDocsTermFreq type conversion error: %v", err)
		}
		log.Printf("%s: %d in doc %s, col %s", tf.Term, tf.Freq, tf.Document, tf.Collection)
		d, ok := docs[tf.Document]
		if ok {
			d = append(d, &tf)
			docs[tf.Document] = d
		} else {
			docs[tf.Document] = []*TermFreqDoc{&tf}
		}
	}
	scores := []BM25Score{}
	for k, v := range docs {
		col := ""
		if len(v) > 0 {
			col = v[0].Collection
		}
		d := BM25Score{
			Document:   k,
			Collection: col,
			Score:      bm25(v),
		}
		scores = append(scores, d)
	}
	return scores, nil
}
