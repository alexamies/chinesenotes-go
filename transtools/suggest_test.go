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

// Unit tests for translate functions
package transtools

import (
	"reflect"
	"strings"
	"testing"
)

const expected = `# Chinese term,expected term present
綠色,green
頭髮,Hair
莊嚴,"solemn,majestic"
`

const replacement = `# Target,replacement
green color,green
`

// TestCorpusDataDir is a trivial query with empty chunk
func TestSuggest(t *testing.T) {
	t.Logf("TestSuggest: Begin unit tests\n")
	green := Note{
		FoundCN: "綠色",
		ExpectedEN: []string{"green"},
	}
	rep := Note{
		FoundEN: "green color",
		Replacement: "green",
	}
	type test struct {
		name        string
		source      string
		translation string
		expected    Results
	}
	tests := []test{
		{
			name:        "Easy pass",
			source:      "綠色",
			translation: "green",
			expected: Results{
				Replacement: "green",
				Notes:       []Note{},
			},
		},
		{
			name:        "Wrong answer",
			source:      "綠色",
			translation: "red",
			expected: Results{
				Replacement: "red",
				Notes:       []Note{green},
			},
		},
		{
			name:        "Make the replacement",
			source:      "綠色",
			translation: "green color",
			expected: Results{
				Replacement: "green",
				Notes:       []Note{rep},
			},
		},
		{
			name:        "Should be case insensitive",
			source:      "綠色頭髮",
			translation: "green hair",
			expected: Results{
				Replacement: "green hair",
				Notes:       []Note{},
			},
		},
		{
			name:        "Expect one of",
			source:      "莊嚴綠色頭髮",
			translation: "majestic green hair",
			expected: Results{
				Replacement: "majestic green hair",
				Notes:       []Note{},
			},
		},
	}
	expectedReader := strings.NewReader(expected)
	replacementReader := strings.NewReader(replacement)
	p := NewProcessor(expectedReader, replacementReader)
	for _, tc := range tests {
		got := p.Suggest(tc.source, tc.translation)
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("%s: got: %v, expected: %v", tc.name, got, tc.expected)
		}
	}
}
