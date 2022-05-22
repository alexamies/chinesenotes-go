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
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/find"
	"github.com/alexamies/chinesenotes-go/fulltext"
	"github.com/alexamies/chinesenotes-go/identity"
	"github.com/alexamies/chinesenotes-go/transmemory"
)

func mockSmallDict() map[string]*dicttypes.Word {
	s1 := "繁体中文"
	t1 := "繁體中文"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "fántǐ zhōngwén",
		Senses:      []dicttypes.WordSense{},
	}
	s2 := "前"
	t2 := "\\N"
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      "qián",
		Senses:      []dicttypes.WordSense{},
	}
	s3 := "不见"
	t3 := "不見"
	hw3 := dicttypes.Word{
		HeadwordId:  3,
		Simplified:  s3,
		Traditional: t3,
		Pinyin:      "bújiàn",
		Senses:      []dicttypes.WordSense{},
	}
	s4 := "古人"
	t4 := "\\N"
	hw4 := dicttypes.Word{
		HeadwordId:  4,
		Simplified:  s4,
		Traditional: t4,
		Pinyin:      "gǔrén",
		Senses:      []dicttypes.WordSense{},
	}
	s5 := "夫"
	t5 := "\\N"
	hw5 := dicttypes.Word{
		HeadwordId:  5,
		Simplified:  s5,
		Traditional: t5,
		Pinyin:      "fú fū",
		Senses:      []dicttypes.WordSense{},
	}
	s6 := "起信论"
	t6 := "起信論"
	hw6 := dicttypes.Word{
		HeadwordId:  6,
		Simplified:  s6,
		Traditional: t6,
		Pinyin:      "Qǐ Xìn Lùn",
		Senses:      []dicttypes.WordSense{},
	}
	s7 := "者"
	t7 := "\\N"
	hw7 := dicttypes.Word{
		HeadwordId:  7,
		Simplified:  s7,
		Traditional: t7,
		Pinyin:      "zhě zhuó",
		Senses:      []dicttypes.WordSense{},
	}
	s8 := "乃是"
	t8 := "\\N"
	hw8 := dicttypes.Word{
		HeadwordId:  8,
		Simplified:  s8,
		Traditional: t8,
		Pinyin:      "nǎishì",
		Senses:      []dicttypes.WordSense{},
	}
	s9 := "莲花"
	t9 := "蓮花"
	hw9 := dicttypes.Word{
		HeadwordId:  9,
		Simplified:  s9,
		Traditional: t9,
		Pinyin:      "liánhuā",
		Senses: []dicttypes.WordSense{
			dicttypes.WordSense{
				HeadwordId:  9,
				Simplified:  s9,
				Traditional: t9,
				Pinyin:      "liánhuā",
				English:     "lotus",
			},
		},
	}
	return map[string]*dicttypes.Word{
		s1: &hw1,
		t1: &hw1,
		s2: &hw2,
		s3: &hw3,
		t3: &hw3,
		s4: &hw4,
		s5: &hw5,
		s6: &hw6,
		t6: &hw6,
		s7: &hw7,
		s8: &hw8,
		s9: &hw9,
		t9: &hw9,
	}
}

type mockReverseIndex struct {
}

func (m mockReverseIndex) Initialized() bool {
	return true
}

func (m mockReverseIndex) FindWordsByEnglish(ctx context.Context, query string) ([]dicttypes.WordSense, error) {
	results := []dicttypes.WordSense{}
	if query == "lotus" {
		s1 := "莲花"
		t1 := "蓮花"
		ws := dicttypes.WordSense{
			HeadwordId:  1,
			Simplified:  s1,
			Traditional: t1,
			Pinyin:      "liánhuā",
			English:     "lotus",
		}
		results = append(results, ws)
	}
	log.Printf("mockReverseIndex.FindWordsByEnglish: query: %s, results: %v", query, results)
	return results, nil
}


// mockDocFinder imitates DocFinder interface for full text search tests
type mockDocFinder struct {
	reverseIndex dictionary.ReverseIndex
	documents []find.Document
}

