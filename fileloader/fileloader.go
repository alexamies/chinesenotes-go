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


//
// Chinese-English dictionary database functions
//
package fileloader

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Loads all words from static files
func LoadDictFile(appConfig config.AppConfig) (map[string]dicttypes.Word, error) {
	fNames := appConfig.LUFileNames
	applog.Infof("LoadDictFile, loading %d files\n", len(fNames))
	wdict := map[string]dicttypes.Word{}
	avoidSub := appConfig.AvoidSubDomains()
	for _, fName := range fNames {
		applog.Infof("fileloader.LoadDictFile: fName: %s", fName)
		wsfile, err := os.Open(fName)
		if err != nil {
			return wdict, fmt.Errorf("fileloader.LoadDictFile, error opening %s: %v",
					fName, err)
		}
		defer wsfile.Close()
		err = loadDictReader(wsfile, wdict, avoidSub)
		if err != nil {
			return wdict, fmt.Errorf("fileloader.LoadDictFile, error reading from %s: %v",
					fName, err)
		}
	}
	applog.Infof("LoadDictFile, loaded %d entries", len(wdict))
	return wdict, nil
}

// Loads all words from a URL
func LoadDictURL(appConfig config.AppConfig, url string) (map[string]dicttypes.Word, error) {
	applog.Info("LoadDictURL loading from URL")
	resp, err := http.Get(url)
	wdict := map[string]dicttypes.Word{}
	if err != nil {
		return wdict, fmt.Errorf("fileloader.LoadDictURL, error GET from %v: %v",
				url, err)
	}
	defer resp.Body.Close()
	avoidSub := appConfig.AvoidSubDomains()
	err = loadDictReader(resp.Body, wdict, avoidSub)
		if err != nil {
			return wdict, fmt.Errorf("fileloader.LoadDictFile, error reading from %s: %v",
					url, err)
	}
	return wdict, nil
}

// loadDictReader ads words from an io.Reader to the given dictionary
func loadDictReader(r io.Reader, wdict map[string]dicttypes.Word,
		avoidSub map[string]bool) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = '#'
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		applog.Error("Could not parse lexical units file", err)
		return err
	}
	for i, row := range rawCSVdata {
		id, err := strconv.ParseInt(row[0], 10, 0)
		if err != nil {
			return fmt.Errorf("Could not parse word id %d for word %v", i, err)
		}
		simp := row[1]
		trad := row[2]
		pinyin := row[3]
		english := row[4]
		grammar := row[5]
		conceptCN := row[6]
		concept := row[7]
		domainCN := row[8]
		domain :=  row[9]
		subdomain :=  row[11]
		// If subdomain, aka parent, should be avoided, then skip
		if _, ok := avoidSub[subdomain]; ok {
			continue
		}
		image := row[12]
		mp3 := row[13]
		notes := row[14]
		if notes == "\\N" {
			notes = ""
		}
		hwId := 0
		if len(row) == 16 {
			hwIdInt, err := strconv.ParseInt(row[15], 10, 0)
			if err != nil {
				applog.Info("loadDictFile, id: %d, simp: %s, trad: %s, " + 
					"pinyin: %s, english: %s, grammar: %s\n",
					id, simp, trad, pinyin, english, grammar,)
				applog.Error("loadDictFile: Could not parse headword id for word ",
					id, err)
			}
			hwId = int(hwIdInt)
		} else {
			applog.Info("loadDictFile, No. cols: %d\n",len(row))
			applog.Info("loadDictFile, id: %d, simp: %s, trad: %s, pinyin: %s, " +
				"english: %s, grammar: %s\n",
				id, simp, trad, pinyin, english, grammar)
			applog.Error("loadDictFile wrong number of columns ", id, err)
		}
		ws := dicttypes.WordSense{}
		ws.Id = hwId
		ws.Simplified =simp
		ws.HeadwordId = hwId
		ws.Traditional = trad
		ws.Pinyin = pinyin
		ws.English = english
		ws.Grammar = grammar
		ws.ConceptCN = conceptCN
		ws.Concept = concept
		ws.DomainCN = domainCN
		ws.Domain = domain
		// applog.Info("loadDictFile, %s domain: %s\n", simp, domain)
		ws.Image = image
		ws.MP3 = mp3
		ws.Notes = notes
		word, ok := wdict[ws.Simplified]
		if ok {
			word.Senses = append(word.Senses, ws)
			wdict[word.Simplified] = word
		} else {
			word = dicttypes.Word{}
			word.Simplified = ws.Simplified
			word.Traditional = ws.Traditional
			word.Pinyin = ws.Pinyin
			word.HeadwordId = ws.HeadwordId
			word.Senses = []dicttypes.WordSense{ws}
			wdict[word.Simplified] = word
		}
		if trad != "\\N" {
			word1, ok1 := wdict[trad]
			if ok1 {
				word1.Senses = append(word1.Senses, ws)
				wdict[word1.Traditional] = word1
			} else {
				word1 = dicttypes.Word{}
				word1.Simplified = ws.Simplified
				word1.Traditional = ws.Traditional
				word1.Pinyin = ws.Pinyin
				word1.HeadwordId = ws.HeadwordId
				word1.Senses = []dicttypes.WordSense{ws}
				wdict[word1.Traditional] = word1
			}
		}
	}
	return nil
}