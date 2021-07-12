package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGoRoutinesAnalyser_NewGoRoutinesAnalyser(t *testing.T) {
	t.Parallel()

	analyser := NewGoRoutinesAnalyser()
	assert.NotNil(t, analyser)

	assert.Equal(t, 0, len(analyser.goRoutinesData[oldRoutine]))
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
}

func TestGoRoutinesAnalyser_DumpGoRoutinesToLogWithTypesRoutineCountOk(t *testing.T) {
	t.Parallel()

	analyser := NewGoRoutinesAnalyser()

	// set baseline
	analyser.DumpGoRoutinesToLogWithTypes()

	ctx, closeFunc := context.WithCancel(context.Background())

	go func() {
		testGoRoutines(ctx)
	}()

	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 1, len(analyser.goRoutinesData[newRoutine]))

	prevRoutinesNumber := len(analyser.goRoutinesData[oldRoutine])
	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, prevRoutinesNumber+1, len(analyser.goRoutinesData[oldRoutine]))

	closeFunc()

	analyser.DumpGoRoutinesToLogWithTypes()
	assert.Equal(t, 0, len(analyser.goRoutinesData[newRoutine]))
	assert.Equal(t, prevRoutinesNumber, len(analyser.goRoutinesData[oldRoutine]))
}

func testGoRoutines(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case <-time.After(1 * time.Second):
			{
				continue
			}
		}
	}
}
