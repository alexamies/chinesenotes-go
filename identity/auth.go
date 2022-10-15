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

// Package for working with corpora with private login required.
package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/alexamies/chinesenotes-go/config"
	_ "github.com/go-sql-driver/mysql"
)

type Authenticator interface {
	ChangePassword(ctx context.Context, userInfo UserInfo, oldPassword, password string) ChangePasswordResult
	CheckLogin(ctx context.Context, username, password string) ([]UserInfo, error)
	CheckSession(ctx context.Context, sessionid string) SessionInfo
	GetUser(ctx context.Context, username string) ([]UserInfo, error)
	Logout(ctx context.Context, sessionid string)
	RequestPasswordReset(ctx context.Context, email string) RequestResetResult
	ResetPassword(ctx context.Context, token, password string) bool
	SaveSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo
	UpdateSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo
}

// Authenticator holds stateful items needed for user authentication.
type AuthenticatorDBImpl struct {
	database               *sql.DB
	changePasswordStmt     *sql.Stmt
	checkSessionStmt       *sql.Stmt
	getResetRequestStmt    *sql.Stmt
	getUserStmt            *sql.Stmt
	getUserByEmailStmt     *sql.Stmt
	loginStmt              *sql.Stmt
	logoutStmt             *sql.Stmt
	requestResetStmt       *sql.Stmt
	saveSessionStmt        *sql.Stmt
	updateSessionStmt      *sql.Stmt
	updateResetRequestStmt *sql.Stmt
}

type ChangePasswordResult struct {
	OldPasswordValid bool
	ChangeSuccessful bool
	ShowNewForm      bool
}

type RequestResetResult struct {
	EmailValid          bool
	RequestResetSuccess bool
	ShowNewForm         bool
	User                UserInfo
	Token               string
}

type RequestResetRecord struct {
	EmailValid          bool   `firestore:"email_valid"`
	RequestResetSuccess bool   `firestore:"request_reset_success"`
	Email               string `firestore:"email"`
	UserName            string `firestore:"username"`
	Token               string `firestore:"token"`
}

type SessionInfo struct {
	Authenticated int
	Valid         bool
	User          UserInfo
}

type SessionRecord struct {
	Authenticated int    `firestore:"authenticated"`
	Valid         bool   `firestore:"valid"`
	UserName      string `firestore:"username"`
}

type UserInfo struct {
	UserID   int    `firestore:"userid"`
	UserName string `firestore:"username"`
	Email    string `firestore:"email"`
	FullName string `firestore:"full_name"`
	Role     string `firestore:"role"`
	Password string `firestore:"password"`
}

