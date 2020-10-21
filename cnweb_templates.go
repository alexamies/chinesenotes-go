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
	"html/template"

	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/webconfig"
)

// Templates from source for zero-config Quickstart
const indexTmpl = `
<!DOCTYPE html>
<html lang="en">
  <body>
    <h1>{{.Title}}</h1>
    <p><a href="/">Home</a></p>
    <form name="findForm" method="post" action="/find/">
      <div>
        <label for="findInput">Search for</label>
        <input type="text" name="query" size="40" required/>
        <button type="submit">Find</button>
      </div>
    </form>
  <body>
</html>
`

const findResultsTmpl = `
<!DOCTYPE html>
<html lang="en">
  <body>
    <h1>{{.Title}}</h1>
    <p><a href="/">Home</a></p>
    <form name="findForm" method="post" action="/find/">
      <div>
        <label for="findInput">Search for</label>
        <input type="text" name="query" size="40" required value="{{.Results.Query}}"/>
        <button type="submit">Find</button>
      </div>
    </form>
    {{if .Results}}
    <h4>Results</h4>
    <ul>
      {{ range $term := .Results.Terms }}
      <li>
        {{ $term.QueryText}} {{ $term.DictEntry.Pinyin}}
        <ul>
        {{ range $ws := $term.DictEntry.Senses }}
          <li>
          {{if ne $ws.English "\\N"}}{{ $ws.English }}{{end}}
          {{if ne $ws.Notes "\\N"}}<div>Notes: {{ $ws.Notes }}</div>{{end}}
          </li>
        {{ end }}
        </ul>
      </li>
      {{ end }}
    </ul>
    {{ end }}
  <body>
</html>
`

// newTemplateMap builds the template map
func newTemplateMap(webConfig webconfig.WebAppConfig) map[string]*template.Template {
	templateMap := make(map[string]*template.Template)
	templDir := webConfig.GetVar("TemplateDir")
	tNames := map[string]string{
		"index.html": indexTmpl,
		"find_results.html": findResultsTmpl,
	}
	if len(templDir) > 0 {
		for tName, defTmpl := range tNames {
			fileName := "templates/" + tName
			var tmpl *template.Template
			var err error
			tmpl, err = template.New(tName).ParseFiles(fileName)
			if err != nil {
				applog.Errorf("newTemplateMap: error parsing template, using default %s: %v",
						tName, err)
				tmpl = template.Must(template.New(tName).Parse(defTmpl))
			}
			templateMap[tName] = tmpl
		}
	} else {
		for tName, defTmpl := range tNames {
			tmpl := template.Must(template.New(tName).Parse(defTmpl))
			templateMap[tName] = tmpl
		}
	}
	return templateMap
}
