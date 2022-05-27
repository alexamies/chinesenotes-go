package config

import (
	"strings"
	"testing"
)

// TestGetPort tests default serving port
func TestGetPort(t *testing.T) {
	port := GetPort()
	if port != 8080 {
		t.Error("TestGetPort: port = ", port)
	}
}

// TestGetVarWithDefault tests the GetVarWithDefault function
func TestGetVarWithDefault(t *testing.T) {
	const expect = "My Title"
	c := WebAppConfig{}
	val := c.GetVarWithDefault("TITLE", expect)
	if expect != val {
		t.Errorf("TestGetVarWithDefault: expect %s vs got %s", expect, val)
	}
}

// TestNotesExtractorPattern tests NotesExtractorPattern
func TestNotesExtractorPattern(t *testing.T) {
	const want = ""
	c := WebAppConfig{}
	got := c.NotesExtractorPattern()
	if got != want {
		t.Errorf("TestNotesExtractorPattern: got %s vs want %s", got, want)
	}
}

// TestWebconfigInit test config initialization
func TestWebconfigInit(t *testing.T) {
	r := strings.NewReader("")
	c := InitWeb(r)
	if c.ConfigVars == nil {
		t.Error("TestWebconfigInit: c.ConfigVars == nil")
	}
}
