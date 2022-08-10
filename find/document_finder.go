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

// Functions for finding documents by full text search
package find

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/fulltext"
)

const (
	maxReturned   = 50
	minSimilarity = -4.75
	avDocLen      = 4497
	intercept     = -5.80042096 // From logistic regression
)

//  From logistic regression
var WEIGHT = []float64{0.3606522, 2.4427158, 3.84494291, 2.74137199} // [BM25 words, BM25 bigrams, bit vector, similar title]
// []float64{0.080, 2.327, 3.040} // old model, did not include similarity of title

// DocFinder finds documents.
type DocFinder interface {
	FindDocuments(ctx context.Context, dictSearcher dictionary.ReverseIndex,
		parser QueryParser, query string, advanced bool) (*QueryResults, error)
	FindDocumentsInCol(ctx context.Context, dictSearcher dictionary.ReverseIndex,
		parser QueryParser, query, col_gloss_file string) (*QueryResults, error)
}

// DocFinder finds documents.
type TermFreqDocFinder interface {
	FindDocsTermFreq(ctx context.Context, terms []string) ([]BM25Score, error)
	FindDocsBigramFreq(ctx context.Context, bigrams []string) ([]BM25Score, error)
	FindDocsTermCo(ctx context.Context, terms []string, col string) ([]BM25Score, error)
	FindDocsBigramCo(ctx context.Context, bigrams []string, col string) ([]BM25Score, error)
}

type BM25Score struct {
	Document      string
	Collection    string
	Score         float64
	BitVector     float64
	ContainsTerms string
}

type Collection struct {
	GlossFile, Title string
}

type Document struct {
	GlossFile, Title, CollectionFile, CollectionTitle, ContainsWords string
	ContainsBigrams                                                  string
	SimTitle, SimWords, SimBigram, SimBitVector, Similarity          float64
	ContainsTerms                                                    []string
	MatchDetails                                                     fulltext.MatchingText
	TitleCNMatch                                                     bool
}

type QueryResults struct {
	Query, CollectionFile        string
	NumCollections, NumDocuments int
	Collections                  []Collection
	Documents                    []Document
	Terms                        []TextSegment
	SimilarTerms                 []TextSegment
}

type docFinder struct {
	tfDocFinder TermFreqDocFinder
	titleFinder TitleFinder
}

// NewDocFinder creates and initializes an implementation of the DocFinder interface
func NewDocFinder(tfDocFinder TermFreqDocFinder, titleFinder TitleFinder) DocFinder {
	return &docFinder{
		tfDocFinder: tfDocFinder,
		titleFinder: titleFinder,
	}
}

type TitleFinder interface {
	CountCollections(ctx context.Context, query string) (int, error)
	FindCollections(ctx context.Context, query string) []Collection
	FindDocsByTitle(ctx context.Context, query string) ([]Document, error)
	FindDocsByTitleInCol(ctx context.Context, query, col_gloss_file string) ([]Document, error)
	ColMap() map[string]string
	DocMap() map[string]DocInfo
}

// convert2DocSim converts a BM25Score struct to a Document for term similarity
func convert4Term(scores []BM25Score) []Document {
	documents := []Document{}
	for _, s := range scores {
		d := Document{
			SimWords:       s.Score,
			SimBitVector:   s.BitVector,
			GlossFile:      s.Document,
			CollectionFile: s.Collection,
			ContainsWords:  s.ContainsTerms,
		}
		documents = append(documents, d)
	}
	return documents
}

// convert2DocSim converts a BM25Score struct to a Document for term similarity
func convert4Bigram(scores []BM25Score) []Document {
	documents := []Document{}
	for _, s := range scores {
		d := Document{
			SimBigram:       s.Score,
			GlossFile:       s.Document,
			CollectionFile:  s.Collection,
			ContainsBigrams: s.ContainsTerms,
		}
		documents = append(documents, d)
	}
	return documents
}

