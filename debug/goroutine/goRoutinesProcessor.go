package goroutine

import (
	"bytes"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-go/debug"
)

const newRoutineMarker = "\n\n"

type goRoutinesProcessor struct {
}

// NewGoRoutinesProcessor creates a new GoRoutinesProcessor
func NewGoRoutinesProcessor() *goRoutinesProcessor {
	return &goRoutinesProcessor{}
}

// ProcessGoRoutineBuffer processes a new go routine stack dump based on the previous data
func (grp *goRoutinesProcessor) ProcessGoRoutineBuffer(
	previousData map[string]debug.GoRoutineHandlerMap,
	buffer *bytes.Buffer,
) map[string]debug.GoRoutineHandlerMap {
	allGoRoutinesString := buffer.String()

	splits := strings.Split(allGoRoutinesString, newRoutineMarker)

	oldGoRoutines := make(debug.GoRoutineHandlerMap)
	newGoRoutines := make(debug.GoRoutineHandlerMap)

	if previousData[oldRoutine] == nil {
		previousData[oldRoutine] = make(debug.GoRoutineHandlerMap)
	}

	for k, val := range previousData[newRoutine] {
		previousData[oldRoutine][k] = val
	}

	currentTime := time.Now()
	for _, st := range splits {
		gId := getGoroutineId(st)
		if len(gId) == 0 {
			continue
		}
		val, ok := previousData[oldRoutine][gId]
		if !ok {
			newGoRoutines[gId] = &goRoutineData{
				id:              gId,
				firstOccurrence: currentTime,
				stackTrace:      st,
			}
			continue
		}
		oldGoRoutines[val.ID()] = val
	}

	latestData := make(map[string]debug.GoRoutineHandlerMap)
	latestData[newRoutine] = newGoRoutines
	latestData[oldRoutine] = oldGoRoutines

	return latestData
}

// IsInterfaceNil checks if the interface is nil
func (grp *goRoutinesProcessor) IsInterfaceNil() bool {
	return grp == nil
}
