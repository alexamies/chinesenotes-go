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
	er := TermFreqDoc{
		Term:   "而",
		Freq:   1,
		IDF:    0.22184874961635637,
		DocLen: 3,
	}
	bai := TermFreqDoc{
		Term:   "敗",
		Freq:   1,
		IDF:    0.3979400086720376,
		DocLen: 3,
	}
	one := []*TermFreqDoc{
		&er,
	}
	oneBM25 := (k + 1.0) * float64(er.Freq) / (float64(er.Freq) + k*(1.0-b+float64(er.DocLen)/float64(avDocLen))) * er.IDF
	two := []*TermFreqDoc{
		&er,
		&bai,
	}
	twoBM25 := oneBM25 + (k+1.0)*float64(bai.Freq)/(float64(bai.Freq)+k*(1.0-b+float64(bai.DocLen)/float64(avDocLen)))*bai.IDF
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
	}
	for _, tc := range tests {
		got := bm25(tc.entries)
		if got != tc.want {
			t.Errorf("TestBM25.%s: got %0.4f, want %0.4f", tc.name, got, tc.want)
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
		_, err := findDocsTermFreq(ctx, client, tc.path, []string{})
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
		_, err := findDocsCol(ctx, client, tc.path, []string{}, "x")
		if !tc.wantError && err != nil {
			t.Fatalf("TestFindDocsCol.%s: unexpected error: %v", tc.name, err)
		}
		if tc.wantError && err == nil {
			t.Fatalf("TestFindDocsCol.%s: expected error but got none", tc.name)
		}
	}
}
