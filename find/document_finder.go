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
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/fulltext"
)

const (
	maxReturned   = 50
	minSimilarity = -4.75
	avDocLen      = 4497
	intercept     = -4.75 // From logistic regression
)

//  From logistic regression
var WEIGHT = []float64{0.080, 2.327, 3.040} // [BM25 words, BM25 bigrams, bit vector]

// DocFinder finds documents.
type DocFinder interface {
	FindDocuments(ctx context.Context, dictSearcher dictionary.ReverseIndex,
		parser QueryParser, query string, advanced bool) (*QueryResults, error)
	FindDocumentsInCol(ctx context.Context, dictSearcher dictionary.ReverseIndex,
		parser QueryParser, query, col_gloss_file string) (*QueryResults, error)
	GetColMap() map[string]string
	Inititialized() bool
}

// databaseDocFinder holds stateful items needed for text search in database.
type databaseDocFinder struct {
	initialized                                                        bool
	countColStmt                                                       *sql.Stmt
	database                                                           *sql.DB
	colMap                                                             map[string]string
	docMap                                                             map[string]DocInfo
	docListStmt                                                        *sql.Stmt
	findAllTitlesStmt, findAllColTitlesStmt                            *sql.Stmt
	findColStmt, findDocStmt, findDocInColStmt, findWordStmt           *sql.Stmt
	simBM251Stmt, simBM252Stmt, simBM253Stmt, simBM254Stmt             *sql.Stmt
	simBM255Stmt, simBM256Stmt                                         *sql.Stmt
	simBM25Col1Stmt, simBM25Col2Stmt, simBM25Col3Stmt, simBM25Col4Stmt *sql.Stmt
	simBM25Col5Stmt, simBM25Col6Stmt                                   *sql.Stmt
	simBigram1Stmt, simBigram2Stmt, simBigram3Stmt, simBigram4Stmt     *sql.Stmt
	simBigram5Stmt                                                     *sql.Stmt
	simBgCol1Stmt, simBgCol2Stmt, simBgCol3Stmt, simBgCol4Stmt         *sql.Stmt
	simBgCol5Stmt                                                      *sql.Stmt
	avdl                                                               int // The average document length
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

// Create and initialize an implementation of the DocFinder interface
func NewDocFinder(ctx context.Context,
	database *sql.DB,
	docMap map[string]DocInfo) DocFinder {
	df := databaseDocFinder{
		database:    database,
		docMap:      docMap,
		initialized: false,
	}
	if database != nil {
		err := df.initFind(ctx)
		if err != nil {
			log.Printf("NewDocFinder, Error: %v", err)
			return &df
		}
	}
	log.Println("NewDocFinder initialized")
	df.initialized = true
	return &df
}

func (df databaseDocFinder) Inititialized() bool {
	return df.initialized
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

// Cache the details of all collecitons by target file name
func (df *databaseDocFinder) cacheColDetails(ctx context.Context) map[string]string {
	if df.findAllColTitlesStmt == nil {
		return map[string]string{}
	}
	df.colMap = map[string]string{}
	results, err := df.findAllColTitlesStmt.QueryContext(ctx)
	if err != nil {
		log.Printf("cacheColDetails, Error for query: %v", err)
		return df.colMap
	}
	defer results.Close()

	for results.Next() {
		var gloss_file, title string
		results.Scan(&gloss_file, &title)
		df.colMap[gloss_file] = title
	}
	log.Printf("cacheColDetails, len(colMap) = %d", len(df.colMap))
	return df.colMap
}

func (df databaseDocFinder) GetColMap() map[string]string {
	return df.colMap
}

// Compute the combined similarity based on logistic regression of document
// relevance for BM25 for words, BM25 for bigrams, and bit vector dot product.
// Raw BM25 values are scaled with 1.0 being the top value
func combineByWeight(doc Document, maxSimWords, maxSimBigram float64) Document {
	similarity := minSimilarity
	if maxSimWords != 0.0 && maxSimBigram != 0.0 {
		similarity = intercept +
			WEIGHT[0]*doc.SimWords/maxSimWords +
			WEIGHT[1]*doc.SimBigram/maxSimBigram +
			WEIGHT[2]*doc.SimBitVector
	}
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

func (df databaseDocFinder) countCollections(ctx context.Context, query string) (int, error) {
	var count int
	results, err := df.countColStmt.QueryContext(ctx, "%"+query+"%")
	if err != nil {
		return 0, fmt.Errorf("countCollections: query %s, error: %v", query, err)
	}
	results.Next()
	results.Scan(&count)
	results.Close()
	return count, nil
}

// findBodyBM25 searches the corpus for document bodies most similar using a BM25 model.
//  Param: terms - The decomposed query string with 0 < num elements < 7
func (df databaseDocFinder) findBodyBM25(ctx context.Context, terms []string) ([]Document, error) {
	log.Println("findBodyBM25, terms = ", terms)
	var results *sql.Rows
	var err error
	if len(terms) == 1 {
		results, err = df.simBM251Stmt.QueryContext(ctx, df.avdl, terms[0])
	} else if len(terms) == 2 {
		results, err = df.simBM252Stmt.QueryContext(ctx, df.avdl, terms[0], terms[1])
	} else if len(terms) == 3 {
		results, err = df.simBM253Stmt.QueryContext(ctx, df.avdl, terms[0], terms[1],
			terms[2])
	} else if len(terms) == 4 {
		results, err = df.simBM254Stmt.QueryContext(ctx, df.avdl, terms[0], terms[1],
			terms[2], terms[3])
	} else if len(terms) == 5 {
		results, err = df.simBM255Stmt.QueryContext(ctx, df.avdl, terms[0], terms[1],
			terms[2], terms[3], terms[4])
	} else {
		// Ignore arguments beyond the first six
		results, err = df.simBM256Stmt.QueryContext(ctx, df.avdl, terms[0], terms[1],
			terms[2], terms[3], terms[4], terms[5])
	}
	if err != nil {

		return nil, fmt.Errorf("findBodyBM25, Error for query %v: %v", terms, err)
	}
	simSlice := []Document{}
	for results.Next() {
		docSim := Document{}
		results.Scan(&docSim.SimWords, &docSim.SimBitVector,
			&docSim.ContainsWords, &docSim.CollectionFile, &docSim.GlossFile)
		//log.Println("findBodyBM25, Similarity, Document = ", docSim)
		simSlice = append(simSlice, docSim)
	}
	return simSlice, nil
}

// Search the corpus for document bodies most similar using a BM25 model in a
// specific collection.
//  Param: terms - The decomposed query string with 1 < num elements < 7
func (df databaseDocFinder) findBodyBM25InCol(ctx context.Context, terms []string,
	col_gloss_file string) ([]Document, error) {
	log.Println("findBodyBM25InCol, terms = ", terms)
	var results *sql.Rows
	var err error
	if len(terms) == 1 {
		results, err = df.simBM25Col1Stmt.QueryContext(ctx, df.avdl, terms[0],
			col_gloss_file)
	} else if len(terms) == 2 {
		results, err = df.simBM25Col2Stmt.QueryContext(ctx, df.avdl, terms[0],
			terms[1], col_gloss_file)
	} else if len(terms) == 3 {
		results, err = df.simBM25Col3Stmt.QueryContext(ctx, df.avdl, terms[0],
			terms[1], terms[2], col_gloss_file)
	} else if len(terms) == 4 {
		results, err = df.simBM25Col4Stmt.QueryContext(ctx, df.avdl, terms[0],
			terms[1], terms[2], terms[3], col_gloss_file)
	} else if len(terms) == 5 {
		results, err = df.simBM25Col5Stmt.QueryContext(ctx, df.avdl, terms[0],
			terms[1], terms[2], terms[3], terms[4], col_gloss_file)
	} else {
		// Ignore arguments beyond the first six
		results, err = df.simBM25Col6Stmt.QueryContext(ctx, df.avdl, terms[0],
			terms[1], terms[2], terms[3], terms[4], terms[5],
			col_gloss_file)
	}
	if err != nil {
		return nil, fmt.Errorf("findBodyBM25InCol, Error for query %v: %v", terms, err)
	}
	simSlice := []Document{}
	for results.Next() {
		docSim := Document{}
		docSim.CollectionFile = col_gloss_file
		results.Scan(&docSim.SimWords, &docSim.SimBitVector,
			&docSim.ContainsWords, &docSim.GlossFile)
		//log.Println("findBodyBM25InCol, Similarity, Document = ", docSim)
		simSlice = append(simSlice, docSim)
	}
	return simSlice, nil
}

// Search the corpus for document bodies most similar using bigrams with a BM25
// model.
//  Param: terms - The decomposed query string with 1 < num elements < 7
func (df databaseDocFinder) findBodyBigram(ctx context.Context, terms []string) ([]Document, error) {
	log.Println("findBodyBigram, terms = ", terms)
	var results *sql.Rows
	var err error
	if len(terms) < 2 {
		return nil, fmt.Errorf("findBodyBigram, too few arguments, len(terms) < 2: %d", len(terms))
	} else if len(terms) == 2 {
		bigram1 := terms[0] + terms[1]
		results, err = df.simBigram1Stmt.QueryContext(ctx, df.avdl, bigram1)
	} else if len(terms) == 3 {
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		results, err = df.simBigram2Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2)
	} else if len(terms) == 4 {
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		bigram3 := terms[2] + terms[3]
		results, err = df.simBigram3Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			bigram3)
	} else if len(terms) == 5 {
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		bigram3 := terms[2] + terms[3]
		bigram4 := terms[3] + terms[4]
		results, err = df.simBigram4Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			bigram3, bigram4)
	} else {
		// Ignore arguments beyond the first six
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		bigram3 := terms[2] + terms[3]
		bigram4 := terms[3] + terms[4]
		bigram5 := terms[4] + terms[5]
		results, err = df.simBigram5Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			bigram3, bigram4, bigram5)
	}
	if err != nil {
		return nil, fmt.Errorf("findBodyBigram, Error for query %v: %v", terms, err)
	}
	simSlice := []Document{}
	for results.Next() {
		docSim := Document{}
		results.Scan(&docSim.SimBigram, &docSim.ContainsBigrams,
			&docSim.CollectionFile, &docSim.GlossFile)
		//log.Println("findBodyBigram, Similarity, Document = ", docSim)
		simSlice = append(simSlice, docSim)
	}
	return simSlice, nil
}

