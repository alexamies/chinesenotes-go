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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestDisplayHome tests the default HTTP handler.
func TestDisplayHome(t *testing.T) {
	t.Logf("TestDisplayHome: Begin unit tests\n")
	os.Unsetenv("PROTECTED")
	type test struct {
		name string
		expectContains string
  }
  tests := []test{
		{
			name: "Show home",
			expectContains: "OK",
		},
  }
  for _, tc := range tests {
  	url := "/"
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		displayHome(w, r)
		result := w.Body.String()
		if !strings.Contains(result, tc.expectContains) {
			t.Errorf("%s: expectContains %q, got %q", tc.name, tc.expectContains, result)
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
			t.Errorf("%s: expectContains %q, got %q", tc.name, tc.expectContains, result)
 		}
 	}
}

// TestTranslationMemory tests translationMemory function.
func TestTranslationMemory(t *testing.T) {
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
			expect: "",
		},
		{
			name: "query with domain many results",
			query: "結實",
			domain: "Buddhism",
			expectMany: true,
			expect: "",
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
