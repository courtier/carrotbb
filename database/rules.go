package database

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"unicode"
)

// Length must be between 1 and 12 chars, only letters, numbers and underscores
func IsUsernameValid(name string) error {
	if len(name) < 1 || len(name) > 12 {
		return errors.New("username length must be between 1 and 12 characters")
	}
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') {
			return errors.New("username can only contain letters, numbers and underscore")
		}
	}
	return nil
}

// No funny unicode stuff allowed and limited length
func IsSignatureValid(signature string) error {
	if len(signature) < 1 || len(signature) > 144 {
		return errors.New("signature length must be between 1 and 144 characters")
	}
	for _, r := range signature {
		if unicode.IsControl(r) {
			return errors.New("disallowed content in signature")
		}
	}
	return nil
}

// No funny unicode stuff allowed
func IsContentValid(content string) error {
	if len(content) < 1 || len(content) > 65535 {
		return errors.New("content length must be between 1 and 65535 characters")
	}
	for _, r := range content {
		if unicode.IsControl(r) {
			return errors.New("disallowed content in content")
		}
	}
	return nil
}

// Salt and hash the password using the username using sha-256
func HashPassword(username, password string) string {
	h := sha256.New()
	h.Write([]byte(username))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
