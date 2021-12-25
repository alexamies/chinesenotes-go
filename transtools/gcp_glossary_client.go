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

	translate "cloud.google.com/go/translate/apiv3"
	translatepb "google.golang.org/genproto/googleapis/cloud/translate/v3"
)

const (
	location   = "us-central1"
	sourceLang = "zh"
	targetLang = "en"
)

type glossaryApiClient struct {
	projectID, glossaryID string
}

// NewGoogleClient creates a Google Translation API that uses a glossary.
func NewGlossaryClient(projectID, glossaryID string) ApiClient {
	return glossaryApiClient{
		projectID:  projectID,
		glossaryID: glossaryID,
	}
}

func (client glossaryApiClient) Translate(sourceText string) (*string, error) {
	ctx := context.Background()
	cl, err := translate.NewTranslationClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewTranslationClient: %v", err)
	}
	defer cl.Close()

	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/%s", client.projectID, location),
		SourceLanguageCode: sourceLang,
		TargetLanguageCode: targetLang,
		MimeType:           "text/plain", // Mime types: "text/plain", "text/html"
		Contents:           []string{sourceText},
		GlossaryConfig: &translatepb.TranslateTextGlossaryConfig{
			Glossary: fmt.Sprintf("projects/%s/locations/%s/glossaries/%s", client.projectID, location, client.glossaryID),
		},
	}

	resp, err := cl.TranslateText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslateText: %v", err)
	}

	targetText := ""
	for _, translation := range resp.GetGlossaryTranslations() {
		targetText = translation.GetTranslatedText()
		// fmt.Printf("Translated text: %v\n", translation.GetTranslatedText())
	}

	return &targetText, nil
}
