package dictionary

import (
	"context"
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// TestAddWordSense2Map does a query expecting empty list
func TestAddWordSense2Map(t *testing.T) {
	wmap := map[string]dicttypes.Word{}
	ws := dicttypes.WordSense{
		Id: 1,
		HeadwordId: 1,
		Simplified: "我",
		Traditional: "",
		Pinyin: "wǒ",
		English: "me",
		Notes: "No notes",
	}
	addWordSense2Map(wmap, ws)
	if len(wmap) != 1 {
		t.Error("TestAddWordSense2Map: unexpected length, ", len(wmap))
	}
}

// TestLookupSubstr tests the SubstringIndexMem implementation of the interface
func TestLookupSubstr(t *testing.T) {
	t.Log("TestLookupSubstr: Begin unit tests")
	ctx := context.Background()
	dictSearcher, err := NewSubstringIndexMem(ctx)
	if err != nil {
		t.Fatalf("could not initialize SubstringIndexMem: %v", err)
	}
	type test struct {
		name string
		query string
		domain string
		expectErr bool
		expectNum int
  }
  tests := []test{
		{	name: "expect error",
			query: "",
			domain: "",
			expectErr: true,
		 	expectNum: 0,
		 },
		{	name: "expect empty",
			query: "我還不知道",
			domain: "",
			expectErr: false,
		 	expectNum: 0,
		 },
		{	name: "invalid domain",
			query: "置",
			domain: "invalid",
			expectErr: false,
		 	expectNum: 0,
		 },
  }
  for _, tc := range tests {
		results, err := dictSearcher.LookupSubstr(ctx, tc.query, "", "")
		if tc.expectErr && err == nil {
			t.Errorf("TestLookupSubstr: %s, expect an error, got none", tc.name)
			continue
		}
		if tc.expectErr {
			continue
		}
		if !tc.expectErr && err != nil {
			t.Errorf("TestLookupSubstr: %s, expect no error, got: %v", tc.name, err)
			continue
		}
		if results == nil {
			t.Errorf("TestLookupSubstr: %s, results nil", tc.name)
			continue
		}
		resNum := len(results.Words)
		if tc.expectNum != resNum {
			t.Errorf("TestLookupSubstr: %s, expected %d results, got: %v", tc.name,
					tc.expectNum, resNum)
		}
	}
}
