package logger

import (
	"elrond/elrond-go-sandbox/config"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

const skip = 2
const (
	LvlDebugString 		= "DEBUG"
	LvlInfoString 		= "INFO"
	LvlWarningString 	= "WARNING"
	LvlErrorString 		= "ERROR"
	LvlPanicString	 	= "PANIC"
)

type ElrondLogger struct {
	logger *log.Logger
}

var EL ElrondLogger
func init() {
	file, _ := currentLogFile()
	EL = *NewElrondLogger(file)
}

// NewElrondLogger will setup the defaults of the application logger.
// If the requested log file is writeble it will setup a MultiWriter on both the file
//  and the standard output. Also sets up the level and the format for the logger
func NewElrondLogger(file *os.File) *ElrondLogger {
	el := &ElrondLogger{
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

// SetLevel sets the log level according to this package's defined levels
func (el *ElrondLogger) SetLevel(level string) {
	switch level {
	case LvlDebugString:
		el.logger.SetLevel(log.DebugLevel)
	case LvlInfoString:
		el.logger.SetLevel(log.InfoLevel)
	case LvlWarningString:
		el.logger.SetLevel(log.WarnLevel)
	case LvlErrorString:
		el.logger.SetLevel(log.ErrorLevel)
	case LvlPanicString:
		el.logger.SetLevel(log.PanicLevel)
	default:
		el.logger.SetLevel(log.ErrorLevel)
	}
}

// SetOutput enables the posibility to change the output of the logger on demand
func (el *ElrondLogger) SetOutput(out io.Writer) {
	el.logger.SetOutput(out)
}

// Debug is an alias for Logrus.Debug, adding some default useful fields
func (el *ElrondLogger) Debug(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"level": LvlDebugString,
		"extra": extra,
	}).Debug(message)
}

// Info is an alias for Logrus.Info, adding some default useful fields
func (el *ElrondLogger) Info(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"level": LvlInfoString,
		"extra": extra,
	}).Info(message)
}

// Warn is an alias for Logrus.Warn, adding some default useful fields
func (el *ElrondLogger) Warn(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"level": LvlWarningString,
		"extra": extra,
	}).Warn(message)
}

// Error is an alias for Logrus.Error, adding some default useful fields
func (el *ElrondLogger) Error(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"level": LvlErrorString,
		"extra": extra,
	}).Error(message)
}

// Panic is an alias for Logrus.Panic, adding some default useful fields
func (el *ElrondLogger) Panic(message string, extra ...interface{}) {
	cl := el.defaultFields()
	cl.WithFields(log.Fields{
		"level": LvlPanicString,
		"extra": extra,
	}).Panic(message)
}

func (el *ElrondLogger) defaultFields() *log.Entry {
	_, file, line, ok := runtime.Caller(skip)
	return el.logger.WithFields(log.Fields{
		"file": file,
		"line_number": line,
		"caller_ok": ok,
	})
}

func currentLogFile() (*os.File, error) {
	os.MkdirAll(config.ElrondLoggerConfig.LogPath, os.ModePerm)
	return os.OpenFile(
		filepath.Join(config.ElrondLoggerConfig.LogPath, time.Now().Format("01-02-2006") + ".log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0666)
}