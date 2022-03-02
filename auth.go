package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/courtier/carrotbb/database"
	"github.com/rs/xid"
	"golang.org/x/crypto/argon2"
)

const (
	DEFAULT_SESSION_EXPIRY = 24 * 7 * time.Hour
)

var (
	ErrSessionNotCached  = errors.New("session not in cache")
	ErrRandReadUnmatched = errors.New("rand read less bytes than required")
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
	if err != nil {
		return "", err
	}
	if n != len(b) {
		return "", ErrRandReadUnmatched
	}
	return hex.EncodeToString(sha256.New().Sum(b)), nil
}

const (
	argonTime    = 4
	argonMemory  = 16 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

// Salt and hash initial using argon2
func saltAndHash(initial, salt string) string {
	key := argon2.IDKey([]byte(initial), []byte(salt), argonTime, argonMemory, argonThreads, argonKeyLen)
	basedKey := base64.RawStdEncoding.EncodeToString(key)
	basedSalt := base64.RawStdEncoding.EncodeToString([]byte(salt))
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads, basedKey, basedSalt)
	return encoded
}

func extractSession(r *http.Request) (token string, err error) {
	c, err := r.Cookie("session_token")
	if err == http.ErrNoCookie {
		return
	}
	token = c.Value
	_, ok := sessionCache[token]
	if !ok {
		err = ErrSessionNotCached
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

func extractUser(r *http.Request) (user database.User, err error) {
	c, err := r.Cookie("session_token")
	if err == http.ErrNoCookie {
		return
	}
	token := c.Value
	sesh, ok := sessionCache[token]
	if !ok {
		err = ErrSessionNotCached
		return
	}
	return db.GetUser(sesh.userID)
}

func extractUsername(r *http.Request) (bool, string) {
	var signedIn bool
	var username string
	user, err := extractUser(r)
	if err == nil {
		signedIn = true
		username = user.Name
	}
	return signedIn, username
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
	if err == nil && sessionCache[token].isExpired() {
		unauthenticateUser(w, token)
	}
	a.handler.ServeHTTP(w, r)
}

func authenticateUser(w http.ResponseWriter, token string, userID xid.ID) {
	sessionCache[token] = session{
		userID: userID,
		expiry: time.Now().Add(DEFAULT_SESSION_EXPIRY),
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(DEFAULT_SESSION_EXPIRY),
		HttpOnly: true,
		Path:     "/",
	})
}

func unauthenticateUser(w http.ResponseWriter, token string) {
	delete(sessionCache, token)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Path:     "/",
	})
}