func (m mockDocFinder) FindDocuments(ctx context.Context, dictSearcher dictionary.ReverseIndex,
	parser find.QueryParser, query string, advanced bool) (*find.QueryResults, error) {
	terms := parser.ParseQuery(query)
	log.Printf("mockDocFinder.FindDocuments, query %s, nTerms %d, ", query, len(terms))
	if m.reverseIndex != nil && !dicttypes.IsCJKChar(query) {
		senses, err := m.reverseIndex.FindWordsByEnglish(ctx, query)
		if err != nil {
			return nil, err
		}
		terms[0].Senses = senses
		return &find.QueryResults{
			Query:          query,
			CollectionFile: "",
			NumCollections: 0,
			NumDocuments:   0,
			Terms:          terms,
		}, nil
	}
	return &find.QueryResults{
		Query:     query,
		Documents: m.documents,
		Terms:     parser.ParseQuery(query),
	}, nil
}

func (m mockDocFinder) FindDocumentsInCol(ctx context.Context,
	dictSearcher dictionary.ReverseIndex,
	parser find.QueryParser, query,
	col_gloss_file string) (*find.QueryResults, error) {
	return nil, fmt.Errorf("Not configured")
}

func (m mockDocFinder) GetColMap() map[string]string {
	cm := make(map[string]string)
	return cm
}

func (m mockDocFinder) Inititialized() bool {
	return true
}

// TestMain runs integration tests if the flag -integration is set
func TestMain(m *testing.M) {
	os.Clearenv()
	os.Exit(m.Run())
}

// TestDisplayHome tests the default HTTP handler.
func TestDisplayHome(t *testing.T) {
	templates = newTemplateMap(webConfig)
	t.Logf("TestDisplayHome: Begin unit tests\n")
	type test struct {
		name           string
		acceptHeader   string
		expectContains string
	}
	tests := []test{
		{
			name:           "Does not accept HTML",
			acceptHeader:   "application/json",
			expectContains: "OK",
		},
		{
			name:           "Show home",
			acceptHeader:   "text/html",
			expectContains: "<title>Chinese Notes Translation Portal</title>",
		},
	}
	for _, tc := range tests {
		u := "/"
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		r.Header.Add("Accept", tc.acceptHeader)
		displayHome(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestDisplayHome %s: expectContains %q, got %q", tc.name,
				tc.expectContains, result)
		}
	}
	templates = nil
}

func TestInitDocTitleFinder(t *testing.T) {
	type test struct {
		name        string
		expectError bool
	}
	tests := []test{
		{
			name:        "Expect error, project home not set",
			expectError: true,
		},
	}
	for _, tc := range tests {
		_, err := initDocTitleFinder()
		if tc.expectError && err == nil {
			t.Errorf("TestInitDocTitleFinder %s: expectError but got nil", tc.name)
		} else if !tc.expectError && err != nil {
			t.Errorf("TestInitDocTitleFinder %s: unexpected error: %v", tc.name, err)
		}
	}
}

func TestAdminHandler(t *testing.T) {
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Expect error, app not configed and user has no session",
			expectContains: "Not authorized",
		},
	}
	for _, tc := range tests {
		u := "/loggedin/admin"
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		adminHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestAdminHandler %s: expectContains %q, got %q", tc.name,
				tc.expectContains, result)
		}
	}
}

func TestChangePasswordHandler(t *testing.T) {
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Expect error, app not configed and user has no session",
			expectContains: "Not authorized",
		},
	}
	for _, tc := range tests {
		u := "/loggedin/submitcpwd"
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		changePasswordHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestChangePasswordHandler %s: expectContains %q, got %q",
				tc.name, tc.expectContains, result)
		}
	}
}

func TestChangePasswordFormHandler(t *testing.T) {
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Expect error, app not configed and user has no session",
			expectContains: "Not authorized",
		},
	}
	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, "/loggedin/changepassword", nil)
		w := httptest.NewRecorder()
		changePasswordFormHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestChangePasswordFormHandler %s: expectContains %q, got %q",
				tc.name, tc.expectContains, result)
		}
	}
}

func TestCustom404(t *testing.T) {
	templates = newTemplateMap(webConfig)
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Expect not found",
			expectContains: "Not found",
		},
	}
	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, "/xyz", nil)
		w := httptest.NewRecorder()
		custom404(w, r, "/xyz")
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestCustom404 %s: expectContains %q, got %q",
				tc.name, tc.expectContains, result)
		}
	}
	templates = nil
}

