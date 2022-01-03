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

// Web application for Chinese-English dictionary lookup, translation memory,
// and finding documents in a corpus. Settings in for the app are controlled
// through the file config.yaml, located in the project home directory, which
// is found through the env variable CNREADER_HOME or the present working
// directory.
//
// Tests for evaluation of glossary with translation quality

package main

import (
    "encoding/csv"
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    "strings"

    "github.com/alexamies/chinesenotes-go/transtools"
)

const (
    deepLKeyName = "DEEPL_AUTH_KEY"
    maxDiff = 0.60
    maxLen = 1.6
    minLen = 0.85
    projectIDKey = "PROJECT_ID"
)

var (
    glossaryApiClient transtools.ApiClient
)

type testCase struct {
    testNo, description, sourceRef, source, model string
    mustInclude []string
}

type testResult struct {
    testNo, source string
    glossaryTarget string
    glossaryPass bool
    glossaryReason string
    deepLTarget string
    deepLPass bool
    deepLReason string
    googleTarget string
    googlePass bool
    googleReason string
}

// Compare similarity of model translation and output from translation engine
func compareSimilarity(model, targetText string, mustInclude []string) (bool, string) {
    modelLower := strings.ToLower(strings.ReplaceAll(model, ".", ""))
    mWords := strings.Split(modelLower, " ")
    targetLower := strings.ToLower(strings.ReplaceAll(targetText, ".", ""))
    targetLower = strings.ReplaceAll(targetLower, "&#", " ")
    targetLower = strings.ReplaceAll(targetLower, ";", " ")
    targetLower = strings.ReplaceAll(targetLower, ",", " ")
    tWords := strings.Split(targetLower, " ")
    if len(tWords) < int(float64(len(mWords)) * minLen) {
        reason := fmt.Sprintf("Target is shorter than %.0f %% in length than model",
            minLen * 100.0)
        return false, reason
    }
    if len(tWords) > int(float64(len(mWords)) * maxLen) {
        reason := fmt.Sprintf("Target (%d) is longer than %.0f %% the length of the model (%d)",
            len(tWords), (maxLen - 1.0) * 100.0, len(mWords))
        return false, reason
    }
    mWSet := make(map[string]bool)
    for _, w := range mWords {
        if len(w) > 1 {
            mWSet[w] = true
        }
    }
    tWSet := make(map[string]bool)
    for _, w := range tWords {
        if len(w) > 1 {
            tWSet[w] = true
        }
    }
    // Find number of words that are in model but not in target
    numDiff := 0
    for w := range mWSet {
        if _, ok := tWSet[w]; !ok {
            numDiff++
        }
    }
    var passing bool
    reason := ""
    if numDiff > int(float64(len(mWSet)) * maxDiff) {
        reason = fmt.Sprintf("%d different words, more than %.0f %%.",
            numDiff, maxDiff * 100.0)
    } else {
        passing = true
    }
    for _, w := range mustInclude {
        term := strings.ToLower(w)
        if !strings.Contains(targetLower, term) {
            passing = false
            if len(reason) > 0 {
                reason += ", "
            }
            reason += fmt.Sprintf("missing '%s'", w)
        }
    }
    return passing, reason
}

// Initializes DeepL translation API client
func initDeepLClient() (transtools.ApiClient, error) {
    deepLKey, ok := os.LookupEnv(deepLKeyName)
    if !ok {
        return nil, fmt.Errorf("%s not set\n", deepLKeyName)
    }
    return transtools.NewDeepLClient(deepLKey), nil
}

// Initializes Google translation API client
func initGoogleClient() (transtools.ApiClient, error) {
    return transtools.NewGoogleClient(), nil
}

// Initializes translation API client with glossary
func initGlossaryClient(glossaryName string) (transtools.ApiClient, error) {
    projectID, ok := os.LookupEnv(projectIDKey)
    if !ok {
        return nil, fmt.Errorf("%s not set\n", projectIDKey)
    }
    return transtools.NewGlossaryClient(projectID, glossaryName), nil
}

func loadTestSuite(f io.Reader) (*[]testCase, error) {
    testSuite := []testCase{}
    r := csv.NewReader(f)
    r.Comma = ','
    r.Comment = '#'
    rows, err := r.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("Error reading test suite data, %v", err)
    }
    for i, row := range rows {
        if len(row) < 6 {
            return nil, fmt.Errorf("Error reading row %d, got %d fields, %v",
                i, len(row), row)
        }
        mustInclude := strings.Split(row[5], ",")
        tc := testCase{
            testNo:         row[0],
            description:    row[1],
            sourceRef:      row[2],
            source:         row[3],
            model:          row[4],
            mustInclude:    mustInclude,
        }
        testSuite = append(testSuite, tc)
    }
    log.Printf("Loaded test suite data from with %d tests", len(testSuite))
    return &testSuite, nil
}

