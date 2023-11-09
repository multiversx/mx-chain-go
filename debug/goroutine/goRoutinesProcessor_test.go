package goroutine

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/debug"
	"github.com/stretchr/testify/assert"
)

const goRoutineFormat = "goroutine %s\n%s"

func TestNewGoRoutinesProcessor(t *testing.T) {
	t.Parallel()

	grp := NewGoRoutinesProcessor()

	assert.NotNil(t, grp)
	assert.False(t, grp.IsInterfaceNil())
}

func TestGoRoutinesProcessor_ProcessGoRoutineBufferWithNilBuffer(t *testing.T) {
	t.Parallel()

	grp := NewGoRoutinesProcessor()

	previousData := make(map[string]debug.GoRoutineHandlerMap)
	result := grp.ProcessGoRoutineBuffer(previousData, &bytes.Buffer{})

	assert.Equal(t, 0, len(result[newRoutine]))
	assert.Equal(t, 0, len(result[oldRoutine]))
}

func TestGoRoutinesProcessor_ProcessGoRoutineBufferOkNew(t *testing.T) {
	t.Parallel()

	grp := NewGoRoutinesProcessor()

	previousData := make(map[string]debug.GoRoutineHandlerMap)

	goRoutineId := "7"
	goRoutineStack := "testStack"
	goRoutineString := fmt.Sprintf(goRoutineFormat, goRoutineId, goRoutineStack)
	buffer := bytes.NewBufferString(goRoutineString)

	result := grp.ProcessGoRoutineBuffer(previousData, buffer)

	assert.Equal(t, 0, len(result[oldRoutine]))
	assert.Equal(t, 1, len(result[newRoutine]))

	gr := result[newRoutine]
	gData := gr[goRoutineId]

	assert.NotNil(t, gData)
	assert.Equal(t, goRoutineId, gData.ID())
	assert.Equal(t, goRoutineString, gData.StackTrace())
	assert.NotNil(t, gData.FirstOccurrence())
}

func TestGoRoutinesProcessor_ProcessGoRoutineBufferOkOld(t *testing.T) {
	t.Parallel()

	grp := NewGoRoutinesProcessor()

	goRoutineId := "7"
	goRoutineStack := "testStack"
	goRoutineString := fmt.Sprintf(goRoutineFormat, goRoutineId, goRoutineStack)

	previousData := make(map[string]debug.GoRoutineHandlerMap)
	previousData[newRoutine] = make(map[string]debug.GoRoutineHandler)
	previousData[newRoutine][goRoutineId] = &goRoutineData{
		id:              goRoutineId,
		firstOccurrence: time.Now(),
		stackTrace:      goRoutineString,
	}

	buffer := bytes.NewBufferString(goRoutineString)

	result := grp.ProcessGoRoutineBuffer(previousData, buffer)

	assert.Equal(t, 1, len(result[oldRoutine]))
	assert.Equal(t, 0, len(result[newRoutine]))

	gr := result[oldRoutine]
	gData := gr[goRoutineId]

	assert.NotNil(t, gData)
	assert.Equal(t, goRoutineId, gData.ID())
	assert.Equal(t, goRoutineString, gData.StackTrace())
	assert.NotNil(t, gData.FirstOccurrence())

	newData := grp.ProcessGoRoutineBuffer(result, &bytes.Buffer{})
	assert.Equal(t, 0, len(newData[oldRoutine]))
	assert.Equal(t, 0, len(newData[newRoutine]))
}
