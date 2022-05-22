package dictionary

import (
	"strings"
	"testing"
)

func simpleValidator() (Validator, error) {
	const posList = "noun\nverb\n"
	posReader := strings.NewReader(posList)
	const domainList = "艺术	Art	\\N	\\N\n佛教	Buddhism	\\N	\\N\n"
	domainReader := strings.NewReader(domainList)
	return NewValidator(posReader, domainReader)
}

// Test for dictionar validation
func TestValidate(t *testing.T) {
	validator, err := simpleValidator()
	if err != nil {
		t.Fatalf("TestNewValidator: Unexpected error: %v", err)
	}
	type test struct {
		name string
		pos string
		domain string
		valid bool
  }
  tests := []test{
		{
			name: "Valid term",
			pos: "noun",
			domain: "Art",
			valid: true,
		},
		{
			name: "Invalid domain",
			pos: "noun",
			domain: "Artistic",
			valid: false,
		},
  }
  for _, tc := range tests {
		err = validator.Validate(tc.pos, tc.domain)		
		if tc.valid && err != nil {
			t.Fatalf("TestNewValidator: unexpected error for %s, %v", tc.name, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("TestNewValidator: expected error for test %s", tc.name)
		}
	}
}
