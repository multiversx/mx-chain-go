package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// These constants are the string representation of the package logging levels.
const (
	LogDebug   = "DEBUG"
	LogInfo    = "INFO"
	LogWarning = "WARNING"
	LogError   = "ERROR"
	LogPanic   = "PANIC"
)

const (
	defaultStackTraceDepth = 2
	maxHeadlineLength      = 100
)

// Logger represents the application logger.
type Logger struct {
	logger          *log.Logger
	file            *LogFileWriter
	stackTraceDepth int
}

// Option represents a functional configuration parameter that can operate
//  over the Logger struct
type Option func(*Logger) error

// NewElrondLogger will setup the defaults of the application logger.
// If the requested log file is writable it will setup a MultiWriter on both the file
//  and the standard output. Also sets up the level and the format for the logger.
func NewElrondLogger(opts ...Option) *Logger {
	el := &Logger{
		logger:          log.New(),
		stackTraceDepth: defaultStackTraceDepth,
		file:            &LogFileWriter{},
	}

	for _, opt := range opts {
		err := opt(el)
		if err != nil {
			el.logger.Error("Could not set opt for logger", err)
		}
	}

	el.logger.SetOutput(el.file)
	el.logger.AddHook(&printerHook{Writer: os.Stdout})
	el.logger.SetFormatter(&log.JSONFormatter{})
	el.logger.SetLevel(log.DebugLevel)

	return el
}

var defaultLogger *Logger
var defaultLoggerMutex = sync.RWMutex{}

// DefaultLogger is a shorthand for instantiating a new logger with default settings.
// If it fails to open the default log file it will return a logger with os.Stdout output.
func DefaultLogger() *Logger {
	defaultLoggerMutex.RLock()
	dl := defaultLogger
	defaultLoggerMutex.RUnlock()

	if dl != nil {
		return dl
	}

	defaultLoggerMutex.Lock()
	dl = defaultLogger
	if dl == nil {
		defaultLogger = NewElrondLogger()
		dl = defaultLogger
	}
	defaultLoggerMutex.Unlock()

	return dl
}

// ApplyOptions can set up different configurable options of a Logger instance
func (el *Logger) ApplyOptions(opts ...Option) error {
	for _, opt := range opts {
		err := opt(el)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
	}
	return nil
}

// File returns the current Logger file
func (el *Logger) File() io.Writer {
	return el.file.Writer()
}

// StackTraceDepth returns the current Logger stackTraceDepth
func (el *Logger) StackTraceDepth() int {
	return el.stackTraceDepth
}

// SetLevel sets the log level according to this package's defined levels.
func (el *Logger) SetLevel(level string) {
	switch level {
	case LogDebug:
		el.logger.SetLevel(log.DebugLevel)
	case LogInfo:
		el.logger.SetLevel(log.InfoLevel)
	case LogWarning:
		el.logger.SetLevel(log.WarnLevel)
	case LogError:
		el.logger.SetLevel(log.ErrorLevel)
	case LogPanic:
		el.logger.SetLevel(log.PanicLevel)
	default:
		el.logger.SetLevel(log.ErrorLevel)
	}
}

// SetOutput enables the possibility to change the output of the logger on demand.
func (el *Logger) SetOutput(out io.Writer) {
	el.logger.SetOutput(out)
}

// Debug is an alias for Logrus.Debug, adding some default useful fields.
func (el *Logger) Debug(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Debug(message)
}

// Info is an alias for Logrus.Info, adding some default useful fields.
func (el *Logger) Info(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Info(message)
}

// Warn is an alias for Logrus.Warn, adding some default useful fields.
func (el *Logger) Warn(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Warn(message)
}

// Error is an alias for Logrus.Error, adding some default useful fields.
func (el *Logger) Error(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Error(message)
}

// Panic is an alias for Logrus.Panic, adding some default useful fields.
func (el *Logger) Panic(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Panic(message)
}

// LogIfError will log if the provided error different than nil
func (el *Logger) LogIfError(err error) {
	if err == nil {
		return
	}
	cl := el.defaultFields()
	cl.Error(err.Error())
}

// Headline will build a headline message given a delimiter string
//  timestamp parameter will be printed before the repeating delimiter
func (el *Logger) Headline(message string, timestamp string, delimiter string) string {
	if len(delimiter) > 1 {
		delimiter = delimiter[:1]
	}
	if len(message) >= maxHeadlineLength {
		return message
	}
	delimiterLength := (maxHeadlineLength - len(message)) / 2
	delimiterText := strings.Repeat(delimiter, delimiterLength)
	return fmt.Sprintf("\n%s %s %s %s\n\n", timestamp, delimiterText, message, delimiterText)
}

func (el *Logger) defaultFields() *log.Entry {
	_, file, line, ok := runtime.Caller(el.stackTraceDepth)
	return el.logger.WithFields(log.Fields{
		"file":        file,
		"line_number": line,
		"caller_ok":   ok,
	})
}

// WithFile sets up the file option for the Logger
func WithFile(file io.Writer) Option {
	return func(el *Logger) error {
		el.file.SetWriter(file)
		return nil
	}
}

// WithStackTraceDepth sets up the stackTraceDepth option for the Logger
func WithStackTraceDepth(depth int) Option {
	return func(el *Logger) error {
		el.stackTraceDepth = depth
		return nil
	}
}
