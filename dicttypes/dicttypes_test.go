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
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// make example data
func makeHW1() Word {
	simp := "严净"
	trad := "嚴淨"
	pinyin := "yán jìng"
	wsArray := []WordSense{}
	return Word{
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// make example data
func makeHW2() Word {
	simp := "缘缘"
	trad := "緣緣"
	pinyin := "yuányuán"
	wsArray := []WordSense{}
	return Word{
		HeadwordId: 2,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// TestAddWordSense2Map does a query expecting empty list
func TestCloneWord(t *testing.T) {
	w1 := Word{
		Simplified: "你好",
		Traditional: "\\N",
		Pinyin: "nǐhǎo",
		HeadwordId: 42,
	}
	w2 := CloneWord(w1)
	if !reflect.DeepEqual(w1, w2) {
		t.Fatalf("not the same, expected %v, got %v", w1, w2)
	}
}

// TestAddWordSense2Map does a query expecting empty list
func TestIsProperNoun(t *testing.T) {
	s := WordSense{
		Simplified: "王",
		Traditional: "\\N",
		Pinyin: "wáng",
		English: "Wang",
		Grammar: "proper noun",
	}
	senses := []WordSense{s}
	w := Word{
		Simplified: "王",
		Traditional: "\\N",
		Pinyin: "wáng",
		HeadwordId: 42,
		Senses: senses,
	}
	got := w.IsProperNoun()
	if !got {
		t.Fatalf("not a proper noun, expected %t, got %t", true, got)
	}
}


// Trival test for headword sorting
func TestWords(t *testing.T) {
	hw0 := makeHW0()
	hw1 := makeHW1()
	hws := Words{hw1, hw0}
	sort.Sort(hws)
	firstWord := hws[0].Pinyin[0]
	pinyinExpected := hw0.Pinyin[0]
	if pinyinExpected != firstWord {
		t.Error("dictionary.TestHeadwords1: Expected pinyin ", pinyinExpected,
			", got", firstWord)
	}
}

// Better test for headword sorting
func TestHeadwords2(t *testing.T) {
	hw0 := makeHW0()
	hw1 := makeHW1()
	hw2 := makeHW2()
	hws := Words{hw2, hw1, hw0}
	sort.Sort(hws)
	firstWord := hws[0].Pinyin[0]
	pinyinExpected := hw0.Pinyin[0]
	if pinyinExpected != firstWord {
		t.Error("dictionary.TestHeadwords2: Expected pinyin ", pinyinExpected,
			", got", firstWord)
	}
	secondWord := hws[1].Pinyin[0]
	secondExpected := hw1.Pinyin[0]
	if secondExpected != secondWord {
		t.Error("dictionary.TestHeadwords2: 2nd expected pinyin ",
			secondExpected,	", got", secondWord)
	}
}

// Test removal of tones from Pinyin
func TestNormalizePinyin0(t *testing.T) {
	pinyin := "guó"
	noTones := normalizePinyin(pinyin)
	expected := "guo"
	if expected != noTones {
		t.Error("dictionary.TestNormalizePinyin0: expected noTones ",
			expected, ", got", noTones)
	}
}

// Test removal of tones from Pinyin
func TestNormalizePinyin1(t *testing.T) {
	pinyin := "Sān Bǎo"
	noTones := normalizePinyin(pinyin)
	expected := "san bao"
	if expected != noTones {
		t.Error("dictionary.TestNormalizePinyin1: expected noTones ",
			expected, ", got", noTones)
	}
}

// Test removal of tones from Pinyin
func TestNormalizePinyin2(t *testing.T) {
	pinyin := "Ēmítuó"
	noTones := normalizePinyin(pinyin)
	expected := "emituo"
	if expected != noTones {
		t.Error("dictionary.TestNormalizePinyin1: expected noTones ",
			expected, ", got", noTones)
	}
}