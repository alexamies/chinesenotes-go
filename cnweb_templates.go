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
	"log"
	"text/template"

	"github.com/alexamies/chinesenotes-go/config"
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

// header block in HTML body
const header = `
<header>
  <h1>{{.Title}}</h1>
</header>
`

// navigation menu
const nav = `
<nav>
  <ul>
    <li><a href="/">Home</a></li>
    <li><a href="/findtm">Translation Memory</a></li>
    <li><a href="/findadvanced/">Full Text Search</a></li>
    <li><a href="/library">Library</a></li>
  </ul>
</nav>
`

// Page footer
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

// Templates from source for zero-config Quickstart
const libraryTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Library</h2>
      <p>
        Placeholder for a library of digital texts.
      </p>
      <form name="findForm" method="post" action="/find/">
        <div>
          <label for="findInput">Search for</label>
          <input type="text" name="query" size="40" required/>
          <input type="hidden" name="title" value="title"/>
          <button type="submit">Find title</button>
        </div>
      </form>
    </main>
    %s
  <body>
</html>
`

const docResultsTmpl = `
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
        {{if .Results.Documents }}
          <h4>Matching documents</h4>
          <div>
            {{ range $doc := .Results.Documents }}
            <div>
              <details open>
                <summary>
                  <span class="dict-entry-headword"><a href='{{$doc.GlossFile}}'>{{ $doc.Title }}</a><</span>
                </summary>
              </details>
            </div>
            {{ end }}
          </div>
          {{ else }}
            {{if .Results.Query}}
              <p>No results</p>
            {{ else }}
              <p>Please enter a query</p>
            {{ end }}
          {{ end }}
      {{ end }}
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
        {{if .Results.Terms }}
        <h4>Terms</h4>
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
        {{ else }}
          {{if .Results.Query}}
            <p>No results</p>
          {{ else }}
            <p>Please enter a query</p>
          {{ end }}
        {{ end }}
        {{if .Results.SimilarTerms}}
        <h4>similar Terms</h4>
        <div>
          {{ range $term := .Results.SimilarTerms }}
          <ol>
            <li>
              <span class="dict-entry-headword">{{ $term.QueryText }}</span>
            </li>
          </ol>
          {{ end }}
        </div>
        {{ end }}
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
          <input type="text" name="query" size="40" required value="{{.Query}}"/>
          <button type="submit">Find</button>
        </div>
      </form>
      {{ end }}
      {{if .TMResults}}
      <h4>Results</h4>
      <ul>
        {{ range $term := .TMResults.Words }}
        <li>
          {{ $term.Simplified}} {{if $term.Traditional}} ({{ $term.Traditional}}) {{ end }} 
          {{ $term.Pinyin }}
          <ol>
            {{ range $ws := $term.Senses }}
            <li>
              {{if ne $ws.English "\\N"}}{{ $ws.English }}{{end}}
              {{if ne $ws.Notes "\\N"}}<div>Notes: {{ $ws.Notes }}</div>{{end}}
            </li>
            {{ end }}
          </ol>
        </li>
        {{ else }}
          <p>No results found</p>
        {{ end }}
      </ul>
      {{ else }}
        <p>No results found</p>
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
          <input type="text" name="query" size="40" required value="{{.Results.Query}}"/>
          <button type="submit">Find</button>
        </div>
      </form>
      {{ end }}
      {{if .Results}}
      <h3>Results</h3>
       {{if .Results.Documents}}
        <ul>
          {{ range $doc := .Results.Documents }}
          <li>
            <h4><a href="{{ $doc.GlossFile }}">{{ $doc.Title }}</a></h4>
            <p>{{ $doc.MatchDetails.Snippet }}</p>
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

const wordDetailTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      {{if .Data.Word }}
      <div>
        <span class="dict-entry-headword">
          {{ .Data.Word.Simplified }}
           {{if .Data.Word.Traditional}} ({{ .Data.Word.Traditional }}) {{ end }}
        </span>
        <span class="dict-entry-pinyin">{{ .Data.Word.Pinyin }}</span>
        <ol>
        {{ range $ws := .Data.Word.Senses }}
          <li>
          {{if ne $ws.Pinyin "\\N"}}<span class="dict-entry-pinyin">{{ $ws.Pinyin }}</span>{{end}}
          {{if ne $ws.Grammar "\\N"}}<span class="dict-entry-grammar">{{ $ws.Grammar }}</span>{{end}}
          {{if ne $ws.English "\\N"}}<span class="dict-entry-definition">{{ $ws.English }}</span>{{end}}
          {{if ne $ws.Domain "\\N"}}<div class="dict-entry-domain">Domain: {{ $ws.Domain }}</div>{{end}}
          {{if ne $ws.Notes "\\N"}}<div class="dict-entry-notes">Notes: {{ $ws.Notes }}</div>{{end}}
          </li>
        {{ end }}
        </ol>
      </div>
      {{ else }}
      <p>Not found</p>
      {{ end }}
      <p>Search for something else</p>
      <form name="findForm" method="post" action="/find/">
        <div>
          <label for="findInput">Search for</label>
          <input type="text" name="query" size="40" required value=""/>
          <button type="submit">Find</button>
        </div>
      </form>
    </main>
    %s
  <body>
</html>
`

// Page not found
const notFoundTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <p>Not found. Sorry we could not find that page</p>
    </main>
    %s
  <body>
</html>
`

// Admin page
const adminPortalTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <p>Under construction</p>
    </main>
    %s
  <body>
</html>
`

// Admin page
const indexAuthTmpl = `
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
      <ul
        <li><a href="/loggedin/changepassword">Change password</a></li>
        <li><a href="/loggedin/logout_form">Logout</a></li>
      </ul>
    </main>
    %s
  <body>
</html>
`

// Admin page
const changePasswordTmpl = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Translation Portal Change Password</h2>
      {{if or .ShowNewForm (not .ChangeSuccessful)}}
      <div id="ChangePasswordBar">
        <form id="ChangePasswordForm" name="login" action="/loggedin/submitcpwd"
              method="post">
          <p>
            <label for="OldPassword">Old Password</label>
            <input id="OldPassword" name="OldPassword" type="password" required/>
          </p>
          <p>
            <label for="Password">New Password</label>
            <input id="Password" name="Password" type="password" required/>
          </p>
          <p>
            <input id="ChangeButton" nam="ChangeButton" type="submit"
                   value="Change Password"/>
          </p>
        </form>
      </div>
      {{ end }}
      {{if .ChangeSuccessful}}
      <div id="PasswordChangedBar">
        You password has been changed
      </div>
      {{else}}
        {{if not .OldPasswordValid}}
        <div id="ErrorDiv">Your old password is incorrect</div>
        {{else}}
        <div id="ErrorDiv">There was an error changing your password</div>
        {{ end }}
      {{ end }}
    </main>
    %s
  <body>
</html>
`

const loggedOutTmp = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Logged out</h2>
      <p>See you again sometime.</p>
      <p><a href="/">Log in</a> again.</p>
    </main>
    %s
  <body>
</html>
`

const logoutTmp = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Logout</h2>
      <p>If you log out then you will have to log back in again to use the portal.</p>
      <form name="logoutForm" method="post" action="/loggedin/logout">
        <div>
          <button type="submit">Logout</button>
        </div>
      </form>
      <p>No, go <a href="/">Home</a> instead.</p>
    </main>
    %s
  <body>
</html>
`

const loginTmp = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Login</h2>
      <div id="LoginBar">
        <form id="LoginForm" name="login" action="/loggedin/login"
              method="post">
          Login:
          <label for="UserName">User Name</label>
          <input id="UserName" name="UserName" type="text" required/>
          <label for="Password">Password</label>
          <input id="Password" name="Password" type="password" required/>
          <input id="LoginButton" nam="LoginButton" type="submit"
               value="Login"/>
        </form>
      </div>
      <div id="ErrorDiv"></div>
      <p>
        <a href="/loggedin/request_reset_form">Forgot username or password</a>
      </p>
    </main>
    %s
  <body>
