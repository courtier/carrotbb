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
		t.Error("Session should be expired")
	}
}

func TestSaltAndHash(t *testing.T) {
	hash := saltAndHash("hello", "world")
	expected := "$argon2id$v=19$m=16384,t=4,p=4$Kmu5BL5wS9ervTy25ilRQCwjj1T2rkwf00ekySVkvQs$d29ybGQ"
	if hash != expected {
		t.Error("Expected:", expected, "got:", hash)
	}
}

func TestExtractSession(t *testing.T) {
	sessionToken := "token"
	sessionCache.Write(sessionToken, session{expiry: time.Now().Add(10 * time.Second)})
	r, err := http.NewRequest("GET", "/signup", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})
	if token, err := extractSessionToken(r); err != nil || token != sessionToken {
		if err != nil {
			t.Fatal(err)
		} else {
			t.Error("Expected:", sessionToken, "got:", token)
		}
	}
}
