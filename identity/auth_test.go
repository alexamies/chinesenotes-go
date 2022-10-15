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

// Unit tests for the identity package
package identity

import (
	"testing"

	"cloud.google.com/go/firestore"
)

type mockFsClient struct {
}

func (m mockFsClient) Collection(path string) *firestore.CollectionRef {
	if len(path) == 0 {
		return nil
	}
	return &firestore.CollectionRef{
		Path:  path,
		Query: firestore.Query{},
	}
}

func TestNewSessionId(t *testing.T) {
	sessionid := NewSessionId()
	if sessionid == "invalid" {
		t.Error("TestNewSessionId: ", sessionid)
	}
}
