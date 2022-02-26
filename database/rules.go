package database

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode"
)

func IsUsernameValid(name string) bool {
	if len(name) < 1 || len(name) > 12 {
		return false
	}
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') {
			return false
		}
	}
	return true
}

func HashPassword(username, password string) string {
	h := sha256.New()
	h.Write([]byte(username))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
