package dictionary

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// DBSearcher looks up Chinese words by either Chinese or English.
// 
// If the dictionary searcher cannot connect to the database then
// it will run in degraded mode by looking up Chinese words form dictionary
// file.
type DBSearcher struct {
	database *sql.DB
	findEnglishStmt *sql.Stmt
	initialized bool
}

// NewSearcher initialize SQL statements
func NewDBSearcher(ctx context.Context, database *sql.DB) ReverseIndex {
	s := DBSearcher{}
	if database != nil {
		var err error
		s.findEnglishStmt, err = initEnglishQuery(ctx, database)
		if err != nil {
			log.Printf("NewSearcher, database statement initializaton error %v", err)
			return &s
		}
	}
	s.initialized = true
	return &s
}

// SubstringIndexDB looks up Chinese words by substring.
type SubstringIndexDB struct {
	database *sql.DB
	findSubstrStmt *sql.Stmt
}

// NewSubstringIndexDB initialize a SubstringIndexDB
func NewSubstringIndexDB(ctx context.Context, database *sql.DB) (SubstringIndex, error) {
	if database != nil {
		findSubstrStmt, err := initSubtrQuery(ctx, database)
		if err != nil {
			return nil, fmt.Errorf("NewSearcher, substr query initializaton error \n%v\n", err)
		}
		return &SubstringIndexDB{
			database: database,
			findSubstrStmt: findSubstrStmt,
		}, nil
	}
	return nil, fmt.Errorf("could not initialize SubstringIndex, database = nil")
}

func initEnglishQuery(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT simplified, traditional, pinyin, english, notes, headword
FROM words
WHERE pinyin = ? OR english LIKE ?
LIMIT 20`)
}

// Initialized returns true if there were no error in initialization.
func (s *DBSearcher) Initialized() bool {
	return s.initialized
}

// FindWordsByEnglish returns the word senses with English approximate or Pinyin exact match
func (searcher *DBSearcher) FindWordsByEnglish(ctx context.Context,
		query string) ([]dicttypes.WordSense, error) {
	log.Printf("findWordsByEnglish, query = %s", query)
	likeEnglish := "%" + query + "%"
	if searcher.findEnglishStmt == nil {
		return nil, fmt.Errorf("FindWordsByEnglish,findEnglishStmt is nil query = %s",
			query)
	}
	results, err := searcher.findEnglishStmt.QueryContext(ctx, query, likeEnglish)
	if err != nil {
		log.Printf("FindWordsByEnglish, Error for query: %s, error %v", query, err)
		// Retry
		results, err = searcher.findEnglishStmt.QueryContext(ctx, query, query)
		if err != nil {
			log.Printf("FindWordsByEnglish, Give up after retry: %s, error: %v", query, err)
			return nil, err
		}
	}
	senses := []dicttypes.WordSense{}
	for results.Next() {
		ws := dicttypes.WordSense{}
		var hw sql.NullInt64
		var trad, pinyin, english, notes sql.NullString
		results.Scan(&ws.Simplified, &trad, &pinyin, &english, &notes, &hw)
		log.Printf("FindWordsByEnglish, simplified, headword = %s, %v",
			ws.Simplified, hw)
		if trad.Valid {
			ws.Traditional = trad.String
		}
		if pinyin.Valid {
			ws.Pinyin = pinyin.String
		}
		if english.Valid {
			ws.English = english.String
		}
		if notes.Valid {
			ws.Notes = notes.String
		}
		if hw.Valid {
			ws.HeadwordId = int(hw.Int64)
		}
		senses = append(senses, ws)
	}
	log.Printf("FindWordsByEnglish, len(senses): %d", len(senses))
	return senses, nil
}

// Lookup a term based on a substring and a topic
func (searcher SubstringIndexDB) LookupSubstr(ctx context.Context, query, topic_en, subtopic_en string) (*Results, error) {
	if query == "" {
		return nil, fmt.Errorf("query string is empty")
	}
	log.Printf("LookupSubstr, query %s, topic = %s", query, topic_en)
	likeTerm := "%" + query + "%"
	results, err := searcher.findSubstrStmt.QueryContext(ctx, likeTerm, likeTerm,
		topic_en, subtopic_en)
	if err != nil {
		return nil, fmt.Errorf("LookupSubstr, Error for query %s: %v", query, err)
	}
	wmap := map[string]dicttypes.Word{}
	for results.Next() {
		ws := dicttypes.WordSense{}
		var hw sql.NullInt64
		var trad, pinyin, english, notes sql.NullString
		results.Scan(&ws.Simplified, &trad, &pinyin, &english, &notes, &hw)
		if trad.Valid {
			ws.Traditional = trad.String
		}
		if pinyin.Valid {
			ws.Pinyin = pinyin.String
		}
		if english.Valid {
			ws.English = english.String
		}
		if notes.Valid {
			ws.Notes = notes.String
		}
		if hw.Valid {
			ws.HeadwordId = int(hw.Int64)
		}
		addWordSense2Map(wmap, ws)
	}
	log.Printf("LookupSubstr, len(wmap): %d", len(wmap))
	words := wordMap2Array(wmap)
	return &Results{words}, nil
}

func initSubtrQuery(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT simplified, traditional, pinyin, english, notes, headword 
FROM words 
WHERE
  (simplified LIKE ? OR traditional LIKE ?)
  AND 
  (topic_en = ? OR parent_en = ?)
LIMIT 100`)
}
