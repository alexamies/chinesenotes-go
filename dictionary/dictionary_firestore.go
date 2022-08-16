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

package dictionary

import (
	"context"
  "errors"
  "fmt"
  "log"
  "strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// fsClient defines Firestore interfaces needed
type fsClient interface {
	Collection(path string) *firestore.CollectionRef
}

// HeadwordSubstrings holds substrings of a headword
type HeadwordSubstrings struct {
	HeadwordId  int64    `firestore:"headword_id"`
	Simplified  string   `firestore:"simplified"`
	Traditional string   `firestore:"traditional"`
	Pinyin      string   `firestore:"pinyin"`
	Substrings  []string `firestore:"substrings"`
}

// substringIndexFS looks up Chinese words using a substring index saved in Firestore.
type substringIndexFS struct {
	client fsClient
	corpus     string
	generation int
	dict *Dictionary
}

// NewSubstringIndexDB initialize an implementation of SubstringIndex using the index saved in Firestore
func NewSubstringIndexFS(client fsClient, corpus string, generation int, dict *Dictionary) (SubstringIndex, error) {
	if client == nil {
		return nil, errors.New("NewSubstringIndexFS, Firestore client must be initialized")
	}
	return substringIndexFS{
		client: client,
		corpus: corpus,
		generation: generation,
		dict: dict,
	}, nil
}

// Lookup a term based on a substring and a topic
func (f substringIndexFS) LookupSubstr(ctx context.Context, query, topic_en, subtopic_en string) (*Results, error) {
	if query == "" {
		return nil, errors.New("query string is empty")
	}
	log.Printf("substringIndexFS.LookupSubstr, query %s, topic = %s", query, topic_en)
	dom := strings.ToLower(topic_en)
	fsCol := fmt.Sprintf("%s_dict_%s_substring_%d", f.corpus, dom, f.generation)
	col := f.client.Collection(fsCol)
	if col == nil {
		return nil, errors.New("substringIndexFS.LookupSubstr collection is empty")
	}
	q := col.Where("substrings", "array-contains", query).Limit(100)
	iter := q.Documents(ctx)
	defer iter.Stop()
	words := []dicttypes.Word{}
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("substringIndexFS.LookupSubstr iteration error: %v", err)
		}
		var d HeadwordSubstrings
		err = ds.DataTo(&d)
		if err != nil {
			return nil, fmt.Errorf("substringIndexFS.LookupSubstr type conversion error: %v", err)
		}
		w := f.dict.HeadwordIds[int(d.HeadwordId)]
		words = append(words, *w)
	}
	log.Printf("substringIndexFS.LookupSubstr, len(wmap): %d", len(words))
	return &Results{words}, nil
}
