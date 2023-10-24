package debug

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEquivalentMessagesDebugger_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var debugger *equivalentMessagesDebugger
	require.True(t, debugger.IsInterfaceNil())

	debugger = NewEquivalentMessagesDebugger()
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

		debugger := NewEquivalentMessagesDebugger()
		debugger.DisplayEquivalentMessagesStatistics(func() map[string]uint64 {
			return make(map[string]uint64)
		})
	})
	t.Run("nil get data handler should early exit", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		debugger := NewEquivalentMessagesDebugger()
		debugger.shouldProcessDataFunc = func() bool {
			return true
		}

		debugger.DisplayEquivalentMessagesStatistics(nil)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		debugger := NewEquivalentMessagesDebugger()
		debugger.shouldProcessDataFunc = func() bool {
			return true
		}

		debugger.DisplayEquivalentMessagesStatistics(func() map[string]uint64 {
			return map[string]uint64{
				"hash1": 1,
				"hash2": 2,
				"hash3": 3,
				"hash4": 4,
			}
		})

	})
}
