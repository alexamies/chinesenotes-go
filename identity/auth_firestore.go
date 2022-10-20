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
	corpus string
}

// Create a new Authenticator with a Firestore client
func NewAuthenticator(client fsClient, corpus string) Authenticator {
	return authenticatorFS{
		client: client,
		corpus: corpus,
	}
}

func (a authenticatorFS) ChangePassword(ctx context.Context, userInfo UserInfo, oldPassword, password string) ChangePasswordResult {
	uPath := a.corpus + "_users"
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
	log.Printf("CheckLogin for username %s in corpus %s", username, a.corpus)
	uPath := a.corpus + "_users"
	if a.client == nil {
		log.Println("CheckLogin, Firestore client is nil")
		return nil, fmt.Errorf("server not configured")
	}
	colRef := a.client.Collection(uPath)
	if colRef == nil {
		log.Println("CheckLogin, colRef is nil")
		return nil, fmt.Errorf("server not configured")
	}
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
		log.Printf("CheckLogin, username %s, password %s does not match", username, hstr)
		return []UserInfo{}, nil
	}
	return []UserInfo{user}, nil
}

func (a authenticatorFS) CheckSession(ctx context.Context, sessionid string) SessionInfo {
	log.Printf("CheckSession, sessionid: %s\n", sessionid)
	uPath := a.corpus + "_sessions"
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
	uPath := a.corpus + "_users"
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
	uPath := a.corpus + "_sessions"
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
	rPath := a.corpus + "_resets"
	colRef := a.client.Collection(rPath)
	docRef := colRef.Doc(token)
	_, err = docRef.Set(ctx, RequestResetRecord{
		EmailValid:          true,
		RequestResetSuccess: true,
		Email:               email,
		UserName:            user.UserName,
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
	log.Println("ResetPassword, token:", token)
	rPath := a.corpus + "_resets"
	colRef := a.client.Collection(rPath)
	docRef := colRef.Doc(token)
	doc, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("ResetPassword, error looking up token %s\n", token)
		return false
	}
	var rRecord RequestResetRecord
	if err := doc.DataTo(&rRecord); err != nil {
		log.Printf("ResetPassword, error reading request record for %s\n", token)
		return false
	}
	if !rRecord.EmailValid {
		log.Printf("ResetPassword, Request record is not valid for %s\n", token)
		return false
	}
	uPath := a.corpus + "_users"
	uColRef := a.client.Collection(uPath)
	uDocRef := uColRef.Doc(rRecord.UserName)
	uDoc, err := uDocRef.Get(ctx)
	if err != nil {
		log.Printf("ResetPassword error getting user record for %s:, %v", rRecord.UserName, err)
		return false
	}
	var user UserInfo
	if err := uDoc.DataTo(&user); err != nil {
		log.Printf("ResetPassword, error reading user record for %s: %v\n", rRecord.UserName, err)
		return false
	}

	h := sha256.New()
	h.Write([]byte(password))
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	_, err = uDocRef.Update(ctx, []firestore.Update{{Path: "Password", Value: hstr}})
	if err != nil {
		log.Printf("ResetPassword, error setting password for %s: %v\n", rRecord.UserName, err)
		return false
	}
	_, err = docRef.Update(ctx, []firestore.Update{{Path: "EmailValid", Value: false}})
	if err != nil {
		log.Printf("ResetPassword, error updating EmailValid for %s: %v\n", rRecord.UserName, err)
		return false
	}
	log.Println("ResetPassword, successful")
	return true
}

func (a authenticatorFS) SaveSession(ctx context.Context, sessionid string, userInfo UserInfo, authenticated int) SessionInfo {
	log.Printf("SaveSession, sessionid: %s\n", sessionid)
	uPath := a.corpus + "_sessions"
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
	uPath := a.corpus + "_users"
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
