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
	"log"
	"reflect"
	"testing"

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

type mockDocFinder struct {
	scores []BM25Score
}

func newMockDocFinder(scores []BM25Score) TermFreqDocFinder {
	return mockDocFinder{
		scores: scores,
	}
}

func mockSmallDict() map[string]*dicttypes.Word {
	s1 := "繁体中文"
	t1 := "繁體中文"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "fántǐ zhōngwén",
		Senses:      []dicttypes.WordSense{},
	}
	s2 := "前"
	t2 := "\\N"
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      "qián",
		Senses:      []dicttypes.WordSense{},
	}
	s3 := "不见"
	t3 := "不見"
	hw3 := dicttypes.Word{
		HeadwordId:  3,
		Simplified:  s3,
		Traditional: t3,
		Pinyin:      "bújiàn",
		Senses:      []dicttypes.WordSense{},
	}
	s4 := "古人"
	t4 := "\\N"
	hw4 := dicttypes.Word{
		HeadwordId:  4,
		Simplified:  s4,
		Traditional: t4,
		Pinyin:      "gǔrén",
		Senses:      []dicttypes.WordSense{},
	}
	s5 := "夫"
	t5 := "\\N"
	hw5 := dicttypes.Word{
		HeadwordId:  5,
		Simplified:  s5,
		Traditional: t5,
		Pinyin:      "fú fū",
		Senses:      []dicttypes.WordSense{},
	}
	s6 := "起信论"
	t6 := "起信論"
	hw6 := dicttypes.Word{
		HeadwordId:  6,
		Simplified:  s6,
		Traditional: t6,
		Pinyin:      "Qǐ Xìn Lùn",
		Senses:      []dicttypes.WordSense{},
	}
	s7 := "者"
	t7 := "\\N"
	hw7 := dicttypes.Word{
		HeadwordId:  7,
		Simplified:  s7,
		Traditional: t7,
		Pinyin:      "zhě zhuó",
		Senses:      []dicttypes.WordSense{},
	}
	s8 := "乃是"
	t8 := "\\N"
	hw8 := dicttypes.Word{
		HeadwordId:  8,
		Simplified:  s8,
		Traditional: t8,
		Pinyin:      "nǎishì",
		Senses:      []dicttypes.WordSense{},
	}
	s9 := "莲花"
	t9 := "蓮花"
	hw9 := dicttypes.Word{
		HeadwordId:  9,
		Simplified:  s9,
		Traditional: t9,
		Pinyin:      "liánhuā",
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  9,
				Simplified:  s9,
				Traditional: t9,
				Pinyin:      "liánhuā",
				English:     "lotus",
			},
		},
	}
	s10 := "北京"
	t10 := "\\N"
	hw10 := dicttypes.Word{
		HeadwordId:  10,
		Simplified:  s10,
		Traditional: t10,
		Pinyin:      "běijīng",
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  10,
				Simplified:  s10,
				Traditional: t10,
				Pinyin:      "běijīng",
				English:     "Beijing",
			},
		},
	}
	return map[string]*dicttypes.Word{
		s1:  &hw1,
		t1:  &hw1,
		s2:  &hw2,
		s3:  &hw3,
		t3:  &hw3,
		s4:  &hw4,
		s5:  &hw5,
		s6:  &hw6,
		t6:  &hw6,
		s7:  &hw7,
		s8:  &hw8,
		s9:  &hw9,
		t9:  &hw9,
		s10: &hw10,
	}
}

func (m mockDocFinder) FindDocsTermFreq(ctx context.Context, terms []string) ([]BM25Score, error) {
	return m.scores, nil
}

func (m mockDocFinder) FindDocsBigramFreq(ctx context.Context, bigrams []string) ([]BM25Score, error) {
	return m.scores, nil
}

func (m mockDocFinder) FindDocsTermCo(ctx context.Context, terms []string, col string) ([]BM25Score, error) {
	return m.scores, nil
}

func (m mockDocFinder) FindDocsBigramCo(ctx context.Context, bigrams []string, col string) ([]BM25Score, error) {
	return m.scores, nil
}

type mockTitleFinder struct {
	collections []Collection
	documents   []Document
	colMap      map[string]string
	docMap      map[string]DocInfo
}

func newMockTitleFinder(collections []Collection, documents []Document, colMap map[string]string, docMap map[string]DocInfo) TitleFinder {
	return mockTitleFinder{
		collections: collections,
		documents:   documents,
		colMap:      colMap,
		docMap:      docMap,
	}
}

func (m mockTitleFinder) CountCollections(ctx context.Context, query string) (int, error) {
	return 0, nil
}

