package goroutine

import (
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/debug"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("goroutine")

const (
	newRoutine = "new"
	oldRoutine = "old"
)

type goRoutinesAnalyser struct {
	goRoutinesProcessor debug.GoRoutinesProcessor
	goRoutinesData      map[string]debug.GoRoutineHandlerMap
}

// NewGoRoutinesAnalyser creates a new structure for go routine statistics analysis
func NewGoRoutinesAnalyser(processor debug.GoRoutinesProcessor) (*goRoutinesAnalyser, error) {
	if check.IfNil(processor) {
		return nil, core.ErrNilGoRoutineProcessor
	}
	emptyData := make(map[string]debug.GoRoutineHandlerMap)
	emptyData[newRoutine] = make(debug.GoRoutineHandlerMap)
	emptyData[oldRoutine] = make(debug.GoRoutineHandlerMap)
	return &goRoutinesAnalyser{
		goRoutinesProcessor: processor,
		goRoutinesData:      emptyData,
	}, nil
}

// DumpGoRoutinesToLogWithTypes will print the currently running go routines stats in the log
func (grd *goRoutinesAnalyser) DumpGoRoutinesToLogWithTypes() {
	buffer := core.GetRunningGoRoutines(log)
	log.Debug("GoRoutinesAnalyser - DumpGoRoutinesToLogWithTypes", "goroutines number", runtime.NumGoroutine())

	newData := grd.goRoutinesProcessor.ProcessGoRoutineBuffer(grd.goRoutinesData, buffer)

	dumpGoRoutinesDataToLog(newData)
	grd.goRoutinesData = newData
}

func dumpGoRoutinesDataToLog(latestData map[string]debug.GoRoutineHandlerMap) {
	oldRoutines := latestData[oldRoutine]
	goRoutines := make([]debug.GoRoutineHandler, 0)

	for _, val := range oldRoutines {
		goRoutines = append(goRoutines, val)
	}

	sort.Slice(goRoutines, func(i, j int) bool {
		if goRoutines[i].FirstOccurrence().Equal(goRoutines[j].FirstOccurrence()) {
			return goRoutines[i].ID() < goRoutines[j].ID()
		}
		return goRoutines[i].FirstOccurrence().Before(goRoutines[j].FirstOccurrence())
	})

	currentTime := time.Now()
	for _, gr := range goRoutines {
		runningTime := currentTime.Sub(gr.FirstOccurrence())
		if runningTime > time.Hour {
			log.Debug("remaining goroutine - more than an hour", "ID", gr.ID(), "running seconds", runningTime)
			continue
		}
		log.Debug("remaining goroutine", "ID", gr.ID(), "running seconds", runningTime)
	}

	for _, val := range latestData[newRoutine] {
		log.Debug("new goroutine", "ID", val.ID(), "\nData", val.StackTrace())
	}
}

func getGoroutineId(goroutineString string) string {
	rg := regexp.MustCompile(`goroutine \d+`)

	matches := rg.FindAllString(goroutineString, 1)
	for _, element := range matches {
		replaced := strings.ReplaceAll(element, "goroutine ", "")
		return replaced
	}

	return ""
}
