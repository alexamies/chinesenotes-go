package config

import (
	"os"
	"reflect"
	"testing"
)

// TestCorpusDataDir is a trivial query with empty chunk
func TestCorpusDataDir(t *testing.T) {
	os.Unsetenv("CNREADER_HOME")
	t.Logf("TestCorpusDataDir: Begin unit tests\n")
	appConfig := InitConfig()
	result := appConfig.CorpusDataDir()
	expect := "./data/corpus"
	if expect != result {
		t.Errorf("expected: %s, got: %s", expect, result)
	}
}

// Test AvoidSubDomains
func TestAvoidSubDomains(t *testing.T) {
	appConfig := InitConfig()
	result := appConfig.AvoidSubDomains()
	expect := make(map[string]bool)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected: %v, got: %v", expect, result)
	}
}
