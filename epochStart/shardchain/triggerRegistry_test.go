package shardchain

import (
	"encoding/json"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cloneTrigger(t *trigger) *trigger {
	rt := &trigger{}

	rt.epoch = t.epoch
	rt.metaEpoch = t.epoch
	rt.currentRoundIndex = t.currentRoundIndex
	rt.epochStartRound = t.epochStartRound
	rt.epochMetaBlockHash = t.epochMetaBlockHash
	rt.triggerStateKey = t.triggerStateKey
	rt.isEpochStart = t.isEpochStart
	rt.chainParametersHandler = t.chainParametersHandler
	rt.validity = t.validity
	rt.epochFinalityAttestingRound = t.epochFinalityAttestingRound
	rt.newEpochHdrReceived = t.newEpochHdrReceived
	rt.mapHashHdr = t.mapHashHdr
	rt.mapNonceHashes = t.mapNonceHashes
	rt.mapEpochStartHdrs = t.mapEpochStartHdrs
	rt.metaHdrStorage = t.metaHdrStorage
	rt.triggerStorage = t.triggerStorage
	rt.metaNonceHdrStorage = t.metaNonceHdrStorage
	rt.uint64Converter = t.uint64Converter
	rt.marshaller = t.marshaller
	rt.hasher = t.hasher
	rt.headerValidator = t.headerValidator
	rt.requestHandler = t.requestHandler
	rt.epochStartNotifier = t.epochStartNotifier
	rt.headersPool = t.headersPool
	rt.epochStartShardHeader = t.epochStartShardHeader
	rt.epochStartMeta = t.epochStartMeta
	rt.shardHdrStorage = t.shardHdrStorage
	rt.peerMiniBlocksSyncer = t.peerMiniBlocksSyncer
	rt.appStatusHandler = t.appStatusHandler
	rt.miniBlocksPool = t.miniBlocksPool
	rt.currentEpochValidatorInfoPool = t.currentEpochValidatorInfoPool
	rt.validatorInfoPool = t.validatorInfoPool
	rt.mapMissingValidatorsInfo = t.mapMissingValidatorsInfo
	rt.mapMissingMiniBlocks = t.mapMissingMiniBlocks
	rt.mapFinalizedEpochs = t.mapFinalizedEpochs
	rt.roundHandler = t.roundHandler
	rt.enableEpochsHandler = t.enableEpochsHandler
	return rt
}

func createDummyEpochStartTriggers(arguments *ArgsShardEpochStartTrigger, key []byte) (*trigger, *trigger) {
	epochStartTrigger1, _ := NewEpochStartTrigger(arguments)
	// create a copy
	epochStartTrigger2 := cloneTrigger(epochStartTrigger1)

	epochStartTrigger1.triggerStateKey = key
	epochStartTrigger1.epoch = 10
	epochStartTrigger1.metaEpoch = 11
	epochStartTrigger1.currentRoundIndex = 800
	epochStartTrigger1.epochStartRound = 650
	epochStartTrigger1.epochMetaBlockHash = []byte("meta block hash")
	epochStartTrigger1.isEpochStart = false
	epochStartTrigger1.epochFinalityAttestingRound = 680
	epochStartTrigger1.cancelFunc = nil
	epochStartTrigger1.epochStartShardHeader = &block.Header{}
	return epochStartTrigger1, epochStartTrigger2
}

func TestTrigger_LoadHeaderV1StateAfterSave(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockShardEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := genericMocks.NewStorerMock()

	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return bootStorer, nil
		},
	}
	key := []byte("key")
	epochStartTrigger1, epochStartTrigger2 := createDummyEpochStartTriggers(arguments, key)
	err := epochStartTrigger1.saveState(key)
	assert.Nil(t, err)
	assert.NotEqual(t, epochStartTrigger1, epochStartTrigger2)

	err = epochStartTrigger2.LoadState(key)
	assert.Nil(t, err)
	assert.Equal(t, epochStartTrigger1, epochStartTrigger2)
}

func TestTrigger_LoadHeaderV2StateAfterSave(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockShardEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := genericMocks.NewStorerMock()

	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return bootStorer, nil
		},
	}

	key := []byte("key")
	epochStartTrigger1, epochStartTrigger2 := createDummyEpochStartTriggers(arguments, key)
	epochStartTrigger1.epochStartShardHeader = &block.HeaderV2{
		Header:            &block.Header{},
		ScheduledRootHash: []byte("scheduled root hash")}
	err := epochStartTrigger1.saveState(key)
	assert.Nil(t, err)
	assert.NotEqual(t, epochStartTrigger1, epochStartTrigger2)

	err = epochStartTrigger2.LoadState(key)
	assert.Nil(t, err)
	assert.Equal(t, epochStartTrigger1, epochStartTrigger2)
}

func TestTrigger_LoadStateBackwardsCompatibility(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockShardEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := genericMocks.NewStorerMock()

	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return bootStorer, nil
		},
	}

	key := []byte("key")
	epochStartTrigger1, epochStartTrigger2 := createDummyEpochStartTriggers(arguments, key)

	trig := createLegacyTriggerRegistryFromTrigger(epochStartTrigger1)
	d, err := json.Marshal(trig)
	require.Nil(t, err)
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)

	err = bootStorer.Put(trigInternalKey, d)
	require.Nil(t, err)

	err = epochStartTrigger2.LoadState(key)
	require.Nil(t, err)
	require.Equal(t, epochStartTrigger1, epochStartTrigger2)
}

type legacyTriggerRegistry struct {
	IsEpochStart                bool
	NewEpochHeaderReceived      bool
	Epoch                       uint32
	MetaEpoch                   uint32
	CurrentRoundIndex           int64
	EpochStartRound             uint64
	EpochFinalityAttestingRound uint64
	EpochMetaBlockHash          []byte
	EpochStartShardHeader       data.HeaderHandler
}

func createLegacyTriggerRegistryFromTrigger(t *trigger) *legacyTriggerRegistry {
	header, _ := t.epochStartShardHeader.(*block.Header)
	return &legacyTriggerRegistry{
		IsEpochStart:                t.isEpochStart,
		NewEpochHeaderReceived:      t.newEpochHdrReceived,
		Epoch:                       t.epoch,
		MetaEpoch:                   t.metaEpoch,
		CurrentRoundIndex:           t.currentRoundIndex,
		EpochStartRound:             t.epochStartRound,
		EpochFinalityAttestingRound: t.epochFinalityAttestingRound,
		EpochMetaBlockHash:          t.epochMetaBlockHash,
		EpochStartShardHeader:       header,
	}
}
