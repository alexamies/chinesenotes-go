package dictionary

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/alexamies/chinesenotes-go/dicttypes"
	_ "github.com/go-sql-driver/mysql"
)

// SubstringIndexDB looks up Chinese words by substring.
type SubstringIndexDB struct {
	database       *sql.DB
	findSubstrStmt *sql.Stmt
}

// NewSubstringIndexDB initialize a SubstringIndexDB
func NewSubstringIndexDB(ctx context.Context, database *sql.DB) (SubstringIndex, error) {
	if database != nil {
		findSubstrStmt, err := initSubtrQuery(ctx, database)
		if err != nil {
			return nil, fmt.Errorf("NewSearcher, substr query initializaton error %v", err)
		}
		return &SubstringIndexDB{
			database:       database,
			findSubstrStmt: findSubstrStmt,
		}, nil
	}
	return nil, fmt.Errorf("could not initialize SubstringIndex, database = nil")
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
