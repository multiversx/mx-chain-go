package counters

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledCounter_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	counter := NewDisabledCounter()
	assert.False(t, check.IfNil(counter))
	counter.ResetCounters()
	counter.SetMaximumValues(nil)
	assert.Nil(t, counter.ProcessMaxBuiltInCounters(nil))
	assert.Nil(t, counter.ProcessCrtNumberOfTrieReadsCounter())
}
