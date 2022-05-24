package dictionary

import (
	"context"
	"reflect"
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
	return map[string]*dicttypes.Word{
		s1: &hw1,
		t1: &hw1,
		s2: &hw2,
	}
}

func TestFindWordsByEnglish(t *testing.T) {
	type test struct {
		name        string
		query       string
		expectCount int
	}
	tests := []test{
		{
			name:        "Simple single word",
			query:       "lotus",
			expectCount: 1,
		},
		{
			name:        "With delimiter",
			query:       "region",
			expectCount: 1,
		},
	}
	for _, tc := range tests {
		ctx := context.Background()
		wdict := mockSmallDict()
		dict := NewDictionary(wdict)
		dictSearcher := NewReverseIndex(dict)
		senses, err := dictSearcher.FindWordsByEnglish(ctx, tc.query)
		if err != nil {
			t.Errorf("%s: unexpected error finding by English: %v", tc.name, err)
		}
		if len(senses) != tc.expectCount {
			t.Errorf("%s: got no results: got %d, want %d - %v", tc.name, len(senses), tc.expectCount, senses)
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
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
