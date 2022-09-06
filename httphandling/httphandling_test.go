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

package httphandling

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/identity"
	"github.com/alexamies/chinesenotes-go/templates"
)


// htmlContent holds content for HTML template
type htmlContent struct {
	Title     string
	Query     string
	ErrorMsg  string
}

func TestDisplayPage(t *testing.T) {
	templates := templates.NewTemplateMap(config.WebAppConfig{})
	pageDisplayer := NewPageDisplayer(templates)
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
		pageDisplayer.DisplayPage(w, tc.template, tc.content)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestDisplayPage %s: got %q, want contains %q, ", tc.name,
				result, tc.expectContains)
		}
	}
}

func TestEnforceValidSession(t *testing.T) {
	templates := templates.NewTemplateMap(config.WebAppConfig{})
	authenticator := identity.Authenticator{}
	pageDisplayer := NewPageDisplayer(templates)
	sessionEnforcer := NewSessionEnforcer(&authenticator, pageDisplayer)
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
		sessionEnforcer.EnforceValidSession(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("TestEnforceValidSession %s: got %q but want contains %q",
				tc.name, result, tc.expectContains)
		}
	}
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
