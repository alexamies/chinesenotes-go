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

// Unit tests for find functions

package find

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/fulltext"
)

type mockReverseIndex struct {
}

func (m mockReverseIndex) Find(ctx context.Context, query string) ([]dicttypes.WordSense, error) {
	results := []dicttypes.WordSense{}
	log.Printf("Find.FindWordsByEnglish: query: %s, results: %v", query, results)
	return results, nil
}

func initDBCon() (*sql.DB, error) {
	if !config.UseDatabase() {
		return nil, nil
	}
	conString := config.DBConfig()
	return sql.Open("mysql", conString)
}

// Test package initialization, which requires a database connection
func TestCacheColDetails(t *testing.T) {
	database, err := initDBCon()
	if err != nil {
		t.Errorf("TestCacheColDetails, Error: %v", err)
		return
	}
	if database == nil {
		t.Skip("TestCacheColDetails, no database skipping")
	}
	df := databaseDocFinder{
		database: database,
	}
	ctx := context.Background()
	err = df.initFind(ctx)
	if err != nil {
		t.Fatalf("TestCacheColDetails, Error: %v", err)
	}
	cMap := df.cacheColDetails(ctx)
	title := cMap["wenxuan.html"]
	if title == "" {
		t.Logf("TestCacheColDetails: got empty title, map size, %d",
			len(cMap))
	}
}

func TestCombineByWeight(t *testing.T) {
	doc := Document{
		GlossFile:    "f2.html",
		Title:        "Very Good doc",
		SimTitle:     1.0,
		SimWords:     0.5,
		SimBigram:    1.5,
		SimBitVector: 1.0,
	}
	maxSimWords := doc.SimWords
	maxBigram := doc.SimBigram
	simDoc := combineByWeight(doc, maxSimWords, maxBigram)
	if simDoc.Similarity == 0.0 {
		t.Error("TestCombineByWeight: simDoc.Similarity == 0.0")
	}
	t.Logf("TestCacheColDetails: simDoc %v\n", simDoc)
	similarity := intercept +
		WEIGHT[0]*doc.SimWords/maxSimWords +
		WEIGHT[1]*doc.SimBigram/maxBigram +
		WEIGHT[2]*doc.SimBitVector
	expectedMin := 0.99 * similarity
	expectedMax := 1.01 * similarity
	if (expectedMin > simDoc.Similarity) ||
		(simDoc.Similarity > expectedMax) {
		t.Errorf("TestCombineByWeight: result out of expected range %v\n",
			simDoc)
	}
}

func TestFindDocuments(t *testing.T) {

	// Setup
	database, err := initDBCon()
	if err != nil {
		t.Errorf("TestFindDocuments, Error: %v", err)
		return
	}
	if database == nil {
		t.Log("TestFindDocuments, not connected to db, skipping")
		return
	}
	df := databaseDocFinder{}
	ctx := context.Background()
	err = df.initFind(ctx)
	if err != nil {
		t.Errorf("TestFindDocuments, Error: %v", err)
		return
	}
	reverseIndex := mockReverseIndex{}
	dict := map[string]*dicttypes.Word{}
	parser := NewQueryParser(dict)

	// Test data
	type test struct {
		name           string
		query          string
		expectError    bool
		expectNoTerms  int
		expectNoSenses int
	}
	tests := []test{
		{
			name:           "Happy pass",
			query:          "Assembly",
			expectError:    false,
			expectNoTerms:  1,
			expectNoSenses: 1,
		},
		{
			name:           "Empty query",
			query:          "",
			expectError:    true,
			expectNoTerms:  0,
			expectNoSenses: 1,
		},
		{
			name:           "No word senses",
			query:          "hello",
			expectError:    false,
			expectNoTerms:  0,
			expectNoSenses: 1,
		},
	}

	for _, tc := range tests {
		qr, err := df.FindDocuments(ctx, reverseIndex, parser, tc.query, false)
		gotError := (err != nil)
		if tc.expectError != gotError {
			t.Errorf("TestFindDocuments, %s: expectError: %t vs got %t",
				tc.name, tc.expectError, gotError)
			if gotError {
				t.Errorf("TestFindDocuments, %s: unexpected error: %v", tc.name, err)
			}
			continue
		}
		if gotError {
			continue
		}
		gotNoTerms := len(qr.Terms)
		if tc.expectNoTerms != gotNoTerms {
			t.Errorf("TestFindDocuments, %s: expectNum: %d vs got %d",
				tc.name, tc.expectNoTerms, gotNoTerms)
		}
	}
}

