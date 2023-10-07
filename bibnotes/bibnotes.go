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

	// Get references for parallel texts for the given collection file name
	GetParallelRefs(fileName string) []ParellelRef

	// Get references for English translations for the given collection file name
	GetTransRefs(fileName string) []TransRef
}

// TransRef holds information on references for English translations of texts
type TransRef struct {
	Kind string // full, partial, or parallel
	Ref  string // Harvard style citation, may have markup
	URL  string // May be a file name if type is parallel (bilingual)
}

// ParellelRef holds information on references for parallel versions of texts
// These may be Chinese-Chinese, Chinese-Sanskrit, etc)
type ParellelRef struct {
	Lang string // Parallel language
	Ref  string // Harvard style citation, may have markup
}

type bibNotesClient struct {
	file2Ref       map[string]string
	refNo2Parallel map[string][]ParellelRef
	refNo2Trans    map[string][]TransRef
}

// Load the bibliographic notes database
func LoadBibNotes(file2RefReader, refNo2ParallelReader, refNo2TransReader io.Reader) (BibNotesClient, error) {
	file2Ref, err := loadFile2Ref(file2RefReader)
	if err != nil {
		return nil, fmt.Errorf("error reading bib notes ref no's, %v", err)
	}
	refNo2Parallel, err := loadParallelRef(refNo2ParallelReader)
	if err != nil {
		return nil, fmt.Errorf("error reading bib notes parallels, %v", err)
	}
	refNo2Trans, err := loadTransRef(refNo2TransReader)
	if err != nil {
		return nil, fmt.Errorf("error reading bib notes translations, %v", err)
	}
	return bibNotesClient{
		file2Ref:       *file2Ref,
		refNo2Parallel: *refNo2Parallel,
		refNo2Trans:    *refNo2Trans,
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
			log.Printf("loadFile2Ref: row %d, expected 2 elements but got %d", i, len(row))
			continue
		}
		file2Ref[row[1]] = row[0]
	}
	log.Printf("loadFile2Ref: loaded %d, rows", len(file2Ref))
	return &file2Ref, nil
}

// Load the parallel publication references
func loadParallelRef(f io.Reader) (*map[string][]ParellelRef, error) {
	r := csv.NewReader(f)
	r.Comma = ','
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading parallel ref: , %v", err)
	}
	refNo2Parallel := make(map[string][]ParellelRef)
	for i, row := range rows {
		if len(row) < 3 {
			log.Printf("loadParallelRef: row %d, expected 3 elements but got %d, row: %v", i, len(row), row)
			continue
		}
		key := row[0]
		ref := ParellelRef{
			Lang: row[1],
			Ref:  row[2],
		}
		refs, ok := refNo2Parallel[key]
		if ok {
			refs = append(refs, ref)
		} else {
			refs = []ParellelRef{ref}
		}
		refNo2Parallel[key] = refs
	}
	log.Printf("loadParallelRef: loaded %d, rows", len(refNo2Parallel))
	return &refNo2Parallel, nil
}

// Load the English translation publication references
func loadTransRef(f io.Reader) (*map[string][]TransRef, error) {
	r := csv.NewReader(f)
	r.Comma = ','
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading Eng trans ref: %v", err)
	}
	refNo2Trans := make(map[string][]TransRef)
	for i, row := range rows {
		if len(row) < 3 {
			log.Printf("loadTransRef: row %d, expected 3 elements but got %d", i, len(row))
			continue
		}
		key := row[0]
		ref := TransRef{
			Kind: row[1],
			Ref:  row[2],
		}
		if len(row) > 3 {
			ref.URL = row[3]
		}
		refs, ok := refNo2Trans[key]
		if ok {
			refs = append(refs, ref)
		} else {
			refs = []TransRef{ref}
		}
		refNo2Trans[key] = refs
	}
	log.Printf("loadTransRef: loaded %d rows", len(refNo2Trans))
	return &refNo2Trans, nil
}

func (client bibNotesClient) GetParallelRefs(fileName string) []ParellelRef {
	refNo, ok := client.file2Ref[fileName]
	if !ok {
		return []ParellelRef{}
	}
	transRefs, ok := client.refNo2Parallel[refNo]
	if !ok {
		return []ParellelRef{}
	}
	return transRefs
}

func (client bibNotesClient) GetTransRefs(fileName string) []TransRef {
	refNo, ok := client.file2Ref[fileName]
	if !ok {
		log.Printf("GetTransRefs: no value for fileName = %s", fileName)
		return []TransRef{}
	}
	transRefs, ok := client.refNo2Trans[refNo]
	if !ok {
		log.Printf("GetTransRefs: no value for refNo = %s", refNo)
		return []TransRef{}
	}
	return transRefs
}
