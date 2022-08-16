package dictionary

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
)

type mockFsClient struct {
}

func (m mockFsClient) Collection(path string) *firestore.CollectionRef {
	return nil
}

func TestFSLookupSubstr(t *testing.T) {
	type test struct {
		name      string
		query     string
		wantError bool
	}
	tests := []test{
		{
			name:    		"Expect error",
			query:   		"柳暗花明",
			wantError:  true,
		},
	}
	for _, tc := range tests {
		ctx := context.Background()
		wdict := mockSmallDict()
		dict := NewDictionary(wdict)
		client := mockFsClient{}
		ssIndex := substringIndexFS{
			client: client,
			corpus: "",
			generation: 0,
			dict: dict,
		}
		_, err := ssIndex.LookupSubstr(ctx, tc.query, "Idiom", "")
		if err != nil {
			if !tc.wantError {
				t.Errorf("%s: unexpected error finding by substring: %v", tc.name, err)
			}
		}
		if tc.wantError && err == nil {
			t.Errorf("%s: expected error but got none", tc.name)
		}
	}
}