func TestFindDocumentsInCol(t *testing.T) {
	database, err := initDBCon()
	if err != nil {
		t.Errorf("TestFindDocumentsInCol, Error: %v", err)
		return
	}
	if database == nil {
		t.Skip("TestFindDocumentsInCol, no database skipping")
	}
	df := databaseDocFinder{
		database: database,
	}
	ctx := context.Background()
	err = df.initFind(ctx)
	if err != nil {
		t.Errorf("TestFindDocumentsInCol, Error: %v", err)
		return
	}
	reverseIndex := mockReverseIndex{}
	dict := map[string]*dicttypes.Word{}
	parser := NewQueryParser(dict)

	// Test data
	type test struct {
		name           string
		query          string
		collection     string
		expectError    bool
		expectNumTerms int
	}
	tests := []test{
		{
			name:           "empty query",
			query:          "",
			collection:     "wenxuan.html",
			expectError:    true,
			expectNumTerms: 0,
		},
		{
			name:           "One term",
			query:          "箴",
			collection:     "wenxuan.html",
			expectError:    false,
			expectNumTerms: 1,
		},
		{
			name:           "Two terms",
			query:          "箴也",
			collection:     "wenxuan.html",
			expectError:    false,
			expectNumTerms: 2,
		},
		{
			name:           "Three terms",
			query:          "箴也所",
			collection:     "wenxuan.html",
			expectError:    false,
			expectNumTerms: 3,
		},
		{
			name:           "Four terms",
			query:          "箴也所以",
			collection:     "wenxuan.html",
			expectError:    false,
			expectNumTerms: 4,
		},
		{
			name:           "Five terms",
			query:          "箴也所以攻",
			collection:     "wenxuan.html",
			expectError:    false,
			expectNumTerms: 5,
		},
		{
			name:           "Six terms",
			query:          "箴也所以攻疾",
			collection:     "wenxuan.html",
			expectError:    false,
			expectNumTerms: 6,
		},
	}

	for _, tc := range tests {
		qr, err := df.FindDocumentsInCol(ctx, reverseIndex, parser, tc.query, tc.collection)
		gotError := (err == nil)
		if tc.expectError != gotError {
			t.Errorf("TestFindDocumentsInCol, %s: expected error %t vs got error: %t",
				tc.name, tc.expectError, gotError)
			if gotError {
				t.Errorf("TestFindDocumentsInCol, %s: unexpected error: %v", tc.name, err)
			}
			continue
		}
		if gotError {
			continue
		}
		if tc.expectNumTerms != len(qr.Terms) {
			t.Errorf("TestFindDocumentsInCol %s:  expected num terms %d vs got %d",
				tc.name, tc.expectNumTerms, len(qr.Terms))
		}
	}
}

