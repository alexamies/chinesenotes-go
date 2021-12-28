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

// Unit tests for bibnotes functions

package bibnotes

import (
	"strings"
	"testing"
)

const (
	ref2FileSmall = `reference_no,file_name
1,example_collection.tsv
`
	refNo2ParallelSmall = `reference_no,type,citation
1,Chinese,"Another version"
`
	refNo2TransSmall = `reference_no,type,citation
1,Full,"Legge 1898"
`
)

// Test the GetParallelRefs function
func TestGetParallelRefs(t *testing.T) {
	type test struct {
		name string
		ref2File string
		refNo2Parallel string
		refNo2Trans string
		fileName string
		wantParallelRefLen int
		wantParallelLangFirst string
		wantParallelRefFirst string
  }

  // Create tests
  tests := []test{
		{
			name: "Happy path",
			ref2File: ref2FileSmall,
			refNo2Parallel: refNo2ParallelSmall,
			refNo2Trans: refNo2TransSmall,
			fileName: "example_collection.tsv",
			wantParallelRefLen: 1,
			wantParallelLangFirst: "Chinese",
			wantParallelRefFirst: "Another version",
		},
  }

  // Run tests
  for _, tc := range tests {
  	ref2FileReader := strings.NewReader(tc.ref2File)
  	refNo2ParallelReader := strings.NewReader(tc.refNo2Parallel)
  	refNo2TransReader := strings.NewReader(tc.refNo2Trans)
  	client, err := LoadBibNotes(ref2FileReader, refNo2ParallelReader, refNo2TransReader)
  	if err != nil {
  		t.Fatalf("TestGetParallelRefs error loading bibnotes: %v", err)
  	}
		got := client.GetParallelRefs(tc.fileName)
		if len(got) !=  tc.wantParallelRefLen {
			t.Errorf("%s: got len %d, want %d", tc.name, len(got), tc.wantParallelRefLen)
			continue
		}
		gotLang := got[0].Lang
		if gotLang != tc.wantParallelLangFirst  {
			t.Errorf("%s: got parrelLangFirst %s, want %s", tc.name, gotLang, tc.wantParallelLangFirst)
		}
		gotRef := got[0].Ref
		if gotRef != tc.wantParallelRefFirst  {
			t.Errorf("%s: got parallelRefFirst %s, want %s", tc.name, gotRef, tc.wantParallelRefFirst)
		}
	}
}

// Test the GetTransRefs function
func TestGetTransRefs(t *testing.T) {
	type test struct {
		name string
		ref2File string
		refNo2Parallel string
		refNo2Trans string
		fileName string
		wantTransRefLen int
		wantTransTypeFirst string
		wantTransRefFirst string
  }

  // Create tests
  tests := []test{
		{
			name: "Happy path",
			ref2File: ref2FileSmall,
			refNo2Parallel: refNo2ParallelSmall,
			refNo2Trans: refNo2TransSmall,
			fileName: "example_collection.tsv",
			wantTransRefLen: 1,
			wantTransTypeFirst: "Full",
			wantTransRefFirst: "Legge 1898",
		},
  }

  // Run tests
  for _, tc := range tests {
  	ref2FileReader := strings.NewReader(tc.ref2File)
  	refNo2ParallelReader := strings.NewReader(tc.refNo2Trans)
  	refNo2TransReader := strings.NewReader(tc.refNo2Trans)
  	client, err := LoadBibNotes(ref2FileReader, refNo2ParallelReader, refNo2TransReader)
  	if err != nil {
  		t.Fatalf("TestGetTranslationRefs error loading bibnotes: %v", err)
  	}
		got := client.GetTransRefs(tc.fileName)
		if len(got) !=  tc.wantTransRefLen {
			t.Errorf("%s: got len %d, want %d", tc.name, len(got), tc.wantTransRefLen)
			continue
		}
		gotKind := got[0].Kind
		if gotKind != tc.wantTransTypeFirst  {
			t.Errorf("%s: got transTypeFirst %s, want %s", tc.name, gotKind, tc.wantTransTypeFirst)
		}
		gotRef := got[0].Ref
		if gotRef != tc.wantTransRefFirst  {
			t.Errorf("%s: got transRefFirst %s, want %s", tc.name, gotRef, tc.wantTransRefFirst)
		}
	}
}
