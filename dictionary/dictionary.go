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
//
// Chinese-English dictionary database search functions

package dictionary

import (
	"context"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// Dictionary is a struct to hold word dictionary indexes
type Dictionary struct {
	// Forward dictionary, lookup by Chinese word
	Wdict map[string]*dicttypes.Word
	HeadwordIds map[int]*dicttypes.Word
}

func NewDictionary(wdict map[string]*dicttypes.Word) *Dictionary {
	hwIdMap := make(map[int]*dicttypes.Word)
	for _, w := range wdict {
		hwIdMap[w.HeadwordId] = w
	}
	return &Dictionary{
		Wdict: wdict,
		HeadwordIds: hwIdMap,
	}
}

// ReverseIndex searches the dictionary by reverse lookup, eg from English to Chinese
type ReverseIndex interface {
	Initialized() bool
	FindWordsByEnglish(ctx context.Context, query string) ([]dicttypes.WordSense, error)
}

type SubstringIndex interface{
	LookupSubstr(ctx context.Context, query, topic_en, subtopic_en string) (*Results, error)
}