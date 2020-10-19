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

// Password recovery
package mail

import (
	"errors"
	"fmt"
	"os"
	
	"github.com/alexamies/chinesenotes-go/applog"
	"github.com/alexamies/chinesenotes-go/identity"
	"github.com/alexamies/chinesenotes-go/webconfig"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendPasswordReset(toUser identity.UserInfo, token string, c webconfig.WebAppConfig) error {
	fromEmail := c.GetVar("FromEmail")
	from := mail.NewEmail("Do Not Reply", fromEmail)
	subject := "Password Reset"
	to := mail.NewEmail(toUser.FullName, toUser.Email)
	passwordResetURL := c.GetPasswordResetURL()
	if len(passwordResetURL) == 0 {
		return errors.New("SendPasswordReset: error, passwordResetURL is empty")
	}
	plainText := "To reset your password, please go to %s?token=%s . Your username is %s."
	plainTextContent := fmt.Sprintf(plainText, passwordResetURL, token, toUser.UserName)
	htmlText := "<p>To reset your password, please click <a href='%s?token=%s'>here</a>. Your username is %s.</p>"
	htmlContent := fmt.Sprintf(htmlText, passwordResetURL, token, toUser.UserName)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("SendPasswordReset: error, %v\n", err)
	} else {
		applog.Infof("SendPasswordReset: sent email code: %v, url: %s",
			response.StatusCode, passwordResetURL)
	}
	return nil
}