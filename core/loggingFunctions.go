package core

import (
	"bytes"
	"runtime"
	"runtime/pprof"
)

// DumpGoRoutinesToLog will print the currently running go routines in the log
func DumpGoRoutinesToLog(goRoutinesNumberStart int) {
	buffer := GetRunningGoRoutines()
	log.Debug("go routines number",
		"start", goRoutinesNumberStart,
		"end", runtime.NumGoroutine())

	log.Debug(buffer.String())
}

func GetRunningGoRoutines() *bytes.Buffer {
	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		log.Error("could not dump goroutines", "error", err)
	}
	return buffer
}
