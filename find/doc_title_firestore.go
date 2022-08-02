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

package find

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
)

// DocTitleRecord holds expanded document title information
// plain_text_file", "gloss_file", "title", "title_cn", "title_en", "col_gloss_file", "col_title", "col_plus_doc_title
type DocTitleRecord struct {
	RawFile         string `firestore:"plain_text_file"`
	GlossFile       string `firestore:"gloss_file"`
	DocTitle        string `firestore:"title"`
	DocTitleZh      string `firestore:"title_zh"`
	DocTitleEn      string `firestore:"title_en"`
	ColGlossFile    string `firestore:"col_gloss_file"`
	ColTitle        string `firestore:"col_title"`
	ColPlusDocTitle string `firestore:"col_plus_doc_title"`
	Substrings      string `firestore:"substrings"`
}

// fsClient defines Firestore interfaces needed
type fsClient interface {
	Collection(path string) *firestore.CollectionRef
}

// firebaseTitleFinder implements the TitleFinder interface with Firestore queries
type firestoreTitleFinder struct {
	client     fsClient
	corpus     string
	generation int
	colMap     map[string]string
	dInfoCN    map[string]DocInfo
	docMap     map[string]DocInfo
}

// NewFirebaseTitleFinder initializes a DocTitleFinder implementation using Firestore queries
func NewFirebaseTitleFinder(client fsClient, corpus string, generation int, colMap map[string]string, dInfoCN, docMap map[string]DocInfo) TitleFinder {
	log.Printf("NewFirebaseTitleFinder len(colMap): %d, len(dInfoCN): %d, len(docMap): %d", len(colMap), len(dInfoCN), len(docMap))
	return firestoreTitleFinder{
		client:     client,
		corpus:     corpus,
		generation: generation,
		colMap:     colMap,
		dInfoCN:    dInfoCN,
		docMap:     docMap,
	}
}

// FileDocTitleFinder finds documents by title using a using Firestore substring query
func (f firestoreTitleFinder) FindDocsByTitle(ctx context.Context, query string) ([]Document, error) {
	fsCol := fmt.Sprintf("%s_doc_title_%d", f.corpus, f.generation)
	results := []Document{}
	col := f.client.Collection(fsCol)
	if col == nil {
		return nil, fmt.Errorf("findDocsTermFreq collection is empty")
	}
	q := col.Where("substring", "array-contains", query).OrderBy("freq", firestore.Desc).Limit(100)
	iter := q.Documents(ctx)
	defer iter.Stop()
	return results, nil
}

func (f firestoreTitleFinder) CountCollections(ctx context.Context, query string) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (f firestoreTitleFinder) FindCollections(ctx context.Context, query string) []Collection {
	return []Collection{}
}

func (f firestoreTitleFinder) FindDocsByTitleInCol(ctx context.Context, query, col_gloss_file string) ([]Document, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f firestoreTitleFinder) ColMap() map[string]string {
	return f.colMap
}

func (f firestoreTitleFinder) DocMap() map[string]DocInfo {
	return f.docMap
}
