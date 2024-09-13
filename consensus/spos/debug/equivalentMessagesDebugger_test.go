package debug

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestNewEquivalentMessagesDebugger_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var debugger *equivalentMessagesDebugger
	require.True(t, debugger.IsInterfaceNil())

	debugger, err := NewEquivalentMessagesDebugger(&dataRetriever.ProofsPoolMock{}, &testscommon.ShardsCoordinatorMock{})
	require.Nil(t, err)

	require.False(t, debugger.IsInterfaceNil())
}

func TestEquivalentMessagesDebugger_DisplayEquivalentMessagesStatistics(t *testing.T) {
	t.Parallel()

	t.Run("log level not trace should early exit", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		debugger, err := NewEquivalentMessagesDebugger(&dataRetriever.ProofsPoolMock{}, &testscommon.ShardsCoordinatorMock{})
		require.Nil(t, err)
		debugger.DisplayEquivalentMessagesStatistics()
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		debugger, err := NewEquivalentMessagesDebugger(&dataRetriever.ProofsPoolMock{}, &testscommon.ShardsCoordinatorMock{})
		require.Nil(t, err)
		debugger.shouldProcessDataFunc = func() bool {
			return true
		}

		debugger.UpsertEquivalentMessage([]byte("hash1"))
		debugger.UpsertEquivalentMessage([]byte("hash2"))

		debugger.DisplayEquivalentMessagesStatistics()
	})
}