// For printing out retrieved document metadata
func (doc Document) String() string {
	return fmt.Sprintf("%s, %s, SimTitle %f, SimWords %f, SimBigram %f, "+
		"SimBitVector %f, Similarity %f, ContainsWords %s, ContainsBigrams %s"+
		", MatchDetails %v",
		doc.GlossFile, doc.CollectionFile, doc.SimTitle, doc.SimWords,
		doc.SimBigram, doc.SimBitVector, doc.Similarity, doc.ContainsWords,
		doc.ContainsBigrams, doc.MatchDetails)
}

// Compute the combined similarity based on logistic regression of document
// relevance for BM25 for words, BM25 for bigrams, and bit vector dot product.
// Raw BM25 values are scaled with 1.0 being the top value
func combineByWeight(doc Document, maxSimWords, maxSimBigram float64) Document {
	similarity := intercept
	if maxSimWords != 0.0 {
		similarity += WEIGHT[0] * doc.SimWords / maxSimWords
	}
	if maxSimBigram != 0.0 {
		similarity += WEIGHT[1] * doc.SimBigram / maxSimBigram
	}
	similarity += WEIGHT[2]*doc.SimBitVector + WEIGHT[3]*doc.SimTitle
	simDoc := Document{
		GlossFile:       doc.GlossFile,
		Title:           doc.Title,
		CollectionFile:  doc.CollectionFile,
		CollectionTitle: doc.CollectionTitle,
		SimTitle:        doc.SimTitle,
		SimWords:        doc.SimWords,
		SimBigram:       doc.SimBigram,
		SimBitVector:    doc.SimBitVector,
		Similarity:      similarity,
		ContainsWords:   doc.ContainsWords,
		ContainsBigrams: doc.ContainsBigrams,
		ContainsTerms:   doc.ContainsTerms,
		MatchDetails:    doc.MatchDetails,
	}
	return simDoc
}

// Bigrams constructs a slice of bigrams from pairs of terms
func Bigrams(terms []string) []string {
	b := []string{}
	if len(terms) < 2 {
		return b
	}
	for i := range terms {
		if i == 0 {
			continue
		}
		b = append(b, terms[i-1]+terms[i])
	}
	return b
}

// findDocuments find documents by both title and contents, and merge the lists
func (df docFinder) findDocuments(ctx context.Context, query string, terms []TextSegment, advanced bool) ([]Document, error) {
	log.Printf("findDocuments, enter: %s, advanced: %t", query, advanced)
	docs, err := df.titleFinder.FindDocsByTitle(ctx, query)
	if err != nil {
		return nil, err
	}
	log.Printf("findDocuments, by title len(docs): %s, %d", query, len(docs))
	queryTerms := toQueryTerms(terms)
	if !advanced {
		return docs, nil
	}

	// For more than one term find docs that are similar body and merge
	simDocMap := toSimilarDocMap(docs) // similarity = 1.0
	log.Printf("findDocuments, len(simDocMap): %d, query: %s", len(simDocMap), query)
	if df.tfDocFinder == nil {
		return nil, fmt.Errorf("full text search is not configured")
	}
	termScores, err := df.tfDocFinder.FindDocsTermFreq(ctx, queryTerms)
	if err != nil {
		return nil, err
	}
	simDocs := convert4Term(termScores)
	mergeDocList(df.titleFinder, simDocMap, simDocs)

	// If less than 2 terms then do not need to check bigrams
	if len(terms) < 2 {
		sortedDocs := toSortedDocList(simDocMap)
		log.Printf("findDocuments, < 2 len(sortedDocs): %s, %d", query, len(sortedDocs))
		relevantDocs := toRelevantDocList(df.titleFinder, sortedDocs, queryTerms)
		return relevantDocs, nil
	}
	qBigrams := Bigrams(queryTerms)
	bigramScores, err := df.tfDocFinder.FindDocsBigramFreq(ctx, qBigrams)
	if err != nil {
		return nil, err
	}
	moreDocs := convert4Bigram(bigramScores)
	mergeDocList(df.titleFinder, simDocMap, moreDocs)
	sortedDocs := toSortedDocList(simDocMap)
	log.Printf("findDocuments, len(sortedDocs): %s, %d", query, len(sortedDocs))
	relevantDocs := toRelevantDocList(df.titleFinder, sortedDocs, queryTerms)
	log.Printf("findDocuments, query: %s,len(relevantDocs):, %d", query, len(relevantDocs))
	return relevantDocs, nil
}

