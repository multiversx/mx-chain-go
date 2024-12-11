package metachain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/require"
)

func createArgsSovereignTrigger() ArgsSovereignTrigger {
	return ArgsSovereignTrigger{
		ArgsNewMetaEpochStartTrigger: createMockEpochStartTriggerArguments(),
		ValidatorInfoSyncer:          &mock.ValidatorInfoSyncerStub{},
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
	t.Run("nil validator info syncer, should error", func(t *testing.T) {
		args := createArgsSovereignTrigger()
		args.ValidatorInfoSyncer = nil
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, sovTrigger)
		require.Equal(t, epochStart.ErrNilValidatorInfoSyncer, err)
	})
	t.Run("nil headers pool, should error", func(t *testing.T) {
		args := createArgsSovereignTrigger()
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return nil
			},
		}
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, sovTrigger)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createArgsSovereignTrigger()

		wasRegisterCalled := false
		args.DataPool = &dataRetrieverMock.PoolsHolderStub{
			CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
				return &vic.ValidatorInfoCacherStub{}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &testscommon.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {
						wasRegisterCalled = true
					},
				}
			},
		}
		sovTrigger, err := NewSovereignTrigger(args)
		require.Nil(t, err)
		require.False(t, sovTrigger.IsInterfaceNil())

		require.Equal(t, "*block.SovereignChainHeader", fmt.Sprintf("%T", sovTrigger.epochStartMeta))
		require.Equal(t, "*metachain.sovereignTriggerRegistryCreator", fmt.Sprintf("%T", sovTrigger.registryHandler))
		require.True(t, wasRegisterCalled)
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
			ValidatorInfoSyncer:          &mock.ValidatorInfoSyncerStub{},
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

func TestSovereignTrigger_receivedBlock(t *testing.T) {
	args := createArgsSovereignTrigger()

	wasNotifyPrepareCalled := false
	wasNotifyEpochChangeCalled := false
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		NotifyAllPrepareCalled: func(hdr data.HeaderHandler, body data.BodyHandler) {
			wasNotifyPrepareCalled = true
		},
		NotifyEpochChangeConfirmedCalled: func(epoch uint32) {
			wasNotifyEpochChangeCalled = true
		},
	}

	valInfoHash := "hash"
	shardValInfo := &state.ShardValidatorInfo{PublicKey: []byte("key")}
	validatorsInfo := map[string]*state.ShardValidatorInfo{
		valInfoHash: shardValInfo,
	}

	wasSyncValidatorsCalled := false
	wasSyncMBCalled := false
	args.ValidatorInfoSyncer = &mock.ValidatorInfoSyncerStub{
		SyncValidatorsInfoCalled: func(body data.BodyHandler) ([][]byte, map[string]*state.ShardValidatorInfo, error) {
			wasSyncValidatorsCalled = true
			return nil, validatorsInfo, nil
		},
		SyncMiniBlocksCalled: func(hdr data.HeaderHandler) ([][]byte, data.BodyHandler, error) {
			wasSyncMBCalled = true
			return nil, nil, nil
		},
	}

	wereValidatorsAdded := false
	valInfoCacher := &vic.ValidatorInfoCacherStub{
		AddValidatorInfoCalled: func(validatorInfoHash []byte, validatorInfo *state.ShardValidatorInfo) {
			require.Equal(t, []byte(valInfoHash), validatorInfoHash)
			require.Equal(t, shardValInfo, validatorInfo)
			wereValidatorsAdded = true
		},
	}
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		CurrEpochValidatorInfoCalled: func() dataRetriever.ValidatorInfoCacher {
			return valInfoCacher
		},
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &testscommon.HeadersCacherStub{}
		},
	}

	sovTrigger, _ := NewSovereignTrigger(args)

	// header is for current epoch, will not notify
	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			Epoch: 4,
		},
		IsStartOfEpoch: true,
	}
	sovTrigger.epoch = 4
	sovTrigger.receivedBlock(sovHdr, nil)
	require.False(t, wasNotifyPrepareCalled)
	require.False(t, wasNotifyEpochChangeCalled)
	require.False(t, wasSyncValidatorsCalled)
	require.False(t, wasSyncMBCalled)
	require.False(t, wereValidatorsAdded)

	sovTrigger.epoch = 3
	sovTrigger.receivedBlock(sovHdr, nil)
	require.True(t, wasNotifyPrepareCalled)
	require.True(t, wasNotifyEpochChangeCalled)
	require.True(t, wasSyncValidatorsCalled)
	require.True(t, wasSyncMBCalled)
	require.True(t, wereValidatorsAdded)
}
