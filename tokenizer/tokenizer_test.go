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

// Test trivial query with empty chunk
func TestTokenize0(t *testing.T) {
	log.Printf("TestTokenize0: Begin unit tests\n")
	tokenizer := DictTokenizer{}
	tokens := tokenizer.Tokenize("")
	if len(tokens) != 0 {
		t.Error("TestTokenize0: expect empty list of tokens, got ", tokens)
	}
}

// Test simple query with one character
func TestTokenize1(t *testing.T) {
	tokenizer := DictTokenizer{}
	chunck := "全"
	tokens := tokenizer.Tokenize(chunck)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunck {
		t.Error("TestTokenize1: expect empty list of one token, got ", tokens)
	}
}

// Test simple query with one character
func TestTokenize2(t *testing.T) {
	dict := map[string]dicttypes.Word{}
	s1 := "你好"
	w := dicttypes.Word{}
	w.Simplified = s1
	w.Traditional = "\\N"
	w.Pinyin = "nǐhǎo"
	w.HeadwordId = 42
	dict["你好"] = w
	tokenizer := DictTokenizer{dict}
	chunck := "你好"
	tokens := tokenizer.Tokenize(chunck)
	expect := 1
	if len(tokens) != expect &&  tokens[0].Token != chunck {
		t.Error("TestTokenize1: expect empty list of one token, got ", tokens)
	}
}