// findDocumentsInCol finds documents in a specific collection by both title and contents, and
// merge the lists
func (df docFinder) findDocumentsInCol(ctx context.Context, query string, terms []TextSegment,
	col_gloss_file string) ([]Document, error) {
	log.Printf("findDocumentsInCol, col_gloss_file, terms: %s, %v",
		col_gloss_file, terms)
	docs, err := df.titleFinder.FindDocsByTitleInCol(ctx, query, col_gloss_file)
	if err != nil {
		return nil, err
	}
	log.Printf("findDocumentsInCol, len(docs) by title: %d", len(docs))
	//log.Println("findDocumentsInCol, docs array by title: ", docs)
	queryTerms := toQueryTerms(terms)

	// For more than one term find docs that are similar body and merge
	simDocMap := toSimilarDocMap(docs) // similarity = 1.0
	//simDocs, err := findBodyBitVector(queryTerms)
	termScores, err := df.tfDocFinder.FindDocsTermCo(ctx, queryTerms, col_gloss_file)
	if err != nil {
		return nil, err
	}
	simDocs := convert4Term(termScores)
	//log.Println("findDocumentsInCol, len(simDocs) by word freq: ", len(simDocs))
	mergeDocList(df.titleFinder, simDocMap, simDocs)

	if len(terms) > 1 {
		// If there are 2 or more terms then check bigrams
		qBigrams := Bigrams(queryTerms)
		bigramScores, err := df.tfDocFinder.FindDocsBigramCo(ctx, qBigrams, col_gloss_file)
		//log.Println("findDocumentsInCol, len(simBGDocs) ", len(simBGDocs))
		if err != nil {
			return nil, fmt.Errorf("findDocumentsInCol, FindDocsBigramCo error: %v",
				err)
		}
		simBGDocs := convert4Bigram(bigramScores)
		mergeDocList(df.titleFinder, simDocMap, simBGDocs)
	}
	sortedDocs := toSortedDocList(simDocMap)
	log.Printf("findDocumentsInCol, len(sortedDocs): %d", len(sortedDocs))
	relevantDocs := toRelevantDocList(df.titleFinder, sortedDocs, queryTerms)
	log.Printf("findDocumentsInCol, len(relevantDocs): %s, %d", query,
		len(relevantDocs))
	return relevantDocs, nil
}

