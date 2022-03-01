package main

import (
	"errors"
	"unicode"
)

var (
	ErrNameBadLength      = errors.New("username must be between 1 and 24 characters")
	ErrNameBadCharacter   = errors.New("username can only contain letters, numbers and underscore")
	ErrTitleBadLength     = errors.New("title must be between 1 and 64 characters")
	ErrDisallowedTitle    = errors.New("disallowed content in title")
	ErrContentBadLength   = errors.New("content must be between 1 and 65535 characters")
	ErrDisallowedContent  = errors.New("disallowed content in content")
	ErrPasswordBadLength  = errors.New("password must be between 1 and 144 characters")
	ErrDisallowedPassword = errors.New("disallowed content in password")
)

// Length must be between 1 and 24 chars, only letters, numbers and underscores
func isUsernameValid(name string) error {
	if len(name) < 1 || len(name) > 24 {
		return ErrNameBadLength
	}
	for _, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') {
			return ErrNameBadCharacter
		}
	}
	return nil
}

func isTitleValid(content string) error {
	if len(content) < 1 || len(content) > 64 {
		return ErrTitleBadLength
	}
	for _, r := range content {
		if unicode.IsControl(r) {
			return ErrDisallowedTitle
		}
	}
	return nil
}

// No funny unicode stuff allowed
func isContentValid(content string) error {
	if len(content) < 1 || len(content) > 65535 {
		return ErrContentBadLength
	}
	for _, r := range content {
		if unicode.IsControl(r) {
			return ErrDisallowedContent
		}
	}
	return nil
}

func isPasswordValid(password string) error {
	if len(password) < 1 || len(password) > 144 {
		return ErrPasswordBadLength
	}
	for _, r := range password {
		if unicode.IsControl(r) {
			return ErrDisallowedPassword
		}
	}
	return nil
}
