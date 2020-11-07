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


// Unit tests for the identity package
package identity

import (
	"context"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
)

// Test check login method
func TestChangePassword(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestChangePassword authenticator not initialized, %v", err)
		return
	}
	userInfo := UserInfo{
		UserID: 1,
		UserName: "guest",
		Email: "",
		FullName: "",
		Role: "",
	}
	result := a.ChangePassword(ctx, userInfo, "guest", "guest")
	if !result.ChangeSuccessful {
		t.Error("TestChangePassword: !result.ChangeSuccessful")
	}
}

// Test check login method
func TestCheckLogin1(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestCheckLogin1 authenticator not initialized, %v", err)
		return
	}
	user, err := a.CheckLogin(ctx, "guest", "guest")
	if err != nil {
		t.Errorf("TestCheckLogin1: error, %v", err)
	}
	if len(user) != 1 {
		t.Errorf("TestCheckLogin1: len(user) != 1, %d", len(user))
	}
}

// Test check login method
func TestCheckLogin2(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestCheckLogin2 authenticator not initialized, %v", err)
		return
	}
	user, err := a.CheckLogin(ctx, "admin", "changeme")
	if err != nil {
		t.Error("TestCheckLogin2: error, ", err)
	}
	if len(user) != 0 {
		t.Error("TestCheckLogin2: len(user) != 0, ", len(user))
	}
}

// Test CheckSession function with expected result that session does not exist
func TestCheckSession1(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestCheckSession1 authenticator not initialized, %v", err)
		return
	}
	sessionid := NewSessionId()
	session := a.CheckSession(ctx, sessionid)
	if session.Valid {
		t.Error("TestCheckSession1: session.Valid, sessionid: ",
			sessionid)
	}
}

// Test CheckSession function with session that does exist
func TestCheckSession2(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestCheckSession2 authenticator not initialized, %v", err)
		return
	}
	sessionid := NewSessionId()
	userInfo := UserInfo{
		UserID: 1,
		UserName: "unittest",
		Email: "",
		FullName: "",
		Role: "",
	}
	a.SaveSession(ctx, sessionid, userInfo, 1)
	session := a.CheckSession(ctx, sessionid)
	if (session.Authenticated != 1) {
		t.Error("TestCheckSession2: session.Authenticated != 1, SessionID: ",
			sessionid)
	}
}

// Test CheckSession function with session that does exist
func TestCheckSession3(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestCheckSession3 authenticator not initialized, %v", err)
		return
	}
	sessionid := NewSessionId()
	userInfo := UserInfo{
		UserID: 1,
		UserName: "guest",
		Email: "",
		FullName: "",
		Role: "",
	}
	a.SaveSession(ctx, sessionid, userInfo, 1)
	session := a.CheckSession(ctx, sessionid)
	if session.Authenticated != 1 {
		t.Error("TestCheckSession3: session.Authenticated != 1, SessionID: ",
			sessionid)
	}
}

func TestGetUser(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestGetUser authenticator not initialized, %v", err)
		return
	}
	username := "guest"
	users, err := a.GetUser(ctx, username)
	if err != nil {
		t.Error("TestGetUser: error: ", err)
		return
	}
	if len(users) == 0 {
		t.Error("TestGetUser: username not found: ", username)
	}
}

// Test check login method
func TestNewSessionId(t *testing.T) {
	sessionid := NewSessionId()
	if sessionid == "invalid" {
		t.Error("TestNewSessionId: ", sessionid)
	}
}

// Test Logout method
func TestLogout(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestGetUser authenticator not initialized, %v", err)
		return
	}
	sessionid := NewSessionId()
	a.Logout(ctx, sessionid)
}

func TestRequestPasswordReset(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestGetUser authenticator not initialized, %v", err)
		return
	}
	result := a.RequestPasswordReset(ctx, "mail.example.com")
	if result.EmailValid {
		t.Error("TestRequestPasswordReset: result.EmailValid not expected")
	}
}

func TestPasswordReset(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestGetUser authenticator not initialized, %v", err)
		return
	}
	result := a.ResetPassword(ctx, "invalid token", "mail.example.com")
	if result {
		t.Error("TestPasswordReset: result true not expected")
	}
}

func TestSaveSession(t *testing.T) {
	if !config.PasswordProtected() {
		return
	}
	ctx := context.Background()
	a, err := NewAuthenticator(ctx)
	if err != nil {
		t.Errorf("TestGetUser authenticator not initialized, %v", err)
		return
	}
	sessionid := NewSessionId()
	userInfo := UserInfo{
		UserID: 1,
		UserName: "testuser",
		Email: "",
		FullName: "",
		Role: "",
	}
	a.SaveSession(ctx, sessionid, userInfo, 1)
}
