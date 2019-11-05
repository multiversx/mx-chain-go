package logger

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

const ellipsisCharacter = "\u2026"

var logMut = &sync.RWMutex{}
var loggers map[string]*logger
var defaultLogOut LogOutputHandler
var defaultLogLevel = LogInfo

func init() {
	logMut.Lock()
	defer logMut.Unlock()

	loggers = make(map[string]*logger)
	defaultLogOut = &logOutputSubject{}
	_ = defaultLogOut.AddObserver(os.Stdout, ConsoleFormatter{})
}

// Get returns a log based on the name provided, generating a new log if there is no log with provided name
func Get(name string) *logger {
	logMut.Lock()
	defer logMut.Unlock()

	logger, ok := loggers[name]
	if !ok {
		logger = newLogger(name, defaultLogLevel, defaultLogOut)
		loggers[name] = logger
	}

	return logger
}

// ConvertHash generates a short-hand of provided bytes slice showing only the first 3 and the last 3 bytes as hex
// in total, the resulting string is maximum 13 characters long
func ConvertHash(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	if len(hash) < 6 {
		return hex.EncodeToString(hash)
	}

	prefix := hex.EncodeToString(hash[:3])
	suffix := hex.EncodeToString(hash[len(hash)-3:])
	return prefix + ellipsisCharacter + suffix
}

// SetLogLevel changes the log level of the contained loggers. The expected format is "LOG_LEVEL|MATCHING_STRING".
// If matching string is * it will change the log levels of all contained loggers and will also set the
// defaultLogLevelProperty. Otherwise, the log level will be modified only on those loggers that will contain the
// matching string on any position.
// For example, having the parameter "DEBUG|process" will set the DEBUG level on all loggers that will contain
// the "process" string in their name ("process/sync", "process/interceptors", "process" and so on).
func SetLogLevel(logLevelAndPattern string) error {
	logLevel, pattern, err := parseLogLevelAndMatchingString(logLevelAndPattern)
	if err != nil {
		return err
	}

	logMut.Lock()
	defer logMut.Unlock()

	setLogLevelOnMap(loggers, pattern, logLevel)
	setLogLevelToVar(&defaultLogLevel, pattern, logLevel)

	return nil
}

// AddLogObserver adds a new observer (writer + formatter) to the already built-in log observers queue
// This method is useful when adding a new output device for logs is needed (such as files, streams, API routes and so on)
func AddLogObserver(w io.Writer, formatter Formatter) error {
	return defaultLogOut.AddObserver(w, formatter)
}

func setLogLevelOnMap(loggers map[string]*logger, pattern string, logLevel LogLevel) {
	for name, log := range loggers {
		isMatching := pattern == "*" || strings.Contains(name, pattern)
		if isMatching {
			log.SetLevel(logLevel)
		}
	}
}

func setLogLevelToVar(dest *LogLevel, pattern string, logLevel LogLevel) {
	if pattern == "*" {
		*dest = logLevel
	}
}

func parseLogLevelAndMatchingString(logLevelAndPattern string) (LogLevel, string, error) {
	input := strings.Split(logLevelAndPattern, "|")
	if len(input) != 2 {
		return LogTrace, "", ErrInvalidLogLevelPattern
	}

	logLevel, err := getLogLevel(input[0])

	return logLevel, input[1], err
}

func getLogLevel(logLevelAsString string) (LogLevel, error) {
	levels := []LogLevel{
		LogTrace,
		LogDebug,
		LogInfo,
		LogWarning,
		LogError,
	}

	providedLogLevelUpper := strings.ToUpper(logLevelAsString)
	for _, level := range levels {
		levelUpper := strings.ToUpper(level.String())
		if levelUpper == providedLogLevelUpper {
			return level, nil
		}
	}

	return LogTrace, errors.New(fmt.Sprintf("unknown log level provided '%s'", logLevelAsString))
}
