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
	"flag"
	"log"
	"testing"

  "cloud.google.com/go/firestore"
)

var (
	projectID = flag.String("project_id", "", "GCP project ID")
)

func TestFindDocsTermFreq(t *testing.T) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	corpus := "cnreader"
	generation := 0
	terms := []string{"古"}
	docs, err := FindDocsTermFreq(ctx, client, corpus, generation, terms)
	if err != nil {
		t.Errorf("Unexpected error in test: %v", err)
	}
	log.Printf("FindDocsTermFreq: found %d docs\n: %v", len(docs), docs)
}