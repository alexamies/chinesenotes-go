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
	"io"
	"log"
)

type DocInfo struct {
	CorpusFile, GlossFile, Title, TitleCN, TitleEN, CollectionFile, CollectionTitle string
}

// DocTitleFinder finds documents by title.
type DocTitleFinder interface {
	FindDocuments(ctx context.Context, query string) (*QueryResults, error)
}

// docTitleFinder implements the DocTitleFinder interface
type docTitleFinder struct {
	infoCache map[string]DocInfo
}

// FileDocTitleFinder finds documents by title using an index loaded from file.
func (f docTitleFinder) FindDocuments(ctx context.Context, query string) (*QueryResults, error) {
	results := QueryResults{
		Query: query,
	}
	if i, ok := f.infoCache[query]; ok {
		d := Document{
			GlossFile:       i.GlossFile,
			Title:           i.Title,
			CollectionFile:  i.CollectionFile,
			CollectionTitle: i.CollectionTitle,
			TitleCNMatch:    true,
		}
		results.NumDocuments = 1
		results.Documents = []Document{d}
	}
	return &results, nil
}

// NewDocTitleFinder initializes a DocTitleFinder implementation
// Params
//   infoCache: key to the map is the Chinese part of the title
func NewDocTitleFinder(infoCache map[string]DocInfo) DocTitleFinder {
	return docTitleFinder{
		infoCache: infoCache,
	}
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
