package dictionary

// Chinese-English dictionary file and network loading functions

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// LoadDictFile loads all words from static files
func LoadDictFile(appConfig config.AppConfig) (*Dictionary, error) {
	fNames := appConfig.LUFileNames
	log.Printf("LoadDictFile, loading %d files", len(fNames))
	wdict := make(map[string]*dicttypes.Word)
	avoidSub := appConfig.AvoidSubDomains()
	for _, fName := range fNames {
		log.Printf("fileloader.LoadDictFile: fName: %s", fName)
		wsfile, err := os.Open(fName)
		if err != nil {
			return nil, fmt.Errorf("fileloader.LoadDictFile, error opening %s: %v",
				fName, err)
		}
		defer wsfile.Close()
		err = loadDictReader(wsfile, wdict, avoidSub)
		if err != nil {
			return nil, fmt.Errorf("fileloader.LoadDictFile, error reading from %s: %v",
				fName, err)
		}
	}
	log.Printf("LoadDictFile, loaded %d entries", len(wdict))
	return NewDictionary(wdict), nil
}

// LoadDictKeys loads the keys only from static files
func LoadDictKeys(appConfig config.AppConfig) (*map[string]bool, error) {
	fNames := appConfig.LUFileNames
	log.Printf("LoadDictFile, loading %d files", len(fNames))
	wdict := make(map[string]bool)
	avoidSub := appConfig.AvoidSubDomains()
	for _, fName := range fNames {
		log.Printf("fileloader.LoadDictKeys: fName: %s", fName)
		wsfile, err := os.Open(fName)
		if err != nil {
			return nil, fmt.Errorf("fileloader.LoadDictKeys, error opening %s: %v",
				fName, err)
		}
		defer wsfile.Close()
		err = loadDictKeys(wsfile, wdict, avoidSub)
		if err != nil {
			return nil, fmt.Errorf("fileloader.LoadDictKeys, error reading from %s: %v",
				fName, err)
		}
	}
	log.Printf("LoadDictKeys, loaded %d entries", len(wdict))
	return &wdict, nil
}

// Loads all words from a URL
func LoadDictURL(appConfig config.AppConfig, url string) (*Dictionary, error) {
	log.Println("LoadDictURL loading from URL")
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fileloader.LoadDictURL, error GET from %v: %v",
			url, err)
	}
	defer resp.Body.Close()
	wdict := make(map[string]*dicttypes.Word)
	avoidSub := appConfig.AvoidSubDomains()
	err = loadDictReader(resp.Body, wdict, avoidSub)
	if err != nil {
		return nil, fmt.Errorf("fileloader.LoadDictFile, error reading from %s: %v",
			url, err)
	}
	return NewDictionary(wdict), nil
}

// loadDictReader ads words from an io.Reader to the given dictionary
func loadDictReader(r io.Reader, wdict map[string]*dicttypes.Word,
	avoidSub map[string]bool) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = '#'
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not parse lexical units file: %v", err)
	}
	for i, row := range rawCSVdata {
		if len(row) < 15 {
			fmt.Printf("only %d elements (less than 15) for row %d, text: %v", len(row), i, row)
			continue
		}
		id, err := strconv.ParseInt(row[0], 10, 0)
		if err != nil {
			fmt.Printf("Could not parse word id %d for word %v", i, err)
		}
		simp := row[1]
		trad := row[2]
		pinyin := row[3]
		english := row[4]
		grammar := row[5]
		conceptCN := row[6]
		if conceptCN == "\\N" {
			conceptCN = ""
		}
		concept := row[7]
		if concept == "\\N" {
			concept = ""
		}
		domainCN := row[8]
		domain := row[9]
		subdomain := row[11]
		if subdomain == "\\N" {
			subdomain = ""
		}
		// If subdomain, aka parent, should be avoided, then skip
		if _, ok := avoidSub[subdomain]; ok {
			continue
		}
		subdomainCN := row[12]
		if subdomainCN == "\\N" {
			subdomainCN = ""
		}
		image := row[12]
		mp3 := row[13]
		notes := ""
		hwId := 0
		if len(row) == 16 {
			notes = row[14]
			if notes == "\\N" {
				notes = ""
			}
			hwIdInt, err := strconv.ParseInt(row[15], 10, 0)
			if err != nil {
				log.Printf("loadDictFile, id: %d, simp: %s, trad: %s, "+
					"pinyin: %s, english: %s, grammar: %s\n",
					id, simp, trad, pinyin, english, grammar)
				log.Printf("loadDictFile: Could not parse headword id for word %d: %v",
					id, err)
			}
			hwId = int(hwIdInt)
		} else {
			log.Printf("loadDictFile, No. cols: %d", len(row))
			log.Printf("loadDictFile, id: %d, simp: %s, trad: %s, pinyin: %s, "+
				"english: %s, grammar: %s\n",
				id, simp, trad, pinyin, english, grammar)
			log.Printf("loadDictFile wrong number of columns %d: %v", id, err)
		}
		ws := dicttypes.WordSense{
			Id:          int(id),
			Simplified:  simp,
			HeadwordId:  hwId,
			Traditional: trad,
			Pinyin:      pinyin,
			English:     english,
			Grammar:     grammar,
			ConceptCN:   conceptCN,
			Concept:     concept,
			DomainCN:    domainCN,
			Domain:      domain,
			Subdomain:   subdomain,
			SubdomainCN: subdomainCN,
			Image:       image,
			MP3:         mp3,
			Notes:       notes,
		}
		word, ok := wdict[simp]
		if ok {
			word.Senses = append(word.Senses, ws)
		} else {
			wdict[simp] = &dicttypes.Word{
				Simplified:  ws.Simplified,
				Traditional: ws.Traditional,
				Pinyin:      ws.Pinyin,
				HeadwordId:  ws.HeadwordId,
				Senses:      []dicttypes.WordSense{ws},
			}
		}
		if trad != "\\N" {
			if ok {
				wdict[trad] = word
			} else {
				wdict[trad] = &dicttypes.Word{
					Simplified:  ws.Simplified,
					Traditional: ws.Traditional,
					Pinyin:      ws.Pinyin,
					HeadwordId:  ws.HeadwordId,
					Senses:      []dicttypes.WordSense{ws},
				}
			}
		}
	}
	return nil
}

// loadDictKeys ads keys only from an io.Reader to the given dictionary
func loadDictKeys(r io.Reader, wdict map[string]bool, avoidSub map[string]bool) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = '#'
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not parse lexical units file: %v", err)
	}
	for i, row := range rawCSVdata {
		if len(row) < 15 {
			fmt.Printf("only %d elements (less than 15) for row %d, text: %v", len(row), i, row)
			continue
		}
		simp := row[1]
		trad := row[2]
		subdomain := row[11]
		if subdomain == "\\N" {
			subdomain = ""
		}
		// If subdomain, aka parent, should be avoided, then skip
		if _, ok := avoidSub[subdomain]; ok {
			continue
		}
		wdict[simp] = true
		if trad != "\\N" {
			wdict[trad] = true
		}
	}
	return nil
}
