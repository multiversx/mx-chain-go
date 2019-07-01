package logger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func (el *Logger) ErrorWithoutFileRoll(message string, extra ...interface{}) {
	el.errorWithoutFileRoll(message, extra)
}

func TestRedirectStderr(t *testing.T) {
	file, _ := newFile("", "logs", "log")
	err := redirectStderr(file)
	assert.Nil(t, err)
}

func TestRedirectStderrWithNilFile(t *testing.T) {
	err := redirectStderr(nil)
	assert.NotNil(t, err)
}

func TestRollFiles(t *testing.T) {
	t.Parallel()
	log := DefaultLogger()
	mockTime := time.Date(2019, 1, 1, 1, 1, 1, 1, time.Local)

	err := log.ApplyOptions(WithFileRotation("", "logs", "log"))
	assert.Nil(t, err)

	for i := 0; i < nrOfFilesToRemember*2; i++ {
		log.file.creationTime = mockTime
		log.rollFiles()
	}

	assert.Equal(t, nrOfFilesToRemember, len(log.logFiles))
}