func (m mockTitleFinder) FindCollections(ctx context.Context, query string) []Collection {
	return m.collections
}

func (m mockTitleFinder) FindDocsByTitle(ctx context.Context, query string) ([]Document, error) {
	return m.documents, nil
}

func (m mockTitleFinder) FindDocsByTitleInCol(ctx context.Context, query, col_gloss_file string) ([]Document, error) {
	return m.documents, nil
}

func (m mockTitleFinder) ColMap() map[string]string {
	return m.colMap
}

func (m mockTitleFinder) DocMap() map[string]DocInfo {
	return m.docMap
}

func TestBigrams(t *testing.T) {
	type test struct {
		name  string
		terms []string
		want  []string
	}
	tests := []test{
		{
			name:  "Zero",
			terms: []string{},
			want:  []string{},
		},
		{
			name:  "One",
			terms: []string{"one"},
			want:  []string{},
		},
		{
			name:  "Two",
			terms: []string{"one", "two"},
			want:  []string{"onetwo"},
		},
		{
			name:  "Three",
			terms: []string{"one", "two", "three"},
			want:  []string{"onetwo", "twothree"},
		},
	}
	for _, tc := range tests {
		got := Bigrams(tc.terms)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TestBigrams.%s with terms %v, got: %v, want %v", tc.name, tc.terms, got, tc.want)
		}
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
		WEIGHT[2]*doc.SimBitVector +
		WEIGHT[3]*doc.SimTitle
	expectedMin := 0.99 * similarity
	expectedMax := 1.01 * similarity
	if (expectedMin > simDoc.Similarity) ||
		(simDoc.Similarity > expectedMax) {
		t.Errorf("TestCombineByWeight: result out of expected range, got %.4f, want in range (%.4f, %.4f), details:  %v\n", simDoc.Similarity, expectedMin, expectedMax, simDoc)
	}
}

func TestFindDocuments(t *testing.T) {

	// Setup
	zeroDocFinder := newMockDocFinder([]BM25Score{})
	oneDocFinder := newMockDocFinder([]BM25Score{
		{
			Document:      "a.html",
			Collection:    "c.html",
			Score:         0.12345,
			BitVector:     1.0,
			ContainsTerms: "前",
		},
	})
	collections := []Collection{}
	documents := []Document{}
	colMap := map[string]string{}
	zeroDocMap := map[string]DocInfo{}
	zeroTitleFinder := newMockTitleFinder(collections, documents, colMap, zeroDocMap)
	oneDocMap := map[string]DocInfo{
		"a.html": {
			GlossFile: "a.html",
		},
	}
	oneTitleFinder := newMockTitleFinder(collections, documents, colMap, oneDocMap)
	ctx := context.Background()
	reverseIndex := mockReverseIndex{}
	emptyDict := map[string]*dicttypes.Word{}
	smallDict := mockSmallDict()

	// Test data
	type test struct {
		name           string
		query          string
		dict           map[string]*dicttypes.Word
		fullText       bool
		expectError    bool
		tdDocFinder    TermFreqDocFinder
		titleFinder    TitleFinder
		expectNoTerms  int
		expectNoSenses int
		expectNDoc     int
	}
	tests := []test{
		{
			name:           "Happy pass",
			query:          "Assembly",
			dict:           emptyDict,
			fullText:       false,
			tdDocFinder:    zeroDocFinder,
			titleFinder:    zeroTitleFinder,
			expectError:    false,
			expectNoTerms:  1,
			expectNoSenses: 0,
			expectNDoc:     0,
		},
		{
			name:           "Empty query",
			query:          "",
			dict:           emptyDict,
			fullText:       false,
			tdDocFinder:    zeroDocFinder,
			titleFinder:    zeroTitleFinder,
			expectError:    true,
			expectNoTerms:  0,
			expectNoSenses: 1,
			expectNDoc:     0,
		},
		{
			name:           "No word senses",
			query:          "hello",
			dict:           emptyDict,
			fullText:       false,
			tdDocFinder:    zeroDocFinder,
			titleFinder:    zeroTitleFinder,
			expectError:    false,
			expectNoTerms:  1,
			expectNoSenses: 0,
			expectNDoc:     0,
		},
		{
			name:           "One term query",
			query:          "前",
			dict:           smallDict,
			fullText:       true,
			tdDocFinder:    oneDocFinder,
			titleFinder:    oneTitleFinder,
			expectError:    false,
			expectNoTerms:  1,
			expectNoSenses: 0,
			expectNDoc:     1,
		},
		{
			name:           "Two term query",
			query:          "前者",
			dict:           smallDict,
			fullText:       true,
			tdDocFinder:    oneDocFinder,
			titleFinder:    oneTitleFinder,
			expectError:    false,
			expectNoTerms:  2,
			expectNoSenses: 0,
			expectNDoc:     1,
		},
	}

	for _, tc := range tests {
		dFinder := docFinder{
			tfDocFinder: tc.tdDocFinder,
			titleFinder: tc.titleFinder,
		}
		parser := NewQueryParser(tc.dict)
		qr, err := dFinder.FindDocuments(ctx, reverseIndex, parser, tc.query, tc.fullText)
		gotError := (err != nil)
		if tc.expectError != gotError {
			t.Errorf("TestFindDocuments.%s: expectError: %t vs got %t",
				tc.name, tc.expectError, gotError)
			if gotError {
				t.Errorf("TestFindDocuments.%s: unexpected error: %v", tc.name, err)
			}
			continue
		}
		if gotError {
			continue
		}
		gotNoTerms := len(qr.Terms)
		if gotNoTerms != tc.expectNoTerms {
			t.Errorf("TestFindDocuments.%s: gotNoTerms %d, want: %d, details: %v", tc.name, gotNoTerms, tc.expectNoTerms, qr.Terms)
		}
		if gotNoTerms > 0 {
			senses := qr.Terms[0].Senses
			gotNoSenses := len(senses)
			if gotNoSenses != tc.expectNoSenses {
				t.Errorf("TestFindDocuments.%s: gotNoSenses %d, want: %d, details: %v", tc.name, gotNoSenses, tc.expectNoSenses, senses)
			}
		}
		if qr.NumDocuments != tc.expectNDoc {
			t.Errorf("TestFindDocuments.%s: qr.NumDocuments %d, want: %d, details: %v", tc.name, qr.NumDocuments, tc.expectNDoc, qr.Documents)
		}
	}
}