func TestDisplayPage(t *testing.T) {
	templates = newTemplateMap(webConfig)
	const query = "邃古"
	tMContent := htmlContent{
		Title: "XYZ",
		Query: query,
	}
	type test struct {
		name           string
		u              string
		template       string
		content        interface{}
		expectContains string
	}
	tests := []test{
		{
			name:           "Translation memory query results shows the query",
			u:              "/findtm",
			template:       "findtm.html",
			content:        tMContent,
			expectContains: query,
		},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		displayPage(w, tc.template, tc.content)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestDisplayPage %s: got %q, want contains %q, ", tc.name,
				result, tc.expectContains)
		}
	}
	templates = nil
}

func TestEnforceValidSession(t *testing.T) {
	type test struct {
		name           string
		u              string
		expectContains string
	}
	tests := []test{
		{
			name:           "Find something",
			u:              "/find/",
			expectContains: "Not authorized",
		},
	}
	for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, tc.u, nil)
		w := httptest.NewRecorder()
		enforceValidSession(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestEnforceValidSession %s: got %q but want contains %q",
				tc.name, result, tc.expectContains)
		}
	}
}

// Mock for testing findDocs by title
type mockDocTitleFinder struct {
	Query     string
	Documents []find.Document
}

func (f mockDocTitleFinder) FindDocuments(ctx context.Context,
	query string) (*find.QueryResults, error) {
	qr := find.QueryResults{
		Query:        f.Query,
		NumDocuments: len(f.Documents),
		Documents:    f.Documents,
	}
	return &qr, nil
}

func TestFindDocs(t *testing.T) {
	templates = newTemplateMap(webConfig)
	const query = "蓮花寺"
	d := find.Document{
		GlossFile:       "lianhuachi.html",
		Title:           "蓮花寺",
		CollectionFile:  "abc.html",
		CollectionTitle: "A B C",
	}
	type test struct {
		name           string
		u              string
		acceptHeader   string
		query          map[string]string
		docs           []find.Document
		expectContains string
		fullText       bool
	}
	tests := []test{
		{
			name:           "No query string",
			u:              "/find/",
			acceptHeader:   "text/html",
			query:          map[string]string{},
			docs:           []find.Document{},
			expectContains: "Please enter a query",
			fullText:       false,
		},
		{
			name:           "Reflect original query",
			u:              "/find/",
			acceptHeader:   "text/html",
			query:          map[string]string{"query": "蓮花寺"},
			docs:           []find.Document{},
			expectContains: "蓮花寺",
			fullText:       false,
		},
		{
			name:           "Return query HTML",
			u:              "/find/",
			acceptHeader:   "text/html",
			query:          map[string]string{"query": "蓮花寺"},
			docs:           []find.Document{},
			expectContains: `value="蓮花寺"`,
			fullText:       false,
		},
		{
			name:           "Simple query - HTML",
			u:              "/find/",
			acceptHeader:   "text/html",
			query:          map[string]string{"query": "蓮花寺"},
			docs:           []find.Document{},
			expectContains: "lotus",
			fullText:       false,
		},
		{
			name:           "Return JSON",
			u:              "/find/",
			acceptHeader:   "application/json",
			query:          map[string]string{"query": "蓮花寺"},
			docs:           []find.Document{},
			expectContains: `"Query":"蓮花寺"`,
			fullText:       false,
		},
		{
			name:           "Search for title, no match",
			u:              "/find/",
			acceptHeader:   "text/html",
			query:          map[string]string{"query": "蓮花寺", "title": "true"},
			docs:           []find.Document{},
			expectContains: `No results`,
			fullText:       false,
		},
		{
			name:           "Search for title, one match",
			u:              "/find/",
			acceptHeader:   "text/html",
			query:          map[string]string{"query": "蓮花寺", "title": "true"},
			docs:           []find.Document{d},
			expectContains: `lianhuachi.html`,
			fullText:       false,
		},
		{
			name:           "Reverse lookup by English equivalent - JSON",
			u:              "/find/",
			acceptHeader:   "application/json",
			query:          map[string]string{"query": "lotus"},
			docs:           []find.Document{},
			expectContains: "蓮花",
			fullText:       false,
		},
		/*{
			name: "Reverse lookup by English - HTML",
			u: "/find/",
			acceptHeader: "text/html",
			query: map[string]string{"query": "lotus"},
			docs: []find.Document{},
			expectContains: "蓮花",
			fullText: false,
		},*/
	}
	ctx := context.Background()
	for _, tc := range tests {
		u := tc.u
		if len(tc.query) > 0 {
			u += "?"
			for k, v := range tc.query {
				u += fmt.Sprintf("%s=%s&", k, v)
			}
		}
		docTitleFinder = mockDocTitleFinder{
			Query:     tc.query["query"],
			Documents: tc.docs,
		}
		dict := dictionary.NewDictionary(mockSmallDict())
		reverseIndex := mockReverseIndex{}
		b := &backends{
			reverseIndex: &reverseIndex,
			df:           mockDocFinder{
				reverseIndex: reverseIndex,
			},
			tmSearcher:   mockTMSearcher{},
			dict:         dict,
			parser:       find.MakeQueryParser(dict.Wdict),
		}
		r := httptest.NewRequest(http.MethodGet, u, nil)
		r.Header.Add("Accept", tc.acceptHeader)
		w := httptest.NewRecorder()
		findDocs(ctx, w, r, b, tc.fullText)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindDocs %s: got %q but want contains %q", tc.name, result, tc.expectContains)
		}
	}
	templates = nil
	docTitleFinder = nil
}

