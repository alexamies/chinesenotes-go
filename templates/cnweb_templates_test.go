package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/find"
)

type htmlContent struct {
	Title   string
	Results find.QueryResults
}

// TestNewTemplateMap building the template map
func TestNewTemplateMap(t *testing.T) {
	const title = "Translation Portal"
	const query = "謹"
	const simplified = "謹"
	const pinyin = "jǐn"
	const english = "to be cautious"
	ws := dicttypes.WordSense{
		Id:          42,
		HeadwordId:  42,
		Simplified:  simplified,
		Traditional: query,
		Pinyin:      pinyin,
		English:     english,
		Grammar:     "verb",
		Concept:     "\\N",
		ConceptCN:   "\\N",
		Domain:      "Literary Chinese",
		DomainCN:    "\\N",
		Subdomain:   "\\N",
		SubdomainCN: "\\N",
		Image:       "\\N",
		MP3:         "\\N",
		Notes:       "\\N",
	}
	w := dicttypes.Word{
		Simplified:  simplified,
		Traditional: "謹",
		Pinyin:      pinyin,
		HeadwordId:  42,
		Senses:      []dicttypes.WordSense{ws},
	}
	term := find.TextSegment{
		QueryText: query,
		DictEntry: w,
	}
	results := find.QueryResults{
		Query:          query,
		CollectionFile: "",
		NumCollections: 0,
		NumDocuments:   0,
		Collections:    []find.Collection{},
		Documents:      []find.Document{},
		Terms:          []find.TextSegment{term},
	}
	type test struct {
		name         string
		templateName string
		content      interface{}
		want         string
	}
	tests := []test{
		{
			name:         "Home page",
			templateName: "index.html",
			content:      map[string]string{"Title": title},
			want:         "<title>" + title + "</title>",
		},
		{
			name:         "Find results",
			templateName: "find_results.html",
			content: htmlContent{
				Title:   title,
				Results: results,
			},
			want: english,
		},
	}
	for _, tc := range tests {
		templates := NewTemplateMap(config.WebAppConfig{})
		tmpl, ok := templates[tc.templateName]
		if !ok {
			t.Errorf("%s, template not found: %s", tc.name, tc.templateName)
		}
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, tc.content)
		if err != nil {
			t.Errorf("%s, error rendering template %v", tc.name, err)
		}
		got := buf.String()
		if !strings.Contains(got, tc.want) {
			t.Errorf("%s, got %s\n bug want %s", tc.name, got, tc.want)
		}
	}
}
