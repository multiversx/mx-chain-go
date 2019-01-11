package logger

import (
	"io"

	log "github.com/sirupsen/logrus"
)

// printerHook is a logrus hook that prints out in the console only the message
// from the logged line. It is used to easlily follow logged messages
// instead of trying to decypher through the full logged json
type printerHook struct {
	Writer    io.Writer
}

// Levels returns the array of levels for which the hook will be applicable
func (h *printerHook) Levels() []log.Level {
	return []log.Level{
		log.DebugLevel,
		log.InfoLevel,
		log.WarnLevel,
		log.ErrorLevel,
		log.FatalLevel,
		log.PanicLevel,
	}
}

// Fire represents the action triggered once a logging function will be called
func (h *printerHook) Fire(entry *log.Entry) (err error) {
	_, err = h.Writer.Write([]byte(entry.Message))
	return err
}