// FindDocuments returns a QueryResults object containing matching collections, documents,
// and dictionary words. For dictionary lookup, a text segment will
// contains the QueryText searched for and possibly a matching
// dictionary entry. There will only be matching dictionary entries for
// Chinese words in the dictionary. If there are no Chinese words in the query
// then the Chinese word senses matching the English or Pinyin will be included
// in the TextSegment.Senses field.
func (df docFinder) FindDocuments(ctx context.Context, reverseIndex dictionary.ReverseIndex, parser QueryParser, query string, advanced bool) (*QueryResults, error) {
	if query == "" {
		return nil, fmt.Errorf("FindDocuments, Empty query string")
	}
	terms := parser.ParseQuery(query)
	log.Printf("FindDocuments, query: %q with %d terms, advanced: %t", query, len(terms), advanced)
	if (len(terms) == 1) && (terms[0].DictEntry.HeadwordId == 0) {
		q := strings.ToLower(query)
		senses, err := reverseIndex.Find(ctx, q)
		if err != nil {
			return nil, err
		}
		log.Printf("FindDocuments, found senses %v matching reverse query: %s", senses, query)
		terms[0].Senses = senses
	}
	nCol := 0
	var err error
	collections := []Collection{}
	documents := []Document{}
	if df.titleFinder != nil {
		nCol, err = df.titleFinder.CountCollections(ctx, query)
		if err != nil {
			log.Printf("FindDocuments, error from CountCollections: %v", err)
		}
		collections = df.titleFinder.FindCollections(ctx, query)
		documents, err = df.findDocuments(ctx, query, terms, advanced)
	}
	if err != nil {
		return nil, fmt.Errorf("FindDocuments, error from findDocuments: %v", err)
	}
	nDoc := len(documents)
	log.Printf("FindDocuments, query %s, nTerms %d, collection %d, doc count %d: ", query, len(terms), nCol, nDoc)
	return &QueryResults{
		Query:          query,
		CollectionFile: "",
		NumCollections: nCol,
		NumDocuments:   nDoc,
		Collections:    collections,
		Documents:      documents,
		Terms:          terms,
		SimilarTerms:   nil,
	}, nil
}

// FindDocumentsInCol returns a QueryResults object containing matching collections, documents,
// and dictionary words within a specific collecion.
// For dictionary lookup, a text segment will
// contains the QueryText searched for and possibly a matching
// dictionary entry. There will only be matching dictionary entries for
// Chinese words in the dictionary. If there are no Chinese words in the query
// then the Chinese word senses matching the English or Pinyin will be included
// in the TextSegment.Senses field.
func (df docFinder) FindDocumentsInCol(ctx context.Context, reverseIndex dictionary.ReverseIndex, parser QueryParser, query, colFile string) (*QueryResults, error) {
	log.Printf("FindDocumentsInCol, Query %q, colFile: %s", query, colFile)
	if len(query) == 0 {
		return nil, fmt.Errorf("FindDocumentsInCol, Empty query string")
	}
	terms := parser.ParseQuery(query)
	if (len(terms) == 1) && (terms[0].DictEntry.HeadwordId == 0) {
		log.Printf("FindDocumentsInCol, Query with no Chinese, look for English and Pinyin matches query: %s", query)
		senses, err := reverseIndex.Find(ctx, terms[0].QueryText)
		if err != nil {
			return nil, err
		} else {
			terms[0].Senses = senses
		}
	}
	documents, err := df.findDocumentsInCol(ctx, query, terms, colFile)
	if err != nil {
		return nil, err
	}
	nDoc := len(documents)
	log.Printf("FindDocumentsInCol, query %s, nTerms %d, collection %d, doc count %d ",
		query, len(terms), 1, nDoc)
	return &QueryResults{
		Query:          query,
		CollectionFile: colFile,
		NumCollections: 1,
		NumDocuments:   nDoc,
		Collections:    []Collection{},
		Documents:      documents,
		Terms:          terms,
		SimilarTerms:   nil,
	}, err
}

