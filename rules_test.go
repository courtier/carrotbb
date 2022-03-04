package main

import (
	"strings"
	"testing"
)

func TestIsUsernameValid(t *testing.T) {
	payloads := map[string]error{
		"":                          ErrNameBadLength,
		"hello":                     nil,
		"hellohellohellohellohello": ErrNameBadLength,
		"....":                      ErrNameBadCharacter,
	}
	for k, v := range payloads {
		if res := isUsernameValid(k); res != v {
			t.Error("Username:", k, "expected:", v, "got:", res)
		}
	}
}

func TestIsTitleValid(t *testing.T) {
	var longB strings.Builder
	for i := 0; i < 100; i++ {
		longB.WriteByte('s')
	}
	long := longB.String()
	payloads := map[string]error{
		"":      ErrTitleBadLength,
		"hello": nil,
		long:    ErrTitleBadLength,
	}
	for k, v := range payloads {
		if res := isTitleValid(k); res != v {
			t.Error("Title:", k, "expected:", v, "got:", res)
		}
	}
}

func TestIsContentValid(t *testing.T) {
	var longB strings.Builder
	for i := 0; i < 65536; i++ {
		longB.WriteByte('s')
	}
	long := longB.String()
	payloads := map[string]error{
		"":      ErrContentBadLength,
		"hello": nil,
		long:    ErrContentBadLength,
	}
	for k, v := range payloads {
		if res := isContentValid(k); res != v {
			t.Error("Content:", k, "expected:", v, "got:", res)
		}
	}
}

func TestIsPasswordValid(t *testing.T) {
	var longB strings.Builder
	for i := 0; i < 146; i++ {
		longB.WriteByte('s')
	}
	long := longB.String()
	payloads := map[string]error{
		"":      ErrPasswordBadLength,
		"hello": nil,
		long:    ErrPasswordBadLength,
	}
	for k, v := range payloads {
		if res := isPasswordValid(k); res != v {
			t.Error("Password:", k, "expected:", v, "got:", res)
		}
	}
}
