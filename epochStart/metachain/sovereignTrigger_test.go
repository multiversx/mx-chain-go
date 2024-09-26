package metachain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/stretchr/testify/require"
)

func createArgsSovereignTrigger() ArgsSovereignTrigger {
	return ArgsSovereignTrigger{
		ArgsNewMetaEpochStartTrigger: createMockEpochStartTriggerArguments(),
		PeerMiniBlocksSyncer:         &mock.ValidatorInfoSyncerStub{},
	}
}

func TestNewSovereignTrigger(t *testing.T) {
	t.Parallel()

	t.Run("nil args meta, should error", func(t *testing.T) {
		args := createArgsSovereignTrigger()
		args.ArgsNewMetaEpochStartTrigger = nil
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, sovTrigger)
		require.Equal(t, epochStart.ErrNilArgsNewMetaEpochStartTrigger, err)
	})
	t.Run("nil peer mb syncer, should error", func(t *testing.T) {
		args := createArgsSovereignTrigger()
		args.PeerMiniBlocksSyncer = nil
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, sovTrigger)
		require.Equal(t, epochStart.ErrNilValidatorInfoProcessor, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createArgsSovereignTrigger()
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, err)
		require.False(t, sovTrigger.IsInterfaceNil())

		require.Equal(t, "*block.SovereignChainHeader", fmt.Sprintf("%T", sovTrigger.epochStartMeta))
		require.Equal(t, "*metachain.sovereignTriggerRegistryCreator", fmt.Sprintf("%T", sovTrigger.registryHandler))
	})
}

func TestSovereignTrigger_SetProcessed(t *testing.T) {
	args := createArgsSovereignTrigger()
	sovTrigger, _ := NewSovereignTrigger(args)

	// wrong header type, should not update internal data
	sovTrigger.SetProcessed(&block.Header{Epoch: 4}, &block.Body{})
	require.Equal(t, uint32(0), sovTrigger.Epoch())

	sovHdr := &block.SovereignChainHeader{Header: &block.Header{Epoch: 4}}
	_ = sovHdr.SetStartOfEpochHeader()
	sovTrigger.SetProcessed(sovHdr, &block.Body{})
	require.Equal(t, uint32(4), sovTrigger.Epoch())
}

func TestSovereignTrigger_RevertStateToBlock(t *testing.T) {
	t.Parallel()

	triggerFactory := func(arguments *ArgsNewMetaEpochStartTrigger) epochStart.TriggerHandler {
		argSovTrigger := ArgsSovereignTrigger{
			ArgsNewMetaEpochStartTrigger: arguments,
			PeerMiniBlocksSyncer:         &mock.ValidatorInfoSyncerStub{},
		}

		sovEpochStartTrigger, _ := NewSovereignTrigger(argSovTrigger)
		return sovEpochStartTrigger
	}
	sovMetaHdrFactory := func(round uint64, epoch uint32, isEpochStart bool) data.MetaHeaderHandler {
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

	testTriggerRevertToEndOfEpochUpdate(t, triggerFactory, sovMetaHdrFactory)
	testTriggerRevertBehindEpochStartBlock(t, triggerFactory, sovMetaHdrFactory)
}
