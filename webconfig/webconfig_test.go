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
package webconfig

import (
	"testing"
)

// Test package initialization, which requires a database connection
func TestInit(t *testing.T) {
	t.Log("TestInit: Begin unit tests")
}

// Test default serving port
func TestGetPort(t *testing.T) {
	port := GetPort()
	if port != 8080 {
		t.Error("TestGetPort: port = ", port)
	}
}

// Test site domain
func TestGetSiteDomain(t *testing.T) {
	domain := GetSiteDomain()
	if domain != "localhost" {
		t.Error("TestGetSiteDomain: domain = ", domain)
	}
}

// Test get app home directory
func TestGetCnWebHome(t *testing.T) {
	cnwebHome := GetCnWebHome()
	if cnwebHome == "" {
		t.Error("TestGetCnWebHome: cnwebHome is empty")
	}
}

// TestGetVarWithDefault tests the GetVarWithDefault function
func TestGetVarWithDefault(t *testing.T) {
	const expect = "My Title"
	val := GetVarWithDefault("TITLE", expect)
	if expect != val {
		t.Errorf("TestGetVarWithDefault: expect %s vs got %s", expect, val)
	}
}