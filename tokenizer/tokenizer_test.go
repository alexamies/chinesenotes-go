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


// Unit tests for tokenizer functions
package tokenizer

import (
	"reflect"
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Test simple query with one character
func TestGreedyLtoR(t *testing.T) {
	t.Log("TestGreedyLtoR: Begin unit tests")
	dict := map[string]dicttypes.Word{}
	s1 := "你好"
	w := dicttypes.Word{}
	w.Simplified = s1
	w.Traditional = "\\N"
	w.Pinyin = "nǐhǎo"
	w.HeadwordId = 42
	dict["你好"] = w
	tokenizer := DictTokenizer{dict}
	chunk := "你好"
	tokens := tokenizer.greedyLtoR(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestTokenize1: expect list of one token, got ", tokens)
	}
}

// Test simple query with one character
func TestGreedyRtoL(t *testing.T) {
	chunk := "全"
	tokenizer := DictTokenizer{}
	tokens := tokenizer.greedyRtoL(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestGreedyRtoL: expect list of one token, got ", tokens)
	}
}

// Test trivial query with empty chunk
func TestTokenize(t *testing.T) {
	quan := dicttypes.Word{
		Simplified:"全",
		Traditional: "",
		Pinyin: "quán",
		HeadwordId: 42,
		Senses: []dicttypes.WordSense{},
	}	
	changahan := dicttypes.Word{
		Simplified:"长阿含经",
		Traditional: "長阿含經",
		Pinyin: "Cháng Āhán Jīng",
		HeadwordId: 43,
		Senses: []dicttypes.WordSense{},
	}	
	xu := dicttypes.Word{
		Simplified:"序",
		Traditional: "",
		Pinyin: "xù",
		HeadwordId: 44,
		Senses: []dicttypes.WordSense{},
	}	
	konglong := dicttypes.Word{
		Simplified:"恐龙",
		Traditional: "恐龍",
		Pinyin: "kǒnglóng",
		HeadwordId: 45,
		Senses: []dicttypes.WordSense{},
	}	
	kong := dicttypes.Word{
		Simplified:"恐",
		Traditional: "恐",
		Pinyin: "kǒnglóng",
		HeadwordId: 46,
		Senses: []dicttypes.WordSense{},
	}	
	longtou := dicttypes.Word{
		Simplified:"龙头蛇尾",
		Traditional: "龍頭蛇尾",
		Pinyin: "lóng tóu shé wěi",
		HeadwordId: 47,
		Senses: []dicttypes.WordSense{},
	}	
	mingyue := dicttypes.Word{
		Simplified:"明月",
		Traditional: "",
		Pinyin: "míngyuè",
		HeadwordId: 48,
		Senses: []dicttypes.WordSense{},
	}	
	qingfeng := dicttypes.Word{
		Simplified:"清风",
		Traditional: "清風",
		Pinyin: "qīngfēng",
		HeadwordId: 49,
		Senses: []dicttypes.WordSense{},
	}	
	wdict := map[string]dicttypes.Word{
		quan.Simplified: quan,
		changahan.Simplified: changahan,
		changahan.Traditional: changahan,
		xu.Simplified: xu,
		konglong.Simplified: konglong,
		konglong.Traditional: konglong,
		kong.Simplified: kong,
		kong.Traditional: kong,
		longtou.Simplified: longtou,
		longtou.Traditional: longtou,
		mingyue.Simplified: mingyue,
		qingfeng.Simplified: qingfeng,
		qingfeng.Traditional: qingfeng,
	}
	q := TextToken{
		Token: "全",
		DictEntry: quan,
		Senses: []dicttypes.WordSense{},
	}
	c := TextToken{
		Token: "長阿含經",
		DictEntry: changahan,
		Senses: []dicttypes.WordSense{},
	}
	x := TextToken{
		Token: "序",
		DictEntry: xu,
		Senses: []dicttypes.WordSense{},
	}
	k := TextToken{
		Token: "恐",
		DictEntry: kong,
		Senses: []dicttypes.WordSense{},
	}
	lt := TextToken{
		Token: "龍頭蛇尾",
		DictEntry: longtou,
		Senses: []dicttypes.WordSense{},
	}
	my := TextToken{
		Token: "明月",
		DictEntry: mingyue,
		Senses: []dicttypes.WordSense{},
	}
	qf := TextToken{
		Token: "清風",
		DictEntry: qingfeng,
		Senses: []dicttypes.WordSense{},
	}
	tokenizer := DictTokenizer{wdict}
	testCases := []struct {
		name string
		in  string
		want []TextToken
	}{
		{
			name: "Empty",
			in: "", 
			want: []TextToken{},
		},
		{
			name: "One token",
			in: "全", 
			want: []TextToken{q},
		},
		{
			name: "Two tokens",
			in: "長阿含經序", 
			want: []TextToken{c, x},
		},
		{
			name: "Overlapping words with R2L winning",
			in: "恐龍頭蛇尾", 
			want: []TextToken{k, lt},
		},
		{
			name: "Two 2 character words",
			in: "明月清風", 
			want: []TextToken{my, qf},
		},
	}
	for _, tc := range testCases {
		got := tokenizer.Tokenize(tc.in)
  	if !reflect.DeepEqual(tc.want, got)  {
  		t.Errorf("%s, expected %v, got %v", tc.name, tc.want, got)
  	}
	}
}

// Three words of different lengths
func TestTokenize3b(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "梁武帝"
	t1 := "\\N"
	w1 := dicttypes.Word{}
	w1.Simplified = s1
	w1.Traditional = t1
	w1.Pinyin = "liáng wǔ dì"
	w1.HeadwordId = 96375
	dict[s1] = w1
	dict[t1] = w1
	s2 := "问"
	t2 := "問"
	w2 := dicttypes.Word{}
	w2.Simplified = s2
	w2.Traditional = t2
	w2.Pinyin = "wèn"
	w2.HeadwordId = 3723
	dict[s2] = w2
	dict[t2] = w2
	s3 := "达磨"
	t3 := "達磨"
	w3 := dicttypes.Word{}
	w3.Simplified = s3
	w3.Traditional = t3
	w3.Pinyin = "Dámó"
	w3.HeadwordId = 17723
	dict[s3] = w3
	dict[t3] = w3
	tokenizer := DictTokenizer{dict}
	chunk := "梁武帝問達磨"
	tokens := tokenizer.Tokenize(chunk)
	expect := 3
	if len(tokens) != expect {
		t.Error("TestTokenize3b: expect list of two tokens, got ", tokens)
	}
	if tokens[0].Token != "梁武帝" {
		t.Error("TestTokenize3b: tokens[0].Token = 梁武帝, got ", tokens[0].Token)
	}
	if tokens[1].Token != "問" {
		t.Error("TestTokenize3b: tokens[1].Token = 問, got ", tokens[1].Token)
	}
	if tokens[2].Token != "達磨" {
		t.Error("TestTokenize3b: tokens[2].Token = 達磨, got ", tokens[2].Token)
	}
}

// Five words 
func TestTokenize5(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "用"
	t1 := "\\N"
	w1 := dicttypes.Word{}
	w1.Simplified = s1
	w1.Traditional = t1
	w1.Pinyin = "yòng"
	w1.HeadwordId = 721
	dict[s1] = w1
	s2 := "功劳"
	t2 := "功勞"
	w2 := dicttypes.Word{}
	w2.Simplified = s2
	w2.Traditional = t2
	w2.Pinyin = "gōngláo"
	w2.HeadwordId = 12162
	dict[s2] = w2
	dict[t2] = w2
	s3 := "来"
	t3 := "來"
	w3 := dicttypes.Word{}
	w3.Simplified = s3
	w3.Traditional = t3
	w3.Pinyin = "lái"
	w3.HeadwordId = 370
	dict[s3] = w3
	dict[t3] = w3
	s4 := "抵消"
	t4 := "\\N"
	w4 := dicttypes.Word{}
	w4.Simplified = s4
	w4.Traditional = t4
	w4.Pinyin = "dǐxiāo"
	w4.HeadwordId = 13197
	dict[s4] = w4
	s5 := "罪过"
	t5 := "罪過"
	w5 := dicttypes.Word{}
	w5.Simplified = s5
	w5.Traditional = t5
	w5.Pinyin = "zuìguò"
	w5.HeadwordId = 69248
	dict[s5] = w5
	dict[t5] = w5
	tokenizer := DictTokenizer{dict}
	chunk := "用功勞來抵消罪過"
	tokens := tokenizer.Tokenize(chunk)
	expect := 5
	if len(tokens) != expect {
		t.Error("TestTokenize5: expect list of two tokens, got ", tokens)
	}
	if tokens[0].Token != s1 {
		t.Error("TestTokenize5: unexpected tokens[0].Token ", tokens[0].Token)
	}
	if tokens[1].Token != t2 {
		t.Error("TestTokenize5: unexpected tokens[1].Token ", tokens[1].Token)
	}
	if tokens[2].Token != t3 {
		t.Error("TestTokenize5: unexpected tokens[2].Token ", tokens[2].Token)
	}
	if tokens[3].Token != s4 {
		t.Error("TestTokenize5: unexpected tokens[3].Token ", tokens[3].Token)
	}
	if tokens[4].Token != t5 {
		t.Error("TestTokenize5: unexpected tokens[4].Token ", tokens[4].Token)
	}
}
