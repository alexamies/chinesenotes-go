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

// Test for docTitleFinder.FindDocuments
func TestDocTitleFinder(t *testing.T) {
	info := docInfo{
		GlossFile: "guanhuazhinan.html",
		Title: "官話指南 A Guide to Mandarin",
		TitleCN: "官話指南",
		TitleEN: "A Guide to Mandarin",
		CollectionFile: "xyz",
	}
	infoCache := map[string]docInfo{
		"官話指南": info,
	}
	tests := []struct {
		name string
		finder docTitleFinder
		query string
		wantNum int
	}{
		{
			name: "Empty",
			finder: docTitleFinder{},
			query: "",
			wantNum: 0,
		},
		{
			name: "Find one doc",
			finder: docTitleFinder{infoCache},
			query: "官話指南",
			wantNum: 1,
		},
	}
	for _, tc := range tests {
		ctx := context.Background()
		results, err := tc.finder.FindDocuments(ctx, tc.query)
		if err != nil {
			t.Fatalf("TestDocTitleFinder %s, unexpected error: %v", tc.name, err)
		}
		numDoc := results.NumDocuments
		if numDoc != tc.wantNum {
			t.Fatalf("TestDocTitleFinder %s, got %d, want %d", tc.name, numDoc, 
					tc.wantNum)
		}
		if numDoc != len(results.Documents) {
			t.Fatalf("TestDocTitleFinder %s, disagreement between numDoc %d, " +
					"and len(Documents) %d", tc.name, numDoc, len(results.Documents))
		}
	}	
}

// Test for loadDocInfo
func TestLoadDocInfo(t *testing.T) {
	line := `guanhuazhinan.txt	guanhuazhinan.html	` +
			`官話指南 A Guide to Mandarin	官話指南	A Guide to Mandarin	xyz.html	` +
			`Collection XYZ	Col: Title\n`
	tests := []struct {
		name string
		input string
		wantNum int
	}{
		{
			name: "Empty",
			input: "",
			wantNum: 0,
		},
		{
			name: "One record",
			input: line,
			wantNum: 1,
		},
	}
	for _, tc := range tests {
		buf := bytes.NewBufferString(tc.input)
		docInfoMap := loadDocInfo(buf)
		if len(docInfoMap) != tc.wantNum {
			t.Fatalf("TestLoadDocInfo %s, got %d, want %d", tc.name, len(docInfoMap), 
					tc.wantNum)
		}
	}	
}