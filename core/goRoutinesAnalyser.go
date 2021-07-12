package core

import (
	"bytes"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type GoRoutineDataMap = map[string]*GoRoutineData

type GoRoutinesAnalyser struct {
	goRoutinesData map[string]GoRoutineDataMap
}

type GoRoutineData struct {
	ID              string
	FirstOccurrence time.Time
	StackTrace      string
}

// NewGoRoutinesAnalyser creates a new structure for go routine statistics analysis
func NewGoRoutinesAnalyser() *GoRoutinesAnalyser {
	latestData := make(map[string]GoRoutineDataMap)
	latestData[newRoutine] = make(GoRoutineDataMap)
	latestData[oldRoutine] = make(GoRoutineDataMap)
	return &GoRoutinesAnalyser{
		latestData,
	}
}

// DumpGoRoutinesToLogWithTypes will print the currently running go routines stats in the log
func (grd *GoRoutinesAnalyser) DumpGoRoutinesToLogWithTypes() {
	buffer := getRunningGoRoutines()
	log.Debug("GoRoutinesAnalyser - DumpGoRoutinesToLogWithTypes", "goroutines number", runtime.NumGoroutine())

	newData := processGoRoutineBuffer(grd.goRoutinesData, buffer)

	dumpGoRoutinesDataToLog(newData)
	grd.goRoutinesData = newData
}

func dumpGoRoutinesDataToLog(latestData map[string]GoRoutineDataMap) {
	oldRoutines := latestData[oldRoutine]
	goRoutines := make([]*GoRoutineData, 0)

	for _, val := range oldRoutines {
		goRoutines = append(goRoutines, val)
	}

	sort.Slice(goRoutines, func(i, j int) bool {
		if goRoutines[i].FirstOccurrence.Equal(goRoutines[j].FirstOccurrence) {
			return strings.Compare(goRoutines[i].ID, goRoutines[j].ID) < 0
		}
		return goRoutines[i].FirstOccurrence.Before(goRoutines[j].FirstOccurrence)
	})

	currentTime := time.Now()
	for _, gr := range goRoutines {
		runningTime := currentTime.Sub(gr.FirstOccurrence).Seconds()
		if runningTime > 3600 {
			log.Debug("\nremaining routine more than an hour", "ID", gr.ID, "running seconds", runningTime)
		}
		log.Debug("\nremaining routine", "ID", gr.ID, "running seconds", runningTime)
	}

	for _, val := range latestData[newRoutine] {
		log.Debug("\nnew routine", "ID", val.ID, "\nData", val.StackTrace+"\n")
	}
}

func processGoRoutineBuffer(previousData map[string]GoRoutineDataMap, buffer *bytes.Buffer) map[string]GoRoutineDataMap {
	allGoRoutinesString := buffer.String()
	splits := strings.Split(allGoRoutinesString, "\n\n")

	oldGoRoutines := make(GoRoutineDataMap)
	newGoRoutines := make(GoRoutineDataMap)

	for k, val := range previousData[newRoutine] {
		previousData[oldRoutine][k] = val
	}

	currentTime := time.Now()
	for _, st := range splits {
		gId := getGoroutineId(st)
		val, ok := previousData[oldRoutine][gId]
		if !ok {
			newGoRoutines[gId] = &GoRoutineData{
				ID:              gId,
				FirstOccurrence: currentTime,
				StackTrace:      st,
			}
			continue
		}
		oldGoRoutines[val.ID] = val
	}

	latestData := make(map[string]GoRoutineDataMap)
	latestData[newRoutine] = newGoRoutines
	latestData[oldRoutine] = oldGoRoutines

	return latestData
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
