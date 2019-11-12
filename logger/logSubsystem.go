package logger

import (
	"encoding/hex"
	"io"
	"os"
	"strings"
	"sync"
)

var logMut = &sync.RWMutex{}
var loggers map[string]*logger
var defaultLogOut LogOutputHandler
var defaultLogLevel = LogInfo
var logPattern = ""

var mutDisplayByteSlice = &sync.RWMutex{}
var displayByteSlice func(slice []byte) string

func init() {
	logMut.Lock()
	logPattern = "*:INFO"
	loggers = make(map[string]*logger)
	defaultLogOut = &logOutputSubject{}
	_ = defaultLogOut.AddObserver(os.Stdout, &ConsoleFormatter{})
	logMut.Unlock()

	mutDisplayByteSlice.Lock()
	displayByteSlice = func(slice []byte) string {
		return hex.EncodeToString(slice)
	}
	mutDisplayByteSlice.Unlock()
}

// GetOrCreate returns a log based on the name provided, generating a new log if there is no log with provided name
func GetOrCreate(name string) *logger {
	logMut.Lock()
	defer logMut.Unlock()

	logger, ok := loggers[name]
	if !ok {
		logger = newLogger(name, defaultLogLevel, defaultLogOut)
		loggers[name] = logger
	}

	return logger
}

// SetLogLevel changes the log level of the contained loggers. The expected format is
// "MATCHING_STRING1:LOG_LEVEL1,MATCHING_STRING2:LOG_LEVEL2".
// If matching string is *, it will change the log levels of all contained loggers and will also set the
// defaultLogLevelProperty. Otherwise, the log level will be modified only on those loggers that will contain the
// matching string on any position.
// For example, having the parameter "DEBUG|process" will set the DEBUG level on all loggers that will contain
// the "process" string in their name ("process/sync", "process/interceptors", "process" and so on).
// The rules are applied in the exact manner as they are provided, starting from left to the right part of the string
// Example: *:INFO,p2p:ERROR,*:DEBUG,data:INFO will result in having the data package logger(s) on INFO log level
// and all other packages on DEBUG level
func SetLogLevel(logLevelAndPattern string) error {
	logLevels, patterns, err := ParseLogLevelAndMatchingString(logLevelAndPattern)
	if err != nil {
		return err
	}

	logMut.Lock()
	setLogLevelOnMap(loggers, &defaultLogLevel, logLevels, patterns)
	logPattern = logLevelAndPattern
	logMut.Unlock()

	return nil
}

// GetLogLevelPattern returns the last set log level pattern.
// The format returned is MATCHING_STRING1:LOG_LEVEL1,MATCHING_STRING2:LOG_LEVEL2".
func GetLogLevelPattern() string {
	logMut.RLock()
	defer logMut.RUnlock()

	return logPattern
}

// AddLogObserver adds a new observer (writer + formatter) to the already built-in log observers queue
// This method is useful when adding a new output device for logs is needed (such as files, streams, API routes and so on)
func AddLogObserver(w io.Writer, formatter Formatter) error {
	return defaultLogOut.AddObserver(w, formatter)
}

// RemoveLogObserver removes an exiting observer by providing the writer pointer.
func RemoveLogObserver(w io.Writer) error {
	return defaultLogOut.RemoveObserver(w)
}

// ClearLogObservers clears the observers lists
func ClearLogObservers() {
	defaultLogOut.ClearObservers()
}

func setLogLevelOnMap(loggers map[string]*logger, dest *LogLevel, logLevels []LogLevel, patterns []string) {
	for i := 0; i < len(logLevels); i++ {
		pattern := patterns[i]
		logLevel := logLevels[i]
		for name, log := range loggers {
			isMatching := pattern == "*" || strings.Contains(name, pattern)
			if isMatching {
				log.SetLevel(logLevel)
			}
		}

		if pattern == "*" {
			*dest = logLevel
		}
	}
}

// ParseLogLevelAndMatchingString can parse a string in the form "MATCHING_STRING1:LOG_LEVEL1,MATCHING_STRING2:LOG_LEVEL2" into its
// corresponding log level and matching string. Errors if something goes wrong.
// For example, having the parameter "DEBUG|process" will set the DEBUG level on all loggers that will contain
// the "process" string in their name ("process/sync", "process/interceptors", "process" and so on).
// The rules are applied in the exact manner as they are provided, starting from left to the right part of the string
// Example: *:INFO,p2p:ERROR,*:DEBUG,data:INFO will result in having the data package logger(s) on INFO log level
// and all other packages on DEBUG level
func ParseLogLevelAndMatchingString(logLevelAndPatterns string) ([]LogLevel, []string, error) {
	splitLevelPatterns := strings.Split(logLevelAndPatterns, ",")

	levels := make([]LogLevel, len(splitLevelPatterns))
	patterns := make([]string, len(splitLevelPatterns))
	for i, levelPattern := range splitLevelPatterns {
		level, pattern, err := parseLevelPattern(levelPattern)
		if err != nil {
			return nil, nil, err
		}

		levels[i] = level
		patterns[i] = pattern
	}

	return levels, patterns, nil
}

func parseLevelPattern(logLevelAndPattern string) (LogLevel, string, error) {
	input := strings.Split(logLevelAndPattern, ":")
	if len(input) != 2 {
		return LogTrace, "", ErrInvalidLogLevelPattern
	}

	logLevel, err := GetLogLevel(input[1])

	return logLevel, input[0], err
}

// SetDisplayByteSlice sets the converter function from byte slice to string
// default, this will call hex.EncodeToString
func SetDisplayByteSlice(f func(slice []byte) string) error {
	if f == nil {
		return ErrNilDisplayByteSliceHandler
	}

	mutDisplayByteSlice.Lock()
	displayByteSlice = f
	mutDisplayByteSlice.Unlock()

	return nil
}
