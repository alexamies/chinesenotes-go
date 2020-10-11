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


// Unit tests for lookup package
package fileloader

import (
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
)

// With no files
func TestLoadNoDictFile(t *testing.T) {
	t.Log("TestLoadNoDictFile: Begin unit tests")
	appConfig := config.AppConfig{
		LUFileNames: []string{},
	}
	dict, err := LoadDictFile(appConfig)
	if err != nil {
		t.Fatalf("TestLoadNoDictFile: Got error %v", err)
	}
	if len(dict) != 0 {
		t.Error("TestLoadNoDictFile: len(dict) != 0")
	}
}
