package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/rs/xid"
)

func TestIsExpired(t *testing.T) {
	s := session{userID: xid.NilID(), expiry: time.Now()}
	time.Sleep(10 * time.Millisecond)
	if !s.isExpired() {
		t.FailNow()
	}
}

func TestSaltAndHash(t *testing.T) {
	hash := saltAndHash("hello", "world")
	if hash != "$argon2id$v=19$m=16384,t=4,p=4$Kmu5BL5wS9ervTy25ilRQCwjj1T2rkwf00ekySVkvQs$d29ybGQ" {
		t.FailNow()
	}
}

func TestExtractSession(t *testing.T) {
	sessionToken := "token"
	sessionCache[sessionToken] = session{expiry: time.Now().Add(10 * time.Second)}
	r, err := http.NewRequest("GET", "/signup", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})
	if token, err := extractSession(r); err != nil || token != sessionToken {
		t.FailNow()
	}
}
