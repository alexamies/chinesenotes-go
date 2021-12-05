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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

const deepLURL = "https://api-free.deepl.com/v2/translate"

type TrResult struct {
	Translations []MT
}

type MT struct {
	DetectedSourceLanguage, Text string
}

type deepLApiClient struct {
	authKey string
}

func NewDeepLClient(authKey string) ApiClient {
	return deepLApiClient{
		authKey: authKey,
	}
}

func (client deepLApiClient) Translate(sourceText string) (*string, error) {
	data := url.Values{
		"auth_key":    {"c57803f2-4eef-e046-c535-1512b67ed5ec:fx"},
		"text":        {sourceText},
		"target_lang": {"EN"},
	}

	resp, err := http.PostForm(deepLURL, data)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var trResult TrResult
	err = json.Unmarshal(body, &trResult)
	if err != nil {
		return nil, err
	}
	if len(trResult.Translations) < 1 {
		return nil, errors.New("no translation returned")
	}
	transText := trResult.Translations[0].Text
	return &transText, nil
}
