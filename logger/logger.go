package logger

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	log "github.com/sirupsen/logrus"
)

// String representation of the package logging levels
const (
	LogDebug   = "DEBUG"
	LogInfo    = "INFO"
	LogWarning = "WARNING"
	LogError   = "ERROR"
	LogPanic   = "PANIC"
)

// Logger represents the application logger
type Logger struct {
	logger *log.Logger
}

// NewElrondLogger will setup the defaults of the application logger.
// If the requested log file is writeble it will setup a MultiWriter on both the file
//  and the standard output. Also sets up the level and the format for the logger
func NewElrondLogger(file *os.File) *Logger {
	el := &Logger{
		log.New(),
	}

	if file != nil {
		mw := io.MultiWriter(os.Stdout, file)
		el.logger.SetOutput(mw)
	} else {
		el.logger.SetOutput(os.Stdout)
	}

	el.logger.SetFormatter(&log.JSONFormatter{})
	el.logger.SetLevel(log.WarnLevel)
	return el
}

// NewDefaultLogger is a shorthand for instantiating a new logger with default settings.
// If it fails to open the default log file it will return a logger with os.Stdout output
func NewDefaultLogger() *Logger {
	file, err := DefaultLogFile()
	if err != nil {
		return NewElrondLogger(nil)
	}
	return NewElrondLogger(file)
}

// SetLevel sets the log level according to this package's defined levels
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

// SetOutput enables the posibility to change the output of the logger on demand
func (el *Logger) SetOutput(out io.Writer) {
	el.logger.SetOutput(out)
}

// Debug is an alias for Logrus.Debug, adding some default useful fields
func (el *Logger) Debug(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Debug(message)
}

// Info is an alias for Logrus.Info, adding some default useful fields
func (el *Logger) Info(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Info(message)
}

// Warn is an alias for Logrus.Warn, adding some default useful fields
func (el *Logger) Warn(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Warn(message)
}

// Error is an alias for Logrus.Error, adding some default useful fields
func (el *Logger) Error(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Error(message)
}

// Panic is an alias for Logrus.Panic, adding some default useful fields
func (el *Logger) Panic(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"extra": extra,
	}).Panic(message)
}

func (el *Logger) defaultFields() *log.Entry {
	_, file, line, ok := runtime.Caller(config.ElrondLoggerConfig.StackTraceDepth)
	return el.logger.WithFields(log.Fields{
		"file": file,
		"line_number": line,
		"caller_ok": ok,
	})
}

// DefaultLogFile returns the default output for the application logger
// The client package can always use another output and provide it in the logger constructor
func DefaultLogFile() (*os.File, error) {
	os.MkdirAll(config.ElrondLoggerConfig.LogPath, os.ModePerm)
	return os.OpenFile(
		filepath.Join(config.ElrondLoggerConfig.LogPath, time.Now().Format("2006-02-01") + ".log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0666)
}