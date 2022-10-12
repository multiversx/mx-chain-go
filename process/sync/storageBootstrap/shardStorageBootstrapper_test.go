package storageBootstrap

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochNotifierMock "github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	storageMock "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type metaBlockInfo struct {
	metablock *block.MetaBlock
	hash      []byte
}

func TestShardStorageBootstrapper_LoadFromStorageShouldWork(t *testing.T) {
	t.Parallel()

	wasCalledBlockchainSetHash := false
	wasCalledBlockchainSetHeader := false
	wasCalledForkDetectorAddHeader := false
	wasCalledBlockTrackerAddTrackedHeader := false
	wasCalledEpochNotifier := false
	savedLastRound := int64(0)

	marshaller := &testscommon.MarshalizerMock{}
	startRound := 4000
	hdr := &block.Header{
		Nonce:    3999,
		Round:    3999,
		RootHash: []byte("roothash"),
		ShardID:  0,
		ChainID:  []byte("1"),
	}
	hdrHash := []byte("header hash")
	hdrBytes, _ := marshaller.Marshal(hdr)
	blockStorerMock := genericMocks.NewStorerMock()
	_ = blockStorerMock.Put(hdrHash, hdrBytes)

	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper{
			BootStorer: &mock.BoostrapStorerMock{
				GetHighestRoundCalled: func() int64 {
					return int64(startRound)
				},
				GetCalled: func(round int64) (bootstrapStorage.BootstrapData, error) {
					return bootstrapStorage.BootstrapData{
						LastHeader: bootstrapStorage.BootstrapHeaderInfo{
							ShardId: hdr.ShardID,
							Epoch:   hdr.Epoch,
							Nonce:   hdr.Nonce,
							Hash:    hdrHash,
						},
						HighestFinalBlockNonce: 3999,
						LastRound:              round - 1,
					}, nil
				},
				SaveLastRoundCalled: func(round int64) error {
					savedLastRound = round
					return nil
				},
			},
			ForkDetector: &mock.ForkDetectorMock{
				AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdrHash, hash)
					assert.Equal(t, process.BHProcessed, state)

					wasCalledForkDetectorAddHeader = true
					return nil
				},
			},
			BlockProcessor: &mock.BlockProcessorMock{},
			ChainHandler: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return nil
				},
				SetCurrentBlockHeaderHashCalled: func(bytes []byte) {
					assert.Equal(t, hdrHash, bytes)
					wasCalledBlockchainSetHash = true
				},
				SetCurrentBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, rootHash []byte) error {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdr.RootHash, rootHash)
					wasCalledBlockchainSetHeader = true

					return nil
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					if wasCalledBlockchainSetHeader {
						return hdr
					}

					return nil
				},
			},
			Marshalizer: &testscommon.MarshalizerMock{},
			Store: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return blockStorerMock
				},
			},
			Uint64Converter:     testscommon.NewNonceHashConverterMock(),
			BootstrapRoundIndex: uint64(startRound - 1),
			ShardCoordinator:    testscommon.NewMultiShardsCoordinatorMock(1),
			NodesCoordinator:    &shardingMocks.NodesCoordinatorMock{},
			EpochStartTrigger:   &mock.EpochStartTriggerStub{},
			BlockTracker: &mock.BlockTrackerMock{
				AddTrackedHeaderCalled: func(header data.HeaderHandler, hash []byte) {
					assert.Equal(t, hdr, header)
					assert.Equal(t, hdrHash, hash)

					wasCalledBlockTrackerAddTrackedHeader = true
				},
			},
			ChainID:                      string(hdr.ChainID),
			ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
			MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
			EpochNotifier: &epochNotifierMock.EpochNotifierStub{
				CheckEpochCalled: func(header data.HeaderHandler) {
					assert.Equal(t, hdr, header)
					wasCalledEpochNotifier = true
				},
			},
			ProcessedMiniBlocksTracker: &testscommon.ProcessedMiniBlocksTrackerStub{},
			AppStatusHandler:           &statusHandler.AppStatusHandlerMock{},
		},
	}

	ssb, err := NewShardStorageBootstrapper(args)
	require.Nil(t, err)

	err = ssb.LoadFromStorage()
	assert.Nil(t, err)
	assert.True(t, wasCalledBlockchainSetHash)
	assert.True(t, wasCalledBlockchainSetHeader)
	assert.True(t, wasCalledForkDetectorAddHeader)
	assert.True(t, wasCalledBlockTrackerAddTrackedHeader)
	assert.Equal(t, int64(3999), savedLastRound)
	assert.True(t, wasCalledEpochNotifier)
}

