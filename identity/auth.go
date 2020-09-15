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

package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/webconfig"
)

// Encapsulates translation memory searcher
type Authenticator struct {
	database *sql.DB
	domain *string
	changePasswordStmt *sql.Stmt
	checkSessionStmt *sql.Stmt
	getResetRequestStmt *sql.Stmt
	getUserStmt *sql.Stmt
	getUserByEmailStmt *sql.Stmt
	loginStmt *sql.Stmt
	logoutStmt *sql.Stmt
	requestResetStmt *sql.Stmt
	saveSessionStmt *sql.Stmt
	updateSessionStmt *sql.Stmt
	updateResetRequestStmt *sql.Stmt
}

type ChangePasswordResult struct {
	OldPasswordValid bool
	ChangeSuccessful bool
	ShowNewForm bool
}

type RequestResetResult struct {
	EmailValid bool
	RequestResetSuccess bool
	ShowNewForm bool
	User UserInfo
	Token string
}

type SessionInfo struct {
	Authenticated int
	Valid bool
	User UserInfo
}

type UserInfo struct {
	UserID int
	UserName, Email, FullName, Role string
}

// NewAuthenticator creates but does not initialize an Authenticator object.
//
// Params:
//   isProtected - set this to true if only the site is password protected
func NewAuthenticator(ctx context.Context) (*Authenticator, error) {
	a := Authenticator{}
	err := a.initStatements(ctx)
	if err != nil {
		return nil, err
	}
	applog.Info("NewAuthenticator: authenticator initialized")
	return &a, nil
}

