package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

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
	fileLifetimeInSeconds  = 3600
	nrOfFilesToRemember    = 24
)

// Logger represents the application logger.
type Logger struct {
	logger          *log.Logger
	file            *LogFileWriter
	logFiles        []*os.File
	roll            bool
	rollLock        sync.Mutex
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
		file:            &LogFileWriter{creationTime: time.Now()},
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
	defaultLoggerMutex.Lock()
	dl := defaultLogger
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
		el.Error("invalid log level")
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
	if el.roll {
		el.rollLock.Lock()
		el.rollFiles()
		el.rollLock.Unlock()
	}

	cl.WithFields(log.Fields{
		"extra": extra,
	}).Debug(message)
}

// Info is an alias for Logrus.Info, adding some default useful fields.
func (el *Logger) Info(message string, extra ...interface{}) {
	cl := el.defaultFields()
	if el.roll {
		el.rollLock.Lock()
		el.rollFiles()
		el.rollLock.Unlock()
	}

	cl.WithFields(log.Fields{
		"extra": extra,
	}).Info(message)
}

// Warn is an alias for Logrus.Warn, adding some default useful fields.
func (el *Logger) Warn(message string, extra ...interface{}) {
	cl := el.defaultFields()
	if el.roll {
		el.rollLock.Lock()
		el.rollFiles()
		el.rollLock.Unlock()
	}

	cl.WithFields(log.Fields{
		"extra": extra,
	}).Warn(message)
}

// Error is an alias for Logrus.Error, adding some default useful fields.
func (el *Logger) Error(message string, extra ...interface{}) {
	cl := el.defaultFields()
	if el.roll {
		el.rollLock.Lock()
		el.rollFiles()
		el.rollLock.Unlock()
	}

	cl.WithFields(log.Fields{
		"extra": extra,
	}).Error(message)
}

func (el *Logger) errorWithoutFileRoll(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Error(message)
}

// Panic is an alias for Logrus.Panic, adding some default useful fields.
func (el *Logger) Panic(message string, extra ...interface{}) {
	cl := el.defaultFields()
	if el.roll {
		el.rollLock.Lock()
		el.rollFiles()
		el.rollLock.Unlock()
	}

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
	if el.roll {
		el.rollLock.Lock()
		el.rollFiles()
		el.rollLock.Unlock()
	}

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

// WithFileRotation sets up the option to roll log files
func WithFileRotation(prefix string, subfolder string, fileExtension string) Option {
	return func(el *Logger) error {
		file, err := newFile(prefix, subfolder, fileExtension)
		if err != nil {
			return err
		}

		el.roll = true
		el.file.prefix = prefix
		el.logFiles = append(el.logFiles, file)
		el.file.SetWriter(file)

		return nil
	}
}

// WithStderrRedirect sets up the option to redirect stderr to file
func WithStderrRedirect() Option {
	return func(el *Logger) error {
		file, err := newFile(el.file.prefix+"_fatalErrors", "logs", "log")
		if err != nil {
			return err
		}

		err = redirectStderr(file)
		if err != nil {
			return err
		}

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

func (el *Logger) rollFiles() {
	duration := time.Now().Sub(el.file.creationTime).Seconds()
	if duration <= fileLifetimeInSeconds {
		return
	}

	file, err := newFile(el.file.prefix, "logs", "log")
	if err != nil {
		el.errorWithoutFileRoll(err.Error())
		return
	}
	el.file.creationTime = time.Now()
	el.file.SetWriter(file)

	err = el.logFiles[len(el.logFiles)-1].Close()
	if err != nil {
		el.errorWithoutFileRoll(err.Error())
	}

	if len(el.logFiles) == nrOfFilesToRemember {
		err = os.Remove(el.logFiles[0].Name())
		if err != nil {
			el.errorWithoutFileRoll(err.Error())
		}

		el.logFiles = el.logFiles[1:]
	}

	el.logFiles = append(el.logFiles, file)
}

func newFile(prefix string, subfolder string, fileExtension string) (*os.File, error) {
	absPath, err := filepath.Abs(subfolder)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(absPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	fileName := time.Now().Format("2006-02-01-15-04-05")
	if prefix != "" {
		fileName = prefix + "-" + fileName
	}

	return os.OpenFile(
		filepath.Join(absPath, fileName+"."+fileExtension),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0666)
}
