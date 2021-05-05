package testscommon

import (
	"bytes"
	"runtime"
	"runtime/pprof"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("testscommon")

// LogGoroutinesNumber will print the currently running go routines
func LogGoroutinesNumber(goRoutinesNumberStart int) {
	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		log.Error("could not dump goroutines")
	}
	log.Debug("go routines number",
		"start", goRoutinesNumberStart,
		"end", runtime.NumGoroutine())

	log.Debug(buffer.String())
}