// initStatements opens a database connection and prepares the statements.
func (a *Authenticator) initStatements(ctx context.Context) error {
	conString := webconfig.DBConfig()
	var err error
	a.database, err = sql.Open("mysql", conString)
	if err != nil {
		applog.Error("FATAL: could not connect to the database, ",
			err)
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
func (a *Authenticator) ChangePassword(ctx context.Context, userInfo UserInfo,
			oldPassword, password string) ChangePasswordResult {
	users, err := a.CheckLogin(ctx, userInfo.UserName, oldPassword)
	if err != nil {
		applog.Error("ChangePassword checking login, ", err)
		return ChangePasswordResult{true, false, false}
	}
	if len(users) != 1 {
		applog.Info("ChangePassword, user or password wrong: ",
			userInfo.UserName)
		return ChangePasswordResult{false, false, false}
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	result, err := a.changePasswordStmt.ExecContext(ctx, hstr, userInfo.UserID)
	if err != nil {
		applog.Error("ChangePassword, Error: ", err)
		return ChangePasswordResult{true, false, false}
	} 
	rowsAffected, _ := result.RowsAffected()
	applog.Info("ChangePassword, rows updated:", rowsAffected)
	return ChangePasswordResult{true, true, false}
}

// CheckLogin checks the password when the user logs in.
func (a *Authenticator) CheckLogin(ctx context.Context,
		username, password string) ([]UserInfo, error) {
	if a.loginStmt == nil {
		return []UserInfo{}, nil
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	//applog.Info("CheckLogin, username, hstr:", username, hstr)
	results, err := a.loginStmt.QueryContext(ctx, username, hstr)
	defer results.Close()
	if err != nil {
		applog.Errorf("CheckLogin, Error for username: %s, %v\n", username, err)
		return []UserInfo{}, err
	}

	users := []UserInfo{}
	for results.Next() {
		user := UserInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role)
		users = append(users, user)
	}
	if len(users) == 0 {
		applog.Infof("CheckLogin, user or password wrong for user %s\n", username)
		u, _ := a.GetUser(ctx, username)
		if len(u) == 0 {
			applog.Infof("CheckLogin, user %s not found\n", username)
		}
	}
	return users, nil
}

// CheckSession checks the session when the user requests a page.
func (a *Authenticator) CheckSession(ctx context.Context, sessionid string) SessionInfo {
	sessions := a.checkSessionStore(ctx, sessionid)
	if len(sessions) != 1 {
		return InvalidSession()
	}
	applog.Infof("CheckSession, Authenticated = %v\n", sessions[0].Authenticated)
	return sessions[0]
}

// checkSessionStore checks the session when the user requests a page
func (a *Authenticator) checkSessionStore(ctx context.Context,
		sessionid string) []SessionInfo {
	applog.Infof("checkSessionStore, sessionid: %s\n", sessionid)
	if a.checkSessionStmt == nil {
		applog.Info("checkSessionStore, checkSessionStmt == nil")
		return []SessionInfo{}
	}
	results, err := a.checkSessionStmt.QueryContext(ctx, sessionid)
	if err != nil {
		applog.Errorf("checkSessionStore, Error: %v\n", err)
	}
	defer results.Close()

	sessions := []SessionInfo{}
	for results.Next() {
		user := UserInfo{}
		session := SessionInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role, &session.Authenticated)
		session.User = user
		session.Valid = true
		sessions = append(sessions, session)
	}
	applog.Infof("checkSessionStore, sessions found: %d\n", len(sessions))
	return sessions
}

// GetUser gets the user information.
func (a *Authenticator) GetUser(ctx context.Context,
		username string) ([]UserInfo, error) {
	applog.Info("getUser, username:", username)
	if a.getUserStmt == nil {
		return []UserInfo{}, nil
	}
	results, err := a.getUserStmt.QueryContext(ctx, username)
	defer results.Close()
	if err != nil {
		applog.Error("getUser, Error for username: ", username, err)
		return []UserInfo{}, err
	}

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
		UserID: 1,
		UserName: "",
		Email: "",
		FullName: "",
		Role: "",
	}
	return SessionInfo{
		Authenticated: 0,
		Valid: false,
		User: userInfo,
	}
}

// Empty session struct for an unauthenticated session
func InvalidUser() UserInfo {
	return UserInfo{
		UserID: 1,
		UserName: "",
		Email: "",
		FullName: "",
		Role: "",
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
func (a *Authenticator) Logout(ctx context.Context, sessionid string) {
	applog.Infof("Logout, sessionid: %s\n", sessionid)
	result, err := a.logoutStmt.ExecContext(ctx, sessionid)
	if err != nil {
		applog.Errorf("Logout, Error: %v\n", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		applog.Infof("Logout, rows updated: %d\n", rowsAffected)
	}
}

// Generate a new session id after login
func NewSessionId() string {
	value := "invalid"
	b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        applog.Error("NewSessionId, Error: ", err)
        return value
    }
    val, err := base64.URLEncoding.EncodeToString(b), err
	if err != nil {
		applog.Info("NewSessionId, Error: ", err)
		return value
	}
	return val
}

// Old password does not match
func OldPasswordDoesNotMatch() ChangePasswordResult {
	return ChangePasswordResult{false, true, false}
}

// RequestPasswordReset requests a password reset, to be sent by email.
func (a *Authenticator) RequestPasswordReset(ctx context.Context,
		email string) RequestResetResult {
	applog.Info("RequestPasswordReset, email:", email)
	b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        applog.Error("RequestPasswordReset, Error: ", err)
        return RequestResetResult{true, false, true, InvalidUser(), ""}
    }
    token, err := base64.URLEncoding.EncodeToString(b), err
	if err != nil {
		applog.Info("RequestPasswordReset, Error: ", err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	results, err := a.getUserByEmailStmt.QueryContext(ctx, email)
	defer results.Close()
	if err != nil {
		applog.Error("RequestPasswordReset, Error for email: ", email, err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	users := []UserInfo{}
	for results.Next() {
		user := UserInfo{}
		results.Scan(&user.UserID, &user.UserName, &user.Email, &user.FullName,
			&user.Role)
		users = append(users, user)
	}

	if len(users) != 1 {
		applog.Error("RequestPasswordReset, No email: ", email)
		return RequestResetResult{false, false, true, InvalidUser(), ""}
	}

	result, err := a.requestResetStmt.ExecContext(ctx, token, users[0].UserID)
	if err != nil {
		applog.Info("RequestPasswordReset, Error for email: ", email, err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	rowsAffected, _ := result.RowsAffected()
	applog.Info("RequestPasswordReset, rows updated: ", rowsAffected)
	return RequestResetResult{true, true, false, users[0], token}
}

// ResetPassword resets a password.
func (a *Authenticator) ResetPassword(ctx context.Context, token, password string) bool {
	applog.Info("ResetPassword, token:", token)
	results, err := a.getResetRequestStmt.QueryContext(ctx, token)
	defer results.Close()
	if err != nil {
		applog.Error("ResetPassword, Error for token: ", token, err)
		return false
	}
	userIds := []string{}
	for results.Next() {
		userId := ""
		results.Scan(&userId)
		userIds = append(userIds, userId)
	}
	if len(userIds) != 1 {
		applog.Error("ResetPassword, No userId: ", token)
		return false
	}
	userId := userIds[0]

	// Change password
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	result, err := a.changePasswordStmt.ExecContext(ctx, hstr, userId)
	if err != nil {
		applog.Error("ResetPassword, Error setting password: ", err)
		return false
	} 
	rowsAffected, _ := result.RowsAffected()
	applog.Info("ResetPassword, rows updated for change pwd:", rowsAffected)

	// Update reset token so that it cannot be used again
	result, err = a.updateResetRequestStmt.ExecContext(ctx, token)
	if err != nil {
		applog.Error("ResetPassword, Error updating reset token: ", err)
	} 
	rowsAffected, _ = result.RowsAffected()
	applog.Info("ResetPassword, rows updated for token:", rowsAffected)

	return true
}

// SaveSession saves an authenticated session to the database
func (a *Authenticator) SaveSession(ctx context.Context,
		sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	applog.Infof("SaveSession, sessionid: %s\n", sessionid)
	result, err := a.saveSessionStmt.ExecContext(ctx, sessionid, userInfo.UserID,
		authenticated)
	if err != nil {
		applog.Infof("SaveSession, Error for user %d, %v\n ", userInfo.UserID, err)
		return InvalidSession()
	}
	rowsAffected, _ := result.RowsAffected()
	applog.Infof("SaveSession, rows updated: %d\n", rowsAffected)
	return SessionInfo{
		Authenticated: authenticated,
		Valid: true,
		User: userInfo,
	}
}

// UpdateSession logs a user in when they already have an unauthenticated session.
func (a *Authenticator) UpdateSession(ctx context.Context,
		sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	result, err := a.updateSessionStmt.ExecContext(ctx, authenticated,
		userInfo.UserID, sessionid)
	if err != nil {
		applog.Error("UpdateSession, Error: ", err)
		return InvalidSession()
	} 
	rowsAffected, _ := result.RowsAffected()
	applog.Infof("UpdateSession, rows updated: %d\n", rowsAffected)
	return SessionInfo{
		Authenticated: authenticated,
		User: userInfo,
	}
}