// findBodyBgInCol searches the corpus for document bodies most similar using bigrams with a BM25
// model within a specific collection
//  Param: terms - The decomposed query string with 1 < num elements < 7
func (df databaseDocFinder) findBodyBgInCol(ctx context.Context, terms []string,
	col_gloss_file string) ([]Document, error) {
	log.Println("findBodyBgInCol, terms = ", terms)
	var results *sql.Rows
	var err error
	if len(terms) < 2 {
		return nil, fmt.Errorf("findBodyBgInCol, too few arguments, len(terms) < 2: %d", len(terms))
	} else if len(terms) == 2 {
		if df.simBgCol1Stmt == nil {
			return []Document{}, nil
		}
		bigram1 := terms[0] + terms[1]
		results, err = df.simBgCol1Stmt.QueryContext(ctx, df.avdl, bigram1,
			col_gloss_file)
	} else if len(terms) == 3 {
		if df.simBgCol2Stmt == nil {
			return []Document{}, nil
		}
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		results, err = df.simBgCol2Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			col_gloss_file)
	} else if len(terms) == 4 {
		if df.simBgCol3Stmt == nil {
			return []Document{}, nil
		}
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		bigram3 := terms[2] + terms[3]
		results, err = df.simBgCol3Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			bigram3, col_gloss_file)
	} else if len(terms) == 5 {
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		bigram3 := terms[2] + terms[3]
		bigram4 := terms[3] + terms[4]
		results, err = df.simBgCol4Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			bigram3, bigram4, col_gloss_file)
	} else {
		// Ignore arguments beyond the first six
		bigram1 := terms[0] + terms[1]
		bigram2 := terms[1] + terms[2]
		bigram3 := terms[2] + terms[3]
		bigram4 := terms[3] + terms[4]
		bigram5 := terms[4] + terms[5]
		results, err = df.simBgCol5Stmt.QueryContext(ctx, df.avdl, bigram1, bigram2,
			bigram3, bigram4, bigram5, col_gloss_file)
	}
	if err != nil {
		return nil, fmt.Errorf("findBodyBgInCol, Error for query %v: %v", terms, err)
	}
	simSlice := []Document{}
	for results.Next() {
		docSim := Document{}
		docSim.CollectionFile = col_gloss_file
		results.Scan(&docSim.SimBigram, &docSim.ContainsBigrams,
			&docSim.GlossFile)
		//log.Println("findBodyBgInCol, Similarity, Document = ", docSim)
		simSlice = append(simSlice, docSim)
	}
	return simSlice, nil
}

