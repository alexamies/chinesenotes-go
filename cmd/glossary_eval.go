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
    maxDiff = 0.25
    maxLen = 1.15
    minLen = 0.9
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
    testNo, source, target, reason string
    pass bool
}

// Compare similarity of model translation and output from translation engine
func compareSimilarity(model, targetText string) (bool, string) {
    log.Printf("Comparing model '%s' with target '%s'", model, targetText)
    mWords := strings.Split(model, " ")
    tWords := strings.Split(targetText, " ")
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
        term := strings.ToLower(strings.ReplaceAll(w, ".", ""))
        if len(term) > 1 {
            mWSet[term] = true
        }
    }
    tWSet := make(map[string]bool)
    for _, w := range tWords {
        term := strings.ToLower(strings.ReplaceAll(w, ".", ""))
        if len(term) > 1 {
            tWSet[term] = true
        }
    }
    // Find number of words that are in model but not in target
    numDiff := 0
    for w := range mWSet {
        if _, ok := tWSet[w]; !ok {
            numDiff++
        }
    }
    if numDiff > int(float64(len(mWSet)) * maxDiff) {
        reason := fmt.Sprintf("%d different words, more than %.0f %%.",
            numDiff, maxDiff * 100.0)
        return false, reason
    }
    return true, ""
}

// Initializes translation API client with glossary
func initTranslationClient(glossaryName string) (transtools.ApiClient, error) {
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
    glossaryApiClient, err := initTranslationClient(glossaryName)
    if err != nil {
        return nil, err
    }
    results := []testResult{}
    for _, tc := range *testSuite {
        trText, err := glossaryApiClient.Translate(tc.source)
        if err != nil {
            return nil, err
        }
        pass, reason := compareSimilarity(tc.model,*trText)
        tr := testResult{
            testNo: tc.testNo,
            source: tc.source,
            target: *trText,
            pass: pass,
            reason: reason,
        }
        for _, w := range tc.mustInclude {
            if !strings.Contains(*trText, w) {
                tr.pass = false
                if len(tr.reason) > 0 {
                    tr.reason += ", "
                }
                tr.reason += fmt.Sprintf("missing '%s'", w)
            }
        }
        results = append(results, tr)
    }
    return &results, nil
}

func writeResults(w io.Writer, results *[]testResult) {
    log.Printf("Writing test results")
    io.WriteString(w,
        "Test no., Source text, Translated text, Pass, Failure reason\n")
    passing := 0
    for _, tr := range *results {
        if tr.pass {
            passing++
        }
        line := fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%t\",\"%s\"\n", tr.testNo,
            tr.source, tr.target, tr.pass, tr.reason)
        io.WriteString(w, line)
    }
    summary := fmt.Sprintf("%d tests run, %d passing", len(*results), passing)
    io.WriteString(w, summary)
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
        log.Fatalf("Error opening %s: %v", testFile, err)
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
        log.Fatalf("Error creating %s: %v", testFile, err)
    }
    defer w.Close()
    writeResults(w, results)
}
