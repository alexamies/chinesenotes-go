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

	"github.com/alexamies/chinesenotes-go/webconfig"
)

// TestNewTemplateMap building the template map
func TestNewTemplateMap(t *testing.T) {
	want := "<title>Home Page</title>"
	type test struct {
		name string
		templateName string
		content interface{}
		want string
  }
  tests := []test{
		{
			name: "Home Page",
			templateName: "index.html",
			content: map[string]string{"Title": "Home Page"},
			want: "<title>Home Page</title>",
		},
  }
  for _, tc := range tests {
		templates := newTemplateMap(webconfig.WebAppConfig{})
		tmpl, ok := templates[tc.templateName]
		if !ok {
			t.Fatalf("%s, template not found: %s", tc.name, tc.templateName)
		}
		content := map[string]string{"Title": "Home Page"}
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, content)
		if err != nil {
			t.Fatalf("error rendering template %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, want) {
			t.Errorf("got %s\n bug want %s", got, want)
		}
	}
}