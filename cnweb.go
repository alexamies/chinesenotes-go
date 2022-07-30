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
//
// Web application for Chinese-English dictionary lookup, translation memory,
// and finding documents in a corpus. Settings in for the app are controlled
// through the file config.yaml, located in the project home directory, which
// is found through the env variable CNREADER_HOME or the present working
// directory.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/find"
	"github.com/alexamies/chinesenotes-go/fulltext"
	"github.com/alexamies/chinesenotes-go/identity"
	"github.com/alexamies/chinesenotes-go/media"
	"github.com/alexamies/chinesenotes-go/templates"
	"github.com/alexamies/chinesenotes-go/termfreq"
	"github.com/alexamies/chinesenotes-go/transmemory"
	"github.com/alexamies/chinesenotes-go/transtools"

	"cloud.google.com/go/firestore"
)

const (
	deepLKeyName         = "DEEPL_AUTH_KEY" // Only needed if using machine translation
	defTitle             = "Chinese Notes Translation Portal"
	glossaryKeyName      = "TRANSLATION_GLOSSARY" // Google Translation API glossary
	projectIDKey         = "PROJECT_ID"           // For GCP project
	colFileName          = "collections.csv"
	titleIndexFN         = "documents.tsv"
	translationTemplFile = "web-resources/translation.html"
)

var (
	b             *backends
	authenticator *identity.Authenticator
	mediaSearcher *media.MediaSearcher
)

// backends holds dependencies that access remote resources
type backends struct {
	appConfig                                             config.AppConfig
	database                                              *sql.DB
	docMap                                                map[string]find.DocInfo
	df                                                    find.DocFinder
	dict                                                  *dictionary.Dictionary
	parser                                                find.QueryParser
	reverseIndex                                          dictionary.ReverseIndex
	substrIndex                                           dictionary.SubstringIndex
	templates                                             map[string]*template.Template
	tmSearcher                                            transmemory.Searcher
	webConfig                                             config.WebAppConfig
	deepLApiClient, translateApiClient, glossaryApiClient transtools.ApiClient
	translationProcessor                                  transtools.Processor
	docTitleFinder                                        find.TitleFinder
}

// htmlContent holds content for HTML template
type htmlContent struct {
	Title     string
	Query     string
	ErrorMsg  string
	Results   find.QueryResults
	TMResults *transmemory.Results
	Data      interface{}
}

// Content for change password page
type ChangePasswordHTML struct {
	Title            string
	OldPasswordValid bool
	ChangeSuccessful bool
	ShowNewForm      bool
}

// Data for displaying the translation page.
type translationPage struct {
	SourceText, TranslatedText, SuggestedText, Message, Title string
	Notes                                                     []transtools.Note
	DeepLChecked, GCPChecked, GlossaryChecked, PostProcessing string
}

func initApp(ctx context.Context) (*backends, error) {
	log.Println("initApp Initializing cnweb")
	appConfig := config.InitConfig()
	cnwebHome := config.GetCnWebHome()
	fileName := fmt.Sprintf("%s/webconfig.yaml", cnwebHome)
	webConfig := config.WebAppConfig{}
	configFile, err := os.Open(fileName)
	if err != nil {
		path, er := os.Getwd()
		if er != nil {
			log.Printf("cannot find cwd: %v", er)
			path = ""
		}
		log.Printf("initApp error loading file '%s' (%s): %v", fileName, path, err)
	} else {
		defer configFile.Close()
		webConfig = config.InitWeb(configFile)
	}
	var database *sql.DB
	log.Printf("initApp UseDatabase %t", config.UseDatabase())
	if config.UseDatabase() {
		database, err = initDBCon()
		if err != nil {
			return nil, fmt.Errorf("initApp unable to connect to database: %v", err)
		}
	}
	var substrIndex dictionary.SubstringIndex
	if database != nil {
		substrIndex, err = dictionary.NewSubstringIndexDB(ctx, database)
		if err != nil {
			log.Printf("initApp, non-fatal error, unable to initialize substrIndex: %v", err)
		}
	}
	cnReaderHome := os.Getenv("CNREADER_HOME")
	var dict *dictionary.Dictionary
	if len(cnReaderHome) > 0 {
		var err error
		dict, err = dictionary.LoadDictFile(appConfig)
		if err != nil {
			return nil, fmt.Errorf("main.initApp() unable to load dictionary locally: %v", err)
		}
	} else {
		// Load from web for zero-config Quickstart
		const url = "https://github.com/alexamies/chinesenotes.com/blob/master/data/cnotes_zh_en_dict.tsv?raw=true"
		var err error
		dict, err = dictionary.LoadDictURL(appConfig, url)
		if err != nil {
			return nil, fmt.Errorf("main.initApp() unable to load dictionary from net: %v", err)
		}
	}
	parser := find.NewQueryParser(dict.Wdict)
	var tms transmemory.Searcher
	var titleFinder find.TitleFinder
	var colMap map[string]string
	var docMap map[string]find.DocInfo
	titleFinder, err = initDocTitleFinder(appConfig)
	if err != nil {
		log.Printf("main.initApp() unable to load titleFinder: %v", err)
	} else {
		colMap = titleFinder.ColMap()
		docMap = titleFinder.DocMap()
		log.Printf("main.initApp() doc map loaded with %d cols and %d docs", len(colMap), len(docMap))
	}
	if database != nil {
		tms, err = transmemory.NewSearcher(ctx, database)
		if err != nil {
			return nil, fmt.Errorf("main.initApp() unable to create new TM searcher: %v", err)
		}
		titleFinder, err = find.NewMysqlTitleFinder(ctx, database, colMap, docMap)
		if err != nil {
			log.Printf("main.initApp() unable to initialize MysqlTitleFinder: %v", err)
		}
	}
	extractor, err := dictionary.NewNotesExtractor(webConfig.NotesExtractorPattern())
	if err != nil {
		log.Printf("initApp, non-fatal error, unable to initialize NotesExtractor: %v", err)
	}
	reverseIndex := dictionary.NewReverseIndex(dict, extractor)

	var tfDocFinder find.TermFreqDocFinder
	projectID, ok := os.LookupEnv(projectIDKey)
	if !ok {
		log.Printf("%s not set, not able to perform full text search", projectIDKey)
	} else {
		client, err := firestore.NewClient(ctx, projectID)
		if err != nil {
			log.Printf("Cannot instantiate Firestore TermFreqDocFinder: %v", err)
		} else {
			indexCorpus, ok := appConfig.IndexCorpus()
			if !ok {
				log.Print("Cannot find value for IndexCorpus")
			} else {
				indexGen := appConfig.IndexGen()
				addDirectory := webConfig.AddDirectoryToCol()
				tfDocFinder = termfreq.NewFirestoreDocFinder(client, indexCorpus, indexGen, addDirectory)
			}
		}
	}
	bends := &backends{
		appConfig:    appConfig,
		database:     database,
		docMap:       docMap,
		df:           find.NewDocFinder(tfDocFinder, titleFinder),
		dict:         dict,
		parser:       parser,
		reverseIndex: reverseIndex,
		substrIndex:  substrIndex,
		templates:    templates.NewTemplateMap(webConfig),
		tmSearcher:   tms,
		webConfig:    webConfig,
	}
	if config.PasswordProtected() {
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			return nil, fmt.Errorf("initApp authenticator not initialized, %v", err)
		}
	}
	return bends, nil
}

