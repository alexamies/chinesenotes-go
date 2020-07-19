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


// Unit tests for main package
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test getChars function
func TestTranslationMemory(t *testing.T) {
	log.Printf("TestTranslationMemory: Begin unit tests\n")
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
  	url := fmt.Sprintf("/translation_memory?query=%s", tc.query)
		r := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
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