func (df databaseDocFinder) findCollections(ctx context.Context, query string) []Collection {
	results, err := df.findColStmt.QueryContext(ctx, "%"+query+"%")
	if err != nil {
		log.Printf("findCollections, Error for query %v: %v", query, err)
		return nil
	}
	defer results.Close()
	collections := []Collection{}
	for results.Next() {
		col := Collection{}
		results.Scan(&col.Title, &col.GlossFile)
		collections = append(collections, col)
	}
	return collections
}

// findDocsByTitle find documents based on a match in title
func (df databaseDocFinder) findDocsByTitle(ctx context.Context, query string) ([]Document, error) {
	results, err := df.findDocStmt.QueryContext(ctx, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("findDocsByTitle, Error for query %v: %v", query, err)
	}
	defer results.Close()

	documents := []Document{}
	for results.Next() {
		doc := Document{}
		results.Scan(&doc.Title, &doc.GlossFile, &doc.CollectionFile,
			&doc.CollectionTitle)
		doc.SimTitle = 1.0
		documents = append(documents, doc)
	}
	return documents, nil
}

// findDocsByTitleInCol find documents based on a match in title within a specific collection
func (df databaseDocFinder) findDocsByTitleInCol(ctx context.Context,
	query, col_gloss_file string) ([]Document, error) {
	results, err := df.findDocInColStmt.QueryContext(ctx, "%"+query+"%",
		col_gloss_file)
	if err != nil {
		return nil, fmt.Errorf("findDocsByTitleInCol, Error for query %v: %v", query, err)
	}
	defer results.Close()

	documents := []Document{}
	for results.Next() {
		doc := Document{}
		doc.CollectionFile = col_gloss_file
		results.Scan(&doc.Title, &doc.GlossFile, &doc.CollectionTitle)
		doc.SimTitle = 1.0
		//log.Println("findDocsByTitleInCol, doc: ", doc)
		documents = append(documents, doc)
	}
	return documents, nil
}

