/*
Web application for finding documents in the corpus
*/
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/find"
	"github.com/alexamies/chinesenotes-go/identity"
	"github.com/alexamies/chinesenotes-go/mail"
	"github.com/alexamies/chinesenotes-go/media"
	"github.com/alexamies/chinesenotes-go/transmemory"
	"github.com/alexamies/chinesenotes-go/webconfig"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const defTitle = "Chinese Notes Translation Portal"

var (
	database *sql.DB
	parser find.QueryParser
	wdict map[string]dicttypes.Word
	dictSearcher *dictionary.Searcher
	tmSearcher *transmemory.Searcher
	df find.DocFinder
	authenticator *identity.Authenticator
)

// Content for HTML template
type HTMLContent struct {
	Title string
	Results *find.QueryResults
	TMResults *transmemory.Results
}

func init() {
	applog.Info("cnweb.main.init Initializing cnweb")
	ctx := context.Background()
	err := initApp(ctx)
	if err != nil {
		applog.Errorf("main.init() error: \n%v\n", err)
	}
}

func initApp(ctx context.Context) error {
	applog.Info("initApp Initializing cnweb")
	var err error
	if webconfig.UseDatabase() {
		database, err = initDBCon()
		if err != nil {
			return fmt.Errorf("initApp unable to connect to database: \n%v\n", err)
		}
	}
	dictSearcher = dictionary.NewSearcher(ctx, database)
	wdict, err = dictionary.LoadDict(ctx, database)
	if err != nil {
		return fmt.Errorf("main.init() unable to load dictionary: \n%v", err)
	}
	parser = find.MakeQueryParser(wdict)
	if database != nil {
		tmSearcher, err = transmemory.NewSearcher(ctx, database)
		if err != nil {
			return fmt.Errorf("main.init() unable to create new TM searcher: \n%v\n", err)
		}
	}
	df = find.NewDocFinder(ctx, database)
	if webconfig.PasswordProtected() {
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			return fmt.Errorf("init authenticator not initialized, \n%v", err)
		}
	}
	return nil
}

// Starting point for the Administration Portal
func adminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("changePasswordHandler authenticator not initialized, \n%v\n", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	}
	if identity.IsAuthorized(sessionInfo.User, "admin_portal") {
		vars := webconfig.GetAll()
		tmpl, err := template.New("admin_portal.html").ParseFiles("templates/admin_portal.html")
		if err != nil {
			applog.Errorf("main.adminHandler: error parsing template %v", err)
		}
		if tmpl == nil {
			applog.Error("main.adminHandler: Template is nil")
		}
		if err != nil {
			applog.Errorf("main.adminHandler: error parsing template %v", err)
		}
		err = tmpl.Execute(w, vars)
		if err != nil {
			applog.Errorf("main.adminHandler: error rendering template %v", err)
		}
	} else {
		applog.Infof("adminHandler, Not authorized: %v", sessionInfo.User)
		http.Error(w, "Not authorized", http.StatusForbidden)
	}
}

// Process a change password request
func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("changePasswordHandler authenticator not initialized, \n%v\n", err)
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
			displayPage(w, "change_password_form.html", result)
		}
	}
}

// Display change password form
func changePasswordFormHandler(w http.ResponseWriter, r *http.Request) {
	sessionInfo := enforceValidSession(w, r)
	if sessionInfo.Authenticated == 1 {
		// fresh form
		result := identity.ChangePasswordResult{false, false, true}
		displayPage(w, "change_password_form.html", result)
	}
}

// Custom 404 page handler
func custom404(w http.ResponseWriter, r *http.Request, url string) {
	applog.Errorf("custom404: sending 404 for %s", url)
	displayPage(w, "404.html", nil)
}