func TestShardStorageBootstrapper_CleanupNotarizedStorageForHigherNoncesIfExist(t *testing.T) {
	baseArgs := createMockShardStorageBoostrapperArgs()

	bForceError := true
	numCalled := 0
	numKeysNotFound := 0
	metaNonce := uint64(2)
	nonceToByteSlice := []byte("nonceToByteSlice")
	metaHash := []byte("meta_hash")

	metaNonceToDelete := metaNonce + maxNumOfConsecutiveNoncesNotFoundAccepted + 2
	metaBlock := &block.MetaBlock{Nonce: metaNonceToDelete}
	marshalledMetaBlock, _ := baseArgs.Marshalizer.Marshal(metaBlock)

	baseArgs.Uint64Converter = &mock.Uint64ByteSliceConverterMock{
		ToByteSliceCalled: func(u uint64) []byte {
			if u == metaNonceToDelete {
				return nonceToByteSlice
			}
			return []byte("")
		},
	}
	baseArgs.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &storageMock.StorerStub{
				RemoveCalled: func(key []byte) error {
					if bForceError {
						return errors.New("forced error")
					}

					if bytes.Equal(key, nonceToByteSlice) {
						numCalled++
						return nil
					}
					if bytes.Equal(key, metaHash) {
						numCalled++
						return nil
					}

					return errors.New("error")
				},
				GetCalled: func(key []byte) ([]byte, error) {
					if bytes.Equal(key, nonceToByteSlice) {
						return metaHash, nil
					}
					if bytes.Equal(key, metaHash) {
						return marshalledMetaBlock, nil
					}
					numKeysNotFound++
					return nil, errors.New("error")
				},
			}
		},
	}

	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	crossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 1})
	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 0, numCalled)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: core.MetachainShardId, Nonce: metaNonce})
	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 0, numCalled)
	assert.Equal(t, maxNumOfConsecutiveNoncesNotFoundAccepted, numKeysNotFound-1)

	numKeysNotFound = 0
	metaNonceToDelete = metaNonce + maxNumOfConsecutiveNoncesNotFoundAccepted + 1
	metaBlock = &block.MetaBlock{Nonce: metaNonceToDelete}
	marshalledMetaBlock, _ = baseArgs.Marshalizer.Marshal(metaBlock)

	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 0, numCalled)
	assert.Equal(t, maxNumOfConsecutiveNoncesNotFoundAccepted*2, numKeysNotFound-1)

	numKeysNotFound = 0
	bForceError = false

	ssb.cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders)
	assert.Equal(t, 2, numCalled)
	assert.Equal(t, maxNumOfConsecutiveNoncesNotFoundAccepted*2, numKeysNotFound-1)
}

func TestShardStorageBootstrapper_GetCrossNotarizedHeaderNonceShouldWork(t *testing.T) {
	crossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Nonce: 1})
	nonce, err := getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders)
	assert.Equal(t, sync.ErrHeaderNotFound, err)
	assert.Equal(t, uint64(0), nonce)

	crossNotarizedHeaders = append(crossNotarizedHeaders, bootstrapStorage.BootstrapHeaderInfo{ShardId: core.MetachainShardId, Nonce: 2})
	nonce, err = getLastCrossNotarizedHeaderNonce(crossNotarizedHeaders)
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), nonce)
}

func TestShardStorageBootstrapper_applyCrossNotarizedHeaders(t *testing.T) {
	t.Parallel()

	t.Run("missing current meta should error", func(t *testing.T) {
		addedCrossNotarized := make(map[string]data.HeaderHandler)
		addedTrackedHeaders := make(map[string]data.HeaderHandler)

		bootstrapper, _, currentMeta, crossNotarizedHeaders := setupForApplyCrossNotarizedHeadersTests(addedCrossNotarized, addedTrackedHeaders)
		_ = bootstrapper.store.GetStorer(dataRetriever.MetaBlockUnit).Remove(currentMeta.hash)

		err := bootstrapper.applyCrossNotarizedHeaders(crossNotarizedHeaders)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "missing header : GetMarshalizedHeaderFromStorage"))
	})
	t.Run("missing prev header should still work", func(t *testing.T) {
		addedCrossNotarized := make(map[string]data.HeaderHandler)
		addedTrackedHeaders := make(map[string]data.HeaderHandler)

		bootstrapper, prevMeta, currentMeta, crossNotarizedHeaders := setupForApplyCrossNotarizedHeadersTests(addedCrossNotarized, addedTrackedHeaders)
		_ = bootstrapper.store.GetStorer(dataRetriever.MetaBlockUnit).Remove(prevMeta.hash)

		err := bootstrapper.applyCrossNotarizedHeaders(crossNotarizedHeaders)
		assert.Nil(t, err)

		assert.Equal(t, 1, len(addedCrossNotarized))
		assert.Equal(t, 1, len(addedTrackedHeaders))

		assert.Equal(t, currentMeta.metablock, addedCrossNotarized[string(currentMeta.hash)])
		assert.Equal(t, currentMeta.metablock, addedTrackedHeaders[string(currentMeta.hash)])
	})
	t.Run("should work", func(t *testing.T) {
		addedCrossNotarized := make(map[string]data.HeaderHandler)
		addedTrackedHeaders := make(map[string]data.HeaderHandler)

		bootstrapper, prevMeta, currentMeta, crossNotarizedHeaders := setupForApplyCrossNotarizedHeadersTests(addedCrossNotarized, addedTrackedHeaders)

		err := bootstrapper.applyCrossNotarizedHeaders(crossNotarizedHeaders)
		assert.Nil(t, err)

		assert.Equal(t, 2, len(addedCrossNotarized))
		assert.Equal(t, 2, len(addedTrackedHeaders))

		assert.Equal(t, currentMeta.metablock, addedCrossNotarized[string(currentMeta.hash)])
		assert.Equal(t, currentMeta.metablock, addedTrackedHeaders[string(currentMeta.hash)])

		assert.Equal(t, prevMeta.metablock, addedCrossNotarized[string(prevMeta.hash)])
		assert.Equal(t, prevMeta.metablock, addedTrackedHeaders[string(prevMeta.hash)])
	})
}

