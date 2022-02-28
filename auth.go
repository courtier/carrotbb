package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/courtier/carrotbb/database"
	"github.com/rs/xid"
)

const (
	DEFAULT_SESSION_EXPIRY = 24 * 7 * time.Hour
)

type session struct {
	userID xid.ID
	expiry time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

// Will be used for csrf and session tokens
func newRandomToken() (string, error) {
	b := make([]byte, 8)
	// We use math/rand here as I think sha256 hashing is
	// sufficient entropy
	n, err := rand.Read(b)
	if n != len(b) {
		return "", errors.New("rand read less bytes than required")
	}
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sha256.New().Sum(b)), nil
}

// Salt and hash the password using the username using sha-256
func hashPassword(username, password string) string {
	h := sha256.New()
	h.Write([]byte(username))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func extractSession(r *http.Request) (token string, err error) {
	c, err := r.Cookie("session_token")
	if err == http.ErrNoCookie {
		return
	}
	token = c.Value
	_, ok := sessionCache[token]
	if !ok {
		err = errors.New("session not in cache")
	}
	return
}

func isRequestAuthenticatedSimple(r *http.Request) bool {
	token, err := extractSession(r)
	if err != nil {
		return false
	}
	return !sessionCache[token].isExpired()
}

func extractUser(r *http.Request) (user *database.User, err error) {
	c, err := r.Cookie("session_token")
	if err == http.ErrNoCookie {
		return
	}
	token := c.Value
	sesh, ok := sessionCache[token]
	if !ok {
		err = errors.New("session not in cache")
		return
	}
	user, err = db.GetUser(sesh.userID)
	return
}

type AuthMiddleware struct {
	handler http.Handler
}

func NewAuthMiddleware(handler http.Handler) *AuthMiddleware {
	return &AuthMiddleware{handler}
}

// what happens if we just dont refresh and hand out weekly tokens
func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := extractSession(r)
	if err != nil {
		a.handler.ServeHTTP(w, r)
		return
	}
	if sessionCache[token].isExpired() {
		delete(sessionCache, token)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now(),
		})
		a.handler.ServeHTTP(w, r)
		return
	}
	if r.URL.EscapedPath() == "/logout" {
		a.handler.ServeHTTP(w, r)
		return
	}
	newToken, err := newRandomToken()
	if err != nil {
		a.handler.ServeHTTP(w, r)
		return
	}
	sessionCache[newToken] = session{
		userID: sessionCache[token].userID,
		expiry: time.Now().Add(DEFAULT_SESSION_EXPIRY),
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newToken,
		Expires: time.Now().Add(DEFAULT_SESSION_EXPIRY),
	})
	a.handler.ServeHTTP(w, r)
	// delete token after this serving this request as to not mess up
	// the authentication for the duration of this request
	delete(sessionCache, token)
}

func authenticateUser(w http.ResponseWriter, token string, userID xid.ID) {
	sessionCache[token] = session{
		userID: userID,
		expiry: time.Now().Add(DEFAULT_SESSION_EXPIRY),
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   token,
		Expires: time.Now().Add(DEFAULT_SESSION_EXPIRY),
	})
}
