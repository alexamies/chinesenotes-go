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

package termfreq

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/alexamies/chinesenotes-go/find"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	er = TermFreqDoc{
		Term:       "而",
		Document:   "sampletest3.html",
		Collection: "testcollection.html",
		Freq:       1,
		IDF:        0.22184874961635637,
		DocLen:     3,
	}
	bu = TermFreqDoc{
		Term:       "不",
		Document:   "sampletest3.html",
		Collection: "testcollection.html",
		Freq:       1,
		IDF:        0.3979400086720376,
		DocLen:     3,
	}
	bai = TermFreqDoc{
		Term:       "敗",
		Document:   "sampletest3.html",
		Collection: "testcollection.html",
		Freq:       1,
		IDF:        0.3979400086720376,
		DocLen:     3,
	}
	oneBM25   = (k + 1.0) * float64(er.Freq) / (float64(er.Freq) + k*(1.0-b+float64(er.DocLen)/float64(avDocLen))) * er.IDF
	twoBM25   = oneBM25 + (k+1.0)*float64(bu.Freq)/(float64(bu.Freq)+k*(1.0-b+float64(bu.DocLen)/float64(avDocLen)))*bu.IDF
	threeBM25 = twoBM25 + (k+1.0)*float64(bai.Freq)/(float64(bai.Freq)+k*(1.0-b+float64(bai.DocLen)/float64(avDocLen)))*bai.IDF
)

type mockFsClient struct {
}

func (m mockFsClient) Collection(path string) *firestore.CollectionRef {
	if len(path) == 0 {
		return nil
	}
	return &firestore.CollectionRef{
		Path:  path,
		Query: firestore.Query{},
	}
}

func TestBM25(t *testing.T) {
	empty := []*TermFreqDoc{}
	one := []*TermFreqDoc{
		&er,
	}
	two := []*TermFreqDoc{
		&er,
		&bu,
	}
	three := []*TermFreqDoc{
		&er,
		&bu,
		&bai,
	}
	type test struct {
		name    string
		entries []*TermFreqDoc
		want    float64
	}
	tests := []test{
		{
			name:    "Empty",
			entries: empty,
			want:    0.0,
		},
		{
			name:    "One",
			entries: one,
			want:    oneBM25,
		},
		{
			name:    "two",
			entries: two,
			want:    twoBM25,
		},
		{
			name:    "three",
			entries: three,
			want:    threeBM25,
		},
	}
	for _, tc := range tests {
		got := bm25(tc.entries)
		if got != tc.want {
			t.Errorf("TestBM25.%s: got %0.4f, want %0.4f", tc.name, got, tc.want)
		}
	}
}

func TestBitVector(t *testing.T) {
	empty := []*TermFreqDoc{}
	one := []*TermFreqDoc{
		&er,
	}
	terms := []string{"而", "敗"}
	two := []*TermFreqDoc{
		&er,
		&bai,
	}
	t3 := []string{"而", "不", "敗"}
	three := []*TermFreqDoc{
		&er,
		&bu,
		&bai,
	}
	type test struct {
		name    string
		terms   []string
		entries []*TermFreqDoc
		want    float64
	}
	tests := []test{
		{
			name:    "Empty",
			terms:   []string{},
			entries: empty,
			want:    0.0,
		},
		{
			name:    "Partial",
			terms:   terms,
			entries: one,
			want:    0.5,
		},
		{
			name:    "Two",
			terms:   terms,
			entries: two,
			want:    1.0,
		},
		{
			name:    "Three",
			terms:   t3,
			entries: three,
			want:    1.0,
		},
		{
			name:    "three",
			terms:   t3,
			entries: three,
			want:    1.0,
		},
	}
	for _, tc := range tests {
		got := bitvector(tc.terms, tc.entries)
		if got != tc.want {
			t.Errorf("TestBitVector.%s: with terms %v, entries %v got %0.3f, want %0.3f", tc.name, tc.terms, tc.entries, got, tc.want)
		}
	}
}

func TestFindDocsTermFreq(t *testing.T) {
	ctx := context.Background()
	client := mockFsClient{}
	type test struct {
		name      string
		path      string
		wantError bool
	}
	tests := []test{
		{
			name:      "Empty",
			path:      "",
			wantError: true,
		},
	}
	for _, tc := range tests {
		_, err := findDocsTermFreq(ctx, client, tc.path, []string{}, false)
		if !tc.wantError && err != nil {
			t.Fatalf("TestFindDocsTermFreq.%s: unexpected error: %v", tc.name, err)
		}
		if tc.wantError && err == nil {
			t.Fatalf("TestFindDocsTermFreq.%s: expected error but got none", tc.name)
		}
	}
}

