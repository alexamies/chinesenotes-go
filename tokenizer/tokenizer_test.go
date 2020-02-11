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
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"log"
	"testing"
)

// Test simple query with one character
func TestGreedyLtoR(t *testing.T) {
	log.Printf("TestGreedyLtoR: Begin unit tests\n")
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
func TestGreedyRtoL1(t *testing.T) {
	chunk := "全"
	tokenizer := DictTokenizer{}
	tokens := tokenizer.greedyRtoL(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestGreedyRtoL1: expect list of one token, got ", tokens)
	}
}

// Test trivial query with empty chunk
func TestTokenize0(t *testing.T) {
	tokenizer := DictTokenizer{}
	tokens := tokenizer.Tokenize("")
	if len(tokens) != 0 {
		t.Error("TestTokenize0: expect empty list of tokens, got ", tokens)
	}
}

// Test simple query with one character
func TestTokenize1(t *testing.T) {
	tokenizer := DictTokenizer{}
	chunk := "全"
	tokens := tokenizer.Tokenize(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestTokenize1: expect list of one token, got ", tokens)
	}
}

// Simple two word test
func TestTokenize2a(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "长阿含经"
	t1 := "長阿含經"
	w1 := dicttypes.Word{}
	w1.Simplified = s1
	w1.Traditional = t1
	w1.Pinyin = "Cháng Āhán Jīng"
	w1.HeadwordId = 29679
	dict[s1] = w1
	dict[t1] = w1
	s2 := "序"
	t2 := "\\N"
	w2 := dicttypes.Word{}
	w2.Simplified = s2
	w2.Traditional = t2
	w2.Pinyin = "xù "
	w2.HeadwordId = 6213
	dict[s2] = w2
	dict[t2] = w2
	tokenizer := DictTokenizer{dict}
	chunk := "長阿含經序"
	tokens := tokenizer.Tokenize(chunk)
	expect := 2
	if len(tokens) != expect {
		t.Error("TestTokenize2a: expect list of two tokens, got ", tokens)
	}
	if tokens[0].Token != "長阿含經" {
		t.Error("TestTokenize2a: tokens[0].Token = 長阿含經, got ", tokens[0].Token)
	}
	if tokens[1].Token != "序" {
		t.Error("TestTokenize2a: tokens[1].Token = 序, got ",
			tokens[1].Token)
	}
}

// Harder test, overlapping words with R2L winning
func TestTokenize2b(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "恐龙"
	t1 := "恐龍"
	w1 := dicttypes.Word{}
	w1.Simplified = s1
	w1.Traditional = t1
	w1.Pinyin = "kǒnglóng"
	w1.HeadwordId = 75439
	dict[s1] = w1
	dict[t1] = w1
	s2 := "龙头蛇尾"
	t2 := "龍頭蛇尾"
	w2 := dicttypes.Word{}
	w2.Simplified = s2
	w2.Traditional = t2
	w2.Pinyin = "lóng tóu shé wěi"
	w2.HeadwordId = 106010
	dict[s2] = w2
	dict[t2] = w2
	tokenizer := DictTokenizer{dict}
	chunk := "恐龍頭蛇尾"
	tokens := tokenizer.Tokenize(chunk)
	expect := 2
	if len(tokens) != expect {
		t.Error("TestTokenize2b: expect list of two tokens, got ", tokens)
	}
	if tokens[0].Token != "恐" {
		t.Error("TestTokenize2b: tokens[0].Token = 恐, got ", tokens[0].Token)
	}
	if tokens[1].Token != "龍頭蛇尾" {
		t.Error("TestTokenize2b: tokens[1].Token = 龍頭蛇尾, got ",
			tokens[1].Token)
	}
}

// Single 3 character word
func TestTokenize3(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "未曾有"
	t1 := "\\N"
	w1 := dicttypes.Word{}
	w1.Simplified = s1
	w1.Traditional = t1
	w1.Pinyin = "wèi céng yǒu"
	w1.HeadwordId = 30356
	dict[s1] = w1
	tokenizer := DictTokenizer{dict}
	chunk := s1
	tokens := tokenizer.Tokenize(chunk)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunk {
		t.Error("TestTokenize3: expect list of one token, got ", tokens)
	}
}

// Two 2 character words
func TestTokenize4(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "明月"
	t1 := "\\N"
	w1 := dicttypes.Word{}
	w1.Simplified = s1
	w1.Traditional = t1
	w1.Pinyin = "míngyuè"
	w1.HeadwordId = 11304
	dict[s1] = w1
	s2 := "清风"
	t2 := "清風"
	w2 := dicttypes.Word{}
	w2.Simplified = s2
	w2.Traditional = t2
	w2.Pinyin = "qīngfēng"
	w2.HeadwordId = 67740
	dict[s2] = w2
	dict[t2] = w2
	tokenizer := DictTokenizer{dict}
	chunk := "明月清風"
	tokens := tokenizer.Tokenize(chunk)
	expect := 2
	if len(tokens) != expect {
		t.Error("TestTokenize4: expect list of two tokens, got ", tokens)
	}
	if tokens[0].Token != "明月" {
		t.Error("TestTokenize4: tokens[0].Token = 明月, got ", tokens[0].Token)
	}
	if tokens[1].Token != "清風" {
		t.Error("TestTokenize4: tokens[1].Token = 清風, got ",
			tokens[1].Token)
	}
}

// Three words of different lengths
func TestTokenize6(t *testing.T) {
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
		t.Error("TestTokenize6: expect list of two tokens, got ", tokens)
	}
	if tokens[0].Token != "梁武帝" {
		t.Error("TestTokenize6: tokens[0].Token = 梁武帝, got ", tokens[0].Token)
	}
	if tokens[1].Token != "問" {
		t.Error("TestTokenize6: tokens[1].Token = 問, got ", tokens[1].Token)
	}
	if tokens[2].Token != "達磨" {
		t.Error("TestTokenize6: tokens[1].Token = 達磨, got ", tokens[2].Token)
	}
}