func TestHighlightMatches(t *testing.T) {
	const s = "故自親事"
	const lm = "親事"
	md := fulltext.MatchingText{
		Snippet:      s,
		LongestMatch: lm,
	}
	d := find.Document{
		MatchDetails: md,
	}
	r := find.QueryResults{
		Documents: []find.Document{d},
	}
	const h = `故自<span class='usage-highlight'>親事</span>`
	type test struct {
		name    string
		results find.QueryResults
		expect  string
	}
	tests := []test{
		{
			name:    "Happy path",
			results: r,
			expect:  h,
		},
	}
	for _, tc := range tests {
		got := highlightMatches(tc.results)
		snippet := got.Documents[0].MatchDetails.Snippet
		if snippet != tc.expect {
			t.Errorf("TestHighlightMatches %s: got %q but want %q", tc.name,
				snippet, tc.expect)
		}
	}
}

func TestFindFullText(t *testing.T) {
	templates = newTemplateMap(webConfig)
	type test struct {
		name           string
		u              string
		query          string
		documents      []find.Document
		acceptHeader   string
		expectContains string
	}
	doc := find.Document{
		GlossFile: "foyasi.html",
		Title:     "Buddha Tooth Temple",
	}
	tests := []test{
		{
			name:           "Return HTML",
			u:              "/findadvanced/",
			query:          "",
			documents:      []find.Document{},
			acceptHeader:   `text/html`,
			expectContains: `<h2>Full Text Search</h2>`,
		},
		{
			name:           "Return JSON",
			u:              "/findadvanced/",
			query:          "佛牙寺",
			documents:      []find.Document{},
			acceptHeader:   `application/json`,
			expectContains: `"Query":"佛牙寺"`,
		},
		{
			name:           "Single result",
			u:              "/findadvanced/",
			query:          "佛牙寺",
			documents:      []find.Document{doc},
			acceptHeader:   `text/html`,
			expectContains: doc.Title,
		},
	}
	for _, tc := range tests {
		u := tc.u + "?query=" + tc.query
		wdict := map[string]*dicttypes.Word{}
		b = &backends{
			reverseIndex: &mockReverseIndex{},
			df: mockDocFinder{
				documents: tc.documents,
			},
			tmSearcher: mockTMSearcher{},
			dict:       dictionary.NewDictionary(wdict),
			parser:     find.MakeQueryParser(wdict),
		}
		r := httptest.NewRequest(http.MethodGet, u, nil)
		r.Header.Add("Accept", tc.acceptHeader)
		w := httptest.NewRecorder()
		findFullText(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindFullText %s: got %q but want contains %q", tc.name,
				result, tc.expectContains)
		}
	}
	templates = nil
	b = nil
}

// TestFindHandler tests finding a word.
func TestFindHandler(t *testing.T) {
	s := "邃古"
	e := "remote antiquity"
	hw := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s,
		Pinyin:     "suìgǔ",
		Senses: []dicttypes.WordSense{
			dicttypes.WordSense{
				Simplified: s,
				English:    e,
			},
		},
	}
	wdict := map[string]*dicttypes.Word{
		s: &hw,
	}
	b = &backends{
		reverseIndex: &mockReverseIndex{},
		df:           mockDocFinder{},
		tmSearcher:   mockTMSearcher{},
		dict:         dictionary.NewDictionary(wdict),
		parser:       find.MakeQueryParser(wdict),
	}
	type test struct {
		name           string
		query          string
		expectContains string
	}
	tests := []test{
		{
			name:           "Find a word",
			query:          "邃古",
			expectContains: "remote antiquity",
		},
	}
	for _, tc := range tests {
		u := "/find/?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		findHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindHandler %s: expectContains %q, got %q", tc.name,
				tc.expectContains, result)
		}
	}
	b = nil
}

