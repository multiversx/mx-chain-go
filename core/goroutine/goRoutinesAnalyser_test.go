package goroutine

import (
	"bytes"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/assert"
)

func TestGoRoutinesAnalyser_NewGoRoutinesAnalyser(t *testing.T) {
	t.Parallel()

	grpm := &mock.GoRoutineProcessorMock{
		ProcessGoRoutineBufferCalled: func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			return make(map[string]mock.GoRoutineHandlerMap)
		},
	}
	analyser, err := NewGoRoutinesAnalyser(grpm)
	assert.NotNil(t, analyser)
	assert.Nil(t, err)
}

func TestGoRoutinesAnalyser_NewGoRoutinesAnalyserNilProcessorShouldErr(t *testing.T) {
	t.Parallel()

	analyser, err := NewGoRoutinesAnalyser(nil)
	assert.Nil(t, analyser)
	assert.Equal(t, core.ErrNilGoRoutineProcessor, err)
}

func TestGoRoutinesAnalyser_DumpGoRoutinesToLogWithTypesRoutineCountOk(t *testing.T) {
	t.Parallel()

	grpm := &mock.GoRoutineProcessorMock{}
	analyser, _ := NewGoRoutinesAnalyser(grpm)

	grpm.ProcessGoRoutineBufferCalled =
		func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			newMap := make(map[string]mock.GoRoutineHandlerMap)
			newMap[newRoutine] = map[string]core.GoRoutineHandler{
				"1": &goRoutineData{
					id:              "1",
					firstOccurrence: time.Now(),
					stackTrace:      "stack",
				},
			}

			return newMap
		}
	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 1, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, 0, len(analyser.goRoutinesData[oldRoutine]))

	grpm.ProcessGoRoutineBufferCalled =
		func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			newMap := make(map[string]mock.GoRoutineHandlerMap)
			newMap[oldRoutine] = map[string]core.GoRoutineHandler{
				"1": &goRoutineData{
					id:              "1",
					firstOccurrence: time.Now(),
					stackTrace:      "stack",
				},
			}

			return newMap
		}
	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, 1, len(analyser.goRoutinesData[oldRoutine]))

	grpm.ProcessGoRoutineBufferCalled =
		func(previousData map[string]mock.GoRoutineHandlerMap, buffer *bytes.Buffer) map[string]mock.GoRoutineHandlerMap {
			newMap := make(map[string]mock.GoRoutineHandlerMap)
			return newMap
		}

	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, 0, len(analyser.goRoutinesData[oldRoutine]))
}
