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

// WebAppConfig holds application configuration data that is specific to the web app
type WebAppConfig struct {

	// A map of project configuration variables
	ConfigVars map[string]string
}

// InitWeb loads the WebAppConfig data. If an error occurs, default values are used
func InitWeb(configFile io.Reader) WebAppConfig {
	log.Println("webconfig.init Initializing webconfig")
	c := WebAppConfig{}
	configVarsPtr, err := readWebConfig(configFile)
	if err != nil {
		log.Printf("webconfig.init error initializing webconfig: %v", err)
		c.ConfigVars = make(map[string]string)
	} else {
		c.ConfigVars = *configVarsPtr
	}
	return c
}

// GetAll gets all configuration variables
func (c WebAppConfig) GetAll() map[string]string {
	return c.ConfigVars
}

// GetFromEmail gets the environment or config variable for sending email from
func (c WebAppConfig) GetFromEmail() string {
	fromEmail := os.Getenv("FROM_EMAIL")
	if len(fromEmail) == 0 {
		fromEmail = c.GetVar("FromEmail")
	}
	return fromEmail
}

// GetPasswordResetURL gets the password reset URL for inclusion in email
func (c WebAppConfig) GetPasswordResetURL() string {
	passwordResetURL := os.Getenv("PASSWORD_RESET_URL")
	if len(passwordResetURL) == 0 {
		passwordResetURL = c.GetVar("PasswordResetURL")
	}
	return passwordResetURL
}

// NotesExtractorPattern gets regular expression for extracting multilingual equivalents in the notes
func (c WebAppConfig) NotesExtractorPattern() string {
	val, ok := c.ConfigVars["NotesExtractorPattern"]
	if !ok {
		log.Println("WebAppConfig.GetVar: could not find NotesExtractorPattern")
		val = ""
	}
	return val
}

// AddDirectoryToCol gets whether to add a directory prefix to collection names in full text search
func (c WebAppConfig) AddDirectoryToCol() bool {
	val, ok := c.ConfigVars["AddDirectoryToCol"]
	if ok && strings.TrimSpace(val) == "True" {
		return true
	}
	return false
}

// GetVar gets a configuration variable value, default empty string
func (c WebAppConfig) GetVar(key string) string {
	val, ok := c.ConfigVars[key]
	if !ok {
		log.Printf("WebAppConfig.GetVar: could not find key: %s", key)
		val = ""
	}
	return val
}

// GetVarWithDefault gets a configuration value with given default
func (c WebAppConfig) GetVarWithDefault(key, defaultVal string) string {
	val, ok := c.ConfigVars[key]
	if !ok {
		return defaultVal
	}
	return val
}

// GetCnReaderHome gets the home directory of the Chinese Notes project
func GetCnReaderHome() string {
	cnReaderHome := os.Getenv("CNREADER_HOME")
	if len(cnReaderHome) == 0 {
		cnReaderHome = "."
	}
	log.Printf("CNREADER_HOME set to %s", cnReaderHome)
	return cnReaderHome
}

// GetCnWebHome gets the home directory of the web application
func GetCnWebHome() string {
	cnWebHome := os.Getenv("CNWEB_HOME")
	if len(cnWebHome) == 0 {
		log.Println("CNWEB_HOME is not defined")
	}
	return cnWebHome
}

// GetEnvIntValue gets a value from the environment
func GetEnvIntValue(key string, defValue int) int {
	if val, ok := os.LookupEnv(key); ok {
		value, err := strconv.Atoi(val)
		if err != nil {
			return defValue
		}
		return value
	}
	return defValue
}

// GetPort get environment variable for serving port
func GetPort() int {
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8080"
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		port = 8080
	}
	return port
}

// Get the domain name of the site
func GetSiteDomain() string {
	domain := "localhost"
	site_domain := os.Getenv("SITEDOMAIN")
	if len(site_domain) != 0 {
		domain = site_domain
	}
	return domain
}

// Reads the configuration file with project variables
func readWebConfig(r io.Reader) (*map[string]string, error) {
	reader := bufio.NewReader(r)
	vars := make(map[string]string)
	eof := false
	for !eof {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			return nil, fmt.Errorf("readWebConfig: error reading file: %v", err)
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
	return &vars, nil
}

// PasswordProtected gets whether the web site is password projected.
func PasswordProtected() bool {
	protected := os.Getenv("PROTECTED")
	if len(protected) > 0 {
		return strings.ToLower(protected) == "true"
	}
	return false
}
