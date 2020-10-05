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
	"testing"
)

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