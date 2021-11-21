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
