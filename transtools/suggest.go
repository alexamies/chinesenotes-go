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

package transtools

import (
	"encoding/csv"
	"io"
	"log"
	"strings"
)

// File name for expected translation elements
const ExpectedDataFile = "data/glossary/expected.csv"

// File name for suggested translation elements
const ReplaceDataFile = "data/glossary/suggestions.csv"

type Processor interface {
	Suggest(source, translation string) Results
}

type processor struct {
	expectedData map[string]string
	replaceData  map[string]string
}

type Results struct {
	Replacement string
	Notes       []Note
}

type Note struct {
	FoundCN string
	ExpectedEN []string
	FoundEN, Replacement string
}

func loadData(f io.Reader) (*map[string]string, error) {
	data := make(map[string]string)
	r := csv.NewReader(f)
	r.Comma = ','
	r.Comment = '#'
	rows, err := r.ReadAll()
	if err != nil {
		log.Printf("Error reading suggestion data, %v", err)
	}
	for i, row := range rows {
		if len(row) < 2 {
			log.Printf("Error reading suggestion row %d, got %d fields, %v", i, len(row), row)
			continue
		}
		data[row[0]] = row[1]
	}
	log.Printf("Loaded suggestion data from with %d rows", len(data))
	return &data, nil
}

func (p *processor) loadExpectedData(r io.Reader) error {
	d, err := loadData(r)
	if err != nil {
		return err
	}
	p.expectedData = *d
	// Lower case the strings
	for k, v := range p.expectedData {
		p.expectedData[k] = strings.ToLower(v)
	}
	return nil
}

func (p *processor) loadReplaceData(r io.Reader) error {
	d, err := loadData(r)
	if err != nil {
		return err
	}
	p.replaceData = *d
	return nil
}

func NewProcessor(eReader, rReader io.Reader) Processor {
	p := processor{}
	p.loadExpectedData(eReader)
	p.loadReplaceData(rReader)
	return p
}

func (p processor) Suggest(source, translation string) Results {
	log.Printf("Suggest replacements with %d rows", len(p.replaceData))
	replacement := translation
	notes := []Note{}
	for k, v := range p.replaceData {
		if strings.Contains(replacement, k) {
			log.Printf("Suggest replacing %s with %s", k, v)
			replacement = strings.Replace(replacement, k, v, -1)
			// note := fmt.Sprintf("Replaced %s with %s", k, v)
			note := Note{
				FoundEN: k,
				Replacement: v,
			}
			notes = append(notes, note)
		}
	}
	rLC := strings.ToLower(replacement)
	for k, v := range p.expectedData {
		// Expected values is a comma separated list
		exTokens := strings.Split(v, ",")
		var gotOne bool
		for _, e := range exTokens {
			if strings.Contains(source, k) && strings.Contains(rLC, e) {
				gotOne = true
			}
		}
		if strings.Contains(source, k) && !gotOne && len(exTokens) == 1 {
			// note := fmt.Sprintf("Expect translation of phrase with %s to include '%s'", k, v)
			note := Note{
				FoundCN: k,
				ExpectedEN: exTokens,
			}
			notes = append(notes, note)
		} else if strings.Contains(source, k) && !gotOne {
			// note := fmt.Sprintf("Expect translation of phrase with %s to include one of '%s'", k, v)
			note := Note{
				FoundCN: k,
				ExpectedEN: exTokens,
			}
			notes = append(notes, note)
		} 
	}
	return Results{
		Replacement: replacement,
		Notes:       notes,
	}
}