func TestMergeDocList(t *testing.T) {
	database, err := initDBCon()
	if err != nil {
		t.Errorf("TestMergeDocList, Error: %v", err)
		return
	}
	if database == nil {
		t.Skip("TestMergeDocList, no database skipping")
	}
	df := databaseDocFinder{
		database: database,
	}
	ctx := context.Background()
	err = df.initFind(ctx)
	if err != nil {
		t.Errorf("TestMergeDocList, Error: %v", err)
		return
	}

	simDocMap := map[string]Document{}
	docList := []Document{}
	doc1 := Document{
		GlossFile: "f1.html",
		Title:     "Good doc by title",
		SimTitle:  1.0,
	}
	simDocMap[doc1.GlossFile] = doc1
	doc2 := Document{
		GlossFile: "f2.html",
		Title:     "Very Good doc",
		SimWords:  0.5,
		SimBigram: 1.5,
	}
	docList = append(docList, doc2)

	simDocMap2 := map[string]Document{}
	docList2 := []Document{}
	doc3 := Document{
		GlossFile: "f1.html",
		Title:     "SAme Very Good doc",
		SimTitle:  1.0,
	}
	simDocMap2[doc3.GlossFile] = doc3
	doc4 := Document{
		GlossFile: "f2.html",
		Title:     "Reasonable by word frequ",
		SimWords:  1.6,
	}
	doc5 := Document{
		GlossFile: "f1.html",
		Title:     "Same Very Good doc",
		SimWords:  1.5,
		SimBigram: 1.5,
	}
	docList2 = append(docList2, doc4)
	docList2 = append(docList2, doc5)

	// Test data
	type test struct {
		name            string
		simDocMap       map[string]Document
		docList         []Document
		expectNum       int
		expectNumDocs   int
		expectGlossFile string
	}
	tests := []test{
		{
			name:            "Basic test",
			simDocMap:       simDocMap,
			docList:         docList,
			expectNum:       2,
			expectNumDocs:   2,
			expectGlossFile: doc2.GlossFile,
		},
		{
			name:            "Harder test",
			simDocMap:       simDocMap2,
			docList:         docList2,
			expectNum:       2,
			expectNumDocs:   2,
			expectGlossFile: doc3.GlossFile,
		},
	}

	for _, tc := range tests {
		mergeDocList(df, tc.simDocMap, tc.docList)
		if tc.expectNum != len(simDocMap) {
			t.Errorf("TestMergeDocList, %s: expected %d vs got %d",
				tc.name, tc.expectNum, len(simDocMap))
			continue
		}
		docs := toSortedDocList(simDocMap)
		if tc.expectNumDocs != len(docs) {
			t.Errorf("TestMergeDocList, %s: expected docs %d vs got %d",
				tc.name, tc.expectNumDocs, len(docs))
			continue
		}
		result := docs[0]
		if tc.expectGlossFile != result.GlossFile {
			t.Errorf("TestMergeDocList: expected %s, got, %v, docs: %v",
				tc.name, tc.expectGlossFile, result.GlossFile)
		}
	}
}

func TestSetContainsTerms1(t *testing.T) {
	terms := []string{"后妃"}
	doc := Document{
		ContainsWords: "后妃",
	}
	docMatch := fulltext.DocMatch{}
	doc = setMatchDetails(doc, terms, docMatch)
	expected0 := "后妃"
	result := doc.ContainsTerms
	if len(result) != 1 {
		t.Errorf("TestSetContainsTerms1: expected len = 1, got, %d\n",
			len(result))
		return
	}
	if result[0] != expected0 {
		t.Errorf("TestSetContainsTerms1: expected %s, got, %v\n", expected0,
			result)
	}
}

func TestSetContainsTerms2(t *testing.T) {
	terms := []string{"后妃", "之"}
	doc := Document{
		ContainsWords: "之,后妃",
	}
	docMatch := fulltext.DocMatch{}
	doc = setMatchDetails(doc, terms, docMatch)
	expected0 := "后妃"
	expected1 := "之"
	result := doc.ContainsTerms
	if len(result) != 2 {
		t.Errorf("TestSetContainsTerms2: expected len = 2, got, %d\n",
			len(result))
		return
	}
	if result[0] != expected0 {
		t.Errorf("TestSetContainsTerms2: expected0 %s, got, %s\n", expected0,
			result[0])
	}
	if result[1] != expected1 {
		t.Errorf("TestSetContainsTerms2: expected1 %s, got, %s\n", expected1,
			result[1])
	}
	t.Logf("TestSetContainsTerms2: %v", result)
}

