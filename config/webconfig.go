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
func InitWeb() WebAppConfig {
	log.Println("webconfig.init Initializing webconfig")
	c := WebAppConfig{}
	cnwebHome := GetCnWebHome()
	fileName := fmt.Sprintf("%s/webconfig.yaml", cnwebHome)
	configFile, err := os.Open(fileName)
	if err != nil {
		path, er := os.Getwd()
		if er != nil {
			log.Printf("cannot find cwd: %v", er)
			path = ""
		}
		log.Printf("InitWeb error loading file '%s' (%s): %v", fileName, path, err)
		return c
	}
	defer configFile.Close()
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
	val = strings.Trim(val, "\"")
	return val
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

// DBConfig gets the configuration string to connect to the database
func DBConfig() string {
	instanceConnectionName := os.Getenv("INSTANCE_CONNECTION_NAME")
	dbUser := "app_user"
	user := os.Getenv("DBUSER")
	if user != "" {
		dbUser = user
	}
	dbpass := os.Getenv("DBPASSWORD")
	dbname := "corpus_index"
	d := os.Getenv("DATABASE")
	if d != "" {
		dbname = d
	}
	// Connection via Unix socket
	if len(instanceConnectionName) > 0 {
		socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
		if !isSet {
			socketDir = "/cloudsql"
		}
		return fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?parseTime=true", dbUser, dbpass,
			socketDir, instanceConnectionName, dbname)
	}
	// Connection via TCP
	dbhost := "localhost"
	host := os.Getenv("DBHOST")
	if host != "" {
		dbhost = host
	}
	dbport := "3306"
	port := os.Getenv("DBPORT")
	if port != "" {
		dbport = port
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbpass, dbhost,
		dbport, dbname)
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

// PasswordProtected gets whether the web site is password projected.
func UseDatabase() bool {
	database := os.Getenv("DATABASE")
	if len(database) > 0 {
		return true
	}
	return false
}
