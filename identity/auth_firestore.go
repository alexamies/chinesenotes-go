package identity

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
)

// fsClient defines Firestore interfaces needed
type fsClient interface {
	Collection(path string) *firestore.CollectionRef
}

// Implements the Authenticator interface with a Firestore client
type authenticatorFS struct {
	client fsClient
	path   string
}

// Create a new Authenticator with a Firestore client
func NewAuthenticatorFS(client fsClient, path string) Authenticator {
	return authenticatorFS{
		client: client,
		path:   path,
	}
}

func (a authenticatorFS) ChangePassword(ctx context.Context, userInfo UserInfo, oldPassword, password string) ChangePasswordResult {
	uPath := a.path + "/users"
	colRef := a.client.Collection(uPath)
	user := colRef.Doc(userInfo.UserName)
	users, err := a.CheckLogin(ctx, userInfo.UserName, oldPassword)
	if err != nil {
		log.Printf("ChangePassword checking login, %v", err)
		return ChangePasswordResult{
			OldPasswordValid: true,
			ChangeSuccessful: false,
			ShowNewForm:      false,
		}
	}
	if len(users) != 1 {
		log.Println("ChangePassword, user or password wrong: ",
			userInfo.UserName)
		return ChangePasswordResult{
			OldPasswordValid: false,
			ChangeSuccessful: false,
			ShowNewForm:      false,
		}
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	_, err = user.Set(ctx, UserInfo{
		Password: hstr,
	})
	if err != nil {
		log.Printf("ChangePassword, Error: %v", err)
		return ChangePasswordResult{
			OldPasswordValid: true,
			ChangeSuccessful: false,
			ShowNewForm:      false,
		}
	}
	log.Println("ChangePassword, successful")
	return ChangePasswordResult{
		OldPasswordValid: true,
		ChangeSuccessful: true,
		ShowNewForm:      false,
	}
}

func (a authenticatorFS) CheckLogin(ctx context.Context, username, password string) ([]UserInfo, error) {
	log.Printf("CheckLogin for username %s", username)
	uPath := a.path + "/users"
	colRef := a.client.Collection(uPath)
	docRef := colRef.Doc(username)
	var user UserInfo
	doc, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	if user.Password != hstr {
		log.Printf("CheckLogin, username %s, hstr %s does not match", username, hstr)
		return []UserInfo{}, nil
	}
	return []UserInfo{user}, nil
}

func (a authenticatorFS) CheckSession(ctx context.Context, sessionid string) SessionInfo {
	return SessionInfo{}
}

func (a authenticatorFS) GetUser(ctx context.Context, username string) ([]UserInfo, error) {
	return []UserInfo{}, nil
}

func (a authenticatorFS) Logout(ctx context.Context, sessionid string) {
	// pass
}

func (a authenticatorFS) RequestPasswordReset(ctx context.Context, email string) RequestResetResult {
	return RequestResetResult{}
}

func (a authenticatorFS) ResetPassword(ctx context.Context, token, password string) bool {
	return false
}

func (a authenticatorFS) SaveSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	return SessionInfo{}
}

func (a authenticatorFS) UpdateSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	return SessionInfo{}
}
