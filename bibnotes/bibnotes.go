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

// Package for 
package bibnotes

import (
  "encoding/csv"
  "fmt"
	"io"
	"log"
)

// Load the bibliographic notes database
type BibNotesClient interface {

	// Get references for English translations for the given collection file name
	GetTransRefs(fileName string) []string
}

type bibNotesClient struct {
	file2Ref map[string]string
	refNo2Trans map[string][]string
}

// Load the bibliographic notes database
func LoadBibNotes(file2RefReader, refNo2TransReader io.Reader) (BibNotesClient, error) {
    file2Ref, err := loadFile2Ref(file2RefReader)
    if err != nil {
        return nil, fmt.Errorf("error reading bib notes, %v", err)
    }
    refNo2Trans, err := loadTransRef(refNo2TransReader)
    if err != nil {
        return nil, fmt.Errorf("error reading bib notes, %v", err)
    }
		return bibNotesClient{
			file2Ref: *file2Ref,
			refNo2Trans: *refNo2Trans,
		}, nil
}

// Load the filename to reference number data
func loadFile2Ref(f io.Reader) (*map[string]string, error) {
    r := csv.NewReader(f)
    r.Comma = ','
    rows, err := r.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("error reading file 2 ref num: , %v", err)
    }
    file2Ref := make(map[string]string)
    for i, row := range rows {
    	if len(row) < 2 {
    		log.Printf("loadFile2Ref: row %d, expected 2 elements but got %d", i,
    				len(row))
    		continue
    	}
    	file2Ref[row[1]] = row[0]
    }
		return &file2Ref, nil
}

// Load the English translation publication references data
func loadTransRef(f io.Reader) (*map[string][]string, error) {
    r := csv.NewReader(f)
    r.Comma = ','
    rows, err := r.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("error reading Eng trans ref: , %v", err)
    }
    refNo2Trans := make(map[string][]string)
    for i, row := range rows {
    	if len(row) < 3 {
    		log.Printf("loadTransRef: row %d, expected 3 elements but got %d", i,
    				len(row))
    		continue
    	}
    	key := row[0]
    	refs, ok := refNo2Trans[key]
    	if ok {
    		refs = append(refs, row[2])
    	} else {
    		refs = []string{row[2]}
    	}
    	refNo2Trans[key] = refs
    }
		return &refNo2Trans, nil
}

func (client bibNotesClient) GetTransRefs(fileName string) []string {
	refNo, ok := client.file2Ref[fileName]
	if !ok {
		return []string{}
	}
	transRefs, ok := client.refNo2Trans[refNo]
	if !ok {
		return []string{}
	}
	return transRefs
}