func displayPage(w http.ResponseWriter, templateName string, content interface{}) {
	tmpl, err := template.New(templateName).ParseFiles("templates/" + templateName)
	if err != nil {
		applog.Errorf("displayPage: error parsing template %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	} else if tmpl == nil {
		applog.Error("displayPage: Template is nil")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, content)
	if err != nil {
		applog.Errorf("displayPage: error rendering template %f", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}	
}

// displayHome shows a simple page, for healthchecks and testing.
// End users may also to see this when accessing direct from the browser
func displayHome(w http.ResponseWriter, r *http.Request) {
	applog.Infof("displayHome: url %s\n", r.URL.Path)

	// Tell healthcheck probes that we are alive
	if !acceptHTML(r) {
		fmt.Fprintf(w, "OK")
    return
	}

	title := webconfig.GetVarWithDefault("Title", defTitle)
	content := HTMLContent{
		Title: title,
	}
	if webconfig.PasswordProtected() {
		ctx := context.Background()
		if authenticator == nil {
			var err error
			authenticator, err = identity.NewAuthenticator(ctx)
			if err != nil {
				applog.Errorf("displayHome: authenticator not initialized, \n%v\n", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
		}
		sessionInfo := identity.InvalidSession()
		cookie, err := r.Cookie("session")
		if err == nil {
			sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
		} else {
			applog.Info("displayHome error getting cookie: %v", err)
			displayPage(w, "login_form.html", content)
			return
		}
		if !sessionInfo.Valid {
			displayPage(w, "login_form.html", content)
			return
		} else {
			displayPage(w, "index_auth.html", content)
			return
		}
	}

	displayPage(w, "index.html", content)
}

// displayPortalHome shows the translation portal home page
func displayPortalHome(w http.ResponseWriter) {
	vars := webconfig.GetAll()
	displayPage(w, "translation_portal.html", vars)
}

// Process a change password request
func enforceValidSession(w http.ResponseWriter, r *http.Request) identity.SessionInfo {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("enforceValidSession authenticator not initialized, \n%v\n", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
		if sessionInfo.Authenticated != 1 {
			http.Error(w, "Not authorized", http.StatusForbidden)
			return sessionInfo
		}
	} else {
		applog.Infof("enforceValidSession, Invalid session %v", sessionInfo.User)
		http.Error(w, "Not authorized", http.StatusForbidden)
		return identity.InvalidSession()
	}
	return sessionInfo
}

// Finds documents matching the given query with search in text body
func findAdvanced(response http.ResponseWriter, request *http.Request) {
	applog.Info("main.findAdvanced, enter")
	findDocs(response, request, true)
}

// findDocs finds documents matching the given query.
func findDocs(response http.ResponseWriter, request *http.Request, advanced bool) {
	q := getSingleValue(request, "query")
	if len(q) == 0 {
		q = getSingleValue(request, "text")
	}

	var results *find.QueryResults
	c := getSingleValue(request, "collection")
	ctx := context.Background()
	if df == nil || !df.Inititialized() || dictSearcher == nil || !dictSearcher.Initialized() {
		err := initApp(ctx)
		if err != nil {
			applog.Errorf("findDocs error: \n%v\n", err)
			http.Error(response, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	var err error
	if len(c) > 0 {
		results, err = df.FindDocumentsInCol(ctx, dictSearcher, parser, q, c)
	} else {
		results, err = df.FindDocuments(ctx, dictSearcher, parser, q, advanced)
	}

	if err != nil {
		applog.Errorf("main.findDocs Error searching docs, %v", err)
		http.Error(response, "Internal error", http.StatusInternalServerError)
		return
	}

	if webconfig.PasswordProtected() {
		sessionInfo := enforceValidSession(response, request)
		if !sessionInfo.Valid {
			return
		}
	}

	// Return HTML if method is post
	if acceptHTML(request) {
		showQueryResults(response, results)
    return
	}

	// Return JSON
	resultsJson, err := json.Marshal(results)
	if err != nil {
		applog.Errorf("main.findDocs error marshalling JSON, %v", err)
		http.Error(response, "Error marshalling results",
			http.StatusInternalServerError)
	} else {
		if (q != "hello" && q != "Eight" ) { // Health check monitoring probe
			applog.Infof("main.findDocs, results: %q", string(resultsJson))
		}
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(response, string(resultsJson))
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

// showQueryResults displays query results on a HTML page
func showQueryResults(w http.ResponseWriter, results *find.QueryResults) {
	title := webconfig.GetVarWithDefault("Title", defTitle)
	content := HTMLContent{
		Title: title,
		Results: results,
	}
	tmpl, err := template.New("find_results.html").ParseFiles("templates/find_results.html")
	if err != nil {
		applog.Errorf("findDocs: error parsing template %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
	if tmpl == nil {
		applog.Error("findDocs: Template is nil")
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
	err = tmpl.Execute(w, content)
	if err != nil {
		applog.Errorf("findDocs: error rendering template %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// findHandler finds documents matching the given query.
func findHandler(response http.ResponseWriter, request *http.Request) {
	applog.Infof("findHandler: url %s\n", request.URL.Path)
	findDocs(response, request, false)
}

// findSubstring finds terms matching the given query with a substring match.
func findSubstring(response http.ResponseWriter, request *http.Request) {
	applog.Info("main.findSubstring, enter")
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
	ctx := context.Background()
	results, err := dictSearcher.LookupSubstr(ctx, q, t, st)
	if err != nil {
		applog.Errorf("main.findSubstring Error looking up term, %v", err)
		http.Error(response, "Error looking up term",
			http.StatusInternalServerError)
		return
	}
	resultsJson, err := json.Marshal(results)
	if err != nil {
		applog.Errorf("main.findSubstring error marshalling JSON, %v", err)
		http.Error(response, "Error marshalling results",
			http.StatusInternalServerError)
	} else {
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(response, string(resultsJson))
	}
}

// Health check for monitoring or load balancing system, checks reachability
func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "healthcheck ok")
}

func initDBCon() (*sql.DB, error) {
	conString := webconfig.DBConfig()
	return sql.Open("mysql", conString)
}

// Display login form for the Translation Portal
func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	displayPage(w, "login_form.html", nil)
}

// Process a login request
func loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("loginHandler authenticator not initialized, \n%v\n", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	err := r.ParseForm()
	if err != nil {
		applog.Errorf("loginHandler: error parsing form: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	username := r.PostFormValue("UserName")
	applog.Infof("loginHandler: username = %s", username)
	password := r.PostFormValue("Password")
	users, err := authenticator.CheckLogin(ctx, username, password)
	if err != nil {
		applog.Errorf("main.loginHandler checking login, %v", err)
		http.Error(w, "Error checking login", http.StatusInternalServerError)
		return
	}
	if len(users) != 1 {
		applog.Error("loginHandler: user not found", username)
	} else {
		cookie, err := r.Cookie("session")
		if err == nil {
			applog.Infof("loginHandler: updating session: %s", cookie.Value)
			sessionInfo = authenticator.UpdateSession(ctx, cookie.Value, users[0], 1)
		}
		if (err != nil) || !sessionInfo.Valid {
			sessionid := identity.NewSessionId()
			domain := webconfig.GetSiteDomain()
			applog.Info("loginHandler: setting new session %s\n for domain %s\n",
				sessionid, domain)
			cookie := &http.Cookie{
        		Name: "session",
        		Value: sessionid,
        		Domain: domain,
        		Path: "/",
        		MaxAge: 86400*30, // One month
        	}
        	http.SetCookie(w, cookie)
        	sessionInfo = authenticator.SaveSession(ctx, sessionid, users[0], 1)
        }
    }
    if strings.Contains(r.Header.Get("Accept"), "application/json") {
    	sendJSON(w, sessionInfo)
	} else {
		if sessionInfo.Authenticated == 1 {
			displayPortalHome(w)
		} else {
			loginFormHandler(w, r)
		}
	}
}

// logoutForm displays a form button to logout the user
func logoutForm(w http.ResponseWriter, r *http.Request) {
	applog.Infof("logoutForm: display form")
	title := webconfig.GetVarWithDefault("Title", defTitle)
	content := HTMLContent{
		Title: title,
	}
	displayPage(w, "logout.html", content)
}

// logoutHandler logs the user out of their session
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	applog.Infof("logoutHandler: process form")
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("loginHandler authenticator not initialized, \n%v\n", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	cookie, err := r.Cookie("session")
	if err != nil {
		// OK, just don't show the contents that require a login
		applog.Error("logoutHandler: no cookie")
	} else {
		authenticator.Logout(ctx, cookie.Value)
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}

	// Return HTML if method is post
	if acceptHTML(r) {
		title := webconfig.GetVarWithDefault("Title", defTitle)
		content := HTMLContent{
			Title: title,
		}
    displayPage(w, "logged_out.html", content)
    return
	}

	message := "Please come back again"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\"message\" :\"%s\"}", message)
}

// Retrieves detail about media objects
func mediaDetailHandler(response http.ResponseWriter, request *http.Request) {
	queryString := request.URL.Query()
	query := queryString["mediumResolution"]
	applog.Infof("mediaDetailHandler: query: %s", query)
	q := "No Query"
	if len(query) > 0 {
		q = query[0]
	}
	results, err := media.FindMedia(q)
	if err != nil {
		applog.Error("main.mediaDetailHandler Error retrieving media detail, ",
			err)
		http.Error(response, "Error retrieving media detail",
			http.StatusInternalServerError)
		return
	}
	resultsJson, err := json.Marshal(results)
	if err != nil {
		applog.Errorf("main.mediaDetailHandler error marshalling JSON, %v", err)
		http.Error(response, "Error marshalling results",
			http.StatusInternalServerError)
	} else {
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(response, string(resultsJson))
	}
}

// portalHandler is the starting point for the Translation Portal
func portalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("portalHandler: authenticator not initialized, \n%v\n", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	} else {
		applog.Info("portalHandler error getting cookie: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	user := sessionInfo.User
	if identity.IsAuthorized(user, "translation_portal") {
		displayPortalHome(w)
	} else {
		applog.Infof("portalHandler %s with role %s not authorized for portal",
			user.UserName, user.Role)
		http.Error(w, "Not authorized", http.StatusForbidden)
	}
}

// portalLibraryHandler handles static but private pages
func portalLibraryHandler(w http.ResponseWriter, r *http.Request) {
	applog.Infof("portalLibraryHandler: url %s\n", r.URL.Path)
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("portalLibraryHandler: authenticator not initialized, \n%v\n", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = authenticator.CheckSession(ctx, cookie.Value)
	} else {
		applog.Infof("portalLibraryHandler error getting cookie: %v\n", err)
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
			applog.Infof("portalLibraryHandler os.Stat error: %v for file %s\n",
					err, filename)
			custom404(w, r, filename)
			return
		}
		applog.Infof("portalLibraryHandler: serving file %s\n", filename)
		http.ServeFile(w, r, filename)
	} else {
		applog.Infof("portalLibraryHandler %s with role %s not authorized\n",
			user.UserName, user.Role)
		http.Error(w, "Not authorized", http.StatusForbidden)
	}
}

// Display form to request a password reset
func requestResetFormHandler(w http.ResponseWriter, r *http.Request) {
	content := identity.RequestResetResult{true, false, true,
		identity.InvalidUser(), ""}
	displayPage(w, "request_reset_form.html", content)
}

// requestResetHandler processes requests for password reset
func requestResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("requestResetHandler: authenticator not initialized, \n%v\n", err)
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
	}
	email := r.PostFormValue("Email")
	result := authenticator.RequestPasswordReset(ctx, email)
	if result.RequestResetSuccess {
		err := mail.SendPasswordReset(result.User, result.Token)
		if err != nil {
			result.RequestResetSuccess = false
		}
	}
    if strings.Contains(r.Header.Get("Accept"), "application/json") {
    	sendJSON(w, result)
	} else {
		displayPage(w, "request_reset_form.html", result)
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
	displayPage(w, "reset_password_form.html", content)
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	applog.Info("resetPasswordHandler enter")
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("resetPasswordHandler: authenticator not initialized, \n%v\n", err)
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
		displayPage(w, "reset_password_confirmation.html", content)
	}
}

func sendJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resultsJson, err := json.Marshal(obj)
	if err != nil {
		applog.Errorf("changePasswordHandler: error marshalling json: %v", err)
		http.Error(w, "Error checking login", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(resultsJson))
}

// sessionHandler checks to see if the user has a session.
// It is used by a JavaScript client to maintain a session.
func sessionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if authenticator == nil {
		var err error
		authenticator, err = identity.NewAuthenticator(ctx)
		if err != nil {
			applog.Errorf("sessionHandler: authenticator not initialized, \n%v\n", err)
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
		applog.Info("sessionHandler: creating a new cookie")
		sessionid := identity.NewSessionId()
		cookie := &http.Cookie{
        	Name: "session",
        	Value: sessionid,
        	Domain: webconfig.GetSiteDomain(),
        	Path: "/",
        	MaxAge: 86400, // One day
        }
        http.SetCookie(w, cookie)
        userInfo := identity.UserInfo{
			UserID: 1,
			UserName: "",
			Email: "",
			FullName: "",
			Role: "",
		}
    authenticator.SaveSession(ctx, sessionid, userInfo, 0)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resultsJson, err := json.Marshal(sessionInfo)
	fmt.Fprintf(w, string(resultsJson))
}

// translationMemory handles requests for for translation memory searches
func translationMemory(w http.ResponseWriter, r *http.Request) {
	q := getSingleValue(r, "query")
	title := webconfig.GetVarWithDefault("Title", defTitle)
	if len(q) == 0 {
		if acceptHTML(r) {
			content := HTMLContent{
				Title: title,
			}
			displayPage(w, "findtm.html", content)
			return
		}
		applog.Error("main.translationMemory Search query string is empty")
		http.Error(w, "Query string is empty", http.StatusInternalServerError)
		return
	}
	d := getSingleValue(r, "domain")
	applog.Infof("main.translationMemory Query: %s, domain: %s", q, d)
	ctx := context.Background()
	if tmSearcher == nil {
		err := initApp(ctx)
		if err != nil {
			applog.Errorf("findDocs error: \n%v\n", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	results, err := tmSearcher.Search(ctx, q, d, wdict)
	if err != nil {
		applog.Errorf("main.translationMemory error searching, %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	if acceptHTML(r) {
		content := HTMLContent{
			Title: title,
			TMResults: results,
		}
		displayPage(w, "findtm.html", content)
		return
	}
	resultsJson, err := json.Marshal(results)
	if err != nil {
		applog.Errorf("main.translationMemory error marshalling JSON, %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(resultsJson))
}

//Entry point for the web application
func main() {
	applog.Info("cnweb.main Iniitalizing cnweb")
	http.HandleFunc("/#", findHandler)
	http.HandleFunc("/find/", findHandler)
	http.HandleFunc("/findadvanced/", findAdvanced)
	http.HandleFunc("/findmedia", mediaDetailHandler)
	http.HandleFunc("/findsubstring", findSubstring)
	http.HandleFunc("/findtm", translationMemory)
	http.HandleFunc("/healthcheck/", healthcheck)
	http.HandleFunc("/loggedin/admin", adminHandler)
	http.HandleFunc("/loggedin/changepassword", changePasswordFormHandler)
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
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", displayHome)
	portStr := ":" + strconv.Itoa(webconfig.GetPort())
	applog.Infof("cnweb.main Starting http server on port %s\n", portStr)
	http.ListenAndServe(portStr, nil)
}