func setupForApplyCrossNotarizedHeadersTests(
	addedCrossNotarized map[string]data.HeaderHandler,
	addedTrackedHeaders map[string]data.HeaderHandler,
) (bootstrapper *shardStorageBootstrapper, prev *metaBlockInfo, current *metaBlockInfo, crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) {
	hasher := &hashingMocks.HasherMock{}

	bootstrapper = &shardStorageBootstrapper{
		storageBootstrapper: &storageBootstrapper{
			store:       genericMocks.NewChainStorerMock(0),
			marshalizer: &testscommon.MarshalizerMock{},
			blockTracker: &mock.BlockTrackerMock{
				AddCrossNotarizedHeaderCalled: func(shardID uint32, crossNotarizedHeader data.HeaderHandler, crossNotarizedHeaderHash []byte) {
					addedCrossNotarized[string(crossNotarizedHeaderHash)] = crossNotarizedHeader
				},
				AddTrackedHeaderCalled: func(header data.HeaderHandler, hash []byte) {
					addedTrackedHeaders[string(hash)] = header
				},
			},
		},
	}

	prevMetaBlock := &block.MetaBlock{
		Nonce:    145,
		Epoch:    0,
		Round:    260,
		PrevHash: []byte("random previous hash"),
	}
	prevMetaBlockBytes, _ := bootstrapper.marshalizer.Marshal(prevMetaBlock)
	prevMetaBlockHash := hasher.Compute(string(prevMetaBlockBytes))
	_ = bootstrapper.store.Put(dataRetriever.MetaBlockUnit, prevMetaBlockHash, prevMetaBlockBytes)
	prev = &metaBlockInfo{
		metablock: prevMetaBlock,
		hash:      prevMetaBlockHash,
	}

	currentMetaBlock := &block.MetaBlock{
		Nonce:    146,
		Epoch:    0,
		Round:    262,
		PrevHash: prevMetaBlockHash,
	}
	currentMetaBlockBytes, _ := bootstrapper.marshalizer.Marshal(currentMetaBlock)
	currentMetaBlockHash := hasher.Compute(string(currentMetaBlockBytes))
	_ = bootstrapper.store.Put(dataRetriever.MetaBlockUnit, currentMetaBlockHash, currentMetaBlockBytes)
	current = &metaBlockInfo{
		metablock: currentMetaBlock,
		hash:      currentMetaBlockHash,
	}

	shardBlock := &block.Header{
		ShardID: 1,
		Nonce:   140,
		Round:   250,
	}
	shardBlockBytes, _ := bootstrapper.marshalizer.Marshal(shardBlock)
	shardBlockHash := hasher.Compute(string(shardBlockBytes))
	_ = bootstrapper.store.Put(dataRetriever.BlockHeaderUnit, shardBlockHash, shardBlockBytes)

	crossNotarizedHeaders = []bootstrapStorage.BootstrapHeaderInfo{
		{
			ShardId: currentMetaBlock.GetShardID(),
			Epoch:   0,
			Nonce:   currentMetaBlock.Nonce,
			Hash:    currentMetaBlockHash,
		},
		{
			ShardId: shardBlock.ShardID,
			Epoch:   0,
			Nonce:   shardBlock.Nonce,
			Hash:    shardBlockHash,
		},
	}

	return
}