func TestSetContainsTerms3(t *testing.T) {
	terms := []string{"后妃", "之"}
	doc := Document{
		ContainsWords:   "之,后妃",
		ContainsBigrams: "后妃之",
	}
	docMatch := fulltext.DocMatch{}
	doc = setMatchDetails(doc, terms, docMatch)
	expected0 := "后妃之"
	result := doc.ContainsTerms
	if len(result) != 1 {
		t.Errorf("TestSetContainsTerms3: expected len = 1, got, %d\n",
			len(result))
		return
	}
	if result[0] != expected0 {
		t.Errorf("TestSetContainsTerms3: expected0 %s, got, %s\n", expected0,
			result[0])
	}
	t.Logf("TestSetContainsTerms3: %v", result)
}

func TestSetContainsTerms4(t *testing.T) {
	terms := []string{"十年", "之", "計"}
	doc := Document{
		ContainsWords:   "十年,之,計",
		ContainsBigrams: "十年之,之計",
	}
	docMatch := fulltext.DocMatch{}
	doc = setMatchDetails(doc, terms, docMatch)
	expected0 := "十年之"
	expected1 := "之計"
	result := doc.ContainsTerms
	if len(result) != 2 {
		t.Errorf("TestSetContainsTerms4: expected len = 2, got, %d\n",
			len(result))
		return
	}
	if result[0] != expected0 {
		t.Errorf("TestSetContainsTerms4: expected0 %s, got, %s\n", expected0,
			result[0])
	}
	if result[1] != expected1 {
		t.Errorf("TestSetContainsTerms4: expected0 %s, got, %s\n", expected1,
			result[1])
	}
	t.Logf("TestSetContainsTerms4: %v", result)
}

// No substring, compare based on similarity
func TestSortMatchingSubstr1(t *testing.T) {
	doc1 := Document{
		GlossFile:  "f1.html",
		Title:      "Good doc",
		Similarity: 1.0,
	}
	doc2 := Document{
		GlossFile:  "f2.html",
		Title:      "Very Good doc",
		Similarity: 1.5,
	}
	doc3 := Document{
		GlossFile:  "f3.html",
		Title:      "Irrelevant doc",
		Similarity: 0.2,
	}
	docs := []Document{doc1, doc2, doc3}
	sortMatchingSubstr(docs)
	expected := "f2.html"
	result := docs[0].GlossFile
	if result != expected {
		t.Errorf("TestSortMatchingSubstr1: expected %s, got, %s", expected,
			result)
	}
}

// Use substring length
func TestSortMatchingSubstr2(t *testing.T) {
	md1 := fulltext.MatchingText{
		Snippet:      "",
		LongestMatch: "好好好",
		ExactMatch:   false,
	}
	doc1 := Document{
		GlossFile:    "f1.html",
		Title:        "Good doc",
		Similarity:   1.0,
		MatchDetails: md1,
	}
	md2 := fulltext.MatchingText{
		Snippet:      "",
		LongestMatch: "好好",
		ExactMatch:   false,
	}
	doc2 := Document{
		GlossFile:    "f2.html",
		Title:        "Very Good doc",
		Similarity:   1.5,
		MatchDetails: md2,
	}
	docs := []Document{doc1, doc2}
	sortMatchingSubstr(docs)
	expected := "f1.html"
	result := docs[0].GlossFile
	if result != expected {
		t.Errorf("TestSortMatchingSubstr2: expected %s, got, %s", expected,
			result)
	}
}

// Equal substring length, use similarity
func TestSortMatchingSubstr3(t *testing.T) {
	md1 := fulltext.MatchingText{
		Snippet:      "",
		LongestMatch: "好好",
		ExactMatch:   false,
	}
	doc1 := Document{
		GlossFile:    "f1.html",
		Title:        "Good doc",
		Similarity:   1.0,
		MatchDetails: md1,
	}
	md2 := fulltext.MatchingText{
		Snippet:      "",
		LongestMatch: "好好",
		ExactMatch:   false,
	}
	doc2 := Document{
		GlossFile:    "f2.html",
		Title:        "Very Good doc",
		Similarity:   1.5,
		MatchDetails: md2,
	}
	docs := []Document{doc1, doc2}
	sortMatchingSubstr(docs)
	expected := "f2.html"
	result := docs[0].GlossFile
	if result != expected {
		t.Errorf("TestSortMatchingSubstr2: expected %s, got, %s", expected,
			result)
	}
}

