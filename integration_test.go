// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/identity"
)

// Test trivial query with empty dictionary
func TestLoadDict(t *testing.T) {
	cnReaderHome := os.Getenv("CNREADER_HOME")
	if len(cnReaderHome) == 0 {
		t.Skip("TestSendPasswordReset: skipping, CNREADER_HOME not defined")
	}
	fnames := []string{"data/testdict.tsv"}
	appConfig := config.AppConfig{
		LUFileNames: fnames,
	}
	dict, err := dictionary.LoadDictFile(appConfig)
	wdict := dict.Wdict
	if err != nil {
		t.Fatalf("TestLoadDict: not able to load dictionary, skipping tests: %v\n", err)
	}
	if len(wdict) == 0 {
		t.Fatalf("TestLoadDict: expected len(wdict) > 0, got %d", len(wdict))
	}
	t.Logf("TestLoadDict: len(wdict): %d", len(wdict))
	trad := "煸"
	w1, ok := wdict[trad]
	if !ok {
		t.Fatalf("TestLoadDict: !ok: %s", trad)
	}
	if w1.HeadwordId == 0 {
		t.Error("TestLoadDict: w.HeadwordId == 0")
	}
	expectPinyin := "biān"
	if expectPinyin != w1.Pinyin {
		t.Errorf("TestLoadDict: expected pinyin: %s, got: %s", expectPinyin,
			w1.Pinyin)
	}
	w2 := wdict["與"]
	if w2.HeadwordId == 0 {
		t.Error("TestLoadDict: w.HeadwordId == 0")
	}
	if w2.Pinyin == "" {
		t.Error("TestLoadDict: w2.Pinyin == ''")
	}
	w3 := wdict["來"]
	if len(w3.Senses) < 2 {
		t.Error("len(w3.Senses) < 2, ", len(w3.Senses))
	}
	w4 := wdict["发"]
	if len(w4.Senses) < 2 {
		t.Error("len(w4.Senses) < 2, ", len(w4.Senses))
	}
}

// TestLoadDictFile tests loading of a dictionary file
func TestLoadDictFile(t *testing.T) {
	cnReaderHome := os.Getenv("CNREADER_HOME")
	if len(cnReaderHome) == 0 {
		t.Skip("TestSendPasswordReset: skipping, CNREADER_HOME not defined")
	}
	fnames := []string{"data/testdict.tsv"}
	appConfig := config.AppConfig{
		LUFileNames: fnames,
	}
	d, err := dictionary.LoadDictFile(appConfig)
	if err != nil {
		t.Fatalf("TestLoadDictFile: Got an error: %v", err)
	}
	dict := d.Wdict
	if len(dict) < 4 {
		t.Fatalf("TestLoadDictFile: excpected at least 4, got %d", len(dict))
	}
	chinese := "邃古"
	word, ok := dict[chinese]
	if !ok {
		t.Fatalf("TestLoadDictFile: could not find word %s", chinese)
	}
	senses := word.Senses
	if len(senses) == 0 {
		t.Fatalf("TestLoadDictFile: expected > 0 senses, got %d", len(senses))
	}
	expectedDom := "Modern Chinese"
	domain := senses[0].Domain
	if expectedDom != domain {
		t.Errorf("TestLoadDictFile: expected domain %s, got %s", expectedDom,
			domain)
	}
}

// TestSendPasswordReset tests sending a password reset
func TestSendPasswordReset(t *testing.T) {
	cnReaderHome := os.Getenv("CNREADER_HOME")
	if len(cnReaderHome) == 0 {
		t.Skip("TestSendPasswordReset: skipping, CNREADER_HOME not defined")
	}
	userInfo := identity.UserInfo{
		UserID:   100,
		UserName: "test",
		Email:    "alex@chinesenotes.com",
		FullName: "Alex Test",
		Role:     "tester",
	}
	r := strings.NewReader("")
	c := config.InitWeb(r)
	err := identity.SendPasswordReset(userInfo, "", c)
	if err != nil {
		t.Fatalf("TestSendPasswordReset: Error, %v", err)
	}
}
