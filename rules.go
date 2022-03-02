package main

import (
	"errors"
	"unicode"
)

var (
	ErrNameBadLength     = errors.New("username must be between 1 and 24 characters")
	ErrNameBadCharacter  = errors.New("username can only contain letters, numbers and underscore")
	ErrTitleBadLength    = errors.New("title must be between 1 and 64 characters")
	ErrContentBadLength  = errors.New("content must be between 1 and 65535 characters")
	ErrPasswordBadLength = errors.New("password must be between 1 and 144 characters")
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
	return nil
}

func isContentValid(content string) error {
	if len(content) < 1 || len(content) > 65535 {
		return ErrContentBadLength
	}
	return nil
}

func isPasswordValid(password string) error {
	if len(password) < 1 || len(password) > 144 {
		return ErrPasswordBadLength
	}
	return nil
}
