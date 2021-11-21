package transtools

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

const expectedDataFile = "data/expected.csv"
const replaceDataFile = "data/suggestions.csv"

type Processor interface {
	Suggest(source, translation string) (*Results, error)
}

type processor struct {
	expectedData map[string]string
	replaceData  map[string]string
}

type Results struct {
	Replacement string
	Notes       []string
}

func loadData(dataFile string) (*map[string]string, error) {
	data := make(map[string]string)
	f, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.Comment = '#'
	rows, err := r.ReadAll()
	if err != nil {
		log.Printf("Error reading suggestion data %s, %v", dataFile, err)
	}
	for i, row := range rows {
		if len(row) < 2 {
			log.Printf("Error reading suggestion row %d, got %d fields, %v", i, len(row), row)
			continue
		}
		data[row[0]] = row[1]
	}
	log.Printf("Loaded suggestion data from %s with %d rows", dataFile, len(data))
	return &data, nil
}

func (p *processor) loadExpectedData() error {
	d, err := loadData(expectedDataFile)
	if err != nil {
		return err
	}
	p.expectedData = *d
	return nil
}

func (p *processor) loadReplaceData() error {
	d, err := loadData(replaceDataFile)
	if err != nil {
		return err
	}
	p.replaceData = *d
	return nil
}

func NewProcessor() Processor {
	p := processor{}
	p.loadExpectedData()
	p.loadReplaceData()
	return p
}

func (p processor) Suggest(source, translation string) (*Results, error) {
	log.Printf("Suggest replacements with %d rows", len(p.replaceData))
	replacement := translation
	notes := []string{}
	for k, v := range p.replaceData {
		if strings.Contains(replacement, k) {
			log.Printf("Suggest replacing %s with %s", k, v)
			replacement = strings.Replace(replacement, k, v, -1)
			note := fmt.Sprintf("Replaced %s with %s", k, v)
			notes = append(notes, note)
		}
	}
	rLC := strings.ToLower(replacement)
	for k, v := range p.expectedData {
		log.Printf("Check translation of string with %s to include %s", k, v)
		if strings.Contains(source, k) && !strings.Contains(rLC, v) {
			note := fmt.Sprintf("Expect translation of phrase with %s to include '%s'", k, v)
			notes = append(notes, note)
		}
	}
	r := Results{
		Replacement: replacement,
		Notes:       notes,
	}
	return &r, nil
}
