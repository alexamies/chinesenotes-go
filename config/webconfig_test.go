package config

import (
	"fmt"
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
	testCases := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "Empty",
			pattern: "",
		},
		{
			name:    "One value",
			pattern: `"Scientific name: (.*?)[\(,\,,\;] Species: (.*?)[\(,\,,\;]"`,
		},
	}
	for _, tc := range testCases {
		r := strings.NewReader(fmt.Sprintf("NotesExtractorPattern: %s", tc.pattern))
		c := InitWeb(r)
		got := c.NotesExtractorPattern()
		if got != tc.pattern {
			t.Errorf("TestNotesExtractorPattern %s: got %s vs want %s", tc.name, got, tc.pattern)
		}
	}
}

func TestAddDirectoryToCol(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "Empty",
			input: "",
			want:  false,
		},
		{
			name:  "Set to True",
			input: `AddDirectoryToCol: True`,
			want:  true,
		},
	}
	for _, tc := range testCases {
		r := strings.NewReader(tc.input)
		c := InitWeb(r)
		got := c.AddDirectoryToCol()
		if got != tc.want {
			t.Errorf("TestAddDirectoryToCol %s: got %t vs want %t", tc.name, got, tc.want)
		}
	}
}

// TestGetFromEmail tests config of from email
func TestGetFromEmail(t *testing.T) {
	const email = "do_not_reply@example.org"
	r := strings.NewReader(fmt.Sprintf("FromEmail: %s", email))
	c := InitWeb(r)
	got := c.GetFromEmail()
	if got != email {
		t.Errorf("TestGetFromEmail: %s != %s", got, email)
	}
}

// TestGetAll tests config GetAll
func TestGetAll(t *testing.T) {
	testCases := []struct {
		name      string
		config    string
		wantLen   int
		key       string
		wantValue string
	}{
		{
			name:      "Empty",
			config:    "",
			wantLen:   0,
			key:       "",
			wantValue: "",
		},
		{
			name:      "One variable",
			config:    "Title: Translation Portal",
			wantLen:   1,
			key:       "Title",
			wantValue: "Translation Portal",
		},
	}
	for _, tc := range testCases {
		r := strings.NewReader(tc.config)
		c := InitWeb(r)
		got := c.GetAll()
		if len(got) != tc.wantLen {
			t.Errorf("TestGetAll.%s: len(got) %d != %d", tc.name, len(got), tc.wantLen)
		}
		if len(got) > 0 {
			gotValue := got[tc.key]
			if gotValue != tc.wantValue {
				t.Errorf("TestGetAll.%s: gotValue %s != %s", tc.name, gotValue, tc.wantValue)
			}
		}
	}
}

// TestGetPasswordResetURL tests value of URL to reset password
func TestGetPasswordResetURL(t *testing.T) {
	const resetURL = "https://fgs-translation.org/loggedin/reset_password"
	r := strings.NewReader(fmt.Sprintf("PasswordResetURL: %s", resetURL))
	c := InitWeb(r)
	got := c.GetPasswordResetURL()
	if got != resetURL {
		t.Errorf("TestGetFromEmail: %s != %s", got, resetURL)
	}
}

// TestWebconfigInit tests config initialization
func TestWebonfigInit(t *testing.T) {
	r := strings.NewReader("")
	c := InitWeb(r)
	if c.ConfigVars == nil {
		t.Error("TestWebconfigInit: c.ConfigVars == nil")
	}
}