// mergeDocList merges a list of documents with map of similar docs, adding the similarity
// for docs that are in both lists
func mergeDocList(df TitleFinder, simDocMap map[string]Document, docList []Document) {
	log.Printf("mergeDocList, len(simDocMap) = %d len(docList) = %d", len(simDocMap), len(docList))
	for _, simDoc := range docList {
		sDoc, ok := simDocMap[simDoc.GlossFile]
		colMap := df.ColMap()
		docMap := df.DocMap()
		if ok {
			sDoc.SimTitle += simDoc.SimTitle
			sDoc.SimWords += simDoc.SimWords
			sDoc.SimBigram += simDoc.SimBigram
			sDoc.SimBitVector += simDoc.SimBitVector
			if sDoc.ContainsWords == "" {
				sDoc.ContainsWords = simDoc.ContainsWords
			} else {
				sDoc.ContainsWords += "," + simDoc.ContainsWords
			}
			if sDoc.ContainsBigrams == "" {
				sDoc.ContainsBigrams = simDoc.ContainsBigrams
			} else {
				sDoc.ContainsBigrams += "," + simDoc.ContainsBigrams
			}
			simDocMap[simDoc.GlossFile] = sDoc
		} else {
			colTitle, ok1 := colMap[simDoc.CollectionFile]
			document, ok2 := docMap[simDoc.GlossFile]
			if ok1 && ok2 {
				doc := Document{CollectionFile: simDoc.CollectionFile,
					CollectionTitle: colTitle,
					GlossFile:       simDoc.GlossFile,
					Title:           document.Title,
					SimTitle:        simDoc.SimTitle,
					SimWords:        simDoc.SimWords,
					SimBigram:       simDoc.SimBigram,
					SimBitVector:    simDoc.SimBitVector,
					Similarity:      simDoc.Similarity,
					ContainsWords:   simDoc.ContainsWords,
					ContainsBigrams: simDoc.ContainsBigrams,
				}
				simDocMap[simDoc.GlossFile] = doc
			} else if ok2 {
				log.Printf("mergeDocList, collection title %s not found: %v", simDoc.CollectionFile, simDoc)
				doc := Document{CollectionFile: "",
					CollectionTitle: "",
					GlossFile:       simDoc.GlossFile,
					Title:           document.Title,
					SimTitle:        simDoc.SimTitle,
					SimWords:        simDoc.SimWords,
					SimBigram:       simDoc.SimBigram,
					SimBitVector:    simDoc.SimBitVector,
					Similarity:      simDoc.Similarity,
					ContainsWords:   simDoc.ContainsWords,
					ContainsBigrams: simDoc.ContainsBigrams,
				}
				simDocMap[simDoc.GlossFile] = doc
			} else {
				log.Printf("mergeDocList, doc title %s not found: %v", simDoc.GlossFile, simDoc)
				simDocMap[simDoc.GlossFile] = simDoc
			}
		}
	}
	log.Printf("mergeDocList, exit with len(simDocMap) = %d len(docList) = %d", len(simDocMap), len(docList))
}

// setMatchDetails organizes the contains terms found of the document in a form
// that helps the user.
//
// doc.ContainsWords is a contained list of terms found in the query and doc
// doc.ContainsBigrams is a contained list of bigrams found in the query and doc
// doc.ContainsTerms is a list of terms found both in the query and the doc
// sorted in the same order as the query terms with words merged to bigrams
func setMatchDetails(doc Document, terms []string, docMatch fulltext.DocMatch) Document {
	log.Printf("setMatchDetails: terms %v, doc %s, snippet: %s", terms, doc.GlossFile, docMatch.MT.Snippet)
	containsTems := []string{}
	for i, term := range terms {
		// log.Printf("sortContainsWords: i = %d", i)
		bigram := ""
		if i > 0 {
			bigram = terms[i-1] + terms[i]
		}
		if (i > 0) && strings.Contains(doc.ContainsBigrams, bigram) {
			j := len(containsTems)
			if (j > 0) && strings.Contains(bigram, containsTems[j-1]) {
				containsTems[j-1] = bigram
			} else {
				containsTems = append(containsTems, bigram)
			}
		} else if strings.Contains(doc.ContainsWords, term) {
			containsTems = append(containsTems, term)
		}
	}
	doc.ContainsTerms = containsTems
	doc.MatchDetails = docMatch.MT
	return doc
}

// Sort firstly based on longest matching substring, then on similarity
func sortMatchingSubstr(docs []Document) {
	sort.Slice(docs, func(i, j int) bool {
		l1 := len(docs[i].MatchDetails.LongestMatch)
		l2 := len(docs[j].MatchDetails.LongestMatch)
		if l1 != l2 {
			return l1 > l2
		}
		return docs[i].Similarity > docs[j].Similarity
	})
}

