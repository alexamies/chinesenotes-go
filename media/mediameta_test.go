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


// Unit tests for query parsing functions
package media

import (
	"context"
	"database/sql"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
)

func initDBCon() (*sql.DB, error) {
	conString := webconfig.DBConfig()
	return sql.Open("mysql", conString)
}

// Test trivial query
func TestFindMedia(t *testing.T) {
	database, err := initDBCon()
	if err != nil {
		t.Skipf("TestFindWordsByEnglish: cannot connect to database: %v", err)
	}
	ctx := context.Background()
	mediaSearcher := NewMediaSearcher(database, ctx)
	if !mediaSearcher.Initialized() {
		t.Skip("TestFindMedia: mediaSearcher cannot be initialized")
	}
	metadata, err := mediaSearcher.FindMedia("hello", ctx)
	if err != nil {
		t.Error("TestFindMedia: encountered error: ", err)
		return
	}
	t.Logf("TestFindMedia: metadata.ObjectId: %v", metadata.ObjectId)
}
