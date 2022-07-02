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

        "google.golang.org/api/iterator"

        "cloud.google.com/go/firestore"
)

type TermFreqDoc struct {
	Term    string  `firestore:"term"`
	Freq int64 `firestore:"freq"`
	Collection string `firestore:"collection"`
	Document string `firestore:"document"`
	IDF float64 `firestore:"idf"`
	DocLen int64 `firestore:"doclen"`
}

// FindDocsForTerm finds documents with occurences of any of the terms given in the corpus
func FindDocsForTerm(ctx context.Context, client *firestore.Client, corpus string, generation int, terms []string) error {
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", corpus, generation)
	entries := client.Collection(fbCol)
	q := entries.Where("term", "in", terms).OrderBy("freq", firestore.Desc)
	iter := q.Documents(ctx)
	defer iter.Stop()
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("FindDocsForTerm iteration error: %v", err)
		}
		var tf TermFreqDoc
		err = ds.DataTo(&tf)
		if err != nil {
			return fmt.Errorf("FindDocsForTerm type conversion error: %v\n", err)
		}
		fmt.Printf("%s: %d in %s", tf.Term, tf.Freq, tf.Document)
	}
	return nil
}
