package debug

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
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
		debugger.DisplayEquivalentMessagesStatistics(func() map[string]*consensus.EquivalentMessageInfo {
			return make(map[string]*consensus.EquivalentMessageInfo)
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

		debugger.DisplayEquivalentMessagesStatistics(func() map[string]*consensus.EquivalentMessageInfo {
			return map[string]*consensus.EquivalentMessageInfo{
				"hash1": {NumMessages: 1, Validated: true, Proof: data.HeaderProof{PubKeysBitmap: []byte("bitmap 1"), AggregatedSignature: []byte("signature 1")}},
				"hash2": {NumMessages: 2, Validated: false},
				"hash3": {NumMessages: 3, Validated: false},
				"hash4": {NumMessages: 4, Validated: true, Proof: data.HeaderProof{PubKeysBitmap: []byte("bitmap 4"), AggregatedSignature: []byte("signature 4")}},
			}
		})

	})
}
