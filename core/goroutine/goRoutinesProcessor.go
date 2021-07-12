package goroutine

import (
	"bytes"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

type goRoutinesProcessor struct {
}

// NewGoRoutinesProcessor creates a new GoRoutinesProcessor
func NewGoRoutinesProcessor() core.GoRoutinesProcessor {
	return &goRoutinesProcessor{}
}

// ProcessGoRoutineBuffer processes a new go routine stack dump based on the previous data
func (grp *goRoutinesProcessor) ProcessGoRoutineBuffer(
	previousData map[string]core.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]core.GoRoutineHandlerMap {
	allGoRoutinesString := buffer.String()

	splits := strings.Split(allGoRoutinesString, "\n\n")

	oldGoRoutines := make(core.GoRoutineHandlerMap)
	newGoRoutines := make(core.GoRoutineHandlerMap)

	if previousData[oldRoutine] == nil {
		previousData[oldRoutine] = make(core.GoRoutineHandlerMap)
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

	latestData := make(map[string]core.GoRoutineHandlerMap)
	latestData[newRoutine] = newGoRoutines
	latestData[oldRoutine] = oldGoRoutines

	return latestData
}

// IsInterfaceNil checks if the interface is nil
func (grp *goRoutinesProcessor) IsInterfaceNil() bool {
	return grp == nil
}