func TestToRelevantDocList(t *testing.T) {
	database, err := initDBCon()
	if err != nil {
		t.Errorf("TestMergeDocList, Error: %v", err)
		return
	}
	if database == nil {
		t.Skip("TestMergeDocList, no database skipping")
		return
	}
	df := databaseDocFinder{
		database: database,
	}
	ctx := context.Background()
	err = df.initFind(ctx)
	if err != nil {
		t.Fatalf("TestMergeDocList, Error: %v", err)
	}

	similarDocMap := map[string]Document{}
	doc1 := Document{
		GlossFile:  "f1.html",
		Title:      "Good doc",
		Similarity: 1.0,
	}
	similarDocMap[doc1.GlossFile] = doc1
	doc2 := Document{
		GlossFile:  "f2.html",
		Title:      "Very Good doc",
		Similarity: 1.5,
	}
	similarDocMap[doc2.GlossFile] = doc2
	doc3 := Document{
		GlossFile:  "f3.html",
		Title:      "Irrelevant doc",
		Similarity: 0.2,
	}
	similarDocMap[doc3.GlossFile] = doc3
	docs := toSortedDocList(similarDocMap)
	queryTerms := []string{}
	docs = toRelevantDocList(df, docs, queryTerms)
	expected := 2
	result := len(docs)
	if result == expected {
		t.Errorf("TestToRelevantDocList: expected %d, got, %d", expected,
			result)
	}
}

func TestToSortedDocList1(t *testing.T) {
	similarDocMap := map[string]Document{}
	doc1 := Document{
		GlossFile: "f1.html",
		Title:     "Good doc",
		SimWords:  1.0,
		SimBigram: 1.0,
	}
	similarDocMap[doc1.GlossFile] = doc1
	doc2 := Document{
		GlossFile: "f2.html",
		Title:     "Very Good doc",
		SimWords:  1.5,
		SimBigram: 1.5,
	}
	similarDocMap[doc2.GlossFile] = doc2
	doc3 := Document{
		GlossFile: "f3.html",
		Title:     "Reasonable doc",
		SimWords:  0.5,
		SimBigram: 0.5,
	}
	similarDocMap[doc3.GlossFile] = doc3
	docs := toSortedDocList(similarDocMap)
	expected := doc2.GlossFile
	result := docs[0]
	if result.Similarity == 0.0 {
		t.Error("TestToSortedDocList1: result.Similarity == 0.0")
	}
	if result.GlossFile != expected {
		t.Errorf("TestToSortedDocList1: expected %s, got, %v", expected, result)
	}
}

func TestToSortedDocList2(t *testing.T) {
	similarDocMap := map[string]Document{}
	doc1 := Document{
		GlossFile: "f1.html",
		Title:     "Good doc",
		SimWords:  0.5,
		SimBigram: 1.0,
	}
	similarDocMap[doc1.GlossFile] = doc1
	doc2 := Document{
		GlossFile: "f2.html",
		Title:     "Very Good doc",
		SimWords:  0.5,
		SimBigram: 1.5,
	}
	similarDocMap[doc2.GlossFile] = doc2
	doc3 := Document{
		GlossFile: "f3.html",
		Title:     "Reasonable doc",
		SimWords:  0.5,
	}
	similarDocMap[doc3.GlossFile] = doc3
	docs := toSortedDocList(similarDocMap)
	expected := doc2.GlossFile
	result := docs[0]
	if result.GlossFile != expected {
		t.Errorf("TestToSortedDocList2: expected %s, got, %v", expected, result)
	}
}