// findDocuments find documents by both title and contents, and merge the lists
func (df databaseDocFinder) findDocuments(
	ctx context.Context,
	query string,
	terms []TextSegment,
	advanced bool) ([]Document, error) {
	log.Printf("findDocuments, enter: %s", query)
	docs, err := df.findDocsByTitle(ctx, query)
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
	log.Printf("findDocuments, len(docMap): %s, %d", query, len(simDocMap))
	simDocs, err := df.findBodyBM25(ctx, queryTerms)
	if err != nil {
		return nil, err
	}
	mergeDocList(df, simDocMap, simDocs)

	// If less than 2 terms then do not need to check bigrams
	if len(terms) < 2 {
		sortedDocs := toSortedDocList(simDocMap)
		log.Printf("findDocuments, < 2 len(sortedDocs): %s, %d", query,
			len(sortedDocs))
		relevantDocs := toRelevantDocList(df, sortedDocs, queryTerms)
		return relevantDocs, nil
	}
	moreDocs, err := df.findBodyBigram(ctx, queryTerms)
	if err != nil {
		return nil, err
	}
	mergeDocList(df, simDocMap, moreDocs)
	sortedDocs := toSortedDocList(simDocMap)
	log.Printf("findDocuments, len(sortedDocs): %s, %d", query, len(sortedDocs))
	relevantDocs := toRelevantDocList(df, sortedDocs, queryTerms)
	log.Printf("findDocuments, len(relevantDocs): %s, %d", query, len(relevantDocs))
	return relevantDocs, nil
}

