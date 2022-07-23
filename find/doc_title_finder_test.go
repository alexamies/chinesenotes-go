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
	"bytes"
	"context"
	"testing"
)

func TestFindDocsByTitle(t *testing.T) {
	info := DocInfo{
		CorpusFile:     "guanhuazhinan.md",
		GlossFile:      "guanhuazhinan.html",
		Title:          "官話指南 A Guide to Mandarin",
		TitleCN:        "官話指南",
		TitleEN:        "A Guide to Mandarin",
		CollectionFile: "xyz",
	}
	doc := Document{
		GlossFile: "guanhuazhinan.html",
		Title:     "官話指南 A Guide to Mandarin",
	}
	infoCache := map[string]DocInfo{
		"官話指南": info,
	}
	collections := []Collection{}
	documents := []Document{}
	colMap := map[string]string{}
	docMap := map[string]DocInfo{}
	titleFinder := newMockTitleFinder(collections, documents, colMap, docMap)
	tests := []struct {
		name    string
		finder  TitleFinder
		query   string
		wantNum int
	}{
		{
			name:    "Empty",
			finder:  titleFinder,
			query:   "",
			wantNum: 0,
		},
		{
			name:    "Find one doc",
			finder:  newMockTitleFinder(collections, []Document{doc}, colMap, infoCache),
			query:   "官話指南",
			wantNum: 1,
		},
	}
	for _, tc := range tests {
		ctx := context.Background()
		results, err := tc.finder.FindDocsByTitle(ctx, tc.query)
		if err != nil {
			t.Fatalf("TestDocTitleFinder %s, unexpected error: %v", tc.name, err)
		}
		numDoc := len(results)
		if numDoc != tc.wantNum {
			t.Fatalf("TestDocTitleFinder.%s with query %q, got %d, want %d", tc.name, tc.query, numDoc, tc.wantNum)
		}
	}
}

// Test for LoadColMap
func TestLoadColMap(t *testing.T) {
	line := `x/y.csv	x/y.html	Classic Title 經	A classic.	x/y_00.txt	A Collection	Classics	100	Ancient\n`
	tests := []struct {
		name      string
		input     string
		wantNum   int
		GlossFile string
		wantTitle string
	}{
		{
			name:      "One record",
			input:     line,
			wantNum:   1,
			GlossFile: "x/y.html",
			wantTitle: "Classic Title 經",
		},
	}
	for _, tc := range tests {
		buf := bytes.NewBufferString(tc.input)
		cMap, err := LoadColMap(buf)
		if err != nil {
			t.Fatalf("TestLoadColMap %s, unexpected error %v", tc.name, err)
		}
		if len(cMap) != tc.wantNum {
			t.Fatalf("TestLoadDocInfo %s, got %d, want %d", tc.name, len(cMap),
				tc.wantNum)
		}
		title := cMap[tc.GlossFile]
		if title != tc.wantTitle {
			t.Fatalf("TestLoadColMap %s, got %s, want %s", tc.name, title, tc.wantTitle)
		}
	}
}

// Test for LoadDocInfo
func TestLoadDocInfo(t *testing.T) {
	line := `guanhuazhinan.txt	guanhuazhinan.html	` +
		`官話指南 A Guide to Mandarin	官話指南	A Guide to Mandarin	xyz.html	` +
		`Collection XYZ	Col: Title\n`
	tests := []struct {
		name           string
		input          string
		wantNum        int
		GlossFile      string
		wantCorpusFile string
	}{
		{
			name:           "Empty",
			input:          "",
			wantNum:        0,
			GlossFile:      "",
			wantCorpusFile: "",
		},
		{
			name:           "One record",
			input:          line,
			wantNum:        1,
			GlossFile:      "guanhuazhinan.html",
			wantCorpusFile: "guanhuazhinan.txt",
		},
	}
	for _, tc := range tests {
		buf := bytes.NewBufferString(tc.input)
		_, docInfoMap := LoadDocInfo(buf)
		if len(docInfoMap) != tc.wantNum {
			t.Fatalf("TestLoadDocInfo %s, got %d, want %d", tc.name, len(docInfoMap),
				tc.wantNum)
		}
		dMap := docInfoMap
		d := dMap[tc.GlossFile]
		if d.CorpusFile != tc.wantCorpusFile {
			t.Fatalf("TestLoadDocInfo %s, got %s, want %s", tc.name, d.CorpusFile,
				tc.wantCorpusFile)
		}
	}
}