func TestFindDocsCol(t *testing.T) {
	ctx := context.Background()
	client := mockFsClient{}
	type test struct {
		name      string
		path      string
		wantError bool
	}
	tests := []test{
		{
			name:      "Empty",
			path:      "",
			wantError: true,
		},
	}
	for _, tc := range tests {
		_, err := findDocsCol(ctx, client, tc.path, []string{}, "x", false)
		if !tc.wantError && err != nil {
			t.Fatalf("TestFindDocsCol.%s: unexpected error: %v", tc.name, err)
		}
		if tc.wantError && err == nil {
			t.Fatalf("TestFindDocsCol.%s: expected error but got none", tc.name)
		}
	}
}

func TestAddDirectoryToCol(t *testing.T) {
	type test struct {
		name string
		col  string
		doc  string
		want string
	}
	tests := []test{
		{
			name: "No prefix needed",
			col:  "b.html",
			doc:  "z.html",
			want: "b.html",
		},
		{
			name: "Prefix is needed",
			col:  "b.html",
			doc:  "a/z.html",
			want: "a/b.html",
		},
	}
	for _, tc := range tests {
		got := addDirectoryToCol(tc.col, tc.doc)
		if got != tc.want {
			t.Errorf("TestAddDirectoryToCol.%s: got %s but want: %s", tc.name, got, tc.want)
		}
	}
}

func TestFindScores(t *testing.T) {
	opt := cmp.Comparer(func(x, y []find.BM25Score) bool {
		if len(x) != len(y) {
			return false
		}
		for i, s1 := range x {
			s2 := y[i]
			if s1.Document != s2.Document {
				return false
			}
			if s1.Collection != s2.Collection {
				return false
			}
			approx := cmpopts.EquateApprox(0.00001, 0.00001)
			if !cmp.Equal(s1.Score, s2.Score, approx) {
				return false
			}
			if !cmp.Equal(s1.BitVector, s2.BitVector, approx) {
				return false
			}
			if s1.ContainsTerms != s2.ContainsTerms {
				return false
			}
		}
		return true
	})
	noDocs := map[string][]*TermFreqDoc{}
	noTerms := []string{}
	one := []*TermFreqDoc{
		&er,
	}
	oneDoc := map[string][]*TermFreqDoc{
		er.Document: one,
	}
	oneTerm := []string{"而"}
	oneS := find.BM25Score{
		Document:      er.Document,
		Collection:    er.Collection,
		Score:         oneBM25,
		BitVector:     1.0,
		ContainsTerms: "而",
	}
	oneScore := []find.BM25Score{
		oneS,
	}
	two := []*TermFreqDoc{
		&er,
		&bu,
	}
	twoDocs := map[string][]*TermFreqDoc{
		er.Document: two,
	}
	twoTerms := []string{"而", "不"}
	twoS := find.BM25Score{
		Document:      bu.Document,
		Collection:    bu.Collection,
		Score:         twoBM25,
		BitVector:     1.0,
		ContainsTerms: "而不",
	}
	twoScores := []find.BM25Score{
		twoS,
	}
	three := []*TermFreqDoc{
		&er,
		&bu,
		&bai,
	}
	threeDocs := map[string][]*TermFreqDoc{
		er.Document: three,
	}
	threeTerms := []string{"而", "不", "敗"}
	threeS := find.BM25Score{
		Document:      bu.Document,
		Collection:    bu.Collection,
		Score:         threeBM25,
		BitVector:     1.0,
		ContainsTerms: "而不敗",
	}
	threeScores := []find.BM25Score{
		threeS,
	}
	const addDirectory = false
	type test struct {
		name  string
		docs  map[string][]*TermFreqDoc
		terms []string
		want  []find.BM25Score
	}
	tests := []test{
		{
			name:  "Empty",
			docs:  noDocs,
			terms: noTerms,
			want:  []find.BM25Score{},
		},
		{
			name:  "One",
			docs:  oneDoc,
			terms: oneTerm,
			want:  oneScore,
		},
		{
			name:  "Two",
			docs:  twoDocs,
			terms: twoTerms,
			want:  twoScores,
		},
		{
			name:  "Three",
			docs:  threeDocs,
			terms: threeTerms,
			want:  threeScores,
		},
	}
	for _, tc := range tests {
		got := findScores(tc.docs, tc.terms, addDirectory)
		if !cmp.Equal(got, tc.want, opt) {
			t.Errorf("TestFindScores.%s: got %v but want: %v", tc.name, got, tc.want)
		}
	}
}
