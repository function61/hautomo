package main

import (
	"github.com/function61/eventhorizon/util/ass"
	"testing"
)

func EqualString(t *testing.T, actual string, expected string) {
	if actual != expected {
		t.Fatalf("exp=%v; got=%v", expected, actual)
	}
}

func TestSaslAuthString(t *testing.T) {
	ass.EqualString(t, saslAuthString("guest@x.com", "guest", "guest"), "Z3Vlc3RAeC5jb20AZ3Vlc3QAZ3Vlc3Q=")
}
