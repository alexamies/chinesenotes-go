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
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// A basic example of the function Segment
func ExampleSegment() {
	segments := Segment("你好 means hello")
  fmt.Printf("Text: %s, Chinese: %t\n", segments[0].Text, segments[0].Chinese)
  fmt.Printf("Text: %s, Chinese: %t\n", strings.TrimSpace(segments[1].Text), segments[1].Chinese)
  // Output: Text: 你好, Chinese: true
  // Text: means hello, Chinese: false
}


// A basic example of the function Segment
func TestSegment(t *testing.T) {
	nihao := TextSegment{"你好", true}
	helloIs := TextSegment{"Hello is ", false}
	period := TextSegment{"。", false}
	testCases := []struct {
		name string
		in  string
		want []TextSegment
	}{
		{
			name: "Empty",
			in: "", 
			want: []TextSegment{},
		},
		{
			name: "One segment",
			in: "你好", 
			want: []TextSegment{nihao},
		},
		{
			name: "Two segments",
			in: "Hello is 你好", 
			want: []TextSegment{helloIs, nihao},
		},
		{
			name: "Three segments",
			in: "Hello is 你好。", 
			want: []TextSegment{helloIs, nihao, period},
		},
	}
	for _, tc := range testCases {
		got := Segment(tc.in)
  	if !reflect.DeepEqual(tc.want, got)  {
  		t.Errorf("%s, expected %v, got %v", tc.name, tc.want, got)
  	}
  }
}