// TestFindSubstring tests search based on a dictionary entry substring.
func TestFindSubstring(t *testing.T) {
	type test struct {
		name           string
		query          string
		expectContains string
	}
	tests := []test{
		{
			name:           "No configured",
			query:          "可思议",
			expectContains: "Server not configured",
		},
	}
	for _, tc := range tests {
		u := "/findsubstring?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		findSubstring(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindSubstring %s: got %q, expectContains %q", tc.name,
				result, tc.expectContains)
		}
	}
}

// Test site domain
func TestGetSiteDomain(t *testing.T) {
	domain := config.GetSiteDomain()
	if domain != "localhost" {
		t.Error("TestGetSiteDomain: domain = ", domain)
	}
}

func TestHealthcheck(t *testing.T) {
	os.Setenv("PROTECTED", "true")
	os.Setenv("DATABASE", "abcd")
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Display OK",
			expectContains: "OK",
		},
		{
			name:           "Check password protected",
			expectContains: "Password protected: true",
		},
		{
			name:           "Check database set",
			expectContains: "Using a database: true",
		},
	}
	for _, tc := range tests {
		const u = "/healthcheck"
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		healthcheck(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestHealthcheck %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	os.Unsetenv("PROTECTED")
	os.Unsetenv("DATABASE")
}

func TestLibrary(t *testing.T) {
	templates = newTemplateMap(webConfig)
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Display library page",
			expectContains: "Library",
		},
	}
	for _, tc := range tests {
		const u = "/library"
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		library(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestLibrary %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	templates = nil
}

func TestLoginFormHandler(t *testing.T) {
	templates = newTemplateMap(webConfig)
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Display login page",
			expectContains: "Login",
		},
	}
	for _, tc := range tests {
		const u = "/loggedin/login_form"
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		loginFormHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestLoginFormHandler %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	templates = nil
}

