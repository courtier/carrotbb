package main

import "testing"

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
