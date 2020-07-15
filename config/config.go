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


//
// Package for command line tool configuration
//
package config

import (
	"bufio"
	"io"
	"fmt"
	"log"
	"os"
	"strings"
)

var projectHome, dictionaryDir string
var configVars map[string]string

func init() {
	projectHome = "."
	log.Println("config.init")
	var err error
	configVars, err = readConfig()
	if err != nil {
		log.Printf("config.init: error reading config: %v", err)
	}
}

// Subdomains to avoid whne loading the dictionary, default: empty
func AvoidSubDomains() map[string]bool {
  avoidSub := make(map[string]bool)
	if val, ok := configVars["AvoidSubDomains"]; ok {
		values := strings.Split(",", val)
		for _, value := range values {
			log.Printf("config.AvoidSubDomains: value: %s", value)
			avoidSub[value] = true
		}
	}
	return avoidSub
}

// Returns the directory where the corpus metadata is stored
func CorpusDataDir() string {
	return projectHome + "/data/corpus"
}

// Returns the directory where the raw corpus text files are read from
func CorpusDir() string {
	return projectHome + "/corpus"
}

// The name of the directory containing the dictionary files
func DictionaryDir() string {
	val, ok := configVars["DictionaryDir"]
	if ok {
		return projectHome + "/" + val
	}
	return projectHome + "/data"
}

// Gets a configuration variable value
func GetVar(key string) string {
	val, ok := configVars[key]
	if !ok {
		log.Printf("config.GetVar: could not find key: '%s'\n", key)
		val = ""
	}
	return val
}

// The name of the text files with lexical units (word senses)
func LUFileNames() []string {
	fileNames := []string{DictionaryDir() + "/words.txt"}
	val, ok := configVars["LUFiles"]
	if ok {
		tokens := strings.Split(val, ",")
		fileNames = []string{}
		for _, token := range tokens {
			fileNames = append(fileNames, DictionaryDir() + "/" + token)
		}
	}
	return fileNames
}

// Reads the configuration file with project variables
func readConfig() (map[string]string, error) {
	vars := make(map[string]string)
	fileName := projectHome + "/config.yaml"
	configFile, err := os.Open(fileName)
	if err != nil {
		projectHome = ".."
		log.Printf("config.readConfig: setting projectHome to: '%s'\n",
			projectHome)
		fileName = projectHome + "/config.yaml"
		configFile, err = os.Open(fileName)
		if err != nil {
			err := fmt.Errorf("error opening config.yaml: %v",err)
			return map[string]string{}, err
		}
	}
	defer configFile.Close()
	reader := bufio.NewReader(configFile)
	eof := false
	for !eof {
		var line string
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			err := fmt.Errorf("error reading config file ", err)
			return map[string]string{}, err
		}
		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, ":")
		if i > 0 {
			varName := line[:i]
			val := strings.TrimSpace(line[i+1:])
			vars[varName] = val
		}
	}
	return vars, nil
}