func TestFindDocumentsInCol(t *testing.T) {
	df := newMockDocFinder([]BM25Score{})
	collections := []Collection{}
	documents := []Document{}
	colMap := map[string]string{}
	docMap := map[string]DocInfo{}
	titleFinder := newMockTitleFinder(collections, documents, colMap, docMap)
	dFinder := docFinder{
		tfDocFinder: df,
		titleFinder: titleFinder,
	}
	ctx := context.Background()
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
		qr, err := dFinder.FindDocumentsInCol(ctx, reverseIndex, parser, tc.query, tc.collection)
		if err != nil {
			if !tc.expectError {
				t.Errorf("TestFindDocumentsInCol.%s: with query %q unexpected error: %v", tc.name, tc.query, err)
			}
			continue
		}
		if err == nil && tc.expectError {
			t.Errorf("TestFindDocumentsInCol.%s: with query %q, no error but want one", tc.name, tc.query)
		}
		if len(qr.Terms) != tc.expectNumTerms {
			t.Errorf("TestFindDocumentsInCol.%s: with query %q, got %d num terms but want %d", tc.name, tc.query, len(qr.Terms), tc.expectNumTerms)
		}
	}
}

func TestMergeDocList(t *testing.T) {
	simDocMap := map[string]Document{}
	docList := []Document{}
	doc1 := Document{
		GlossFile:      "f1.html",
		CollectionFile: "collection.html",
		Title:          "Good doc by title",
		SimTitle:       1.0,
	}
	simDocMap[doc1.GlossFile] = doc1
	doc2 := Document{
		GlossFile:      "f2.html",
		CollectionFile: "collection.html",
		Title:          "Very Good doc",
		SimWords:       0.5,
		SimBigram:      1.5,
	}
	docList = append(docList, doc2)

	simDocMap2 := map[string]Document{}
	docList2 := []Document{}
	doc3 := Document{
		GlossFile:      "f1.html",
		CollectionFile: "collection.html",
		Title:          "Same Very Good doc",
		SimTitle:       1.0,
	}
	simDocMap2[doc3.GlossFile] = doc3
	doc4 := Document{
		CollectionFile: "collection.html",
		GlossFile:      "f2.html",
		Title:          "Reasonable by word frequ",
		SimWords:       1.6,
	}
	doc5 := Document{
		GlossFile:      "f1.html",
		CollectionFile: "collection.html",
		Title:          "Same Very Good doc",
		SimWords:       1.5,
		SimBigram:      1.5,
	}
	docList2 = append(docList2, doc4, doc5)

	collections := []Collection{}
	documents := []Document{}
	colMap := map[string]string{
		"collection.html": "collection.html",
	}
	docMap := map[string]DocInfo{
		"f1.html": {
			GlossFile:      "f1.html",
			CollectionFile: "collection.html",
		},
		"f2.html": {
			GlossFile:      "f2.html",
			CollectionFile: "collection.html",
		},
		"f3.html": {
			GlossFile:      "f3.html",
			CollectionFile: "collection.html",
		},
	}
	titleFinder := newMockTitleFinder(collections, documents, colMap, docMap)

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
			expectGlossFile: doc2.GlossFile,
		},
	}

	for _, tc := range tests {
		mergeDocList(titleFinder, tc.simDocMap, tc.docList)
		if tc.expectNum != len(simDocMap) {
			t.Errorf("TestMergeDocList.%s: expected %d vs got %d",
				tc.name, tc.expectNum, len(simDocMap))
			continue
		}
		docs := toSortedDocList(simDocMap)
		if tc.expectNumDocs != len(docs) {
			t.Errorf("TestMergeDocList.%s: expected docs %d vs got %d",
				tc.name, tc.expectNumDocs, len(docs))
			continue
		}
		result := docs[0]
		if result.GlossFile != tc.expectGlossFile {
			t.Errorf("TestMergeDocList.%s: got result.GlossFile %v, want: %v", tc.name, result.GlossFile, tc.expectGlossFile)
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
	collections := []Collection{}
	documents := []Document{}
	colMap := map[string]string{}
	docMap := map[string]DocInfo{}
	titleFinder := newMockTitleFinder(collections, documents, colMap, docMap)

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
	docs = toRelevantDocList(titleFinder, docs, queryTerms)
	expected := 2
	result := len(docs)
	if result == expected {
		t.Errorf("TestToRelevantDocList: expected %d, got, %d", expected,
			result)
	}
}

func TestToSortedDocList(t *testing.T) {
	similarDocMap1 := map[string]Document{}
	doc11 := Document{
		GlossFile: "f1.html",
		Title:     "Good doc",
		SimWords:  1.0,
		SimBigram: 1.0,
	}
	similarDocMap1[doc11.GlossFile] = doc11
	doc12 := Document{
		GlossFile: "f2.html",
		Title:     "Very Good doc",
		SimWords:  1.5,
		SimBigram: 1.5,
	}
	similarDocMap1[doc12.GlossFile] = doc12
	doc13 := Document{
		GlossFile: "f3.html",
		Title:     "Reasonable doc",
		SimWords:  0.5,
		SimBigram: 0.5,
	}
	similarDocMap1[doc13.GlossFile] = doc13

	similarDocMap2 := map[string]Document{}
	doc21 := Document{
		GlossFile: "f1.html",
		Title:     "Good doc",
		SimWords:  0.5,
		SimBigram: 1.0,
	}
	similarDocMap2[doc21.GlossFile] = doc21
	doc22 := Document{
		GlossFile: "f2.html",
		Title:     "Very Good doc",
		SimWords:  0.5,
		SimBigram: 1.5,
	}
	similarDocMap2[doc22.GlossFile] = doc22
	doc23 := Document{
		GlossFile: "f3.html",
		Title:     "Reasonable doc",
		SimWords:  0.5,
	}
	similarDocMap2[doc23.GlossFile] = doc23

	similarDocMap3 := map[string]Document{}
	doc31 := Document{
		GlossFile: "f1.html",
		Title:     "Good doc",
		SimWords:  0.5,
		SimBigram: 0.0,
		SimTitle:  1.0,
	}
	similarDocMap3[doc31.GlossFile] = doc31
	doc32 := Document{
		GlossFile: "f2.html",
		Title:     "Very Good doc",
		SimWords:  0.5,
		SimBigram: 0.5,
	}
	similarDocMap3[doc32.GlossFile] = doc32
	doc33 := Document{
		GlossFile: "f3.html",
		Title:     "Reasonable doc",
		SimWords:  0.5,
	}
	similarDocMap3[doc33.GlossFile] = doc33

	type test struct {
		name          string
		similarDocMap map[string]Document
		want          string
	}
	tests := []test{
		{
			name:          "Strong match for both terms and bigrams",
			similarDocMap: similarDocMap1,
			want:          doc12.GlossFile,
		},
		{
			name:          "Bigrams win",
			similarDocMap: similarDocMap2,
			want:          doc22.GlossFile,
		},
		{
			name:          "Similar title wins",
			similarDocMap: similarDocMap3,
			want:          doc31.GlossFile,
		},
	}
	for _, tc := range tests {
		docs := toSortedDocList(tc.similarDocMap)
		result := docs[0]
		if result.Similarity == 0.0 {
			t.Error("TestToSortedDocList: result.Similarity == 0.0")
		}
		if result.GlossFile != tc.want {
			t.Errorf("TestToSortedDocList %s: got, %s but want %s, details: %v", tc.name, result.GlossFile, tc.want, result)
		}
	}
}
