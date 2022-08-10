package dictionary

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

func mockSmallDict() map[string]*dicttypes.Word {
	s1 := "莲花"
	t1 := "蓮花"
	p1 := "liánhuā"
	e1 := "lotus"
	s2 := "域"
	t2 := "\\N"
	p2 := "yù"
	e2 := "district; region"
	s3 := "喜马拉雅雪松"
	t3 := "喜馬拉雅雪松"
	p3 := "xǐmǎlāyǎ xuěsōng"
	e3 := "deodar cedar"
	n3 := "Scientific name: Cedrus deodara, aka: Himalayan cedar; a native to the Himalayas (Wikipedia '喜马拉雅雪松')"
	s4 := "北京"
	t4 := "\\N"
	p4 := "běijīng"
	e4 := "Beijing"
	s5 := "炼化"
	t5 := "煉化"
	p5 := "liànhuà"
	e5 := "to refine; refining (e.g. oil, chemicals etc)"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      p1,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  1,
				Simplified:  s1,
				Traditional: t1,
				Pinyin:      p1,
				English:     e1,
			},
		},
	}
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      p2,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  2,
				Simplified:  s2,
				Traditional: t2,
				Pinyin:      p2,
				English:     e2,
			},
		},
	}
	hw3 := dicttypes.Word{
		HeadwordId:  3,
		Simplified:  s3,
		Traditional: t3,
		Pinyin:      p3,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  3,
				Simplified:  s3,
				Traditional: t3,
				Pinyin:      p3,
				English:     e3,
				Notes:       n3,
			},
		},
	}
	hw4 := dicttypes.Word{
		HeadwordId:  4,
		Simplified:  s4,
		Traditional: t4,
		Pinyin:      p4,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  4,
				Simplified:  s4,
				Traditional: t4,
				Pinyin:      p4,
				English:     e4,
			},
		},
	}
	hw5 := dicttypes.Word{
		HeadwordId:  5,
		Simplified:  s5,
		Traditional: t5,
		Pinyin:      p5,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  5,
				Simplified:  s5,
				Traditional: t5,
				Pinyin:      p5,
				English:     e5,
			},
		},
	}
	return map[string]*dicttypes.Word{
		s1: &hw1,
		t1: &hw1,
		s2: &hw2,
		s3: &hw3,
		t3: &hw3,
		s4: &hw4,
		s5: &hw5,
		t5: &hw5,
	}
}

func TestFind(t *testing.T) {
	type test struct {
		name        string
		extractRe   string
		query       string
		expectCount int
		expectTrad  string
		expectHwId  int
	}
	tests := []test{
		{
			name:        "Simple single word",
			extractRe:   "",
			query:       "lotus",
			expectCount: 1,
			expectTrad:  "蓮花",
			expectHwId:  1,
		},
		{
			name:        "With delimiter",
			extractRe:   `"Scientific name: (.*?)[\(,\,,\;]","Species: (.*?)[\(,\,,\;]"`,
			query:       "region",
			expectCount: 1,
			expectTrad:  "",
			expectHwId:  2,
		},
		{
			name:        "From pinyin",
			extractRe:   "",
			query:       "lianhua",
			expectCount: 2,
			expectTrad:  "蓮花",
			expectHwId:  1,
		},
		{
			name:        "Equivalent from notes",
			extractRe:   `"Scientific name: (.*?)[\(,\,,\;]","Species: (.*?)[\(,\,,\;]"`,
			query:       "cedrus deodara",
			expectCount: 1,
			expectTrad:  "喜馬拉雅雪松",
			expectHwId:  3,
		},
		{
			name:        "English with upper case",
			extractRe:   "",
			query:       "beijing",
			expectCount: 1,
			expectTrad:  "",
			expectHwId:  4,
		},
		{
			name:        "English with paretheses",
			extractRe:   "",
			query:       "refining",
			expectCount: 1,
			expectTrad:  "煉化",
			expectHwId:  5,
		},
	}
	for _, tc := range tests {
		ctx := context.Background()
		wdict := mockSmallDict()
		dict := NewDictionary(wdict)
		extractor, err := NewNotesExtractor(tc.extractRe)
		if err != nil {
			t.Errorf("TestFind %s: could not create extractor: %v", tc.name, err)
		}
		dictSearcher := NewReverseIndex(dict, extractor)
		senses, err := dictSearcher.Find(ctx, tc.query)
		if err != nil {
			t.Errorf("%s: unexpected error finding by English: %v", tc.name, err)
		}
		if len(senses) != tc.expectCount {
			t.Errorf("TestFind %s: num results: got %d, want %d - %v", tc.name, len(senses), tc.expectCount, senses)
		}
		if len(senses) == 1 && senses[0].Traditional != tc.expectTrad {
			t.Errorf("TestFind %s: got traditional %s, want %s", tc.name, senses[0].Traditional, tc.expectTrad)
		}
		if len(senses) == 1 && senses[0].HeadwordId != tc.expectHwId {
			t.Errorf("TestFind %s: got Headword Id %d, want %d", tc.name, senses[0].HeadwordId, tc.expectHwId)
		}
	}
}

func TestSplitEnglish(t *testing.T) {
	type test struct {
		name       string
		equivalent string
		want       []string
	}
	tests := []test{
		{
			name:       "One equivalent",
			equivalent: "item",
			want:       []string{"item"},
		},
		{
			name:       "Two equivalents",
			equivalent: "item; thing",
			want:       []string{"item", "thing"},
		},
		{
			name:       "With stopwords",
			equivalent: "an item; a thing",
			want:       []string{"item", "thing"},
		},
	}
	for _, tc := range tests {
		got := splitEnglish(tc.equivalent)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: splitEnglish(%s) got %q, want %q", tc.name, tc.equivalent, got, tc.want)
		}
	}
}

func TestStripParen(t *testing.T) {
	type test struct {
		name       string
		equivalent string
		want       string
	}
	tests := []test{
		{
			name:       "No parentheses",
			equivalent: "item",
			want:       "item",
		},
		{
			name:       "One parentheses",
			equivalent: "item (a thing)",
			want:       "item",
		},
	}
	for _, tc := range tests {
		got := stripParen(tc.equivalent)
		if got != tc.want {
			t.Errorf("%s: stripParen(%s) got %q, want %q", tc.name, tc.equivalent, got, tc.want)
		}
	}
}

func TestNgrams(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		minLen int
		want   []string
	}{
		{
			name:   "Empty",
			input:  "",
			minLen: 2,
			want:   []string{},
		},
		{
			name:   "One character",
			input:  "世",
			minLen: 2,
			want:   []string{},
		},
		{
			name:   "Min length 1",
			input:  "世",
			minLen: 1,
			want:   []string{"世"},
		},
		{
			name:   "Two characters",
			input:  "世界",
			minLen: 2,
			want:   []string{"世界"},
		},
		{
			name:   "Three characters",
			input:  "看世界",
			minLen: 2,
			want:   []string{"看世界", "看世", "世界"},
		},
		{
			name:   "Four characters",
			input:  "看看世界",
			minLen: 2,
			want:   []string{"看看世界", "看看世", "看看", "看世界", "看世", "世界"},
		},
		{
			name:   "Five characters",
			input:  "看整個世界",
			minLen: 2,
			want:   []string{"看整個世界", "看整個世", "看整個", "看整", "整個世界", "整個世", "整個", "個世界", "個世", "世界"},
		},
	}
	for _, tc := range tests {
		chars := strings.Split(tc.input, "")
		got := Ngrams(chars, tc.minLen)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s, got %v\n but want %v", tc.name, got, tc.want)
		}
	}
}
