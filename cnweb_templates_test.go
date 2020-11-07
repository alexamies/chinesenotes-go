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


package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/find"
)

// TestNewTemplateMap building the template map
func TestNewTemplateMap(t *testing.T) {
	const title = "Translation Portal"
	const query = "謹"
	const simplified = "謹"
	const pinyin = "jǐn"
	const english = "to be cautious"
	ws := dicttypes.WordSense{
		Id: 42,
		HeadwordId: 42,
		Simplified: simplified,
		Traditional: query,
		Pinyin: pinyin,
		English: english,
		Grammar: "verb",
		Concept: "\\N",
		ConceptCN: "\\N",
		Domain: "Literary Chinese",
		DomainCN: "\\N",
		Subdomain: "\\N",
		SubdomainCN: "\\N",
		Image: "\\N",
		MP3: "\\N",
		Notes: "\\N",
	}
	w := dicttypes.Word{
		Simplified: simplified,
		Traditional: "謹",
		Pinyin: pinyin,
		HeadwordId: 42,
		Senses: []dicttypes.WordSense{ws},
	}
	term := find.TextSegment{
		QueryText: query,
		DictEntry: w,
	}
	results := find.QueryResults{
		Query: query,
		CollectionFile: "",
		NumCollections: 0,
		NumDocuments: 0,
		Collections: []find.Collection{},
		Documents: []find.Document{},
		Terms: []find.TextSegment{term},
	}
	type test struct {
		name string
		templateName string
		content interface{}
		want string
  }
  tests := []test{
		{
			name: "Home page",
			templateName: "index.html",
			content: map[string]string{"Title": title},
			want: "<title>" + title + "</title>",
		},
		{
			name: "Find results",
			templateName: "find_results.html",
			content: htmlContent{
				Title: title,
				Results: &results,
			},
			want: english,
		},
  }
  for _, tc := range tests {
		templates := newTemplateMap(config.WebAppConfig{})
		tmpl, ok := templates[tc.templateName]
		if !ok {
			t.Fatalf("%s, template not found: %s", tc.name, tc.templateName)
		}
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, tc.content)
		if err != nil {
			t.Fatalf("%s, error rendering template %v", tc.name, err)
		}
		got := buf.String()
		if !strings.Contains(got, tc.want) {
			t.Errorf("%s, got %s\n bug want %s", tc.name, got, tc.want)
		}
	}
}