// findDocumentsInCol finds documents in a specific collection by both title and contents, and
// merge the lists
func (df databaseDocFinder) findDocumentsInCol(ctx context.Context, query string, terms []TextSegment,
	col_gloss_file string) ([]Document, error) {
	log.Printf("findDocumentsInCol, col_gloss_file, terms: %s, %v",
		col_gloss_file, terms)
	docs, err := df.findDocsByTitleInCol(ctx, query, col_gloss_file)
	if err != nil {
		return nil, err
	}
	log.Printf("findDocumentsInCol, len(docs) by title: %d", len(docs))
	//log.Println("findDocumentsInCol, docs array by title: ", docs)
	queryTerms := toQueryTerms(terms)

	// For more than one term find docs that are similar body and merge
	simDocMap := toSimilarDocMap(docs) // similarity = 1.0
	//simDocs, err := findBodyBitVector(queryTerms)
	simDocs, err := df.findBodyBM25InCol(ctx, queryTerms, col_gloss_file)
	if err != nil {
		return nil, err
	}
	//log.Println("findDocumentsInCol, len(simDocs) by word freq: ", len(simDocs))
	mergeDocList(df, simDocMap, simDocs)

	if len(terms) > 1 {
		// If there are 2 or more terms then check bigrams
		simBGDocs, err := df.findBodyBgInCol(ctx, queryTerms, col_gloss_file)
		//log.Println("findDocumentsInCol, len(simBGDocs) ", len(simBGDocs))
		if err != nil {
			return nil, fmt.Errorf("findDocumentsInCol, findBodyBgInCol error: %v",
				err)
		}
		mergeDocList(df, simDocMap, simBGDocs)
	}
	sortedDocs := toSortedDocList(simDocMap)
	log.Printf("findDocumentsInCol, len(sortedDocs): %d", len(sortedDocs))
	relevantDocs := toRelevantDocList(df, sortedDocs, queryTerms)
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
func (df databaseDocFinder) FindDocuments(ctx context.Context, reverseIndex dictionary.ReverseIndex,
	parser QueryParser, query string,
	advanced bool) (*QueryResults, error) {
	if query == "" {
		return nil, fmt.Errorf("FindDocuments, Empty query string")
	}
	terms := parser.ParseQuery(query)
	log.Printf("FindDocuments, got: %d terms", len(terms))
	if (len(terms) == 1) && (terms[0].DictEntry.HeadwordId == 0) {
		log.Printf("FindDocuments, no Chinese in query, look for English and Pinyin matches: %s", query)
		senses, err := reverseIndex.Find(ctx, terms[0].QueryText)
		log.Printf("FindDocuments, found senses %v matching reverse query: %s", senses, query)
		if err != nil {
			return nil, err
		}
		terms[0].Senses = senses
	}

	if df.database == nil {
		return &QueryResults{
			Query:          query,
			CollectionFile: "",
			NumCollections: 0,
			NumDocuments:   0,
			Collections:    nil,
			Documents:      nil,
			Terms:          terms,
			SimilarTerms:   nil,
		}, nil
	}

	nCol, err := df.countCollections(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("FindDocuments, error from countCollections: %v", err)
	}
	collections := df.findCollections(ctx, query)
	documents, err := df.findDocuments(ctx, query, terms, advanced)
	if err != nil {
		return nil, fmt.Errorf("FindDocuments, error from findDocuments: %v", err)
	}
	nDoc := len(documents)
	log.Printf("FindDocuments, query %s, nTerms %d, collection %d, doc count %d: ",
		query, len(terms), nCol, nDoc)
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
func (df databaseDocFinder) FindDocumentsInCol(ctx context.Context,
	reverseIndex dictionary.ReverseIndex, parser QueryParser, query,
	col_gloss_file string) (*QueryResults, error) {
	if query == "" {
		return nil, fmt.Errorf("FindDocumentsInCol, Empty query string")
	}
	terms := parser.ParseQuery(query)
	if (len(terms) == 1) && (terms[0].DictEntry.HeadwordId == 0) {
		log.Printf("FindDocumentsInCol, Query does not contain Chinese, "+
			"look for English and Pinyin matches query: %s", query)
		senses, err := reverseIndex.Find(ctx, terms[0].QueryText)
		if err != nil {
			return nil, err
		} else {
			terms[0].Senses = senses
		}
	}
	documents, err := df.findDocumentsInCol(ctx, query, terms, col_gloss_file)
	if err != nil {
		return nil, err
	}
	nDoc := len(documents)
	log.Printf("FindDocumentsInCol, query %s, nTerms %d, collection %d, doc count %d ",
		query, len(terms), 1, nDoc)
	return &QueryResults{
		Query:          query,
		CollectionFile: col_gloss_file,
		NumCollections: 1,
		NumDocuments:   nDoc,
		Collections:    []Collection{},
		Documents:      documents,
		Terms:          terms,
		SimilarTerms:   nil,
	}, err
}

// Open database connection and prepare statements. Allows for re-initialization
// at most every minute
func (df *databaseDocFinder) initFind(ctx context.Context) error {
	log.Println("find.initFind Initializing document_finder")
	df.avdl = config.GetEnvIntValue("AVG_DOC_LEN", avDocLen)
	err := df.initStatements(ctx)
	if err != nil {
		conString := config.DBConfig()
		return fmt.Errorf("find.initFind: got error with conString %s: \n%v", conString, err)
	}
	if err != nil {
		return err
	}
	df.colMap = df.cacheColDetails(ctx)
	return nil
}

func (df *databaseDocFinder) initStatements(ctx context.Context) error {
	var err error
	if df.database == nil {
		return fmt.Errorf("initStatements, database is nil")
	}

	df.docListStmt, err = df.database.PrepareContext(ctx,
		"SELECT plain_text_file, gloss_file "+
			"FROM document")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for docListStmt: %v", err)
	}

	df.findColStmt, err = df.database.PrepareContext(ctx,
		"SELECT title, gloss_file FROM collection WHERE title LIKE ? LIMIT 20")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error preparing collection stmt: %v",
			err)
	}

	df.countColStmt, err = df.database.PrepareContext(ctx,
		"SELECT count(title) FROM collection WHERE title LIKE ?")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error preparing cstmt: %v", err)
	}

	// Search documents by title substring
	df.findDocStmt, err = df.database.PrepareContext(ctx,
		"SELECT title, gloss_file, col_gloss_file, col_title "+
			"FROM document "+
			"WHERE col_plus_doc_title LIKE ? LIMIT 20")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error preparing dstmt: %v", err)
	}

	// Search documents by title substring within a collection
	df.findDocInColStmt, err = df.database.PrepareContext(ctx,
		"SELECT title, gloss_file, col_title "+
			"FROM document "+
			"WHERE col_plus_doc_title LIKE ? "+
			"AND col_gloss_file = ? "+
			"LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error preparing dstmt: %v", err)
	}

	df.findWordStmt, err = df.database.PrepareContext(ctx,
		"SELECT simplified, traditional, pinyin, headword FROM words WHERE "+
			"simplified = ? OR traditional = ? LIMIT 1")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error preparing fwstmt: %v", err)
	}

	// Document similarity with BM25 using 1-6 terms, k = 1.5, b = 0.65
	df.simBM251Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" collection, document "+
			"FROM word_freq_doc "+
			"WHERE word = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM251Stmt: %v", err)
	}

	df.simBM252Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 2.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" collection, document "+
			"FROM word_freq_doc "+
			"WHERE word = ? OR word = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM252Stmt: %s", err)
	}

	df.simBM253Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 3.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" collection, document "+
			"FROM word_freq_doc "+
			"WHERE word = ? OR word = ? OR word = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM253Stmt: %v", err)
	}

	df.simBM254Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 4.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" collection, document "+
			"FROM word_freq_doc "+
			"WHERE word = ? OR word = ? OR word = ? OR word = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM254Stmt: %v", err)
	}

	df.simBM255Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 5.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" collection, document "+
			"FROM word_freq_doc "+
			"WHERE word = ? OR word = ? OR word = ? OR word = ? OR word = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM255Stmt: %v", err)
	}

	df.simBM256Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 5.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" collection, document "+
			"FROM word_freq_doc "+
			"WHERE word = ? OR word = ? OR word = ? OR word = ? OR word = ? "+
			"OR word = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM256Stmt: %v", err)
	}

	// Document similarity with BM25 using 2-6 terms, for a specific collection
	df.simBM25Col1Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" SUM(2.5 * frequency * idf / (frequency + 1.5)) AS bm25, "+
			" COUNT(frequency) / 1.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" document "+
			"FROM word_freq_doc "+
			"WHERE "+
			" (word = ?) AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM25Col1Stmt: %v", err)
	}

	df.simBM25Col2Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 2.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" document "+
			"FROM word_freq_doc "+
			"WHERE (word = ? OR word = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM252Stmt: %v", err)
	}

	df.simBM25Col3Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 3.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" document "+
			"FROM word_freq_doc "+
			"WHERE (word = ? OR word = ? OR word = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM253Stmt: %v", err)
	}

	df.simBM25Col4Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 4.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" document "+
			"FROM word_freq_doc "+
			"WHERE (word = ? OR word = ? OR word = ? OR word = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM254Stmt: %v", err)
	}

	df.simBM25Col5Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 5.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" document "+
			"FROM word_freq_doc "+
			"WHERE (word = ? OR word = ? OR word = ? OR word = ? OR word = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM255Stmt: %v", err)
	}

	df.simBM25Col6Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" COUNT(frequency) / 5.0 AS bitvector, "+
			" GROUP_CONCAT(word) AS contains_words, "+
			" document "+
			"FROM word_freq_doc "+
			"WHERE (word = ? OR word = ? OR word = ? OR word = ? OR word = ? "+
			"OR word = ?) "+
			"AND collection = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM256Stmt: %v", err)
	}

	// Document similarity with Bigram using 1-6 bigrams, k = 1.5, b = 0
	df.simBigram1Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" collection, document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBigram1Stmt: %v", err)
	}

	df.simBigram2Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" collection, document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? OR bigram = ? GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBM252Stmt: %v", err)
	}

	df.simBigram3Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" collection, document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? OR bigram = ? OR bigram = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBigram3Stmt: %v", err)
	}

	df.simBigram4Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" collection, document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? OR bigram = ? OR bigram = ? OR bigram = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBigram4Stmt: %v", err)
	}

	df.simBigram5Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" collection, document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? OR bigram = ? OR bigram = ? OR bigram = ? "+
			"OR bigram = ? "+
			"GROUP BY collection, document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBigram5Stmt: %v", err)
	}

	// Document similarity with Bigram using 1-6 bigrams, within a specific
	// collection
	df.simBgCol1Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBgCol1Stmt: %v", err)
	}

	df.simBgCol2Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" document "+
			"FROM bigram_freq_doc "+
			"WHERE (bigram = ? OR bigram = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBgCol2Stmt: %v", err)
	}

	df.simBgCol3Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" document "+
			"FROM bigram_freq_doc "+
			"WHERE bigram = ? OR bigram = ? OR bigram = ? "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBgCol3Stmt: %v", err)
	}

	df.simBgCol4Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" document "+
			"FROM bigram_freq_doc "+
			"WHERE (bigram = ? OR bigram = ? OR bigram = ? OR bigram = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBgCol4Stmt: %v", err)
	}

	df.simBgCol5Stmt, err = df.database.PrepareContext(ctx,
		"SELECT "+
			" SUM((1.5 + 1) * frequency * idf / "+
			"  (frequency + 1.5 * (1 - 0.65 + 0.65 * (doc_len / ?)))) AS bm25, "+
			" GROUP_CONCAT(bigram) AS contains_bigrams, "+
			" document "+
			"FROM bigram_freq_doc "+
			"WHERE (bigram = ? OR bigram = ? OR bigram = ? OR bigram = ? "+
			"OR bigram = ?) "+
			"AND collection = ? "+
			"GROUP BY document "+
			"ORDER BY bm25 DESC LIMIT 500")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for simBgCol5Stmt: %v", err)
	}

	// Find the titles of all documents
	df.findAllTitlesStmt, err = df.database.PrepareContext(ctx,
		"SELECT gloss_file, title, col_gloss_file, col_title "+
			"FROM document LIMIT 5000000")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for findAllTitlesStmt: %v", err)
	}

	// Find the titles of all documents
	df.findAllColTitlesStmt, err = df.database.PrepareContext(ctx,
		"SELECT gloss_file, title FROM collection LIMIT 500000")
	if err != nil {
		return fmt.Errorf("find.initStatements() Error for findAllColTitlesStmt: %v",
			err)
	}

	return nil
}

