package logging

import (
	"io"
	"log"
	"net/url"
	"os"

	"github.com/chuxorg/chux-datastore/errors"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

var logger *log.Logger
var logLevel LogLevel

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags)
	logLevel = LogLevelInfo
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

func SetLogLevel(level LogLevel) {
	logLevel = level
}

func Debug(v ...interface{}) {
	if logLevel <= LogLevelDebug {
		logger.SetPrefix("[DEBUG] ")
		logger.Println(v...)
	}
}

func Info(v ...interface{}) {
	if logLevel <= LogLevelInfo {
		logger.SetPrefix("[INFO] ")
		logger.Println(v...)
	}
}

func Warning(v ...interface{}) {
	if logLevel <= LogLevelWarning {
		logger.SetPrefix("[WARNING] ")
		logger.Println(v...)
	}
}

func Error(v ...interface{}) {
	if logLevel <= LogLevelError {
		logger.SetPrefix("[ERROR] ")
		logger.Println(v...)
	}
}

func MaskUri(uri string) (string, error) {

	parsedURI, err := url.Parse(uri)
	if err != nil {
		return "", errors.NewChuxDataStoreError("failed to parse uri: %v", 1000, err)
	}

	if parsedURI.User != nil {
		username := parsedURI.User.Username()
		password, _ := parsedURI.User.Password()

		usernameMask := MaskString(username, '*')
		passwordMask := MaskString(password, '*')

		parsedURI.User = url.UserPassword(usernameMask, passwordMask)
	}

	maskedURI := parsedURI.String()
	return maskedURI, nil
}

func MaskString(s string, mask rune) string {
	masked := make([]rune, len(s))
	for i := range masked {
		masked[i] = mask
	}
	return string(masked)
}