// NewAuthenticator creates but does not initialize an Authenticator object.
//
// Params:
//   isProtected - set this to true if only the site is password protected
func NewAuthenticator(ctx context.Context) (Authenticator, error) {
	a := AuthenticatorDBImpl{}
	err := a.initStatements(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("NewAuthenticator: authenticator initialized")
	return &a, nil
}

// initStatements opens a database connection and prepares the statements.
func (a *AuthenticatorDBImpl) initStatements(ctx context.Context) error {
	conString := config.DBConfig()
	var err error
	a.database, err = sql.Open("mysql", conString)
	if err != nil {
		log.Printf("FATAL: could not connect to the database, %v", err)
		return err
	}

	a.loginStmt, err = a.database.PrepareContext(ctx,
		`SELECT user.UserID, UserName, Email, FullName, Role 
		FROM user, passwd 
		WHERE UserName = ? 
		AND user.UserID = passwd.UserID
		AND Password = ?
		LIMIT 1`)
	if err != nil {
		return err
	}

	a.saveSessionStmt, err = a.database.PrepareContext(ctx,
		`INSERT INTO
		  session (SessionID, UserID, Authenticated)
		VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}

	// Need to fix use of username in session table. Should be UserId
	a.checkSessionStmt, err = a.database.PrepareContext(ctx,
		`SELECT user.UserID, UserName, Email, FullName, Role, Authenticated
		FROM user, session 
		WHERE SessionID = ? 
		AND user.UserID = session.UserID
		LIMIT 1`)
	if err != nil {
		return err
	}

	a.logoutStmt, err = a.database.PrepareContext(ctx,
		`UPDATE session SET
		Authenticated = 0
		WHERE SessionID = ?`)
	if err != nil {
		return err
	}

	a.updateSessionStmt, err = a.database.PrepareContext(ctx,
		`UPDATE session SET
		Authenticated = ?,
		UserID = ?
		WHERE SessionID = ?`)
	if err != nil {
		return err
	}

	a.changePasswordStmt, err = a.database.PrepareContext(ctx,
		`UPDATE passwd SET
		Password = ?
		WHERE UserID = ?`)
	if err != nil {
		return err
	}

	a.getUserStmt, err = a.database.PrepareContext(ctx,
		`SELECT user.UserID, UserName, Email, FullName, Role 
		FROM user
		WHERE UserName = ? 
		LIMIT 1`)
	if err != nil {
		return err
	}

	a.requestResetStmt, err = a.database.PrepareContext(ctx,
		`INSERT INTO
		passwdreset (Token, UserID)
		VALUES (?, ?)`)
	if err != nil {
		return err
	}

	a.getUserByEmailStmt, err = a.database.PrepareContext(ctx,
		`SELECT user.UserID, UserName, Email, FullName, Role 
		FROM user
		WHERE Email = ? 
		LIMIT 1`)
	if err != nil {
		return err
	}

	a.getResetRequestStmt, err = a.database.PrepareContext(ctx,
		`SELECT UserID
		FROM passwdreset
		WHERE Token = ?
		AND Valid = 1
		LIMIT 1`)
	if err != nil {
		return err
	}

	a.updateResetRequestStmt, err = a.database.PrepareContext(ctx,
		`UPDATE passwdreset SET
		Valid = 0
		WHERE Token = ?`)
	if err != nil {
		return err
	}

	return nil
}

// ChangePassword enables the user to change passwords.
func (a *AuthenticatorDBImpl) ChangePassword(ctx context.Context, userInfo UserInfo,
	oldPassword, password string) ChangePasswordResult {
	users, err := a.CheckLogin(ctx, userInfo.UserName, oldPassword)
	if err != nil {
		log.Printf("ChangePassword checking login, %v", err)
		return ChangePasswordResult{true, false, false}
	}
	if len(users) != 1 {
		log.Println("ChangePassword, user or password wrong: ",
			userInfo.UserName)
		return ChangePasswordResult{false, false, false}
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	result, err := a.changePasswordStmt.ExecContext(ctx, hstr, userInfo.UserID)
	if err != nil {
		log.Printf("ChangePassword, Error: %v", err)
		return ChangePasswordResult{true, false, false}
	}
	rowsAffected, _ := result.RowsAffected()
	log.Println("ChangePassword, rows updated:", rowsAffected)
	return ChangePasswordResult{true, true, false}
}

// CheckLogin checks the password when the user logs in.
func (a *AuthenticatorDBImpl) CheckLogin(ctx context.Context,
	username, password string) ([]UserInfo, error) {
	if a.loginStmt == nil {
		return []UserInfo{}, nil
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	//log.Println("CheckLogin, username, hstr:", username, hstr)
	results, err := a.loginStmt.QueryContext(ctx, username, hstr)
	if err != nil {
		log.Printf("CheckLogin, Error for username: %s, %v\n", username, err)
		return []UserInfo{}, err
	}
	defer results.Close()

	users := []UserInfo{}
	for results.Next() {
		user := UserInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role)
		users = append(users, user)
	}
	if len(users) == 0 {
		log.Printf("CheckLogin, user or password wrong for user %s\n", username)
		u, _ := a.GetUser(ctx, username)
		if len(u) == 0 {
			log.Printf("CheckLogin, user %s not found\n", username)
		}
	}
	return users, nil
}

// CheckSession checks the session when the user requests a page.
func (a *AuthenticatorDBImpl) CheckSession(ctx context.Context, sessionid string) SessionInfo {
	sessions := a.checkSessionStore(ctx, sessionid)
	if len(sessions) != 1 {
		return InvalidSession()
	}
	log.Printf("CheckSession, Authenticated = %v\n", sessions[0].Authenticated)
	return sessions[0]
}

// checkSessionStore checks the session when the user requests a page
func (a *AuthenticatorDBImpl) checkSessionStore(ctx context.Context,
	sessionid string) []SessionInfo {
	log.Printf("checkSessionStore, sessionid: %s\n", sessionid)
	if a.checkSessionStmt == nil {
		log.Println("checkSessionStore, checkSessionStmt == nil")
		return []SessionInfo{}
	}
	results, err := a.checkSessionStmt.QueryContext(ctx, sessionid)
	if err != nil {
		log.Printf("checkSessionStore, Error: %v\n", err)
	}
	defer results.Close()

	sessions := []SessionInfo{}
	for results.Next() {
		user := UserInfo{}
		session := SessionInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role, &session.Authenticated)
		session.User = user
		if session.Authenticated == 1 {
			session.Valid = true
			sessions = append(sessions, session)
		}
	}
	log.Printf("checkSessionStore, sessions found: %d\n", len(sessions))
	return sessions
}

// GetUser gets the user information.
func (a *AuthenticatorDBImpl) GetUser(ctx context.Context,
	username string) ([]UserInfo, error) {
	log.Println("getUser, username:", username)
	if a.getUserStmt == nil {
		return []UserInfo{}, nil
	}
	results, err := a.getUserStmt.QueryContext(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("getUser, Error for username %s: %v", username, err)
	}
	defer results.Close()

	users := []UserInfo{}
	for results.Next() {
		user := UserInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role)
		users = append(users, user)
	}
	return users, nil
}

// InvalidSession creates an empty session struct.
func InvalidSession() SessionInfo {
	userInfo := UserInfo{
		UserID:   1,
		UserName: "",
		Email:    "",
		FullName: "",
		Role:     "",
	}
	return SessionInfo{
		Authenticated: 0,
		Valid:         false,
		User:          userInfo,
	}
}

// Empty session struct for an unauthenticated session
func InvalidUser() UserInfo {
	return UserInfo{
		UserID:   1,
		UserName: "",
		Email:    "",
		FullName: "",
		Role:     "",
	}
}

// Generate a new session id after login
func IsAuthorized(user UserInfo, permission string) bool {
	if user.Role == "admin" || user.Role == "editor" || user.Role == "translator" {
		return true
	}
	return false
}

// Logout logs the user out of the current session.
func (a *AuthenticatorDBImpl) Logout(ctx context.Context, sessionid string) {
	log.Printf("Logout, sessionid: %s\n", sessionid)
	result, err := a.logoutStmt.ExecContext(ctx, sessionid)
	if err != nil {
		log.Printf("Logout, Error: %v\n", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Printf("Logout, rows updated: %d\n", rowsAffected)
	}
}

// Generate a new session id after login
func NewSessionId() string {
	value := "invalid"
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("NewSessionId, Error: %v", err)
		return value
	}
	val, err := base64.URLEncoding.EncodeToString(b), err
	if err != nil {
		log.Println("NewSessionId, Error: ", err)
		return value
	}
	return val
}

// Old password does not match
func OldPasswordDoesNotMatch() ChangePasswordResult {
	return ChangePasswordResult{false, true, false}
}

// RequestPasswordReset requests a password reset, to be sent by email.
func (a *AuthenticatorDBImpl) RequestPasswordReset(ctx context.Context,
	email string) RequestResetResult {
	log.Println("RequestPasswordReset, email:", email)
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("RequestPasswordReset, Error: %v", err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	token, err := base64.URLEncoding.EncodeToString(b), err
	if err != nil {
		log.Println("RequestPasswordReset, Error: ", err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	results, err := a.getUserByEmailStmt.QueryContext(ctx, email)
	if err != nil {
		log.Printf("RequestPasswordReset, Error for email %s: %v", email, err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	defer results.Close()
	users := []UserInfo{}
	for results.Next() {
		user := UserInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role)
		users = append(users, user)
	}

	if len(users) != 1 {
		log.Printf("RequestPasswordReset, No email: %v", email)
		return RequestResetResult{false, false, true, InvalidUser(), ""}
	}

	result, err := a.requestResetStmt.ExecContext(ctx, token, users[0].UserID)
	if err != nil {
		log.Println("RequestPasswordReset, Error for email: ", email, err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	rowsAffected, _ := result.RowsAffected()
	log.Println("RequestPasswordReset, rows updated: ", rowsAffected)
	return RequestResetResult{true, true, false, users[0], token}
}

// ResetPassword resets a password.
func (a *AuthenticatorDBImpl) ResetPassword(ctx context.Context, token, password string) bool {
	log.Println("ResetPassword, token:", token)
	results, err := a.getResetRequestStmt.QueryContext(ctx, token)
	if err != nil {
		log.Printf("ResetPassword, Error for token %s: %v", token, err)
		return false
	}
	defer results.Close()
	userIds := []string{}
	for results.Next() {
		userId := ""
		results.Scan(&userId)
		userIds = append(userIds, userId)
	}
	if len(userIds) != 1 {
		log.Printf("ResetPassword, No userId: %s", token)
		return false
	}
	userId := userIds[0]

	// Change password
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	result, err := a.changePasswordStmt.ExecContext(ctx, hstr, userId)
	if err != nil {
		log.Printf("ResetPassword, Error setting password: %v", err)
		return false
	}
	rowsAffected, _ := result.RowsAffected()
	log.Println("ResetPassword, rows updated for change pwd:", rowsAffected)

	// Update reset token so that it cannot be used again
	result, err = a.updateResetRequestStmt.ExecContext(ctx, token)
	if err != nil {
		log.Printf("ResetPassword, Error updating reset token: %v", err)
	}
	rowsAffected, _ = result.RowsAffected()
	log.Println("ResetPassword, rows updated for token:", rowsAffected)

	return true
}

// SaveSession saves an authenticated session to the database
func (a *AuthenticatorDBImpl) SaveSession(ctx context.Context,
	sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	log.Printf("SaveSession, sessionid: %s\n", sessionid)
	result, err := a.saveSessionStmt.ExecContext(ctx, sessionid, userInfo.UserID,
		authenticated)
	if err != nil {
		log.Printf("SaveSession, Error for user %d, %v\n ", userInfo.UserID, err)
		return InvalidSession()
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("SaveSession, rows updated: %d\n", rowsAffected)
	return SessionInfo{
		Authenticated: authenticated,
		Valid:         true,
		User:          userInfo,
	}
}

// UpdateSession logs a user in when they already have an unauthenticated session.
func (a *AuthenticatorDBImpl) UpdateSession(ctx context.Context,
	sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	result, err := a.updateSessionStmt.ExecContext(ctx, authenticated,
		userInfo.UserID, sessionid)
	if err != nil {
		log.Printf("UpdateSession, Error: %v", err)
		return InvalidSession()
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("UpdateSession, rows updated: %d", rowsAffected)
	return SessionInfo{
		Authenticated: authenticated,
		User:          userInfo,
	}
}
