package dictionary

import (
	"log"
	"regexp"
	"strings"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Processes notes with a regular expression
type NotesProcessor struct {
	patterns []*regexp.Regexp
	replaces []string
}

// newNotesProcessor creates a new notesProcessor
// Param
//   patternList a list of patterns to match regular expressions, quoted and delimited by commas
//   replaceList a list of replacement regular expressions, same cardinality
func NewNotesProcessor(patternList, replaceList string) NotesProcessor{
	log.Printf("analysis.NewNotesProcessor: patternList: %s replaceList: %s",
			patternList, replaceList)
	p := strings.Split(patternList, `","`)
	patterns := []*regexp.Regexp{}
	for _, t := range p {
		pattern := strings.Trim(t, ` "`)
		log.Printf("notes.newNotesProcessor: compiling %s ", pattern)
		re := regexp.MustCompile(pattern)
		patterns =  append(patterns, re)
	}
	r := strings.Split(replaceList, ",")
	replaces := []string{}
	for _, t := range r {
		replace := strings.Trim(t, ` "`)
		log.Printf("notes.newNotesProcessor: adding replacement %s ", replace)
		replaces =  append(replaces, replace)
	}
	log.Printf("analysis.newNotesProcessor: got %d patterns ", len(patterns))
	return NotesProcessor {
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
			Id: ws.Id,
			HeadwordId: ws.HeadwordId,
			Simplified: ws.Simplified,
			Traditional: ws.Traditional,
			Pinyin: ws.Pinyin,
			English: ws.English,
			Grammar: ws.Grammar,
			Concept: ws.Concept,
			ConceptCN: ws.ConceptCN,
			Domain: ws.Domain,
			DomainCN: ws.DomainCN,
			Subdomain: ws.Subdomain,
			SubdomainCN: ws.SubdomainCN,
			Image: ws.Image,
			MP3: ws.MP3,
			Notes: n,
		}
		senses = append(senses, s)
	}
	return dicttypes.Word{
		Simplified: w.Simplified,
		Traditional: w.Traditional,
		Pinyin: w.Pinyin,
		HeadwordId: w.HeadwordId,
		Senses: senses,
	}
}
