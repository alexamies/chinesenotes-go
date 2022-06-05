package dictionary

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// NotesExtractor is an interface for extracting multilingual equivalents using
// regular expressions in the notes.
type NotesExtractor struct {
	patterns []*regexp.Regexp
}

// NewNotesExtractor creates a new NotesExtractor.
func NewNotesExtractor(patternList string) (*NotesExtractor, error) {
	log.Printf("dictionary.NewNotesExtractor: patternStr: %s", patternList)
	patterns := []*regexp.Regexp{}
	patternStr := strings.Split(patternList, `","`)
	for _, p := range patternStr {
		p = strings.Trim(p, "\"")
		pattern, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("could not compile notes extractor regex %s: %v", patternStr, err)
		}
		patterns = append(patterns, pattern)
	}
	return &NotesExtractor{
		patterns: patterns,
	}, nil
}

// Extract extracts multilingual equivalents from the given note
func (n NotesExtractor) Extract(notes string) []string {
	extracted := []string{}
	for _, pattern := range n.patterns {
		g := pattern.FindStringSubmatch(notes)
		// log.Printf("NotesExtractor.Extract: notes: %s, %d groups: %v", notes, len(g), g)
		if len(g) > 0 {
			for _, t := range g[1:] {
				extracted = append(extracted, strings.Trim(t, " "))
			}
		}
	}
	return extracted
}

// NotesProcessor processes notes with a regular expression
type NotesProcessor struct {
	patterns []*regexp.Regexp
	replaces []string
}

// newNotesProcessor creates a new notesProcessor
// Param
//   patternList a list of patterns to match regular expressions, quoted and delimited by commas
//   replaceList a list of replacement regular expressions, same cardinality
func NewNotesProcessor(patternList, replaceList string) NotesProcessor {
	log.Printf("dictionary.NewNotesProcessor: patternList: %s replaceList: %s",
		patternList, replaceList)
	p := strings.Split(patternList, `","`)
	patterns := []*regexp.Regexp{}
	for _, t := range p {
		pattern := strings.Trim(t, ` "`)
		log.Printf("notes.newNotesProcessor: compiling %s ", pattern)
		re := regexp.MustCompile(pattern)
		patterns = append(patterns, re)
	}
	r := strings.Split(replaceList, ",")
	replaces := []string{}
	for _, t := range r {
		replace := strings.Trim(t, ` "`)
		log.Printf("notes.newNotesProcessor: adding replacement %s ", replace)
		replaces = append(replaces, replace)
	}
	log.Printf("dictionary.newNotesProcessor: got %d patterns ", len(patterns))
	return NotesProcessor{
		patterns: patterns,
		replaces: replaces,
	}
}

// processes the notes
func (p NotesProcessor) process(notes string) string {
	s := notes
	for i, re := range p.patterns {
		if re.MatchString(notes) {
			s = re.ReplaceAllString(s, p.replaces[i])
		}
	}
	return s
}

// Process checks all senses in the word and replaces note using the regex
func (p NotesProcessor) Process(w dicttypes.Word) dicttypes.Word {
	senses := []dicttypes.WordSense{}
	for _, ws := range w.Senses {
		n := p.process(ws.Notes)
		if n == ws.Notes {
			senses = append(senses, ws)
			continue
		}
		s := dicttypes.WordSense{
			Id:          ws.Id,
			HeadwordId:  ws.HeadwordId,
			Simplified:  ws.Simplified,
			Traditional: ws.Traditional,
			Pinyin:      ws.Pinyin,
			English:     ws.English,
			Grammar:     ws.Grammar,
			Concept:     ws.Concept,
			ConceptCN:   ws.ConceptCN,
			Domain:      ws.Domain,
			DomainCN:    ws.DomainCN,
			Subdomain:   ws.Subdomain,
			SubdomainCN: ws.SubdomainCN,
			Image:       ws.Image,
			MP3:         ws.MP3,
			Notes:       n,
		}
		senses = append(senses, s)
	}
	return dicttypes.Word{
		Simplified:  w.Simplified,
		Traditional: w.Traditional,
		Pinyin:      w.Pinyin,
		HeadwordId:  w.HeadwordId,
		Senses:      senses,
	}
}