</html>
`

const requestResetTmp = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Request Reset Password</h2>
      <p>
        If you have forgotton your username or password, you may use this page to
        find your username or request of your password.
      </p>
      {{if .Data.ShowNewForm}}
      <div id="RequestResetBar">
        <form id="RequestResetForm" name="RequestResetForm"
              action="/loggedin/request_reset" method="post">
          <p>
            <label for="Email">Email Address</label>
            <input id="Email" name="Email" type="text" required/>
          </p>
          <p>
            <input id="RequestResetButton" nam="RequestResetButton" type="submit"
                   value="Request Password Reset"/>
          </p>
        </form>
      </div>
      {{ end }}
      {{if .Data.RequestResetSuccess}}
      <div id="SentResetSuccessfulBar">
        An email has been sent for password reset. Please check your inbox.
      </div>
      {{else}}
        {{if not .Data.EmailValid}}
        <div id="ErrorDiv">We do not have that email address on file.</div>
        {{end}}
        {{if not .Data.ShowNewForm}}
        <div id="ErrorDiv">There was an error with your request.</div>
        {{end}}
      {{end}}
    </main>
    %s
  <body>
</html>
`

const resetConfTmp = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Reset Password Confirmation</h2>
      {{if .ResetPasswordSuccessful}}
      <div id="ResetPasswordSuccessfulBar">
        Your password has been reset. <a href="/">Login</a>.
      </div>
      {{else}}
        <div id="ErrorDiv">
          There was an error with your request.
          <a href="/">Try again</a>.
        </div>
      {{ end }}
    </main>
    %s
  <body>
</html>
`

const resetFormTmp = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>Reset Password</h2>
      {{if eq .Token ""}}
      <div id="NoToken">
        There was a problem with your request. Please try again.
      </div>
      {{else}}
      <div id="ResetPasswordBar">
        <form id="ResetPasswordForm" name="ResetPasswordForm"
              action="/loggedin/reset_password_submit" method="post">
          <input id="Token" name="Token" type="hidden" value="{{.Token}}"/>
          <p>
            <label for="NewPassword">New Password</label>
            <input id="NewPassword" name="NewPassword" type="password" required/>
          </p>
          <p>
            <input id="ResetPasswordButton" nam="ResetPasswordButton" type="submit"
                   value="Reset Password"/>
          </p>
        </form>
      </div>
      {{ end }}
    </main>
    %s
  <body>
</html>
`

// newTemplateMap builds the template map
func newTemplateMap(webConfig config.WebAppConfig) map[string]*template.Template {
	tNames := map[string]string{
		"404.html":                         notFoundTmpl,
		"admin_portal.html":                adminPortalTmpl,
		"change_password_form.html":        changePasswordTmpl,
		"doc_results.html":                 docResultsTmpl,
		"find_results.html":                findResultsTmpl,
		"findtm.html":                      findTMTmpl,
		"full_text_search.html":            fullTextSearchTmpl,
		"index.html":                       indexTmpl,
		"index_auth.html":                  indexAuthTmpl,
		"library.html":                     libraryTmpl,
		"logged_out.html":                  loggedOutTmp,
		"login_form.html":                  loginTmp,
		"logout.html":                      logoutTmp,
		"request_reset_form.html":          requestResetTmp,
		"reset_password_confirmation.html": resetConfTmp,
		"reset_password_form.html":         resetFormTmp,
		"translation.html":                 notFoundTmpl,
		"word_detail.html":                 wordDetailTmpl,
	}
	templateMap := make(map[string]*template.Template)
	templDir := webConfig.GetVar("TemplateDir")
	if len(templDir) > 0 {
		log.Printf("newTemplateMap, using TemplateDir: %s", templDir)
		for tName, defTmpl := range tNames {
			fileName := templDir + "/" + tName
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
