package main

import (
	"errors"
	"unicode"
)

var (
	ErrNameTooShort     = errors.New("username must have at least 1 character")
	ErrNameTooLong      = errors.New("username must have at most 24 characters")
	ErrNameBadCharacter = errors.New("username can only contain letters, numbers and underscore")
)

// Length must be between 1 and 24 chars, only letters, numbers and underscores
func isUsernameValid(name string) error {
	if len(name) < 1 {
		return ErrNameTooShort
	}
	if len(name) > 24 {
		return ErrNameTooLong
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
		return errors.New("title length must be between 1 and 64 characters")
	}
	for _, r := range content {
		if unicode.IsControl(r) {
			return errors.New("disallowed content in title")
		}
	}
	return nil
}

// No funny unicode stuff allowed
func isContentValid(content string) error {
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

func isPasswordValid(password string) error {
	if len(password) < 1 || len(password) > 144 {
		return errors.New("password length must be between 1 and 144 characters")
	}
	for _, r := range password {
		if unicode.IsControl(r) {
			return errors.New("disallowed content in password")
		}
	}
	return nil
}
