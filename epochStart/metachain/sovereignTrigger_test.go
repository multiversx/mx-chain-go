package metachain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignTrigger(t *testing.T) {
	t.Parallel()

	t.Run("nil input, should error", func(t *testing.T) {
		sovTrigger, err := NewSovereignTrigger(nil)
		require.Nil(t, sovTrigger)
		require.Equal(t, epochStart.ErrNilArgsNewMetaEpochStartTrigger, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createMockEpochStartTriggerArguments()
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, err)
		require.False(t, sovTrigger.IsInterfaceNil())

		require.Equal(t, "*block.SovereignChainHeader", fmt.Sprintf("%T", sovTrigger.epochStartMeta))
		require.Equal(t, "*metachain.sovereignTriggerRegistryCreator", fmt.Sprintf("%T", sovTrigger.registryHandler))
	})
}

func TestSovereignTrigger_SetProcessed(t *testing.T) {
	args := createMockEpochStartTriggerArguments()
	sovTrigger, _ := NewSovereignTrigger(args)

	// wrong header type, should not update internal data
	sovTrigger.SetProcessed(&block.Header{Epoch: 4}, &block.Body{})
	require.Equal(t, uint32(0), sovTrigger.Epoch())

	sovHdr := &block.SovereignChainHeader{Header: &block.Header{Epoch: 4}}
	_ = sovHdr.SetStartOfEpochHeader()
	sovTrigger.SetProcessed(sovHdr, &block.Body{})
	require.Equal(t, uint32(4), sovTrigger.Epoch())
}

func TestSovereignTrigger_RevertBehindEpochStartBlock(t *testing.T) {
	t.Parallel()

	triggerFactory := func(arguments *ArgsNewMetaEpochStartTrigger) epochStart.TriggerHandler {
		sovEpochStartTrigger, _ := NewSovereignTrigger(arguments)
		return sovEpochStartTrigger
	}
	metaHdrFactory := func(round uint64, epoch uint32, isEpochStart bool) data.MetaHeaderHandler {
		sovHdr := &block.SovereignChainHeader{
			Header: &block.Header{
				Round: round,
				Epoch: epoch,
			},
		}

		if isEpochStart {
			sovHdr.IsStartOfEpoch = true
		}

		return sovHdr
	}

	testTriggerRevertBehindEpochStartBlock(t, triggerFactory, metaHdrFactory)
}
