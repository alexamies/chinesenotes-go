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

package find

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
)

type DocInfo struct {
	CorpusFile, GlossFile, Title, TitleCN, TitleEN, CollectionFile, CollectionTitle string
}

// docTitleFinder implements the TitleFinder interface
type fileTitleFinder struct {
	colMap  *map[string]string
	dInfoCN *map[string]DocInfo
	docMap  *map[string]DocInfo
}

// NewDocTitleFinder initializes a DocTitleFinder implementation
// Params
//   infoCache: key to the map is the Chinese part of the title
func NewFileTitleFinder(colMap *map[string]string, dInfoCN, docMap *map[string]DocInfo) TitleFinder {
	return fileTitleFinder{
		colMap:  colMap,
		dInfoCN: dInfoCN,
		docMap:  docMap,
	}
}

// FileDocTitleFinder finds documents by title using an index loaded from file.
func (f fileTitleFinder) FindDocsByTitle(ctx context.Context, query string) ([]Document, error) {
	results := []Document{}
	dInfoCN := *f.dInfoCN
	if i, ok := dInfoCN[query]; ok {
		d := Document{
			GlossFile:       i.GlossFile,
			Title:           i.Title,
			CollectionFile:  i.CollectionFile,
			CollectionTitle: i.CollectionTitle,
			TitleCNMatch:    true,
		}
		results = append(results, d)
	}
	return results, nil
}

func (f fileTitleFinder) CountCollections(ctx context.Context, query string) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (f fileTitleFinder) FindCollections(ctx context.Context, query string) []Collection {
	return []Collection{}
}

func (f fileTitleFinder) FindDocsByTitleInCol(ctx context.Context, query, col_gloss_file string) ([]Document, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f fileTitleFinder) ColMap() *map[string]string {
	return f.colMap
}

func (f fileTitleFinder) DocMap() *map[string]DocInfo {
	return f.docMap
}

// Load title info for all documents
func LoadDocInfo(r io.Reader) (map[string]DocInfo, map[string]DocInfo) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 8
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	dInfoCN := make(map[string]DocInfo)
	dInfoGlossFN := make(map[string]DocInfo)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("loadDocInfo, error reading doc titles: %v", err)
		return dInfoCN, dInfoGlossFN
	}
	log.Printf("loadDocInfo, reading collections")
	for _, r := range records {
		glossFN := r[1]
		titleCN := r[3]
		d := DocInfo{
			CorpusFile:      r[0],
			GlossFile:       r[1],
			Title:           r[2],
			TitleCN:         r[3],
			TitleEN:         r[4],
			CollectionFile:  r[5],
			CollectionTitle: r[6],
		}
		dInfoCN[titleCN] = d
		dInfoGlossFN[glossFN] = d
	}
	return dInfoCN, dInfoGlossFN
}
