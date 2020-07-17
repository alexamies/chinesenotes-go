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


//
// Chinese-English dictionary database functions
//
package dictionary

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/fileloader"
	"github.com/alexamies/chinesenotes-go/webconfig"
	"time"
)

// Encapsulates dictionary searcher
type Searcher struct {
	database *sql.DB
	findEnglishStmt *sql.Stmt
	findSubstrStmt *sql.Stmt
}

// Initialize SQL statements
func NewSearcher(ctx context.Context, database *sql.DB) (*Searcher, error) {
	stmt, err := initEnglishQuery(ctx, database)
	if err != nil {
		return nil, err
	}
	substStmt, err := initSubtrQuery(ctx, database)
	if err != nil {
		return nil, err
	}
	return &Searcher{
		database: database,
		findEnglishStmt: stmt,
		findSubstrStmt: substStmt,
	}, nil
}

func InitDBCon() (*sql.DB, error) {
	conString := webconfig.DBConfig()
	return sql.Open("mysql", conString)
}

func initEnglishQuery(ctx context.Context, database *sql.DB) (*sql.Stmt, error) {
	return database.PrepareContext(ctx, 
`SELECT simplified, traditional, pinyin, english, notes, headword
FROM words
WHERE pinyin = ? OR english LIKE ?
LIMIT 20`)
}

// Returns the word senses with English approximate or Pinyin exact match
func (searcher *Searcher) FindWordsByEnglish(ctx context.Context,
		query string) ([]dicttypes.WordSense, error) {
	applog.Info("findWordsByEnglish, query = ", query)
	likeEnglish := "%" + query + "%"
	if searcher.findEnglishStmt == nil {
		return nil, fmt.Errorf("FindWordsByEnglish,findEnglishStmt is nil query = ", query)
	}
	results, err := searcher.findEnglishStmt.QueryContext(ctx, query, likeEnglish)
	if err != nil {
		applog.Error("FindWordsByEnglish, Error for query: ", query, err)
		// Retry
		results, err = searcher.findEnglishStmt.QueryContext(ctx, query, query)
		if err != nil {
			applog.Error("FindWordsByEnglish, Give up after retry: ", query, err)
			return nil, err
		}
	}
	senses := []dicttypes.WordSense{}
	for results.Next() {
		ws := dicttypes.WordSense{}
		var hw sql.NullInt64
		var trad, pinyin, english, notes sql.NullString
		results.Scan(&ws.Simplified, &trad, &pinyin, &english, &notes, &hw)
		applog.Info("FindWordsByEnglish, simplified, headword = ",
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
	applog.Info("FindWordsByEnglish, len(senses): ", len(senses))
	return senses, nil
}

// Loads all words from the database
func LoadDict(ctx context.Context, database *sql.DB) (map[string]dicttypes.Word, error) {
	start := time.Now()
	wdict := map[string]dicttypes.Word{}
	avoidSub := config.AvoidSubDomains()
	stmt, err := database.PrepareContext(ctx, 
		"SELECT id, simplified, traditional, pinyin, english, parent_en, notes, headword FROM words")
    if err != nil {
        applog.Error("find.load_dict Error preparing stmt: ", err)
        return loadDictFile()
    }
	results, err := stmt.QueryContext(ctx)
	if err != nil {
		applog.Error("find.load_dict, Error for query: ", err)
        return loadDictFile()
	}
	for results.Next() {
		ws := dicttypes.WordSense{}
		var wsId, hw sql.NullInt64
		var trad, notes, pinyin, english, parent_en sql.NullString
		results.Scan(&wsId, &ws.Simplified, &trad, &pinyin, &english, &parent_en, &notes,
			&hw)
		if wsId.Valid {
			ws.Id = int(wsId.Int64)
		}
		if hw.Valid {
			ws.HeadwordId = int(hw.Int64)
		}
		if trad.Valid {
			ws.Traditional = trad.String
		}
		if pinyin.Valid {
			ws.Pinyin = pinyin.String
		}
		if english.Valid {
			ws.English = english.String
		}
		// If subdomain, aka parent, should be avoided, then skip
		if parent_en.Valid {
			if _, ok := avoidSub[parent_en.String]; ok {
				continue
			}
		}
		if notes.Valid {
			ws.Notes = notes.String
		}
		word, ok := wdict[ws.Simplified]
		if ok {
			word.Senses = append(word.Senses, ws)
			wdict[word.Simplified] = word
		} else {
			word = dicttypes.Word{}
			word.Simplified = ws.Simplified
			word.Traditional = ws.Traditional
			word.Pinyin = ws.Pinyin
			word.HeadwordId = ws.HeadwordId
			word.Senses = []dicttypes.WordSense{ws}
			wdict[word.Simplified] = word
		}
		if trad.Valid {
			word1, ok1 := wdict[trad.String]
			if ok1 {
				word1.Senses = append(word1.Senses, ws)
				wdict[word1.Traditional] = word1
			} else {
				word1 = dicttypes.Word{}
				word1.Simplified = ws.Simplified
				word1.Traditional = ws.Traditional
				word1.Pinyin = ws.Pinyin
				word1.HeadwordId = ws.HeadwordId
				word1.Senses = []dicttypes.WordSense{ws}
				wdict[word1.Traditional] = word1
			}
		}
	}
	applog.Info("LoadDict, loading time: ", time.Since(start))
	return wdict, nil
}

// Loads all words from a static file included in the Docker image
func loadDictFile() (map[string]dicttypes.Word, error) {
	applog.Info("loadDictFile, enter")
	wsFilenames := config.LUFileNames()
	cnReaderHome := webconfig.GetCnReaderHome()
	fnames := []string{}
	for _, wsfilename := range wsFilenames {
		fName := cnReaderHome + "/" + wsfilename
		fnames = append(fnames, fName)
	}
	return fileloader.LoadDictFile(fnames)
}
