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
	"log"
	"testing"
	"cnweb/identity"
)

// Test package initialization, which requires a database connection
func TestSendPasswordReset(t *testing.T) {
	log.Printf("TestSendPasswordReset: Begin unit tests\n")
	userInfo := identity.UserInfo{
		UserID: 100,
		UserName: "test",
		Email: "alex@chinesenotes.com",
		FullName: "Alex Test",
		Role: "tester",
	}
	err := SendPasswordReset(userInfo)
	if err != nil {
		log.Println("TestSendPasswordReset: Error, ", err)
	}
}