// Merge a list of documents with map of similar docs, adding the similarity
// for docs that are in both lists
func mergeDocList(df databaseDocFinder, simDocMap map[string]Document, docList []Document) {
	for _, simDoc := range docList {
		sDoc, ok := simDocMap[simDoc.GlossFile]
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
			colTitle, ok1 := df.GetColMap()[simDoc.CollectionFile]
			document, ok2 := df.docMap[simDoc.GlossFile]
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
				log.Println("mergeDocList, collection title not found: ",
					simDoc)
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
				log.Printf("mergeDocList, doc title not found: %v", simDoc)
				simDocMap[simDoc.GlossFile] = simDoc
			}
		}
	}
}

// Organizes the contains terms found of the document in a form that helps
// the user.
//
// doc.ContainsWords is a contained list of terms found in the query and doc
// doc.ContainsBigrams is a contained list of bigrams found in the query and doc
// doc.ContainsTerms is a list of terms found both in the query and the doc
// sorted in the same order as the query terms with words merged to bigrams
func setMatchDetails(doc Document, terms []string, docMatch fulltext.DocMatch) Document {
	log.Printf("sortContainsWords: %v", terms)
	containsTems := []string{}
	for i, term := range terms {
		//fmt.Printf("sortContainsWords: i = %d", i)
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
func toRelevantDocList(df databaseDocFinder, docs []Document, terms []string) []Document {
	if len(docs) < 1 {
		return docs
	}
	keys := []string{}
	for _, doc := range docs {
		d, ok := df.docMap[doc.GlossFile]
		if !ok {
			log.Printf("find.toRelevantDocList could not find %s", doc.GlossFile)
			continue
		}
		keys = append(keys, d.CorpusFile)
	}
	docMatches := fulltext.GetMatches(keys, terms)
	relDocs := []Document{}
	for _, doc := range docs {
		log.Printf("toRelevantDocList, check Similarity %f, min %f, gloss %s, "+
			"title: %s", doc.Similarity, minSimilarity, doc.GlossFile,
			doc.Title)
		d, ok := df.docMap[doc.GlossFile]
		if !ok {
			log.Printf("find.toRelevantDocList 2 could not find %s", doc.GlossFile)
			continue
		}
		docMatch := docMatches[d.CorpusFile]
		doc = setMatchDetails(doc, terms, docMatch)
		if doc.Similarity < minSimilarity {
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
		simDocs = append(simDocs, simDoc)
	}
	// Sort again by combined similarity
	sort.Slice(simDocs, func(i, j int) bool {
		return simDocs[i].Similarity > simDocs[j].Similarity
	})
	if len(simDocs) > maxReturned {
		return simDocs[:maxReturned]
	}
	return simDocs
}