func TestLoginHandler(t *testing.T) {
	templates = newTemplateMap(webConfig)
	authenticator = &identity.Authenticator{}
	type test struct {
		name           string
		expectContains string
	}
	tests := []test{
		{
			name:           "Display login page",
			expectContains: "Login",
		},
	}
	for _, tc := range tests {
		const u = "/loggedin/login"
		r := httptest.NewRequest(http.MethodPost, u, nil)
		w := httptest.NewRecorder()
		loginHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestLoginHandler %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	templates = nil
	authenticator = nil
}

func TestShowQueryResults(t *testing.T) {
	templates = newTemplateMap(webConfig)
	type test struct {
		name           string
		query          string
		template       string
		expectContains string
	}
	tests := []test{
		{
			name:           "Query is shown in results",
			query:          "大庾嶺",
			template:       "find_results.html",
			expectContains: "大庾嶺",
		},
		{
			name:           "Query is shown in full text results",
			query:          "大庾嶺",
			template:       "full_text_search.html",
			expectContains: "大庾嶺",
		},
	}
	for _, tc := range tests {
		w := httptest.NewRecorder()
		results := find.QueryResults{
			Query: tc.query,
		}
		showQueryResults(w, results, tc.template)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestShowQueryResults %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	templates = nil
}

func TestGetStaticFileName(t *testing.T) {
	tests := []struct {
		name   string
		u      string
		expect string
	}{
		{
			name:   "empty query",
			u:      "app.js",
			expect: "./web/app.js",
		},
	}
	for _, tc := range tests {
		u, err := url.Parse(tc.u)
		if err != nil {
			t.Fatalf("TestGetStaticFileName %s: cannot parse %s, error: %v", tc.name,
				tc.u, err)
		}
		got := getStaticFileName(*u)
		if got != tc.expect {
			t.Errorf("TestGetStaticFileName %s: got %q, want %q: ", tc.name, got,
				tc.expect)
		}
	}
}

type mockTMSearcher struct {
	words []dicttypes.Word
}

func (s mockTMSearcher) Search(ctx context.Context,
	query string,
	domain string,
	includeSubstrings bool,
	wdict map[string]*dicttypes.Word) (*transmemory.Results, error) {
	r := transmemory.Results{
		Words: s.words,
	}
	return &r, nil
}

// TestTranslationMemory tests translationMemory function.
func TestTranslationMemory(t *testing.T) {
	jieshi := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  "结实",
		Traditional: "結實",
		Pinyin:      "jiēshi",
	}
	kaihuajieshi := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  "结实",
		Traditional: "開花結實",
		Pinyin:      "kāi huā jiē shi",
	}
	type test struct {
		name           string
		query          string
		domain         string
		words          []dicttypes.Word
		expectContains string
	}
	tests := []test{
		{
			name:           "empty query",
			query:          "",
			domain:         "",
			words:          []dicttypes.Word{},
			expectContains: "Query string is empty\n",
		},
		{
			name:           "query with no results",
			query:          "hello",
			domain:         "",
			words:          []dicttypes.Word{},
			expectContains: `{"Words":[]}`,
		},
		{
			name:           "query two results",
			query:          "結實",
			domain:         "",
			words:          []dicttypes.Word{jieshi, kaihuajieshi},
			expectContains: "结实",
		},
	}
	for _, tc := range tests {
		wdict := map[string]*dicttypes.Word{}
		b = &backends{
			reverseIndex: &mockReverseIndex{},
			df:           mockDocFinder{},
			tmSearcher:   mockTMSearcher{tc.words},
			dict:         dictionary.NewDictionary(wdict),
			parser:       find.MakeQueryParser(wdict),
		}
		u := "/findtm?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		translationMemory(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestTranslationMemory %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	b = nil
}

func TestGetHeadwordId(t *testing.T) {
	type test struct {
		name        string
		path        string
		expectError bool
		expectHwId  int
	}
	tests := []test{
		{
			name:        "Basic match",
			path:        "/words/1234.html",
			expectError: false,
			expectHwId:  1234,
		},
		{
			name:        "No match",
			path:        "/words/abcd.html",
			expectError: true,
			expectHwId:  -1,
		},
	}
	for _, tc := range tests {
		hwId, err := getHeadwordId(tc.path)
		if tc.expectError && err == nil {
			t.Fatalf("TestGetHeadwordId %s: expected error", tc.name)
		}
		if !tc.expectError && err != nil {
			t.Fatalf("TestGetHeadwordId %s: unexpected error: %v", tc.name, err)
		}
		if hwId != tc.expectHwId {
			t.Errorf("TestGetHeadwordId %s: got %d, want %d", tc.name, hwId,
				tc.expectHwId)
		}
	}
}

func TestWordDetail(t *testing.T) {
	templates = newTemplateMap(webConfig)
	smallDict := mockSmallDict()
	s := "一時三相"
	ws := dicttypes.WordSense{
		Id:         1,
		Simplified: s,
		Notes:      "FGDB entry 1",
	}
	hw := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s,
		Senses:     []dicttypes.WordSense{ws},
	}
	dictWNotes := map[string]*dicttypes.Word{
		s: &hw,
	}
	webConfig = config.WebAppConfig{
		ConfigVars: map[string]string{
			"NotesReMatch": `FGDB entry ([0-9]*)`,
			"NotesReplace": `<a href="/web/${1}.html">FGDB entry</a>`,
		},
	}
	type test struct {
		name           string
		hwId           int
		wdict          map[string]*dicttypes.Word
		expectContains string
	}
	tests := []test{
		{
			name:           "Not found",
			hwId:           123,
			wdict:          map[string]*dicttypes.Word{},
			expectContains: "Not found: 123",
		},
		{
			name:           "Word Found",
			hwId:           1,
			wdict:          smallDict,
			expectContains: "繁體中文",
		},
		{
			name:           "Contains corpus doc link in notes",
			hwId:           1,
			wdict:          dictWNotes,
			expectContains: `<a href="/web/1.html">FGDB entry</a>`,
		},
	}
	for _, tc := range tests {
		u := fmt.Sprintf("/words/%d.html", tc.hwId)
		b = &backends{
			reverseIndex: &mockReverseIndex{},
			df:           mockDocFinder{},
			tmSearcher:   mockTMSearcher{},
			dict:         dictionary.NewDictionary(tc.wdict),
			parser:       find.MakeQueryParser(tc.wdict),
		}
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		wordDetail(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestWordDetail %s: got %q, want %q, ", tc.name, result,
				tc.expectContains)
		}
	}
	templates = nil
	webConfig = config.WebAppConfig{}
}
