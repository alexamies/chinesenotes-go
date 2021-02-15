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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/find"
	"github.com/alexamies/chinesenotes-go/identity"
)

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
		name string
		acceptHeader string
		expectContains string
  }
  tests := []test{
		{
			name: "Does not accept HTML",
			acceptHeader: "application/json",
			expectContains: "OK",
		},
		{
			name: "Show home",
			acceptHeader: "text/html",
			expectContains: "<title>Chinese Notes Translation Portal</title>",
		},
  }
  for _, tc := range tests {
  	url := "/"
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
		name string
		expectError bool
  }
  tests := []test{
		{
			name: "Expect error, project home not set",
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
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Expect error, app not configed and user has no session",
			expectContains: "Not authorized",
		},
  }
  for _, tc := range tests {
  	url := "/loggedin/admin"
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Expect error, app not configed and user has no session",
			expectContains: "Not authorized",
		},
  }
  for _, tc := range tests {
  	url := "/loggedin/submitcpwd"
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Expect error, app not configed and user has no session",
			expectContains: "Not authorized",
		},
  }
  for _, tc := range tests {
  	url := "/loggedin/changepassword"
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Expect not found",
			expectContains: "Not found",
		},
  }
  for _, tc := range tests {
  	url := "/xyz"
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
		name string
		url string
		template string
		content interface{}
		expectContains string
  }
  tests := []test{
		{
			name: "Translation memory query results shows the query",
			url: "/findtm",
			template: "findtm.html",
			content: tMContent,
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
		name string
		url string
		expectContains string
  }
  tests := []test{
		{
			name: "Find something",
			url: "/find/",
			expectContains: "Not authorized",
		},
  }
  for _, tc := range tests {
		r := httptest.NewRequest(http.MethodGet, tc.url, nil)
		w := httptest.NewRecorder()
		enforceValidSession(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestEnforceValidSession %s: got %q but want contains %q",
					tc.name, result, tc.expectContains)
 		}
 	}
}

type mockDocTitleFinder struct{}
func (f mockDocTitleFinder) FindDocuments(ctx context.Context, query string) (*find.QueryResults, error) {
	d := find.Document{
		GlossFile: "lianhuachi.html",
		Title: "蓮花寺",
		CollectionFile: "abc.html",
		CollectionTitle: "A B C",
	}
	qr := find.QueryResults{
		Query: "蓮花寺",
		NumDocuments: 1,
		Documents: []find.Document{d},
	}
	return &qr, nil
}

func TestFindDocs(t *testing.T) {
	docTitleFinder = mockDocTitleFinder{}
	type test struct {
		name string
		url string
		acceptHeader string
		query map[string]string
		expectContains string
		fullText bool
  }
  tests := []test{
		{
			name: "Reflect original query",
			url: "/find/",
			acceptHeader: "text/html",
			query: map[string]string{"query": "蓮花寺"},
			expectContains: "蓮花寺",
			fullText: false,
		},
		{
			name: "Return HTML",
			url: "/find/",
			acceptHeader: "text/html",
			query: map[string]string{"query": "蓮花寺"},
			expectContains: `value="蓮花寺"`,
			fullText: false,
		},
		{
			name: "Return JSON",
			url: "/find/",
			acceptHeader: "application/json",
			query: map[string]string{"query": "蓮花寺"},
			expectContains: `"Query":"蓮花寺"`,
			fullText: false,
		},
		{
			name: "Search for title",
			url: "/find/",
			acceptHeader: "text/html",
			query: map[string]string{"query": "蓮花寺", "title": "true"},
			expectContains: `<a href='/web/lianhuachi.html'>蓮花寺</a>`,
			fullText: false,
		},
  }
  for _, tc := range tests {
  	url := tc.url
  	if len(tc.query) > 0 {
  		url += "?"
  		for k, v := range tc.query {
  			url += fmt.Sprintf("%s=%s&", k, v)
  		}
  	}
		r := httptest.NewRequest(http.MethodGet, url, nil)
		r.Header.Add("Accept", tc.acceptHeader)
		w := httptest.NewRecorder()
		findDocs(w, r, tc.fullText)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindDocs %s: got %q but want contains %q", tc.name,
					result, tc.expectContains)
 		}
 	}
 	docTitleFinder = nil
}

