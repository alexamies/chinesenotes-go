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
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// isEqual checks equality since introduction of generics seems to break reflect.DeepEqual
func isEqual(got, want []TextToken) bool {
	if len(got) != len(want) {
		return false
	}
	for i, _ := range got {
		if got[i].Token != want[i].Token {
			return false
		}
	}
	return true
}

// TestGreedyLtoR tests simple query with one character
func TestNewDictTokenizer(t *testing.T) {
	dict := map[string]bool{"你好": true}
	tokenizer := NewDictTokenizer(dict)
	chunk := "你好"
	tokens := tokenizer.Tokenize(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestNewDictTokenizer: expect list of one token, got ", tokens)
	}
}

// TestGreedyLtoR tests simple query with one character
func TestGreedyLtoR(t *testing.T) {
	dict := map[string]*dicttypes.Word{}
	s1 := "你好"
	w := dicttypes.Word{}
	w.Simplified = s1
	w.Traditional = "\\N"
	w.Pinyin = "nǐhǎo"
	w.HeadwordId = 42
	dict["你好"] = &w
	tokenizer := NewDictTokenizer(dict)
	chunk := "你好"
	tokens := tokenizer.greedyLtoR(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestGreedyLtoR: expect list of one token, got ", tokens)
	}
}

// TestTerm tests the generic method
func TestTerm(t *testing.T) {
	emptyDict := map[string]bool{}
	simpleDict := map[string]bool{"全": true}
	testCases := []struct {
		name string
		dict map[string]bool
		in  string
		wantOK bool
	}{
		{
			name: "Empty",
			dict: emptyDict,
			in: "全", 
			wantOK: false,
		},
		{
			name: "Happy pass",
			dict: simpleDict,
			in: "全",
			wantOK: true,
		},
	}
  for _, tc := range testCases {
		_, ok := term(tc.dict, tc.in)
  	if ok != tc.wantOK {
  		t.Errorf("TestTerm %s, got %t, wantOK %t", tc.name, ok, tc.wantOK)
  	}
	}
}

// TestGreedyRtoL tests simple query with one character
func TestGreedyRtoL(t *testing.T) {
	chunk := "全"
	tokenizer := DictTokenizer[bool]{}
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
	liangwudi := dicttypes.Word{
		Simplified:"梁武帝",
		Traditional: "",
		Pinyin: "liáng wǔ dì",
		HeadwordId: 50,
		Senses: []dicttypes.WordSense{},
	}	
	wen := dicttypes.Word{
		Simplified:"问",
		Traditional: "問",
		Pinyin: "wèn",
		HeadwordId: 51,
		Senses: []dicttypes.WordSense{},
	}	
	damo := dicttypes.Word{
		Simplified:"达磨",
		Traditional: "達磨",
		Pinyin: "Dámó",
		HeadwordId: 52,
		Senses: []dicttypes.WordSense{},
	}	
	yong := dicttypes.Word{
		Simplified:"用",
		Traditional: "",
		Pinyin: "yòng",
		HeadwordId: 53,
		Senses: []dicttypes.WordSense{},
	}	
	gonglao := dicttypes.Word{
		Simplified:"功劳",
		Traditional: "功勞",
		Pinyin: "gōngláo",
		HeadwordId: 54,
		Senses: []dicttypes.WordSense{},
	}	
	lai := dicttypes.Word{
		Simplified:"来",
		Traditional: "來",
		Pinyin: "lái",
		HeadwordId: 55,
		Senses: []dicttypes.WordSense{},
	}	
	dixiao := dicttypes.Word{
		Simplified:"抵消",
		Traditional: "",
		Pinyin: "dǐxiāo",
		HeadwordId: 56,
		Senses: []dicttypes.WordSense{},
	}	
	zuiguo := dicttypes.Word{
		Simplified:"罪过",
		Traditional: "罪過",
		Pinyin: "zuìguò",
		HeadwordId: 57,
		Senses: []dicttypes.WordSense{},
	}	
	wdict := map[string]*dicttypes.Word{
		quan.Simplified: &quan,
		changahan.Simplified: &changahan,
		changahan.Traditional: &changahan,
		xu.Simplified: &xu,
		konglong.Simplified: &konglong,
		konglong.Traditional: &konglong,
		kong.Simplified: &kong,
		kong.Traditional: &kong,
		longtou.Simplified: &longtou,
		longtou.Traditional: &longtou,
		mingyue.Simplified: &mingyue,
		qingfeng.Simplified: &qingfeng,
		qingfeng.Traditional: &qingfeng,
		liangwudi.Simplified: &liangwudi,
		wen.Simplified: &wen,
		wen.Traditional: &wen,
		damo.Simplified: &damo,
		damo.Traditional: &damo,
		yong.Simplified: &yong,
		gonglao.Simplified: &gonglao,
		gonglao.Traditional: &gonglao,
		lai.Simplified: &lai,
		lai.Traditional: &lai,
		dixiao.Simplified: &dixiao,
		zuiguo.Simplified: &zuiguo,
		zuiguo.Traditional: &zuiguo,
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
	lwd := TextToken{
		Token: "梁武帝",
		DictEntry: liangwudi,
		Senses: []dicttypes.WordSense{},
	}
	w := TextToken{
		Token: "問",
		DictEntry: wen,
		Senses: []dicttypes.WordSense{},
	}
	dm := TextToken{
		Token: "達磨",
		DictEntry: damo,
		Senses: []dicttypes.WordSense{},
	}
	y := TextToken{
		Token: "用",
		DictEntry: yong,
		Senses: []dicttypes.WordSense{},
	}
	gl := TextToken{
		Token: "功勞",
		DictEntry: gonglao,
		Senses: []dicttypes.WordSense{},
	}
	l := TextToken{
		Token: "來",
		DictEntry: lai,
		Senses: []dicttypes.WordSense{},
	}
	dx := TextToken{
		Token: "抵消",
		DictEntry: dixiao,
		Senses: []dicttypes.WordSense{},
	}
	zg := TextToken{
		Token: "罪過",
		DictEntry: zuiguo,
		Senses: []dicttypes.WordSense{},
	}
	period := TextToken{
		Token: "。",
	}
	nums := TextToken{
		Token: "1234",
	}
	comma := TextToken{
		Token: "，",
	}
	tokenizer := NewDictTokenizer(wdict)
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
		{
			name: "Three words of different lengths",
			in: "梁武帝問達磨", 
			want: []TextToken{lwd, w, dm},
		},
		{
			name: "Five words",
			in: "用功勞來抵消罪過", 
			want: []TextToken{y, gl, l, dx, zg},
		},
		{
			name: "Chinese with punctuation",
			in: "用功勞來抵消罪過。", 
			want: []TextToken{y, gl, l, dx, zg, period},
		},
		{
			name: "Numbers and words",
			in: "明月1234清風。", 
			want: []TextToken{my, nums, qf, period},
		},
		{
			name: "With comma",
			in: "明月，清風。", 
			want: []TextToken{my, comma, qf, period},
		},
	}
	for _, tc := range testCases {
		got := tokenizer.Tokenize(tc.in)
  	if !isEqual(got, tc.want) {
  		t.Errorf("%s, got %v, want %v", tc.name, got, tc.want)
  	}
	}
}
