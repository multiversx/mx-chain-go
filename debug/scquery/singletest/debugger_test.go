package singletest

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/debug/scquery"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/goroutines"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScQueryDebugger_CloseShouldNotLeaveDanglingGoRoutine(t *testing.T) {
	t.Parallel()

	args := scquery.ArgsSCQueryDebugger{
		IntervalAutoPrintInSeconds: 1,
		LoggerInstance:             &testscommon.LoggerStub{},
	}

	time.Sleep(time.Millisecond * 100) // wait for the test to start

	goRoutineCounter := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	firstIndex, err := goRoutineCounter.Snapshot()
	require.Nil(t, err)

	debugger, _ := scquery.NewSCQueryDebugger(args)
	time.Sleep(time.Millisecond * 100) // wait for the go routine to start

	startedIndex, err := goRoutineCounter.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 1, len(goRoutineCounter.DiffGoRoutines(firstIndex, startedIndex)))

	_ = debugger.Close()
	time.Sleep(time.Millisecond * 100) // wait for the go routine to stop

	finishedIndex, err := goRoutineCounter.Snapshot()
	require.Nil(t, err)
	assert.Equal(t, 0, len(goRoutineCounter.DiffGoRoutines(firstIndex, finishedIndex)))
}
