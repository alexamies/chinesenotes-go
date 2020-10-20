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

package tokenizer

import (
	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// A text segment that contains either Chinese or non-Chinese text
type TextSegment struct{

	// The text contained in the segment
	Text string

	// False if punctuation or non-Chinese text
	Chinese bool
}

// Segment a text document into segments of Chinese separated by either
// puncuation or non-Chinese text.
func Segment(text string) []TextSegment {
	segments := []TextSegment{}
	cjk := ""
	noncjk := ""
	for _, character := range text {
		if dicttypes.IsCJKChar(string(character)) {
			if noncjk != "" {
				s := TextSegment{noncjk, false}
				segments = append(segments, s)
				noncjk = ""
			}
			cjk += string(character)
		} else if cjk != "" {
			s := TextSegment{cjk, true}
			segments = append(segments, s)
			cjk = ""
			noncjk += string(character)
		} else {
			noncjk += string(character)
		}
	}
	if cjk != "" {
		s := TextSegment{cjk, true}
		segments = append(segments, s)
	}
	if noncjk != "" {
		s := TextSegment{noncjk, false}
		segments = append(segments, s)
	}
	return segments
}
