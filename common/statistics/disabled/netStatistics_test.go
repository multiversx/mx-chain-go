package disabled

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledNetStatistics(t *testing.T) {
	t.Parallel()

	stats := NewDisabledNetStatistics()
	assert.False(t, check.IfNil(stats))
}

func TestNetStatistics_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	stats := NewDisabledNetStatistics()
	message := stats.TotalReceivedInCurrentEpoch()
	assert.Equal(t, notAvailable, message)

	message = stats.TotalSentInCurrentEpoch()
	assert.Equal(t, notAvailable, message)

	stats.EpochConfirmed(0, 0)

	assert.Nil(t, stats.Close())
}
