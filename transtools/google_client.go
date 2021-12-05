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

package transtools

import (
	"context"
	"fmt"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

type googleApiClient struct {
}

// NewGoogleClient creates a Google Translation API that does not use a glossary.
func NewGoogleClient() ApiClient {
	return googleApiClient{}
}

func (client googleApiClient) Translate(sourceText string) (*string, error) {
	// text := "The Go Gopher is cute"
	ctx := context.Background()

	lang, err := language.Parse("en")
	if err != nil {
		return nil, fmt.Errorf("language.Parse: %v", err)
	}

	apiClient, err := translate.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer apiClient.Close()

	resp, err := apiClient.Translate(ctx, []string{sourceText}, lang, nil)
	if err != nil {
		return nil, fmt.Errorf("Translate: %v", err)
	}
	if len(resp) == 0 {
		return nil, fmt.Errorf("Translate returned empty response to text: %s", sourceText)
	}
	return &resp[0].Text, nil
}
