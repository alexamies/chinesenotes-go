package dicttypes

import (
	"reflect"
	"sort"
	"testing"
)

// make example data
func makeHW0() Word {
	simp := "国"
	trad := "國"
	pinyin := "guó"
	wsArray := []WordSense{}
	return Word{
		HeadwordId:  1,
		Simplified:  simp,
		Traditional: trad,
		Pinyin:      pinyin,
		Senses:      wsArray,
	}
}

// make example data
func makeHW1() Word {
	simp := "严净"
	trad := "嚴淨"
	pinyin := "yán jìng"
	wsArray := []WordSense{}
	return Word{
		HeadwordId:  1,
		Simplified:  simp,
		Traditional: trad,
		Pinyin:      pinyin,
		Senses:      wsArray,
	}
}

// make example data
func makeHW2() Word {
	simp := "缘缘"
	trad := "緣緣"
	pinyin := "yuányuán"
	wsArray := []WordSense{}
	return Word{
		HeadwordId:  2,
		Simplified:  simp,
		Traditional: trad,
		Pinyin:      pinyin,
		Senses:      wsArray,
	}
}

// TestCloneWord does a query expecting empty list
func TestCloneWord(t *testing.T) {
	w1 := Word{
		Simplified:  "你好",
		Traditional: "\\N",
		Pinyin:      "nǐhǎo",
		HeadwordId:  42,
	}
	w2 := CloneWord(w1)
	if !reflect.DeepEqual(w1, w2) {
		t.Fatalf("not the same, expected %v, got %v", w1, w2)
	}
}

// TestIsProperNoun does a query expecting empty list
func TestIsProperNoun(t *testing.T) {
	s := WordSense{
		Simplified:  "王",
		Traditional: "\\N",
		Pinyin:      "wáng",
		English:     "Wang",
		Grammar:     "proper noun",
	}
	senses := []WordSense{s}
	w := Word{
		Simplified:  "王",
		Traditional: "\\N",
		Pinyin:      "wáng",
		HeadwordId:  42,
		Senses:      senses,
	}
	got := w.IsProperNoun()
	if !got {
		t.Fatalf("not a proper noun, expected %t, got %t", true, got)
	}
}

// Trival test for headword sorting
func TestWords(t *testing.T) {
	type test struct {
		name         string
		input        Words
		expectFirst  string
		expectSecond string
	}
	hw0 := makeHW0()
	hw1 := makeHW1()
	hw2 := makeHW2()
	tests := []test{
		{
			name:         "happy path",
			input:        Words{hw1, hw0},
			expectFirst:  hw0.Pinyin,
			expectSecond: hw1.Pinyin,
		},
		{
			name:         "slightly longer",
			input:        Words{hw2, hw1, hw0},
			expectFirst:  hw0.Pinyin,
			expectSecond: hw1.Pinyin,
		},
	}
	for _, tc := range tests {
		sort.Sort(tc.input)
		first := tc.input[0].Pinyin
		if tc.expectFirst != first {
			t.Errorf("TestWords %s: expectFirst got %s, expected %s", tc.name, first,
				tc.expectFirst)
		}
		second := tc.input[1].Pinyin
		if tc.expectSecond != second {
			t.Errorf("TestWords %s: expectSecond got %s, expected %s", tc.name,
				second, tc.expectSecond)
		}
	}
}

// Test removal of tones from Pinyin
func TestNormalizePinyin(t *testing.T) {
	type test struct {
		name   string
		input  string
		expect string
	}
	tests := []test{
		{
			name:   "happy path",
			input:  "guó",
			expect: "guo",
		},
		{
			name:   "two syllables",
			input:  "Sān Bǎo",
			expect: "san bao",
		},
		{
			name:   "accent on upper case letter",
			input:  "Ēmítuó",
			expect: "emituo",
		},
	}
	for _, tc := range tests {
		noTones := NormalizePinyin(tc.input)
		if noTones != tc.expect {
			t.Errorf("TestNormalizePinyin %s: got %s but expected %s ", tc.name,
				noTones, tc.expect)
		}
	}
}

// Test hasNotesLabel
func TestHasNotesLabel(t *testing.T) {
	ws1 := WordSense{
		Notes: "Sanskrit equivalent: prajñā",
	}
	w1 := Word{
		HeadwordId:  1,
		Simplified:  "般若",
		Traditional: "\\N",
		Pinyin:      "bōrě",
		Senses:      []WordSense{ws1},
	}
	ws2 := WordSense{
		Notes: "Something else",
	}
	w2 := Word{
		HeadwordId:  2,
		Simplified:  "心与道一",
		Traditional: "心與道一",
		Pinyin:      "xīn yǔ dào yī",
		Senses:      []WordSense{ws2},
	}
	ws3 := WordSense{
		Notes: "FGDB entry 27920",
	}
	ws4 := WordSense{
		Notes: "Some other source",
	}
	w3 := Word{
		HeadwordId: 3,
		Simplified: "白象",
		Pinyin:     "báixiàng",
		Senses:     []WordSense{ws3, ws4},
	}
	tests := []struct {
		name   string
		label  string
		input  Word
		expect bool
	}{
		{
			name:   "Has Sanskrit",
			label:  "Sanskrit equivalent:",
			input:  w1,
			expect: true,
		},
		{
			name:   "Has nothing interesting",
			label:  "Sanskrit equivalent:",
			input:  w2,
			expect: false,
		},
		{
			name:   "Multiple senses",
			label:  "FGDB entry",
			input:  w3,
			expect: true,
		},
	}
	for _, tc := range tests {
		got := tc.input.HasNotesLabel(tc.label)
		if got != tc.expect {
			t.Errorf("TestHasNotesLabel %s: got %t but expected %t ", tc.name,
				got, tc.expect)
		}
	}
}

// Test IsQuote
func TestIsQuote(t *testing.T) {
	type test struct {
		name   string
		input  Word
		expect bool
	}
	w := Word{
		HeadwordId:  1,
		Simplified:  "国",
		Traditional: "國",
		Pinyin:      "guó",
		Senses:      []WordSense{},
	}
	ws := WordSense{
		Notes: "Quote: something in a book",
	}
	quote := Word{
		HeadwordId:  2,
		Simplified:  "心与道一",
		Traditional: "心與道一",
		Pinyin:      "xīn yǔ dào yī",
		Senses:      []WordSense{ws},
	}
	tests := []test{
		{
			name:   "Not a quote",
			input:  w,
			expect: false,
		},
		{
			name:   "Is a quote",
			input:  quote,
			expect: true,
		},
	}
	for _, tc := range tests {
		got := tc.input.IsQuote()
		if got != tc.expect {
			t.Errorf("TestIsQuote %s: got %t but expected %t ", tc.name,
				got, tc.expect)
		}
	}
}

// TestIsCJKChar tests IsCJKChar
func TestIsCJKChar(t *testing.T) {
	type test struct {
		name   string
		input  string
		expect bool
	}
	tests := []test{
		{
			name:   "Chinese",
			input:  "中国",
			expect: true,
		},
		{
			name:   "English",
			input:  "USA",
			expect: false,
		},
		{
			name:   "Hiragana",
			input:  "あ",
			expect: true,
		},
		{
			name:   "Katakana",
			input:  "ア",
			expect: true,
		},
	}
	for _, tc := range tests {
		got := IsCJKChar(tc.input)
		if got != tc.expect {
			t.Errorf("%s: IsCJKChar(%s) got %t but expected %t ", tc.name, tc.input, got, tc.expect)
		}
	}
}