func runTestSuite(glossaryName string, testSuite *[]testCase) (*[]testResult, error) {
    glossaryApiClient, err := initGlossaryClient(glossaryName)
    if err != nil {
        return nil, err
    }
    deepLApiClient, err := initDeepLClient()
    if err != nil {
        return nil, err
    }
    googleApiClient, err := initGoogleClient()
    if err != nil {
        return nil, err
    }
    results := []testResult{}
    for _, tc := range *testSuite {
        glossaryTarget, err := glossaryApiClient.Translate(tc.source)
        if err != nil {
            return nil, err
        }
        glossaryPass, glossaryReason := compareSimilarity(tc.model,*glossaryTarget, tc.mustInclude)

        deepLTarget, err := deepLApiClient.Translate(tc.source)
        if err != nil {
            return nil, err
        }
        deepLPass, deepLReason := compareSimilarity(tc.model,*deepLTarget, tc.mustInclude)

        googleTarget, err := googleApiClient.Translate(tc.source)
        if err != nil {
            return nil, err
        }
        googlePass, googleReason := compareSimilarity(tc.model,*googleTarget, tc.mustInclude)

        tr := testResult{
            testNo: tc.testNo,
            source: tc.source,
            glossaryTarget: *glossaryTarget,
            glossaryPass: glossaryPass,
            glossaryReason: glossaryReason,
            deepLTarget: *deepLTarget,
            deepLPass: deepLPass,
            deepLReason: deepLReason,
            googleTarget: *googleTarget,
            googlePass: googlePass,
            googleReason: googleReason,
        }
        results = append(results, tr)
    }
    return &results, nil
}

func writeResults(w io.Writer, results *[]testResult) {
    io.WriteString(w,
        "Test no., Source text, " +
        "Glossary translated text,Glossary Pass,Glossary Failure reason," +
        "DeepL translated text,DeepL Pass, DeepL Failure reason" +
        "Google translated text,Google Pass,Google Failure reason\n")
    glossaryPass := 0
    deepLPass := 0
    googlePass := 0
    for _, tr := range *results {
        if tr.glossaryPass {
            glossaryPass++
        }
        if tr.deepLPass {
            deepLPass++
        }
        if tr.googlePass {
            googlePass++
        }
        line := fmt.Sprintf("\"%s\",\"%s\",\"%s\",%t,\"%s\",\"%s\",%t,\"%s\",\"%s\",%t,\"%s\"\n",
            tr.testNo, tr.source,
            tr.glossaryTarget, tr.glossaryPass, tr.glossaryReason,
            tr.deepLTarget, tr.deepLPass, tr.deepLReason,
            tr.googleTarget, tr.googlePass, tr.googleReason,
        )
        io.WriteString(w, line)
    }
    summary := fmt.Sprintf("Total test: %d. API with glossary passing: %d. DeepL passing: %d. Google passing: %d.",
        len(*results), glossaryPass, deepLPass, googlePass)
    io.WriteString(w, summary)
    log.Printf(summary)
}

func main() {
    var glossary = flag.String("glossary", "",
        "Name of Google Translate API custom glossary")
    var testFile = flag.String("test_file", "",
        "Input file containing test suite")
    var outFile = flag.String("out_file", "",
        "Output file containing test results")
    flag.Parse()
    log.Printf("Starting translation evaluation with glossary: %s, test file %s, output file %s",
        *glossary, *testFile, *outFile)
    r, err := os.Open(*testFile)
    if err != nil {
        log.Fatalf("Error opening %s: %v", *testFile, err)
    }
    defer r.Close()
    testSuite, err := loadTestSuite(r)
    if err != nil {
        log.Fatalf("Error loading test suite: %v", err)
    }
    results, err := runTestSuite(*glossary, testSuite)
    if err != nil {
        log.Fatalf("Error running tests %v", err)
    }
    w, err := os.Create(*outFile)
    if err != nil {
        log.Fatalf("Error creating %s: %v", *testFile, err)
    }
    defer w.Close()
    writeResults(w, results)
}
