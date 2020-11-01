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

package config

import (
	"testing"
)

// Test default serving port
func TestGetPort(t *testing.T) {
	port := GetPort()
	if port != 8080 {
		t.Error("TestGetPort: port = ", port)
	}
}

// Test get app home directory
func TestGetCnWebHome(t *testing.T) {
	cnwebHome := GetCnWebHome()
	if len(cnwebHome) == 0 {
		t.Error("TestGetCnWebHome: cnwebHome is empty")
	}
}

// TestGetVarWithDefault tests the GetVarWithDefault function
func TestGetVarWithDefault(t *testing.T) {
	const expect = "My Title"
	c := WebAppConfig{}
	val := c.GetVarWithDefault("TITLE", expect)
	if expect != val {
		t.Errorf("TestGetVarWithDefault: expect %s vs got %s", expect, val)
	}
}