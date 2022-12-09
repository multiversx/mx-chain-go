package counters

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func generateMapsOfValues(value uint64) map[string]uint64 {
	return map[string]uint64{
		maxTrieReads:    value,
		maxTransfers:    value,
		maxBuiltinCalls: value,
	}
}

func TestNewUsageCounter(t *testing.T) {
	t.Parallel()

	t.Run("nil esdt transfer parser should error", func(t *testing.T) {
		t.Parallel()

		counter, err := NewUsageCounter(nil)
		assert.True(t, check.IfNil(counter))
		assert.Equal(t, process.ErrNilESDTTransferParser, err)
	})
	t.Run("nil esdt transfer parser should error", func(t *testing.T) {
		t.Parallel()

		counter, err := NewUsageCounter(&testscommon.ESDTTransferParserStub{})
		assert.False(t, check.IfNil(counter))
		assert.Nil(t, err)
	})
}

func TestUsageCounter_ProcessCrtNumberOfTrieReadsCounter(t *testing.T) {
	t.Parallel()

	counter, _ := NewUsageCounter(&testscommon.ESDTTransferParserStub{})
	counter.SetMaximumValues(generateMapsOfValues(3))

	assert.Nil(t, counter.ProcessCrtNumberOfTrieReadsCounter())                                        // counter is now 1
	assert.Nil(t, counter.ProcessCrtNumberOfTrieReadsCounter())                                        // counter is now 2
	assert.Nil(t, counter.ProcessCrtNumberOfTrieReadsCounter())                                        // counter is now 3
	assert.ErrorIs(t, counter.ProcessCrtNumberOfTrieReadsCounter(), process.ErrMaxBuiltInCallsReached) // counter is now 4, error signalled
	assert.ErrorIs(t, counter.ProcessCrtNumberOfTrieReadsCounter(), process.ErrMaxBuiltInCallsReached) // counter is now 5, error signalled

	counter.ResetCounters()

	assert.Nil(t, counter.ProcessCrtNumberOfTrieReadsCounter()) // counter is now 1
}

func TestUsageCounter_ProcessMaxBuiltInCounters(t *testing.T) {
	t.Parallel()

	t.Run("builtin functions exceeded", func(t *testing.T) {
		t.Parallel()

		counter, _ := NewUsageCounter(&testscommon.ESDTTransferParserStub{})
		counter.SetMaximumValues(generateMapsOfValues(3))

		vmInput := &vmcommon.ContractCallInput{}

		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput))                                        // counter is now 1
		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput))                                        // counter is now 2
		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput))                                        // counter is now 3
		assert.ErrorIs(t, counter.ProcessMaxBuiltInCounters(vmInput), process.ErrMaxBuiltInCallsReached) // counter is now 4, error signalled
		assert.ErrorIs(t, counter.ProcessMaxBuiltInCounters(vmInput), process.ErrMaxBuiltInCallsReached) // counter is now 5, error signalled

		counter.ResetCounters()

		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput)) // counter is now 1
	})
	t.Run("number of transfers exceeded", func(t *testing.T) {
		t.Parallel()

		counter, _ := NewUsageCounter(&testscommon.ESDTTransferParserStub{
			ParseESDTTransfersCalled: func(sndAddr []byte, rcvAddr []byte, function string, args [][]byte) (*vmcommon.ParsedESDTTransfers, error) {
				return &vmcommon.ParsedESDTTransfers{
					ESDTTransfers: make([]*vmcommon.ESDTTransfer, 2),
				}, nil
			},
		})
		values := generateMapsOfValues(6)
		values[maxBuiltinCalls] = 1000
		counter.SetMaximumValues(values)

		vmInput := &vmcommon.ContractCallInput{}

		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput))                                        // counter is now 2
		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput))                                        // counter is now 4
		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput))                                        // counter is now 6
		assert.ErrorIs(t, counter.ProcessMaxBuiltInCounters(vmInput), process.ErrMaxBuiltInCallsReached) // counter is now 8, error signalled
		assert.ErrorIs(t, counter.ProcessMaxBuiltInCounters(vmInput), process.ErrMaxBuiltInCallsReached) // counter is now 10, error signalled

		counter.ResetCounters()

		assert.Nil(t, counter.ProcessMaxBuiltInCounters(vmInput)) // counter is now 2
	})
}
