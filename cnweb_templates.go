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
  "fmt"
	"html/template"
  "log"

	"github.com/alexamies/chinesenotes-go/webconfig"
)

// HTML fragment for page head
const head = `
  <head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <link href="https://fonts.googleapis.com/css?family=Noto+Sans" rel="stylesheet">
    <link rel="stylesheet" href="/web/styles.css">
  </head>
`

const header = `
<header>
  <h1>{{.Title}}</h1>
</header>
`

const nav = `
<nav>
  <ul>
    <li><a href="/">Home</a></li>
    <li><a href="/findtm">Translation Memory</a></li>
    <li><a href="/findadvanced/">Full Text Search</a></li>
    <li><a href="/web/texts.html">Library</a></li>
  </ul>
</nav>
`

const footer = `
    <footer>
      <p>
        Copyright Fo Guang Shan 佛光山 2020.
        The Chinese-English dictionary is reproduced from the <a 
        href="http://ntireader.org/" target="_blank"
        > NTI Buddhist Text Reader</a> under the <a 
        href="https://creativecommons.org/licenses/by-sa/3.0/" target="_blank"
        >Creative Commons Attribution-Share Alike 3.0 License</a>
        (CCASE 3.0). 
        The site is powered by open source
        software under an <a 
        href="http://www.apache.org/licenses/LICENSE-2.0.html"
        >Apache 2.0 license</a>.
        Other content shown in password protected versions of this site is
        copyright protected.
      </p>
    </footer>
`

// Templates from source for zero-config Quickstart
const indexTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <p>
        Enter Chinese text into the input field below to find each word and its
        English equivalent.
      </p>
      <form name="findForm" method="post" action="/find/">
        <div>
          <label for="findInput">Search for</label>
          <input type="text" name="query" size="40" required/>
          <button type="submit">Find</button>
        </div>
      </form>
    </main>
    %s
  <body>
</html>
`

const findResultsTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <form name="findForm" method="post" action="/find/">
        <div>
          <label for="findInput">Search for</label>
          <input type="text" name="query" size="40" required value="{{.Results.Query}}"/>
          <button type="submit">Find</button>
        </div>
      </form>
      {{if .Results}}
      <h4>Results</h4>
      <div>
        {{ range $term := .Results.Terms }}
        <div>
          <details open>
            <summary>
              <span class="dict-entry-headword">{{ $term.QueryText }}</span>
              <span class="dict-entry-pinyin">{{ $term.DictEntry.Pinyin }}</span>
            </summary>
            <ol>
            {{ range $ws := $term.DictEntry.Senses }}
              <li>
              {{if ne $ws.Pinyin "\\N"}}<span class="dict-entry-pinyin">{{ $ws.Pinyin }}</span>{{end}}
              {{if ne $ws.Grammar "\\N"}}<span class="dict-entry-grammar">{{ $ws.Grammar }}</span>{{end}}
              {{if ne $ws.English "\\N"}}<span class="dict-entry-definition">{{ $ws.English }}</span>{{end}}
              {{if ne $ws.Domain "\\N"}}<div class="dict-entry-domain">Domain: {{ $ws.Domain }}</div>{{end}}
              {{if ne $ws.Notes "\\N"}}<div class="dict-entry-notes">Notes: {{ $ws.Notes }}</div>{{end}}
              </li>
            {{ end }}
            </ol>
          </details>
          </div>
        {{ end }}
      </div>
      {{ end }}
    </main>
    %s
  <body>
</html>
`

const findTMTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Translation Memory</h2>
      {{if .ErrorMsg}}
        <p>Error: {{ .ErrorMsg }}</p>
      {{ else }}
      <p>Enter Chinese text into to the most closely related names and phrases</p>
      <form name="findForm" method="post" action="/findtm">
        <div>
          <label for="findInput">Search for</label>
          <input type="text" name="query" size="40" required/>
          <button type="submit">Find</button>
        </div>
      </form>
      {{ end }}
      {{if .TMResults}}
      <h4>Results</h4>
      <ul>
        {{ range $term := .TMResults.Words }}
        <li>
          {{ $term.Traditional}} {{ $term.Pinyin }}
          <ol>
            {{ range $ws := $term.Senses }}
            <li>
              {{if ne $ws.English "\\N"}}{{ $ws.English }}{{end}}
              {{if ne $ws.Notes "\\N"}}<div>Notes: {{ $ws.Notes }}</div>{{end}}
            </li>
            {{ end }}
          </ol>
        </li>
        {{ end }}
      </ul>
      {{ end }}
    </main>
    %s
  <body>
</html>
`

const fullTextSearchTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Full Text Search</h2>
      {{if .ErrorMsg}}
        <p>Error: {{ .ErrorMsg }}</p>
      {{ else }}
      <p>Enter Chinese text into to the most relevant documents</p>
      <form name="findForm" method="post" action="/findadvanced/">
        <div>
          <label for="findInput">Search for</label>
          <input type="text" name="query" size="40" required/>
          <button type="submit">Find</button>
        </div>
      </form>
      {{ end }}
      {{if .Results}}
      <h4>Results</h4>
       {{if .Results.Documents}}
        <ul>
          {{ range $doc := .Results.Documents }}
          <li>
            <a href="{{ $doc.GlossFile }}">{{ $doc.Title }}</a>
          </li>
          {{ end }}
        </ul>
        {{ else }}
        <p>No results found</p>
        {{ end }}
      {{ end }}
    </main>
    %s
  <body>
</html>
`

// newTemplateMap builds the template map
func newTemplateMap(webConfig webconfig.WebAppConfig) map[string]*template.Template {
	tNames := map[string]string{
		"index.html": indexTmpl,
		"find_results.html": findResultsTmpl,
    "findtm.html": findTMTmpl,
    "full_text_search.html": fullTextSearchTmpl,
	}
  templateMap := make(map[string]*template.Template)
  templDir := webConfig.GetVar("TemplateDir")
	if len(templDir) > 0 {
		for tName, defTmpl := range tNames {
			fileName := "templates/" + tName
			var tmpl *template.Template
			var err error
			tmpl, err = template.New(tName).ParseFiles(fileName)
			if err != nil {
				log.Printf("newTemplateMap: error parsing template, using default %s: %v",
						tName, err)
        t := fmt.Sprintf(defTmpl, head, header, nav, footer)
				tmpl = template.Must(template.New(tName).Parse(t))
			}
			templateMap[tName] = tmpl
		}
	} else {
		for tName, defTmpl := range tNames {
      t := fmt.Sprintf(defTmpl, head, header, nav, footer)
			tmpl := template.Must(template.New(tName).Parse(t))
			templateMap[tName] = tmpl
		}
	}
	return templateMap
}
