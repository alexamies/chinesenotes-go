package dictionary

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Performs validation of dictionary entries.
// Use NewValidator to create a Validator.
type Validator interface {
	Validate(pos, domain string) error
}

// Performs validation of dictionary entries.
// Use NewValidator to create a Validator.
type validator struct {
	validPos     map[string]bool
	validDomains map[string]bool
}

// Crates a Validator with the given readers
// Params:
//   posReader Reader to load the valid parts of speech from
//   domainReader Reader to load the valid subject domains
// Returns:
//   An initialized Validator
func NewValidator(posReader io.Reader, domainReader io.Reader) (Validator, error) {
	validPos := make(map[string]bool)
	posFScanner := bufio.NewScanner(posReader)
	for posFScanner.Scan() {
		pos := posFScanner.Text()
		validPos[pos] = true
	}
	if err := posFScanner.Err(); err != nil {
		return nil, fmt.Errorf("could not read list of valid parts of speech: %v", err)
	}
	validDomains := make(map[string]bool)
	dFScanner := bufio.NewScanner(domainReader)
	for dFScanner.Scan() {
		line := dFScanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			return nil, fmt.Errorf("could not parse list of valid domains: %s", line)
		}
		validDomains[parts[1]] = true
	}
	if err := dFScanner.Err(); err != nil {
		return nil, fmt.Errorf("could not read list of valid domains: %v", err)
	}
	return validator{
		validPos:     validPos,
		validDomains: validDomains,
	}, nil
}

func (val validator) Validate(pos, domain string) error {
	if pos != "\\N" {
		if _, ok := val.validPos[pos]; !ok {
			return fmt.Errorf("%s is not a recognized part of speech", pos)
		}
	}
	if _, ok := val.validDomains[domain]; !ok {
		return fmt.Errorf("'%s' is not a recognized domain", domain)
	}
	return nil
}

// ValidateDict check the Chinese-English for errors
func ValidateDict(wdict map[string]*dicttypes.Word, validator Validator) error {
	for _, word := range wdict {
		for _, ws := range word.Senses {
			if err := validator.Validate(ws.Grammar, ws.Domain); err != nil {
				log.Printf("ValidateDict, Line: %v", ws)
				return fmt.Errorf("invalid entry %s: %v", ws.Simplified, err)
			}
		}
	}
	return nil
}
