package statusHandler

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewProcessStatusHandler(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	assert.False(t, check.IfNil(psh))
}

func TestProcessStatusHandler_AllMethods(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	assert.True(t, psh.IsIdle())

	psh.SetToBusy("reason 1")
	assert.False(t, psh.IsIdle())

	psh.SetToIdle()
	assert.True(t, psh.IsIdle())

	psh.SetToBusy("reason 2")
	assert.False(t, psh.IsIdle())

	psh.SetToIdle()
	assert.True(t, psh.IsIdle())
}

func TestNewProcessStatusHandler_ParallelCalls(t *testing.T) {
	t.Parallel()

	psh := NewProcessStatusHandler()
	numCalls := 1000
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		operationIndex := i % 3

		go func() {
			switch operationIndex {
			case 0:
				psh.SetToBusy("reason")
			case 1:
				psh.SetToIdle()
			case 2:
				psh.IsIdle()
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
