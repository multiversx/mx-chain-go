package process

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledDebugger(t *testing.T) {
	t.Parallel()

	debugger := NewDisabledDebugger()
	assert.False(t, check.IfNil(debugger))
}

func TestDisabledDebugger_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
		}
	}()

	debugger := NewDisabledDebugger()
	debugger.SetLastCommittedBlockRound(0)
	debugger.SetLastCommittedBlockRound(1)
	err := debugger.Close()
	assert.Nil(t, err)

	debugger.SetLastCommittedBlockRound(1)
}