// Initialize the document title finder
func initDocTitleFinder(appConfig config.AppConfig) (find.TitleFinder, error) {
	if b != nil && b.docTitleFinder != nil {
		return b.docTitleFinder, nil
	}
	colFileName := appConfig.CorpusDataDir() + "/" + colFileName
	cr, err := os.Open(colFileName)
	if err != nil {
		return nil, fmt.Errorf("initDocTitleFinder: Error opening %s: %v", colFileName, err)
	}
	defer cr.Close()
	colMap, err := find.LoadColMap(cr)
	if err != nil {
		return nil, fmt.Errorf("initDocTitleFinder: Error loading col map: %v", err)
	}
	titleFileName := appConfig.IndexDir() + "/" + titleIndexFN
	r, err := os.Open(titleFileName)
	if err != nil {
		return nil, fmt.Errorf("initDocTitleFinder: Error opening %s: %v", titleFileName, err)
	}
	defer r.Close()
	var dInfoCN, docMap map[string]find.DocInfo
	dInfoCN, docMap = find.LoadDocInfo(r)
	log.Printf("initDocTitleFinder loaded %d cols and  %d docs", len(colMap), len(docMap))
	docTitleFinder := find.NewFileTitleFinder(colMap, dInfoCN, docMap)
	if b != nil {
		b.docTitleFinder = docTitleFinder
	}
	return docTitleFinder, nil
}

