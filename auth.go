package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/courtier/carrotbb/database"
	"github.com/rs/xid"
	"golang.org/x/crypto/argon2"
)

const (
	DEFAULT_SESSION_EXPIRY = 24 * 7 * time.Hour
)

var (
	ErrSessionNotCached    = errors.New("session token not in cache")
	ErrExpiredSessionToken = errors.New("session token has expired")
	ErrRandReadUnmatched   = errors.New("rand read less bytes than required")
)

type MapCache struct {
	cache map[string]session
	lock  sync.RWMutex
}

var (
	sessionCache = &MapCache{
		cache: make(map[string]session),
	}
)

func (m *MapCache) ReadOK(key string) (session, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	s, ok := m.cache[key]
	return s, ok
}

func (m *MapCache) Read(key string) session {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.cache[key]
}

func (m *MapCache) Write(key string, value session) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cache[key] = value
}

func (m *MapCache) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.cache, key)
}

type session struct {
	userID xid.ID
	expiry time.Time
}

// isExpired checks if a session is expired
func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

// newRandomToken should be used for generating session/csrf tokens
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

// saltAndHash returns the Argon2 hash of initial using salt as salt
func saltAndHash(initial, salt string) string {
	key := argon2.IDKey([]byte(initial), []byte(salt), argonTime, argonMemory, argonThreads, argonKeyLen)
	basedKey := base64.RawStdEncoding.EncodeToString(key)
	basedSalt := base64.RawStdEncoding.EncodeToString([]byte(salt))
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads, basedKey, basedSalt)
	return encoded
}

// TODO: these two functions could be done in a better way

// extractSessionToken extracts the session token from the session_token cookie
// if possible, returns an error if the cookie does not exist.
// this function does not do any sanitizing/validating, the token string could be anything
func extractSessionToken(r *http.Request) (string, error) {
	c, err := r.Cookie("session_token")
	if err == http.ErrNoCookie {
		return "", err
	}
	return c.Value, err
}

// extractUser extracts a user from a request using the session_token cookie
// if the token is in the cache. if it is not it returns ErrSessionNotCached
// iif the token is cached, but it is expired it returns ErrExpiredSessionToken
func extractUser(r *http.Request) (token string, user database.User, err error) {
	token, err = extractSessionToken(r)
	if err == http.ErrNoCookie {
		return
	}
	sesh, ok := sessionCache.ReadOK(token)
	if !ok {
		err = ErrSessionNotCached
		return
	}
	if sesh.isExpired() {
		err = ErrExpiredSessionToken
		return
	}
	user, err = db.GetUser(sesh.userID)
	return
}

type ContextString string

// AuthMiddleware checks for expired tokens and deletes them if it catches any
type AuthMiddleware struct {
	handler http.Handler
}

func NewAuthMiddleware(handler http.Handler) *AuthMiddleware {
	return &AuthMiddleware{handler}
}

func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqCtx := r.Context()
	token, user, err := extractUser(r)
	if err != nil {
		if err == ErrExpiredSessionToken {
			unauthenticateUser(w, token)
		}
	} else {
		reqCtx = context.WithValue(reqCtx, ContextString("user"), user)
	}
	a.handler.ServeHTTP(w, r.WithContext(reqCtx))
}

// authenticateUser puts the token and session in the cache and sets the cookie
func authenticateUser(w http.ResponseWriter, token string, userID xid.ID) {
	sessionCache.Write(token, session{
		userID: userID,
		expiry: time.Now().Add(DEFAULT_SESSION_EXPIRY),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(DEFAULT_SESSION_EXPIRY),
		HttpOnly: true,
		Path:     "/",
	})
}

// unauthenticateUser removes the token and session from the cache and removes the cookie
func unauthenticateUser(w http.ResponseWriter, token string) {
	sessionCache.Delete(token)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Path:     "/",
	})
}
