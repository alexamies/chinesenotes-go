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


// Unit tests for the mail package
package mail

import (
	"testing"

	"github.com/alexamies/chinesenotes-go/identity"
	"github.com/alexamies/chinesenotes-go/webconfig"
)

// Test package initialization, which requires a database connection
func TestSendPasswordResetExpectError(t *testing.T) {
	t.Log("TestSendPasswordResetExpectError: Begin unit test")
	userInfo := identity.UserInfo{
		UserID: 100,
		UserName: "test",
		Email: "alex@chinesenotes.com",
		FullName: "Alex Test",
		Role: "tester",
	}
	c := webconfig.WebAppConfig{}
	err := SendPasswordReset(userInfo, "", c)
	if err == nil {
		t.Fatal("TestSendPasswordResetExpectError: Expect error with no config")
	}
}
