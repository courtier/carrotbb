package main

import (
	"testing"
)

func TestPathIntoArray(t *testing.T) {
	payloads := map[string][]string{
		"":              {},
		"/hello":        {"hello"},
		"/hello/":       {"hello"},
		"hello/":        {"hello"},
		"/hello/world":  {"hello", "world"},
		"/hello/world/": {"hello", "world"},
		"hello/world/":  {"hello", "world"},
	}
	for k, v := range payloads {
		got := pathIntoArray(k)
		if len(got) != len(v) {
			t.Error("Content:", k, "expected:", v, "got:", got)
		}
		for i := 0; i < len(v); i++ {
			if got[i] != v[i] {
				t.Error("Content:", k, "expected:", v, "got:", got)
			}
		}
	}
}
