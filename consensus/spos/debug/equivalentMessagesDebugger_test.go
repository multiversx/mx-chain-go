package debug

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
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

		debugger := NewEquivalentMessagesDebugger()
		debugger.shouldProcessDataFunc = func() bool {
			return true
		}

		debugger.SetValidEquivalentProof([]byte("hash1"), data.HeaderProof{PubKeysBitmap: []byte("bitmap 1"), AggregatedSignature: []byte("signature 1")})
		debugger.SetValidEquivalentProof([]byte("hash2"), data.HeaderProof{PubKeysBitmap: []byte("bitmap 2"), AggregatedSignature: []byte("signature 2")})

		debugger.DisplayEquivalentMessagesStatistics()
	})
}
