package goroutine

import (
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

var log = logger.GetOrCreate("goroutine")

const (
	newRoutine = "new"
	oldRoutine = "old"
)

type GoRoutinesAnalyser struct {
	goRoutinesProcessor core.GoRoutinesProcessor
	goRoutinesData      map[string]core.GoRoutineHandlerMap
}

// NewGoRoutinesAnalyser creates a new structure for go goroutine statistics analysis
func NewGoRoutinesAnalyser(processor core.GoRoutinesProcessor) (*GoRoutinesAnalyser, error) {
	if check.IfNil(processor) {
		return nil, core.ErrNilGoRoutineProcessor
	}
	emptyData := make(map[string]core.GoRoutineHandlerMap)
	emptyData[newRoutine] = make(core.GoRoutineHandlerMap)
	emptyData[oldRoutine] = make(core.GoRoutineHandlerMap)
	return &GoRoutinesAnalyser{
		goRoutinesProcessor: processor,
		goRoutinesData:      emptyData,
	}, nil
}

// DumpGoRoutinesToLogWithTypes will print the currently running go routines stats in the log
func (grd *GoRoutinesAnalyser) DumpGoRoutinesToLogWithTypes() {
	buffer := core.GetRunningGoRoutines()
	log.Debug("GoRoutinesAnalyser - DumpGoRoutinesToLogWithTypes", "goroutines number", runtime.NumGoroutine())

	newData := grd.goRoutinesProcessor.ProcessGoRoutineBuffer(grd.goRoutinesData, buffer)

	dumpGoRoutinesDataToLog(newData)
	grd.goRoutinesData = newData
}

func dumpGoRoutinesDataToLog(latestData map[string]core.GoRoutineHandlerMap) {
	oldRoutines := latestData[oldRoutine]
	goRoutines := make([]core.GoRoutineHandler, 0)

	for _, val := range oldRoutines {
		goRoutines = append(goRoutines, val)
	}

	sort.Slice(goRoutines, func(i, j int) bool {
		if goRoutines[i].FirstOccurrence().Equal(goRoutines[j].FirstOccurrence()) {
			return strings.Compare(goRoutines[i].ID(), goRoutines[j].ID()) < 0
		}
		return goRoutines[i].FirstOccurrence().Before(goRoutines[j].FirstOccurrence())
	})

	currentTime := time.Now()
	for _, gr := range goRoutines {
		runningTime := currentTime.Sub(gr.FirstOccurrence()).Seconds()
		if runningTime > 3600 {
			log.Debug("\nremaining goroutine more than an hour", "ID", gr.ID(), "running seconds", runningTime)
		}
		log.Debug("\nremaining goroutine", "ID", gr.ID(), "running seconds", runningTime)
	}

	for _, val := range latestData[newRoutine] {
		log.Debug("\nnew goroutine", "ID", val.ID(), "\nData", val.StackTrace()+"\n")
	}
}

func getGoroutineId(goroutineString string) string {
	rg := regexp.MustCompile(`goroutine \d+`)

	matches := rg.FindAllString(goroutineString, -1)
	for _, element := range matches {
		replaced := strings.ReplaceAll(element, "goroutine ", "")
		return replaced
	}

	return ""
}
