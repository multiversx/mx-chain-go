package shardchain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
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
	rt.finality = t.finality
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
	rt.marshalizer = t.marshalizer
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
	rt.mapMissingMiniblocks = t.mapMissingMiniblocks
	rt.mapFinalizedEpochs = t.mapFinalizedEpochs
	rt.roundHandler = t.roundHandler
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
	bootStorer := mock.NewStorerMock()

	arguments.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return bootStorer
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
	bootStorer := mock.NewStorerMock()

	arguments.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return bootStorer
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
