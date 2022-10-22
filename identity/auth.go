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
	"encoding/base64"
	"log"
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
	FullName string `firestore:"fullname"`
	Role     string `firestore:"role"`
	Password string `firestore:"password"`
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
