package disabled

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestProcessStatusHandler_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	psh := NewProcessStatusHandler()
	assert.False(t, check.IfNil(psh))
	psh.SetBusy("")
	psh.SetIdle()
	assert.True(t, psh.IsIdle())
}
