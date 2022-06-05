package dictionary

import (
	"reflect"
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// TestExtract tests extraction of equilalents from notes
func TestExtract(t *testing.T) {
	// const replace = `"Scientific name: Cedrus (CC-CEDICT '雪松'; Wikipedia '雪松')"`
	testCases := []struct {
		name      string
		extractRe string
		note      string
		want      []string
	}{
		{
			name:      "Empty",
			extractRe: ``,
			note:      "",
			want:      []string{},
		},
		{
			name:      "Single equivalent",
			extractRe: `"Scientific name: (.*?)[\(,\,,\;]"`,
			note:      "Scientific name: Cedrus (CC-CEDICT '雪松'; Wikipedia '雪松')",
			want:      []string{"Cedrus"},
		},
		{
			name:      "Eqiuvalent has spaces",
			extractRe: `"Scientific name: (.*)[\(,\,,\;]"`,
			note:      "Scientific name: Cedrus deodara (Wikipedia '喜马拉雅雪松')",
			want:      []string{"Cedrus deodara"},
		},
		{
			name:      "Two equivalents",
			extractRe: `"Scientific name: (.*?)[\(,\,,\;]","aka: (.*?)[\(,\,,\;]"`,
			note:      "Scientific name: Cedrus deodara, aka: Himalayan cedar; a species of cedar native to the Himalayas (Wikipedia '喜马拉雅雪松')",
			want:      []string{"Cedrus deodara", "Himalayan cedar"},
		},
		{
			name:      "Two regex, one equivalent",
			extractRe: `"Scientific name: (.*?)[\(,\,,\;]","Species: (.*?)[\(,\,,\;]"`,
			note:      "Species: Caprimulgus indicus (Unihan '鷏')",
			want:      []string{"Caprimulgus indicus"},
		},
	}
	for _, tc := range testCases {
		extractor, err := NewNotesExtractor(tc.extractRe)
		if err != nil {
			t.Errorf("TestExtract %s: could not create extractor: %v", tc.name, err)
		}
		got := extractor.Extract(tc.note)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TestExtract %s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}

// TestNotesProcessor tests processing of notes
func TestNotesProcessor(t *testing.T) {
	const match = `"(T ([0-9]))(\)|,|;)","(T ([0-9]{2}))(\)|,|;)","(T ([0-9]{3}))(\)|,|;)","(T ([0-9]{4}))(\)|,|;)"`
	const replace = `"<a href="/taisho/t000${2}.html">${1}</a>${3}","<a href="/taisho/t00${2}.html">${1}</a>${3}","<a href="/taisho/t0${2}.html">${1}</a>${3}","<a href="/taisho/t${2}.html">${1}</a>${3}"`
	testCases := []struct {
		name    string
		match   string
		replace string
		notes   string
		expect  string
	}{
		{
			name:    "Empty",
			match:   "",
			replace: "",
			notes:   "hello",
			expect:  "hello",
		},
		{
			name:    "Basic capture",
			match:   `(T 1)`,
			replace: "<a>${1}</a>",
			notes:   "T 1",
			expect:  "<a>T 1</a>",
		},
		{
			name:    "Single digit",
			match:   `(T ([0-9]))(\)|,|;)`,
			replace: `<a href="/taisho/t000${2}.html">${1}</a>${3}`,
			notes:   "; T 2)",
			expect:  `; <a href="/taisho/t0002.html">T 2</a>)`,
		},
		{
			name:    "Two digits",
			match:   `(T ([0-9]{2}))(\)|,|;)`,
			replace: `<a href="/taisho/t00${2}.html">${1}</a>${3}`,
			notes:   " T 12,",
			expect:  ` <a href="/taisho/t0012.html">T 12</a>,`,
		},
		{
			name:    "No match",
			match:   `(T ([0-9]{2}))(\)|,|;)`,
			replace: `<a href="/taisho/t00${2}.html">${1}</a>${3}`,
			notes:   "Testing 123",
			expect:  `Testing 123`,
		},
		{
			name:    "One Taisho text",
			match:   match,
			replace: replace,
			notes:   `(T 123)`,
			expect:  `(<a href="/taisho/t0123.html">T 123</a>)`,
		},
		{
			name:    "Replace Taisho abbreviations",
			match:   match,
			replace: replace,
			notes:   `(T 1; T 23; T 456; T 1234)`,
			expect:  `(<a href="/taisho/t0001.html">T 1</a>; <a href="/taisho/t0023.html">T 23</a>; <a href="/taisho/t0456.html">T 456</a>; <a href="/taisho/t1234.html">T 1234</a>)`,
		},
		{
			name:    "FGDB entry",
			match:   `FGDB entry ([0-9]*)`,
			replace: `<a href="/web/${1}.html">FGDB entry</a>`,
			notes:   `FGDB entry 15`,
			expect:  `<a href="/web/15.html">FGDB entry</a>`,
		},
	}
	for _, tc := range testCases {
		processor := NewNotesProcessor(tc.match, tc.replace)
		got := processor.process(tc.notes)
		if got != tc.expect {
			t.Errorf("TestNotesProcessor %s: got %s, want %s", tc.name, got,
				tc.expect)
		}
	}
}

// TestProcess tests processing of a headword
func TestProcess(t *testing.T) {
	s := "一時三相"
	ws := dicttypes.WordSense{
		Id:         1,
		Simplified: s,
		Notes:      "FGDB entry 9412",
	}
	hw := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s,
		Senses:     []dicttypes.WordSense{ws},
	}
	testCases := []struct {
		name    string
		match   string
		replace string
		word    dicttypes.Word
		expect  string
	}{
		{
			name:    "FGDB entry",
			match:   `FGDB entry ([0-9]*)`,
			replace: `<a href="/web/${1}.html">FGDB entry</a>`,
			word:    hw,
			expect:  `<a href="/web/9412.html">FGDB entry</a>`,
		},
	}
	for _, tc := range testCases {
		processor := NewNotesProcessor(tc.match, tc.replace)
		got := processor.Process(tc.word)
		gotNotes := got.Senses[0].Notes
		if gotNotes != tc.expect {
			t.Errorf("TestProcess %s: got %s, want %s", tc.name, gotNotes,
				tc.expect)
		}
	}
}