func TestFindFullText(t *testing.T) {
	type test struct {
		name string
		url string
		query string
		acceptHeader string
		expectContains string
  }
  tests := []test{
		{
			name: "Return HTML",
			url: "/findadvanced/",
			query: "",
			acceptHeader: `text/html`,
			expectContains: `<h2>Full Text Search</h2>`,
		},
		{
			name: "Return JSON",
			url: "/findadvanced/",
			query: "佛牙寺",
			acceptHeader: `application/json`,
			expectContains: `"Query":"佛牙寺"`,
		},
  }
  for _, tc := range tests {
  	url := tc.url + "?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, url, nil)
		r.Header.Add("Accept", tc.acceptHeader)
		w := httptest.NewRecorder()
		findFullText(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindFullText %s: got %q but want contains %q", tc.name,
					result, tc.expectContains)
 		}
 	}
}

// TestFindHandler tests finding a word.
func TestFindHandler(t *testing.T) {
	type test struct {
		name string
		query string
		expectContains string
  }
  tests := []test{
		{
			name: "Find a word",
			query: "邃古",
			expectContains: "remote antiquity",
		},
  }
  for _, tc := range tests {
  	url := "/find/?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		findHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestFindHandler %s: expectContains %q, got %q", tc.name,
					tc.expectContains, result)
 		}
 	}
}

// TestFindSubstring tests search based on a dictionary entry substring.
func TestFindSubstring(t *testing.T) {
	type test struct {
		name string
		query string
		expectContains string
  }
  tests := []test{
		{
			name: "No configured",
			query: "可思议",
			expectContains: "Server not configured",
		},
  }
  for _, tc := range tests {
  	url := "/findsubstring?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Display OK",
			expectContains: "OK",
		},
		{
			name: "Check password protected",
			expectContains: "Password protected: true",
		},
		{
			name: "Check database set",
			expectContains: "Using a database: true",
		},
	}
  for _, tc := range tests {
  	const url = "/healthcheck"
		r := httptest.NewRequest(http.MethodGet, url, nil)
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
	type test struct {
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Display library page",
			expectContains: "Library",
		},
	}
  for _, tc := range tests {
  	const url = "/library"
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		library(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestLibrary %s: got %q, want %q, ", tc.name, result,
					tc.expectContains)
 		}
  }
}

func TestLoginFormHandler(t *testing.T) {
	type test struct {
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Display login page",
			expectContains: "Login",
		},
	}
  for _, tc := range tests {
  	const url = "/loggedin/login_form"
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		loginFormHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestLoginFormHandler %s: got %q, want %q, ", tc.name, result,
					tc.expectContains)
 		}
  }
}

func TestLoginHandler(t *testing.T) {
	authenticator = &identity.Authenticator{}
	type test struct {
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Display login page",
			expectContains: "Login",
		},
	}
  for _, tc := range tests {
  	const url = "/loggedin/login"
		r := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
		loginHandler(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestLoginHandler %s: got %q, want %q, ", tc.name, result,
					tc.expectContains)
 		}
  }
  authenticator = nil
}

func TestShowQueryResults(t *testing.T) {
	type test struct {
		name string
		query string
		template string
		expectContains string
  }
  tests := []test{
		{
			name: "Query is shown in results",
			query: "大庾嶺",
			template: "find_results.html",
			expectContains: "大庾嶺",
		},
		{
			name: "Query is shown in full text results",
			query: "大庾嶺",
			template: "full_text_search.html",
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
}

// TestTranslationMemory tests translationMemory function.
func TestTranslationMemory(t *testing.T) {
	db := os.Getenv("DATABASE")
	if len(db) == 0 {
		t.Skip("TestTranslationMemory: skipping, DATABASE not defined")
	}
	type test struct {
		name string
		query string
		domain string
		expectMany bool
		expect string
  }
  tests := []test{
		{
			name: "empty query",
			query: "",
			domain: "",
			expectMany: false,
			expect: "Query string is empty\n",
		},
		{
			name: "query with no results",
			query: "hello",
			domain: "",
			expectMany: false,
			expect: "{\"Words\":null}",
		},
		{
			name: "query many results",
			query: "結實",
			domain: "",
			expectMany: true,
			expect: "结实",
		},
		{
			name: "query with domain many results",
			query: "結實",
			domain: "Buddhism",
			expectMany: true,
			expect: "结实",
		},
  }
  for _, tc := range tests {
  	url := "/translation_memory?query=" + tc.query
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		if (tmSearcher == nil) || !tmSearcher.DatabaseInitialized() {
			t.Skip("TestTranslationMemory: database not initialized, skippining unit test")
			return
		}
		translationMemory(w, r)
		result := w.Body.String()
		if !tc.expectMany && tc.expect != result {
			t.Errorf("%s: expect %q, got %q", tc.name, tc.expect, result)
 		}
		if tc.expectMany && len(result) < 10 {
			t.Errorf("%s: expectMany but got only %d chars", tc.name, len(result))
 		}
 	}
}
