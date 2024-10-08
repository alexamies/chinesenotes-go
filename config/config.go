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
// Package for configuration of command line tool and web apps.
package config

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// AppConfig holds application configuration data that is general to the API
//
// These variables are common to API and web app usage, especially loading
// dictionary files.
type AppConfig struct {

	// The top level directory for the project
	ProjectHome string

	// A map of project configuration variables
	ConfigVars map[string]string

	// A list of files to read the lexical units in the dictionary from
	LUFileNames []string
}

// InitConfig sets application configuration data
func InitConfig() AppConfig {
	projectHome := "."
	cnReaderHome := os.Getenv("CNREADER_HOME")
	if len(cnReaderHome) != 0 {
		projectHome = cnReaderHome
	}
	log.Printf("InitConfig projectHome: %s\n", projectHome)
	c := AppConfig{
		ProjectHome: projectHome,
	}
	log.Printf("config.init projectHome = %s\n", projectHome)
	var err error
	configVars, err := readConfig(projectHome)
	if err != nil {
		log.Printf("InitConfig, error reading config, using defaults: %v\n", err)
		configVars = make(map[string]string)
	}
	c.ConfigVars = configVars
	c.LUFileNames = readLUFileNames(configVars, c.DictionaryDir())
	return c
}

// AvoidSubDomains get the subdomains to avoid whne loading the dictionary

// Default: empty
func (c AppConfig) AvoidSubDomains() map[string]bool {
	avoidSub := make(map[string]bool)
	if val, ok := c.ConfigVars["AvoidSubDomains"]; ok {
		values := strings.Split(",", val)
		for _, value := range values {
			log.Printf("config.AvoidSubDomains: value: %s", value)
			avoidSub[value] = true
		}
		return avoidSub
	}
	log.Print("config.AvoidSubDomains: no values")
	return avoidSub
}

// CorpusDataDir returns the directory where the corpus metadata is stored
func (c AppConfig) CorpusDataDir() string {
	return c.ProjectHome + "/data/corpus"
}

// CorpusDir gets the directory where the raw corpus text files are read from
func (c AppConfig) CorpusDir() string {
	return c.ProjectHome + "/corpus"
}

// DictionaryDir gets the name of the directory containing the dictionary files
func (c AppConfig) DictionaryDir() string {
	val, ok := c.ConfigVars["DictionaryDir"]
	if ok {
		return c.ProjectHome + "/" + val
	}
	return c.ProjectHome + "/data"
}

// IndexDir gets the name of the directory containing the dictionary files
func (c AppConfig) IndexDir() string {
	return c.ProjectHome + "/index"
}

// IndexCorpus gets the name of the corpus to test the term frequency index in Firestore.
func (c AppConfig) IndexCorpus() (string, bool) {
	val, ok := c.ConfigVars["IndexCorpus"]
	return val, ok
}

// IndexGen gets the generation number for the term frequency index in Firestore.
func (c AppConfig) IndexGen() int {
	val, ok := c.ConfigVars["IndexGen"]
	if !ok {
		val = "0"
		log.Printf("config.IndexGen: no value found for IndexGen using %s", val)
	}
	gen, err := strconv.Atoi(val)
	if err != nil {
		gen = 0
		log.Printf("config.IndexGen: bad value %s found for IndexGen using %d", val, gen)
	}
	return gen
}

// GetVar gets a configuration variable value
func (c AppConfig) GetVar(key string) string {
	val, ok := c.ConfigVars[key]
	if !ok {
		// log.Printf("config.GetVar: could not find key: '%s'\n", key)
		val = ""
	}
	return val
}

// readLUFileNames gets the name of the text files with lexical units (word senses)
func readLUFileNames(configVars map[string]string, dictionaryDir string) []string {
	fileNames := []string{}
	val, ok := configVars["LUFiles"]
	if ok {
		tokens := strings.Split(val, ",")
		fileNames = []string{}
		for _, token := range tokens {
			fileNames = append(fileNames, dictionaryDir+"/"+token)
		}
	}
	return fileNames
}

// readConfig reads the configuration file with project variables
func readConfig(projectHome string) (map[string]string, error) {
	vars := make(map[string]string)
	sep := "/"
	if strings.HasSuffix(projectHome, "/") {
		sep = ""
	}
	fileName := projectHome + sep + "config.yaml"
	configFile, err := os.Open(fileName)
	if err != nil {
		projectHome = "."
		log.Printf("config.readConfig: setting projectHome to: '%s'\n",
			projectHome)
		fileName = projectHome + "/config.yaml"
		configFile, err = os.Open(fileName)
		if err != nil {
			err := fmt.Errorf("error opening config.yaml: %v", err)
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
			err := fmt.Errorf("error reading config file: %v", err)
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