// Starting point for the Administration Portal
func adminHandler(w http.ResponseWriter, r *http.Request) {
	d := os.Getenv("DATABASE")
	if len(d) == 0 {
		log.Print("adminHandler database not initialized")
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("adminHandler authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
			return
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	}
	if identity.IsAuthorized(sessionInfo.User, "admin_portal") {
		vars := b.webConfig.GetAll()
		tmpl, err := template.New("admin_portal.html").ParseFiles("templates/admin_portal.html")
		if err != nil {
			log.Printf("main.adminHandler: error parsing template %v", err)
		}
		if tmpl == nil {
			log.Println("main.adminHandler: Template is nil")
		}
		if err != nil {
			log.Printf("main.adminHandler: error parsing template %v", err)
		}
		err = tmpl.Execute(w, vars)
		if err != nil {
			log.Printf("main.adminHandler: error rendering template %v", err)
		}
	} else {
		log.Printf("adminHandler, Not authorized: %v", sessionInfo.User)
		http.Error(w, "Not authorized", http.StatusForbidden)
	}
}

// Process a change password request
func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	d := os.Getenv("DATABASE")
	if len(d) == 0 {
		log.Print("changePasswordHandler database not initialized")
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("changePasswordHandler authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := enforceValidSession(w, r)
	if sessionInfo.Authenticated == 1 {
		oldPassword := r.PostFormValue("OldPassword")
		password := r.PostFormValue("Password")
		result := authenticator.ChangePassword(ctx, sessionInfo.User, oldPassword,
			password)
		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			sendJSON(w, result)
		} else {
			title := b.webConfig.GetVarWithDefault("Title", defTitle)
			content := ChangePasswordHTML{
				Title:            title,
				OldPasswordValid: result.OldPasswordValid,
				ChangeSuccessful: result.ChangeSuccessful,
				ShowNewForm:      result.ShowNewForm,
			}
			displayPage(w, b, "change_password_form.html", content)
		}
	}
}

// Display change password form
func changePasswordFormHandler(w http.ResponseWriter, r *http.Request) {
	d := os.Getenv("DATABASE")
	if len(d) == 0 {
		log.Print("changePasswordFormHandler database not initialized")
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}
	sessionInfo := enforceValidSession(w, r)
	if sessionInfo.Authenticated == 1 {
		title := b.webConfig.GetVarWithDefault("Title", defTitle)
		result := ChangePasswordHTML{
			Title:            title,
			OldPasswordValid: false,
			ChangeSuccessful: false,
			ShowNewForm:      true,
		}
		displayPage(w, b, "change_password_form.html", result)
	}
}

// Custom 404 page handler
func custom404(w http.ResponseWriter, r *http.Request, url string) {
	log.Printf("custom404: sending 404 for %s", url)
	displayPage(w, b, "404.html", nil)
}

func displayPage(w http.ResponseWriter, b *backends, templateName string, content interface{}) {
	tmpl, ok := b.templates[templateName]
	if !ok {
		log.Printf("displayPage: template not found %s", templateName)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	err := tmpl.Execute(w, content)
	if err != nil {
		log.Printf("displayPage: error rendering template %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}
}

// displayHome shows a simple page, for health checks and testing.
// End users may also to see this when accessing direct from the browser
func displayHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("displayHome: url %s", r.URL.Path)

	// Tell health check probes that we are alive
	if !acceptHTML(r) {
		fmt.Fprintln(w, "OK")
		return
	}

	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	content := htmlContent{
		Title: title,
	}
	if config.PasswordProtected() {
		ctx := context.Background()
		if authenticator == nil {
			var err error
			authenticator, err = identity.NewAuthenticator(ctx)
			if err != nil {
				log.Printf("displayHome: authenticator not initialized, %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
		}
		sessionInfo := identity.InvalidSession()
		cookie, err := r.Cookie("session")
		if err == nil {
			sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
		} else {
			log.Printf("displayHome error getting cookie: %v", err)
			displayPage(w, b, "login_form.html", content)
			return
		}
		if !sessionInfo.Valid {
			displayPage(w, b, "login_form.html", content)
			return
		} else {
			displayPage(w, b, "index_auth.html", content)
			return
		}
	}

	displayPage(w, b, "index.html", content)
}

// Process a change password request
func enforceValidSession(w http.ResponseWriter, r *http.Request) identity.SessionInfo {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("enforceValidSession authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
		if sessionInfo.Authenticated != 1 {
			if acceptHTML(r) {
				displayPage(w, b, "login_form.html", nil)
			} else {
				http.Error(w, "Not authorized", http.StatusForbidden)
			}
			return sessionInfo
		}
	} else {
		log.Printf("enforceValidSession, Invalid session %v", sessionInfo.User)
		http.Error(w, "Not authorized", http.StatusForbidden)
		return identity.InvalidSession()
	}
	return sessionInfo
}

// Finds documents matching the given query with search in text body
func findFullText(response http.ResponseWriter, request *http.Request) {
	log.Println("findFullText, enter")
	q := getSingleValue(request, "query")
	if len(q) == 0 {
		q = getSingleValue(request, "text")
	}
	if len(q) == 0 {
		if acceptHTML(request) {
			title := b.webConfig.GetVarWithDefault("Title", defTitle)
			content := htmlContent{
				Title: title,
			}
			displayPage(response, b, "full_text_search.html", content)
			return
		}
	}
	ctx := context.Background()
	if b == nil {
		log.Println("main.findFullText re-initializing app")
		var err error
		b, err = initApp(ctx)
		if err != nil {
			log.Printf("main.findFullText error initializing app: %v", err)
			http.Error(response, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	findDocs(ctx, response, request, b, true)
}

// findDocs finds documents matching the given query.
func findDocs(ctx context.Context, response http.ResponseWriter, request *http.Request, b *backends, fullText bool) {

	if config.PasswordProtected() {
		sessionInfo := enforceValidSession(response, request)
		if !sessionInfo.Valid {
			return
		}
	}

	q := getSingleValue(request, "query")
	if len(q) == 0 {
		q = getSingleValue(request, "text")
	}
	// No query, eg someone nativated directly to the HTML page, redisplay it
	var err error
	if len(q) == 0 && acceptHTML(request) {
		log.Print("main.findDocs No query provided")
		templateFile := "find_results.html"
		if fullText {
			templateFile = "full_text_search.html"
		}
		err = showQueryResults(response, b, find.QueryResults{}, templateFile)
		if err != nil {
			log.Printf("main.findDocs error displaying empty results %v", err)
			http.Error(response, "Internal error", http.StatusInternalServerError)
			return
		}
		return
	}

	findTitle := getSingleValue(request, "title")
	log.Printf("main.findDocs q: %s, title: %s", q, findTitle)

	var results *find.QueryResults
	c := getSingleValue(request, "collection")
	if len(c) > 0 {
		results, err = b.df.FindDocumentsInCol(ctx, b.reverseIndex, b.parser, q, c)
	} else if len(findTitle) > 0 {
		docTitleFinder, err := initDocTitleFinder(b.appConfig)
		if err == nil {
			docs, err := docTitleFinder.FindDocsByTitle(ctx, q)
			results = &find.QueryResults{
				Query:     q,
				Documents: docs,
			}
			if err != nil {
				log.Printf("main.findDocs Error finding docs, %v", err)
				http.Error(response, "Internal error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		results, err = b.df.FindDocuments(ctx, b.reverseIndex, b.parser, q, fullText)
	}

	if err != nil {
		log.Printf("main.findDocs Error searching docs, %v", err)
		http.Error(response, "Internal error", http.StatusInternalServerError)
		return
	}

	// Add similar results from translation memory, only do this when more than
	// one term is found and when the query string is between 2 and 8 characters
	// in length
	if !fullText && (b != nil) && (len([]rune(q)) > 1) && (len([]rune(q)) < 9) && (len(results.Terms) > 1) && (b.tmSearcher != nil) {
		log.Println("main.findDocs similar results from translation memory")
		tmResults, err := b.tmSearcher.Search(ctx, q, "", false, b.dict.Wdict)
		if err != nil {
			// Not essential to the main request
			log.Printf("main.findDocs translation memory error, ignoring: %v", err)
		} else if len(tmResults.Words) > 0 {
			similarTerms := []find.TextSegment{}
			for _, w := range tmResults.Words {
				chinese := w.Simplified
				if (len(w.Traditional) > 0) && (w.Traditional != "\\N") {
					chinese += " (" + w.Traditional + ")"
				}
				seg := find.TextSegment{
					QueryText: chinese,
					DictEntry: w,
				}
				similarTerms = append(similarTerms, seg)
			}
			results.SimilarTerms = similarTerms
			log.Printf("main.findDocs, for query %s, found %d similar phrases",
				q, len(results.SimilarTerms))
		}
	}

	// Return HTML if method is post
	if acceptHTML(request) {
		templateFile := "find_results.html"
		if len(findTitle) > 0 {
			templateFile = "doc_results.html"
		} else if fullText {
			templateFile = "full_text_search.html"
			r := highlightMatches(*results)
			results = &r

			// Transform notes field with regular expressions
		} else if len(results.Terms) > 0 {
			log.Println("main.findDocs, processing notes")
			match := b.webConfig.GetVar("NotesReMatch")
			replace := b.webConfig.GetVar("NotesReplace")
			processor := dictionary.NewNotesProcessor(match, replace)
			terms := []find.TextSegment{}
			for _, t := range results.Terms {
				word := processor.Process(t.DictEntry)
				term := find.TextSegment{
					QueryText: t.QueryText,
					DictEntry: word,
					Senses:    t.Senses,
				}
				terms = append(terms, term)
			}
			results.Terms = terms
		}
		err = showQueryResults(response, b, *results, templateFile)
		if err != nil {
			log.Printf("main.findDocs Error displaying results: %v", err)
			http.Error(response, "Internal error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Return JSON
	resultsJson, err := json.Marshal(results)
	if err != nil {
		log.Printf("main.findDocs error marshalling JSON, %v", err)
		http.Error(response, "Error marshalling results",
			http.StatusInternalServerError)
	} else {
		if q != "hello" && q != "Eight" { // Health check monitoring probe
			log.Printf("main.findDocs, results: %q", string(resultsJson))
		}
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(response, string(resultsJson))
	}
}

func acceptHTML(r *http.Request) bool {
	acceptEnc := r.Header.Get("Accept")
	if len(acceptEnc) > 0 && strings.Contains(acceptEnc, "text/html") {
		return true
	}
	return false
}

func getSingleValue(r *http.Request, key string) string {
	var q string
	if r.Method == http.MethodPost {
		q = r.FormValue(key)
	} else {
		url := r.URL
		queryString := url.Query()
		query := queryString[key]
		if len(query) > 0 {
			q = query[0]
		}
	}
	return q
}

// highlightMatches adds a HTML span element with highlight for matches in the
// snippets of full texts search results
func highlightMatches(r find.QueryResults) find.QueryResults {
	results := find.QueryResults{
		Query:          r.Query,
		CollectionFile: r.CollectionFile,
		NumCollections: r.NumCollections,
		NumDocuments:   r.NumDocuments,
		Collections:    r.Collections,
		Terms:          r.Terms,
		SimilarTerms:   r.SimilarTerms,
	}
	documents := []find.Document{}
	for _, d := range r.Documents {
		lm := d.MatchDetails.LongestMatch
		span := fmt.Sprintf("<span class='usage-highlight'>%s</span>", lm)
		s := strings.Replace(d.MatchDetails.Snippet, lm, span, 1)
		md := fulltext.MatchingText{
			Snippet:      s,
			LongestMatch: lm,
			ExactMatch:   d.MatchDetails.ExactMatch,
		}
		doc := find.Document{
			GlossFile:       d.GlossFile,
			Title:           d.Title,
			CollectionFile:  d.CollectionFile,
			CollectionTitle: d.CollectionTitle,
			ContainsWords:   d.ContainsWords,
			ContainsBigrams: d.ContainsBigrams,
			SimTitle:        d.SimTitle,
			SimWords:        d.SimWords,
			SimBigram:       d.SimBigram,
			SimBitVector:    d.SimBigram,
			Similarity:      d.Similarity,
			ContainsTerms:   d.ContainsTerms,
			MatchDetails:    md,
			TitleCNMatch:    d.TitleCNMatch,
		}
		documents = append(documents, doc)
	}
	results.Documents = documents
	return results
}

// initTranslationClients initializes translation API clients and processing utility.
func initTranslationClients(b *backends) {
	deepLKey, ok := os.LookupEnv(deepLKeyName)
	if !ok {
		log.Printf("%s not set\n", deepLKeyName)
	} else {
		b.deepLApiClient = transtools.NewDeepLClient(deepLKey)
	}
	b.translateApiClient = transtools.NewGoogleClient()
	glossaryName, ok := os.LookupEnv(glossaryKeyName)
	if !ok {
		log.Printf("%s not set\n", glossaryKeyName)
	} else {
		projectID, ok := os.LookupEnv(projectIDKey)
		if !ok {
			log.Printf("%s not set\n", projectIDKey)
		} else {
			b.glossaryApiClient = transtools.NewGlossaryClient(projectID, glossaryName)
		}
	}
	fExpected, err := os.Open(transtools.ExpectedDataFile)
	if err != nil {
		log.Printf("initTranslationClients: Error opening expected file: %v", err)
		return
	}
	fReplace, err := os.Open(transtools.ReplaceDataFile)
	if err != nil {
		log.Printf("initTranslationClients: Error opening replace file: %v", err)
		return
	}
	defer func() {
		if err = fExpected.Close(); err != nil {
			log.Printf("Error closing expected file: %v", err)
		}
		if err = fReplace.Close(); err != nil {
			log.Printf("Error closing replace file: %v", err)
		}
	}()
	b.translationProcessor = transtools.NewProcessor(fExpected, fReplace)
}

// processTranslation performs translation and post processing of source text.
func processTranslation(w http.ResponseWriter, r *http.Request) {
	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	if b.translationProcessor == nil {
		p := &translationPage{
			Message:        "Translation service not initialized",
			Title:          title,
			PostProcessing: "on",
		}
		showTranslationPage(w, b, p)
	}
	source := r.FormValue("source")
	translated := ""
	message := ""
	notes := []transtools.Note{}
	deepLChecked := "checked"
	gcpChecked := ""
	glossaryChecked := ""
	platform := r.FormValue("platform")
	if platform == "gcp" {
		deepLChecked = ""
		gcpChecked = "checked"
		glossaryChecked = ""
	} else if platform == "withGlossary" {
		deepLChecked = ""
		gcpChecked = ""
		glossaryChecked = "checked"
	}
	log.Printf("processTranslation, glossaryChecked %s, source: %s", glossaryChecked, source)
	processingChecked := r.FormValue("processing")
	if len(source) > 0 {
		log.Printf("platform: %s", platform)
		trText, err := translate(b, source, platform)
		if err != nil {
			log.Printf("Translation error: %v", err)
			message = err.Error()
		} else {
			log.Printf("Translation result: %s", *trText)
			translated = *trText
		}
	} else {
		message = "Please enter translated text or click Translate for a machine translation"
	}
	if len(translated) > 0 && processingChecked == "on" {
		result := b.translationProcessor.Suggest(source, translated)
		translated = result.Replacement
		notes = result.Notes
		log.Printf("suggestion notes: %s, suggested translation: %s", notes, translated)
	}
	log.Printf("deepLChecked: %s, gcpChecked: %s, glossaryChecked: %s, processingChecked: %s, len(translated) = %d",
		deepLChecked, gcpChecked, glossaryChecked, processingChecked, len(translated))
	if config.PasswordProtected() {
		sessionInfo := enforceValidSession(w, r)
		if !sessionInfo.Valid {
			return
		}
	}
	postProcessing := ""
	if processingChecked == "on" {
		postProcessing = "checked"
	}
	p := &translationPage{
		SourceText:      source,
		TranslatedText:  translated,
		Message:         message,
		Title:           title,
		Notes:           notes,
		DeepLChecked:    deepLChecked,
		GCPChecked:      gcpChecked,
		GlossaryChecked: glossaryChecked,
		PostProcessing:  postProcessing,
	}
	showTranslationPage(w, b, p)
}

// showQueryResults displays query results on a HTML page
func showQueryResults(w io.Writer, b *backends, results find.QueryResults, templateFile string) error {
	res := results
	staticDir := b.appConfig.GetVar("GoStaticDir")
	if len(staticDir) > 0 && len(results.Documents) > 0 {
		log.Printf("showQueryResults, len(Documents): %d", len(results.Documents))
		docs := []find.Document{}
		for _, doc := range results.Documents {
			d := find.Document{
				GlossFile:       "/" + staticDir + "/" + doc.GlossFile,
				Title:           doc.Title,
				CollectionFile:  "/" + staticDir + "/" + doc.CollectionFile,
				CollectionTitle: doc.CollectionTitle,
				ContainsWords:   doc.ContainsWords,
				ContainsBigrams: doc.ContainsBigrams,
				SimTitle:        doc.SimTitle,
				SimWords:        doc.SimWords,
				SimBigram:       doc.SimBigram,
				SimBitVector:    doc.SimBitVector,
				Similarity:      doc.Similarity,
				ContainsTerms:   doc.ContainsTerms,
				MatchDetails:    doc.MatchDetails,
				TitleCNMatch:    doc.TitleCNMatch,
			}
			log.Printf("showQueryResults, adding: %s", d.Title)
			docs = append(docs, d)
		}
		res = find.QueryResults{
			Query:          results.Query,
			CollectionFile: staticDir + "/" + results.CollectionFile,
			NumCollections: results.NumCollections,
			NumDocuments:   results.NumDocuments,
			Collections:    results.Collections,
			Documents:      docs,
			Terms:          results.Terms,
			SimilarTerms:   results.SimilarTerms,
		}
	}
	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	content := htmlContent{
		Title:   title,
		Results: res,
	}
	var tmpl *template.Template
	var err error
	tmpl = b.templates[templateFile]
	if err != nil {
		return fmt.Errorf("showQueryResults: error parsing template %v", err)
	}
	if tmpl == nil {
		return fmt.Errorf("showQueryResults: %s", "Template is nil")
	}
	err = tmpl.Execute(w, content)
	if err != nil {
		return fmt.Errorf("showQueryResults: error rendering template %v", err)
	}
	return nil
}

// Displays the translation page.
func showTranslationPage(w http.ResponseWriter, b *backends, p *translationPage) {
	displayPage(w, b, "translation.html", p)
}

// findHandler finds documents matching the given query.
func findHandler(response http.ResponseWriter, request *http.Request) {
	log.Printf("findHandler: url %s", request.URL.Path)
	ctx := context.Background()
	findDocs(ctx, response, request, b, false)
}

// findSubstring finds terms matching the given query with a substring match.
func findSubstring(response http.ResponseWriter, request *http.Request) {
	log.Println("main.findSubstring, enter")
	url := request.URL
	queryString := url.Query()
	query := queryString["query"]
	q := ""
	if len(query) > 0 {
		q = query[0]
	}
	topic := queryString["topic"]
	t := ""
	if len(topic) > 0 {
		t = topic[0]
	}
	subtopic := queryString["subtopic"]
	st := "placeholder"
	if len(subtopic) > 0 {
		st = subtopic[0]
	}
	d := os.Getenv("DATABASE")
	if len(d) == 0 {
		log.Print("findSubstring database not initialized")
		http.Error(response, "Server not configured", http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	results, err := b.substrIndex.LookupSubstr(ctx, q, t, st)
	if err != nil {
		log.Printf("main.findSubstring Error looking up term, %v", err)
		http.Error(response, "Error looking up term",
			http.StatusInternalServerError)
		return
	}
	resultsJson, err := json.Marshal(results)
	if err != nil {
		log.Printf("main.findSubstring error marshalling JSON, %v", err)
		http.Error(response, "Error marshalling results",
			http.StatusInternalServerError)
	} else {
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(response, string(resultsJson))
	}
}

// Health check for monitoring or load balancing system, checks reachability
func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
	fmt.Fprintf(w, "Using a database: %t", config.UseDatabase())
	fmt.Fprintf(w, "Password protected: %t", config.PasswordProtected())
}

func initDBCon() (*sql.DB, error) {
	conString := config.DBConfig()
	return sql.Open("mysql", conString)
}

// Display library page for digital texts
func library(w http.ResponseWriter, r *http.Request) {
	log.Printf("library: url %s", r.URL.Path)

	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	content := htmlContent{
		Title: title,
	}
	if config.PasswordProtected() {
		ctx := context.Background()
		if authenticator == nil {
			var err error
			authenticator, err = identity.NewAuthenticator(ctx)
			if err != nil {
				log.Printf("displayHome: authenticator not initialized, %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
		}
		sessionInfo := identity.InvalidSession()
		cookie, err := r.Cookie("session")
		if err == nil {
			sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
		} else {
			log.Printf("displayHome error getting cookie: %v", err)
			displayPage(w, b, "login_form.html", content)
			return
		}
		if !sessionInfo.Valid {
			displayPage(w, b, "login_form.html", content)
			return
		} else {
			displayPage(w, b, "library.html", content)
			return
		}
	}

	displayPage(w, b, "library.html", content)
}

// Display login form for the Translation Portal
func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	displayPage(w, b, "login_form.html", nil)
}

// Process a login request
func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("loginHandler authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	err := r.ParseForm()
	if err != nil {
		log.Printf("loginHandler: error parsing form: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	username := r.PostFormValue("UserName")
	log.Printf("loginHandler: username = %s", username)
	password := r.PostFormValue("Password")
	users, err := authenticator.CheckLogin(ctx, username, password)
	if err != nil {
		log.Printf("main.loginHandler checking login, %v", err)
		http.Error(w, "Error checking login", http.StatusInternalServerError)
		return
	}
	if len(users) != 1 {
		log.Println("loginHandler: user not found", username)
	} else {
		cookie, err := r.Cookie("session")
		if err == nil {
			log.Printf("loginHandler: updating session: %s", cookie.Value)
			sessionInfo = authenticator.UpdateSession(ctx, cookie.Value, users[0], 1)
		}
		if (err != nil) || !sessionInfo.Valid {
			sessionid := identity.NewSessionId()
			domain := config.GetSiteDomain()
			log.Printf("loginHandler: setting new session %s for domain %s",
				sessionid, domain)
			cookie := &http.Cookie{
				Name:   "session",
				Value:  sessionid,
				Domain: domain,
				Path:   "/",
				MaxAge: 86400 * 30, // One month
			}
			http.SetCookie(w, cookie)
			sessionInfo = authenticator.SaveSession(ctx, sessionid, users[0], 1)
		}
	}
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		sendJSON(w, sessionInfo)
	} else {
		if sessionInfo.Authenticated == 1 {
			title := b.webConfig.GetVarWithDefault("Title", defTitle)
			content := htmlContent{
				Title: title,
			}
			displayPage(w, b, "index.html", content)
		} else {
			loginFormHandler(w, r)
		}
	}
}

// logoutForm displays a form button to logout the user
func logoutForm(w http.ResponseWriter, r *http.Request) {
	log.Print("logoutForm: display form")
	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	content := htmlContent{
		Title: title,
	}
	displayPage(w, b, "logout.html", content)
}

// logoutHandler logs the user out of their session
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("logoutHandler: process form")
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("loginHandler authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	cookie, err := r.Cookie("session")
	if err != nil {
		// OK, just don't show the contents that require a login
		log.Println("logoutHandler: no cookie")
	} else {
		authenticator.Logout(ctx, cookie.Value)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}

	// Return HTML if method is post
	if acceptHTML(r) {
		title := b.webConfig.GetVarWithDefault("Title", defTitle)
		content := htmlContent{
			Title: title,
		}
		displayPage(w, b, "logged_out.html", content)
		return
	}

	message := "Please come back again"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"message\" :\"%s\"}", message)
}

// Retrieves detail about media objects
func mediaDetailHandler(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	if mediaSearcher == nil {
		mediaSearcher = media.NewMediaSearcher(b.database, ctx)
		if !mediaSearcher.Initialized() {
			log.Println("main.mediaDetailHandler initializing media searcher")
			http.Error(response, "Error marshalling results",
				http.StatusInternalServerError)
			return
		}
	}
	if !mediaSearcher.Initialized() {
		mediaSearcher.InitQuery(ctx)
	}
	queryString := request.URL.Query()
	query := queryString["mediumResolution"]
	log.Printf("mediaDetailHandler: query: %s", query)
	q := "No Query"
	if len(query) > 0 {
		q = query[0]
	}
	results, err := mediaSearcher.FindMedia(q, ctx)
	if err != nil {
		log.Println("main.mediaDetailHandler Error retrieving media detail, ",
			err)
		http.Error(response, "Error retrieving media detail",
			http.StatusInternalServerError)
		return
	}
	resultsJson, err := json.Marshal(results)
	if err != nil {
		log.Printf("main.mediaDetailHandler error marshalling JSON, %v", err)
		http.Error(response, "Error marshalling results",
			http.StatusInternalServerError)
	} else {
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(response, string(resultsJson))
	}
}

// portalHandler is the starting point for the Translation Portal
func portalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("portalHandler: authenticator not initialized, %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	} else {
		log.Printf("portalHandler error getting cookie: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	user := sessionInfo.User
	if identity.IsAuthorized(user, "translation_portal") {
		displayHome(w, r)
	} else {
		log.Printf("portalHandler %s with role %s not authorized for portal",
			user.UserName, user.Role)
		http.Error(w, "Not authorized", http.StatusForbidden)
	}
}

// portalLibraryHandler handles static but private pages
func portalLibraryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("portalLibraryHandler: url %s", r.URL.Path)
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("portalLibraryHandler: authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	} else {
		log.Printf("portalLibraryHandler error getting cookie: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	user := sessionInfo.User
	if identity.IsAuthorized(user, "translation_portal") {
		portalLibHome := os.Getenv("PORTAL_LIB_HOME")
		filepart := r.URL.Path[len("/loggedin/portal_library/"):]
		filename := portalLibHome + "/" + filepart
		_, err := os.Stat(filename)
		if err != nil {
			log.Printf("portalLibraryHandler os.Stat error: %v for file %s",
				err, filename)
			custom404(w, r, filename)
			return
		}
		log.Printf("portalLibraryHandler: serving file %s", filename)
		http.ServeFile(w, r, filename)
	} else {
		log.Printf("portalLibraryHandler %s with role %s not authorized",
			user.UserName, user.Role)
		http.Error(w, "Not authorized", http.StatusForbidden)
	}
}

// Display form to request a password reset
func requestResetFormHandler(w http.ResponseWriter, r *http.Request) {
	data := identity.RequestResetResult{
		EmailValid:          true,
		RequestResetSuccess: false,
		ShowNewForm:         true,
		User:                identity.InvalidUser(),
		Token:               "",
	}
	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	content := htmlContent{
		Title:     title,
		ErrorMsg:  "",
		TMResults: nil,
		Data:      data,
	}
	displayPage(w, b, "request_reset_form.html", content)
}

// requestResetHandler processes requests for password reset
func requestResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("requestResetHandler: authenticator not initialized: %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	email := r.PostFormValue("Email")
	result := authenticator.RequestPasswordReset(ctx, email)
	if result.RequestResetSuccess {
		err := identity.SendPasswordReset(result.User, result.Token, b.webConfig)
		if err != nil {
			log.Printf("requestResetHandler: could not send password reset: %v", err)
			result.RequestResetSuccess = false
		}
	}
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		sendJSON(w, result)
	} else {
		title := b.webConfig.GetVarWithDefault("Title", defTitle)
		content := htmlContent{
			Title:     title,
			ErrorMsg:  "",
			TMResults: nil,
			Data:      result,
		}
		displayPage(w, b, "request_reset_form.html", content)
	}
}

func resetPasswordFormHandler(w http.ResponseWriter, r *http.Request) {
	queryString := r.URL.Query()
	token := queryString["token"]
	content := make(map[string]string)
	if len(token) == 1 {
		content["Token"] = token[0]
	} else {
		content["Token"] = ""
	}
	displayPage(w, b, "reset_password_form.html", content)
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("resetPasswordHandler enter")
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("resetPasswordHandler: authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	token := r.PostFormValue("Token")
	newPassword := r.PostFormValue("NewPassword")
	result := authenticator.ResetPassword(ctx, token, newPassword)
	content := make(map[string]bool)
	if result {
		content["ResetPasswordSuccessful"] = true
	}
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		sendJSON(w, result)
	} else {
		displayPage(w, b, "reset_password_confirmation.html", content)
	}
}

func sendJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resultsJson, err := json.Marshal(obj)
	if err != nil {
		log.Printf("changePasswordHandler: error marshalling json: %v", err)
		http.Error(w, "Error checking login", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(resultsJson))
}

// sessionHandler checks to see if the user has a session.
// It is used by a JavaScript client to maintain a session.
func sessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			log.Printf("sessionHandler: authenticator not initialized, %v", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	}
	if (err != nil) || (!sessionInfo.Valid) {
		// OK, just don't show the contents that don't require a login
		log.Println("sessionHandler: creating a new cookie")
		sessionid := identity.NewSessionId()
		cookie := &http.Cookie{
			Name:   "session",
			Value:  sessionid,
			Domain: config.GetSiteDomain(),
			Path:   "/",
			MaxAge: 86400, // One day
		}
		http.SetCookie(w, cookie)
		userInfo := identity.UserInfo{
			UserID:   1,
			UserName: "",
			Email:    "",
			FullName: "",
			Role:     "",
		}
		authenticator.SaveSession(ctx, sessionid, userInfo, 0)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resultsJson, err := json.Marshal(sessionInfo)
	if err != nil {
		log.Printf("sessionHandler: error marshalling JSON, %v", err)
	}
	fmt.Fprint(w, string(resultsJson))
}

func getStaticFileName(u url.URL) string {
	log.Printf("getStaticFileName path: %s", u.Path)
	return "./web/" + u.Path
}

type StaticHandler struct{}

// serveStatic handles requests for static files
func (h StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if config.PasswordProtected() {
		sessionInfo := enforceValidSession(w, r)
		if !sessionInfo.Valid {
			return
		}
	}
	fname := getStaticFileName(*r.URL)
	http.ServeFile(w, r, fname)
}

// Call the relevant API to translate text.
func translate(b *backends, sourceText, platform string) (*string, error) {
	if platform == "DeepL" {
		if b.deepLApiClient == nil {
			return nil, fmt.Errorf("DeepL API client not initialized: %s", platform)
		}
		return b.deepLApiClient.Translate(sourceText)
	}
	if platform == "gcp" {
		if b.translateApiClient == nil {
			return nil, fmt.Errorf("GCP API client not initialized: %s", platform)
		}
		return b.translateApiClient.Translate(sourceText)
	}
	if b.glossaryApiClient == nil {
		return nil, fmt.Errorf("API client still not initialized: %s", platform)
	}
	return b.glossaryApiClient.Translate(sourceText)
}

// Initialzie an empty translation page and display it.
func translationHome(w http.ResponseWriter, r *http.Request) {
	if config.PasswordProtected() {
		sessionInfo := enforceValidSession(w, r)
		if !sessionInfo.Valid {
			return
		}
	}
	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	p := &translationPage{
		SourceText:      "",
		TranslatedText:  "",
		Message:         "",
		Title:           title,
		DeepLChecked:    "",
		GCPChecked:      "",
		GlossaryChecked: "checked",
		PostProcessing:  "on",
	}
	showTranslationPage(w, b, p)
}

// translationMemory handles requests for translation memory searches
func translationMemory(w http.ResponseWriter, r *http.Request) {
	if config.PasswordProtected() {
		sessionInfo := enforceValidSession(w, r)
		if !sessionInfo.Valid {
			return
		}
	}

	q := getSingleValue(r, "query")
	title := b.webConfig.GetVarWithDefault("Title", defTitle)
	if len(q) == 0 {
		if acceptHTML(r) {
			content := htmlContent{
				Title: title,
			}
			if !config.UseDatabase() {
				log.Println("translationMemory database is needed for this feature")
				content.ErrorMsg = "Translation memory not configured"
			}
			displayPage(w, b, "findtm.html", content)
			return
		}
		log.Println("translationMemory Search query string is empty")
		http.Error(w, "Query string is empty", http.StatusInternalServerError)
		return
	}
	d := getSingleValue(r, "domain")
	log.Printf("main.translationMemory Query: %s, domain: %s", q, d)
	ctx := context.Background()
	if b == nil {
		var err error
		b, err = initApp(ctx)
		if err != nil {
			log.Printf("translationMemory initalizing app, error: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	if b.tmSearcher == nil || b.dict == nil {
		log.Printf("main.translationMemory b.tmSearcher == nil || dict == nil: %v, %v", b.tmSearcher, b.dict)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	results, err := b.tmSearcher.Search(ctx, q, d, true, b.dict.Wdict)
	if err != nil {
		log.Printf("main.translationMemory error searching, %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if acceptHTML(r) {
		content := htmlContent{
			Title:     title,
			Query:     q,
			TMResults: results,
		}
		displayPage(w, b, "findtm.html", content)
		return
	}
	resultsJson, err := json.Marshal(results)
	if err != nil {
		log.Printf("main.translationMemory error marshalling JSON, %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(resultsJson))
}

var wordsRe *regexp.Regexp = regexp.MustCompile(`[0-9]+`)

// getHeadwordId extracts the headword id from the URL
// URL format: domain.com/words/1234.html where 1234 is the headword id
// the headword id or an error if it cannot be determined
func getHeadwordId(path string) (int, error) {
	hwIdStr := wordsRe.FindString(path)
	if len(hwIdStr) == 0 {
		return -1, fmt.Errorf("no headword id provided: %s", path)
	}
	hwId, err := strconv.Atoi(hwIdStr)
	if err != nil {
		return -1, err
	}
	return hwId, nil
}

// wordDetail shows details for a single word entry, returns HTML
func wordDetail(w http.ResponseWriter, r *http.Request) {
	if config.PasswordProtected() {
		sessionInfo := enforceValidSession(w, r)
		if !sessionInfo.Valid {
			return
		}
	}

	log.Printf("main.wordDetail path: %s", r.URL.Path)
	hwId, err := getHeadwordId(r.URL.Path)
	if err != nil {
		log.Printf("main.wordDetail headword not found: %v", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if hw, ok := b.dict.HeadwordIds[hwId]; ok {
		title := b.webConfig.GetVarWithDefault("Title", defTitle)
		match := b.webConfig.GetVar("NotesReMatch")
		replace := b.webConfig.GetVar("NotesReplace")
		processor := dictionary.NewNotesProcessor(match, replace)
		word := processor.Process(*hw)
		content := htmlContent{
			Title: title,
			Data: struct {
				Word dicttypes.Word
			}{
				Word: word,
			},
		}
		displayPage(w, b, "word_detail.html", content)
		return
	}

	msg := fmt.Sprintf("Not found: %d", hwId)
	http.Error(w, msg, http.StatusNotFound)
}

//Entry point for the web application
func main() {
	log.Println("cnweb.main Iniitalizing cnweb")
	ctx := context.Background()
	var err error
	b, err = initApp(ctx)
	if err != nil {
		log.Printf("main() error for initApp, will retry on subsequent HTTP requests: %v", err)
	}

	http.HandleFunc("/", displayHome)
	http.HandleFunc("/#", findHandler)
	http.HandleFunc("/find/", findHandler)
	http.HandleFunc("/findadvanced/", findFullText)
	http.HandleFunc("/findmedia", mediaDetailHandler)
	http.HandleFunc("/findsubstring", findSubstring)
	http.HandleFunc("/findtm", translationMemory)
	http.HandleFunc("/healthcheck", healthcheck)
	http.HandleFunc("/loggedin/admin", adminHandler)
	http.HandleFunc("/loggedin/changepassword", changePasswordFormHandler)
	http.HandleFunc("/library", library)
	http.HandleFunc("/loggedin/login", loginHandler)
	http.HandleFunc("/loggedin/login_form", loginFormHandler)
	http.HandleFunc("/loggedin/logout_form", logoutForm)
	http.HandleFunc("/loggedin/logout", logoutHandler)
	http.HandleFunc("/loggedin/session", sessionHandler)
	http.HandleFunc("/loggedin/portal", portalHandler)
	http.HandleFunc("/loggedin/portal_library/", portalLibraryHandler)
	http.HandleFunc("/loggedin/request_reset", requestResetHandler)
	http.HandleFunc("/loggedin/request_reset_form", requestResetFormHandler)
	http.HandleFunc("/loggedin/reset_password", resetPasswordFormHandler)
	http.HandleFunc("/loggedin/reset_password_submit", resetPasswordHandler)
	http.HandleFunc("/loggedin/submitcpwd", changePasswordHandler)
	initTranslationClients(b)
	http.HandleFunc("/translateprocess", processTranslation)
	http.HandleFunc("/translate", translationHome)
	http.Handle("/web/", http.StripPrefix("/web/", StaticHandler{}))
	http.HandleFunc("/words/", wordDetail)

	portStr := ":" + strconv.Itoa(config.GetPort())
	log.Printf("cnweb.main Starting http server at http://localhost%s", portStr)
	err = http.ListenAndServe(portStr, nil)
	if err != nil {
		log.Printf("main() error for starting server: %v", err)
		os.Exit(1)
	}
}