// Filter documents that are not similar
func toRelevantDocList(df TitleFinder, docs []Document, terms []string) []Document {
	if len(docs) < 1 {
		return docs
	}
	keys := []string{}
	docMap := df.DocMap()
	for _, doc := range docs {
		d, ok := df.DocMap()[doc.GlossFile]
		if !ok {
			log.Printf("find.toRelevantDocList could not find %s", doc.GlossFile)
			continue
		}
		keys = append(keys, d.CorpusFile)
	}
	docMatches := fulltext.GetMatches(keys, terms)
	relDocs := []Document{}
	for _, doc := range docs {
		// log.Printf("toRelevantDocList, check Similarity %f, min %f, gloss %s, title: %s", doc.Similarity, minSimilarity, doc.GlossFile,doc.Title)
		d, ok := docMap[doc.GlossFile]
		if !ok {
			log.Printf("find.toRelevantDocList 2 could not find %s", doc.GlossFile)
			continue
		}
		docMatch := docMatches[d.CorpusFile]
		doc = setMatchDetails(doc, terms, docMatch)
		if doc.Similarity < minSimilarity {
			log.Printf("find.toRelevantDocList doc %s Similarity %4f < minSimilarity %4f, returning %d docs", doc.GlossFile, doc.Similarity, minSimilarity, len(relDocs))
			return relDocs
		}
		relDocs = append(relDocs, doc)
	}
	sortMatchingSubstr(relDocs)
	return relDocs
}

// Convert list to a map of similar docs with similarity set to 1.0
func toSimilarDocMap(docs []Document) map[string]Document {
	similarDocMap := map[string]Document{}
	for _, doc := range docs {
		// log.Printf("find.toSimilarDocMap find %s, SimTitle = %4f", doc.GlossFile, doc.SimTitle)
		simDoc := Document{
			GlossFile:       doc.GlossFile,
			Title:           doc.Title,
			CollectionFile:  doc.CollectionFile,
			CollectionTitle: doc.CollectionTitle,
			SimTitle:        doc.SimTitle,
			SimWords:        doc.SimWords,
			SimBigram:       doc.SimBigram,
			SimBitVector:    doc.SimBitVector,
			ContainsWords:   doc.ContainsWords,
			ContainsBigrams: doc.ContainsBigrams,
			Similarity:      doc.Similarity,
		}
		similarDocMap[doc.GlossFile] = simDoc
	}
	return similarDocMap
}

// Convert a map of similar docs into a sorted list, and truncate
func toSortedDocList(similarDocMap map[string]Document) []Document {
	docs := []Document{}
	if len(similarDocMap) < 1 {
		return docs
	}
	for _, similarDoc := range similarDocMap {
		docs = append(docs, similarDoc)
	}
	// First sort by BM25 bigrams
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].SimBigram > docs[j].SimBigram
	})
	maxSimWords := docs[0].SimWords
	maxSimBigram := docs[0].SimBigram
	simDocs := []Document{}
	for _, doc := range docs {
		simDoc := combineByWeight(doc, maxSimWords, maxSimBigram)
		// log.Printf("find.toSortedDocList doc %s SimTitle = %.4f, Similarity = %.4f", doc.GlossFile, simDoc.SimTitle, simDoc.Similarity)
		simDocs = append(simDocs, simDoc)
	}
	// Sort again by combined similarity
	sort.Slice(simDocs, func(i, j int) bool {
		return simDocs[i].Similarity > simDocs[j].Similarity
	})
	if len(simDocs) > maxReturned {
		log.Printf("find.toSortedDocList got %d results, truncating to = %d with min Similarity = %.4f", len(simDocs), maxReturned, simDocs[maxReturned-1].Similarity)
		return simDocs[:maxReturned]
	}
	return simDocs
}
