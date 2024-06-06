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

// Package for handling of HTTP content
package httphandling

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"cloud.google.com/go/storage"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/identity"
)

// StaticHandler knows how to serve static HTTP pages
type StaticHandler interface {

	// ServeHTTP service a static page
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// staticHandler implements the StaticHandler interface
type staticHandler struct {
	enforcer SessionEnforcer
}

// NewStaticHandler creates an implementation of the StaticHandler interface
func NewStaticHandler(enforcer SessionEnforcer) StaticHandler {
	return staticHandler{
		enforcer: enforcer,
	}
}

// ServeHTTP handles requests for static files from the local file system.
func (h staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if config.PasswordProtected() {
		sessionInfo := h.enforcer.EnforceValidSession(ctx, w, r)
		if !sessionInfo.Valid {
			return
		}
	}
	fname := getStaticFileName(*r.URL)
	http.ServeFile(w, r, fname)
}

// gcsStaticHandler implements the StaticHandler interface by reading from GCS
type gcsHandler struct {
	client   *storage.Client
	bucket   string
	enforcer SessionEnforcer
}

// NewStaticHandler creates an implementation of the StaticHandler interface
func NewGcsHandler(client *storage.Client, bucket string, enforcer SessionEnforcer) StaticHandler {
	return gcsHandler{
		client:   client,
		bucket:   bucket,
		enforcer: enforcer,
	}
}

// ServeHTTP handles requests for static files from GCS
func (h gcsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	fname := strings.TrimPrefix(r.URL.Path, "/")
	if len(fname) == 0 {
		fname = "index.html"
	}
	if config.PasswordProtected() && !strings.HasSuffix(fname, ".css") && !strings.HasSuffix(fname, ".js") && !strings.HasSuffix(fname, ".ico") && !strings.HasSuffix(fname, ".png") && !strings.HasSuffix(fname, ".jpg") {
		sessionInfo := h.enforcer.EnforceValidSession(ctx, w, r)
		if !sessionInfo.Valid {
			// Forward to login page
			fname = "login_form.html"
		}
	}
	log.Printf("gcsHandler.ServeHTTP, fname = %s", fname)
	rc, err := h.client.Bucket(h.bucket).Object(fname).NewReader(ctx)
	if err != nil {
		log.Printf("gcsHandler.ServeHTTP, error reading file %s, %v", fname, err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	body, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("gcsHandler.ServeHTTP, error reading file %s, %v", fname, err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	if strings.HasSuffix(fname, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if strings.HasSuffix(fname, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(fname, ".js") {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	} else if strings.HasSuffix(fname, ".json") {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else if strings.HasSuffix(fname, ".png") {
		w.Header().Set("Content-Type", "image/png")
	} else if strings.HasSuffix(fname, ".jpg") {
		w.Header().Set("Content-Type", "image/jpeg")
	}
	fmt.Fprint(w, string(body))
}

func getStaticFileName(u url.URL) string {
	log.Printf("getStaticFileName path: %s", u.Path)
	return "./web/" + u.Path
}

// SessionEnforcer defines an interface for enforcing valid HTTP sessions
type SessionEnforcer interface {

	// EnforceValidSession ensures that, if authentication is required, the client has done it
	EnforceValidSession(ctx context.Context, w http.ResponseWriter, r *http.Request) identity.SessionInfo
}

type sessionEnforcer struct {
	authenticator identity.Authenticator
	pd            PageDisplayer
}

func NewSessionEnforcer(authenticator identity.Authenticator, pd PageDisplayer) SessionEnforcer {
	return sessionEnforcer{
		authenticator: authenticator,
		pd:            pd,
	}
}

func (s sessionEnforcer) EnforceValidSession(ctx context.Context, w http.ResponseWriter, r *http.Request) identity.SessionInfo {
	sessionInfo := identity.InvalidSession()
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionInfo = s.authenticator.CheckSession(ctx, cookie.Value)
		if sessionInfo.Authenticated != 1 {
			if AcceptHTML(r) {
				s.pd.DisplayPage(w, "login_form.html", nil)
			} else {
				http.Error(w, "Not authorized", http.StatusForbidden)
			}
			return sessionInfo
		}
	} else {
		log.Printf("EnforceValidSession, Invalid session %v, err: %v, AcceptHTML(r): %t", sessionInfo.User, err, AcceptHTML(r))
		if !AcceptHTML(r) {
			http.Error(w, "Not authorized", http.StatusForbidden)
		}
		return identity.InvalidSession()
	}
	return sessionInfo
}

type PageDisplayer interface {
	DisplayPage(w http.ResponseWriter, templateName string, content interface{})
}

type pageDisplayer struct {
	templates map[string]*template.Template
}

func NewPageDisplayer(templates map[string]*template.Template) PageDisplayer {
	return pageDisplayer{
		templates: templates,
	}
}

func (pd pageDisplayer) DisplayPage(w http.ResponseWriter, templateName string, content interface{}) {
	tmpl, ok := pd.templates[templateName]
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

func AcceptHTML(r *http.Request) bool {
	acceptEnc := r.Header.Get("Accept")
	if len(acceptEnc) > 0 && strings.Contains(acceptEnc, "text/html") {
		return true
	}
	return false
}
