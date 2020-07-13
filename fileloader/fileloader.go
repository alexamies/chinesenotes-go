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
	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"os"
	"strconv"
)

// Loads all words from a static file included in the Docker image
func LoadDictFile(fNames []string) (map[string]dicttypes.Word, error) {
	applog.Info("LoadDictFile, enter")
	wdict := map[string]dicttypes.Word{}
	avoidSub := config.AvoidSubDomains()
	for _, fName := range fNames {
		applog.Info("dictionary.loadDictFile: fName: ", fName)
		wsfile, err := os.Open(fName)
		if err != nil {
			applog.Error("dictionary.loadDictFile, error: ", err)
			return wdict, err
		}
		defer wsfile.Close()
		reader := csv.NewReader(wsfile)
		reader.FieldsPerRecord = -1
		reader.Comma = rune('\t')
		reader.Comment = '#'
		rawCSVdata, err := reader.ReadAll()
		if err != nil {
			applog.Error("Could not parse lexical units file", err)
			return wdict, err
		}
		for i, row := range rawCSVdata {
			id, err := strconv.ParseInt(row[0], 10, 0)
			if err != nil {
				applog.Error("Could not parse word id for word ", i, err)
				return wdict, err
			}
			simp := row[1]
			trad := row[2]
			pinyin := row[3]
			english := row[4]
			grammar := row[5]
			parent_en :=  row[11]
			// If subdomain, aka parent, should be avoided, then skip
			if _, ok := avoidSub[parent_en]; ok {
				continue
			}
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
	}
	applog.Info("LoadDictFile, loaded")
	return wdict, nil
}
