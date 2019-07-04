package logger

import (
	"os"
	"time"
)

func (el *Logger) ErrorWithoutFileRoll(message string, extra ...interface{}) {
	el.errorWithoutFileRoll(message, extra)
}

func RedirectStderr(f *os.File) error {
	return redirectStderr(f)
}

func (el *Logger) RollFiles() {
	el.rollLock.Lock()
	el.rollFiles()
	el.rollLock.Unlock()
}

func (el *Logger) SetCreationTime(t time.Time) {
	el.file.creationTime = t
}

func (el *Logger) GetNrOfLogFiles() int {
	return len(el.logFiles)
}

func NrOfFilesToRemember() int {
	return nrOfFilesToRemember
}
