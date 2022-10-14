package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
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
		log.Printf("ChangePassword, user or password wrong: %s", userInfo.UserName)
		return ChangePasswordResult{
			OldPasswordValid: false,
			ChangeSuccessful: false,
			ShowNewForm:      false,
		}
	}
	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	_, err = user.Update(ctx, []firestore.Update{{Path: "Password", Value: hstr}})
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
	log.Printf("CheckSession, sessionid: %s\n", sessionid)
	uPath := a.path + "/sessions"
	colRef := a.client.Collection(uPath)
	docRef := colRef.Doc(sessionid)
	doc, err := docRef.Get(ctx)
	if err != nil {
		return SessionInfo{}
	}
	var sessionRecord SessionRecord
	if err := doc.DataTo(&sessionRecord); err != nil {
		return SessionInfo{}
	}
	return SessionInfo{
		Authenticated: sessionRecord.Authenticated,
		Valid:         sessionRecord.Valid,
		User: UserInfo{
			UserName: sessionRecord.UserName,
		},
	}
}

func (a authenticatorFS) GetUser(ctx context.Context, username string) ([]UserInfo, error) {
	log.Println("GetUser, username:", username)
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
	return []UserInfo{user}, nil
}

func (a authenticatorFS) Logout(ctx context.Context, sessionid string) {
	log.Printf("Logout, sessionid: %s\n", sessionid)
	uPath := a.path + "/sessions"
	colRef := a.client.Collection(uPath)
	docRef := colRef.Doc(sessionid)
	_, err := docRef.Delete(ctx)
	if err != nil {
		log.Printf("CheckSession, error logging out for sessionid: %s\n", sessionid)
	}
}

func (a authenticatorFS) RequestPasswordReset(ctx context.Context, email string) RequestResetResult {
	log.Printf("RequestPasswordReset, email: %s\n", email)
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
	user, err := a.getUserByEmail(ctx, email)
	if err != nil {
		log.Println("RequestPasswordReset, Error: ", err)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	uPath := a.path + "/resets"
	colRef := a.client.Collection(uPath)
	docRef := colRef.Doc(email)
	_, err = docRef.Set(ctx, RequestResetRecord{
		EmailValid:          true,
		RequestResetSuccess: true,
		Email:               email,
		Token:               token,
	})
	if err != nil {
		log.Printf("RequestPasswordReset, error setting record for email %s\n", email)
		return RequestResetResult{true, false, true, InvalidUser(), ""}
	}
	return RequestResetResult{
		EmailValid:          true,
		RequestResetSuccess: true,
		ShowNewForm:         false,
		User:                user,
		Token:               token,
	}
}

func (a authenticatorFS) ResetPassword(ctx context.Context, token, password string) bool {
	return false
}

func (a authenticatorFS) SaveSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	log.Printf("SaveSession, sessionid: %s\n", sessionid)
	uPath := a.path + "/sessions"
	colRef := a.client.Collection(uPath)
	docRef := colRef.Doc(sessionid)
	_, err := docRef.Set(ctx, SessionRecord{
		Authenticated: 1,
		Valid:         true,
		UserName:      userInfo.UserName,
	})
	if err != nil {
		log.Printf("SaveSession, error saving session for sessionid: %s, user: %s\n", sessionid, userInfo.UserName)
		return SessionInfo{}
	}
	return SessionInfo{
		Authenticated: 1,
		Valid:         true,
		User:          userInfo,
	}
}

func (a authenticatorFS) UpdateSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	return a.SaveSession(ctx, sessionid, userInfo, authenticated)
}

func (a authenticatorFS) getUserByEmail(ctx context.Context, email string) (UserInfo, error) {
	log.Printf("getUserByEmail, email: %s", email)
	uPath := a.path + "/users"
	colRef := a.client.Collection(uPath)
	q := colRef.Where("substrings", "array-contains", email).Limit(100)
	iter := q.Documents(ctx)
	defer iter.Stop()
	for {
		ds, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return UserInfo{}, fmt.Errorf("getUserByEmail iteration error: %v", err)
		}
		var d UserInfo
		err = ds.DataTo(&d)
		if err != nil {
			return UserInfo{}, fmt.Errorf("getUserByEmail type conversion error: %v", err)
		}
		if len(d.UserName) > 0 {
			return d, nil
		}
	}
	return UserInfo{}, fmt.Errorf("user not found for email: %s", email)
}
