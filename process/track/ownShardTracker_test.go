package track

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
)

func TestNewOwnShardTracker(t *testing.T) {
	t.Parallel()

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewOwnShardTracker(nil, 10)
		assert.Nil(t, tracker)
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})
	t.Run("invalid maxNonceDifference", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewOwnShardTracker(&enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 3)
		assert.Nil(t, tracker)
		assert.Equal(t, process.ErrInvalidMaxNonceDifference, err)
	})
	t.Run("valid parameters", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewOwnShardTracker(&enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 10)
		assert.NotNil(t, tracker)
		assert.Nil(t, err)
		assert.False(t, tracker.IsOwnShardStuck())
	})
}

func TestComputeOwnShardStuck(t *testing.T) {
	t.Parallel()

	t.Run("supernova flag disabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return false
			},
		}
		tracker, _ := NewOwnShardTracker(enableEpochsHandler, 10)
		tracker.ComputeOwnShardStuck(&block.BaseExecutionResult{HeaderNonce: 10}, 30)
		assert.False(t, tracker.IsOwnShardStuck())
	})

	t.Run("supernova flag enabled, nonce difference within limit", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return true
			},
		}
		tracker, _ := NewOwnShardTracker(enableEpochsHandler, 10)
		tracker.ComputeOwnShardStuck(&block.BaseExecutionResult{HeaderNonce: 10}, 15)
		assert.False(t, tracker.IsOwnShardStuck())
	})

	t.Run("supernova flag enabled, nonce difference exceeds limit", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return true
			},
		}
		tracker, _ := NewOwnShardTracker(enableEpochsHandler, 10)
		tracker.ComputeOwnShardStuck(&block.BaseExecutionResult{HeaderNonce: 10}, 30)
		assert.True(t, tracker.IsOwnShardStuck())
	})
}
