package bootstrap

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func createStorageHandlerArgs() StorageHandlerArgs {
	return StorageHandlerArgs{
		GeneralConfig:                   testscommon.GetGeneralConfig(),
		PreferencesConfig:               config.PreferencesConfig{},
		ShardCoordinator:                &mock.ShardCoordinatorStub{},
		PathManagerHandler:              &testscommon.PathManagerStub{},
		Marshaller:                      &mock.MarshalizerMock{},
		Hasher:                          &hashingMocks.HasherMock{},
		CurrentEpoch:                    0,
		Uint64Converter:                 &mock.Uint64ByteSliceConverterMock{},
		NodeTypeProvider:                &nodeTypeProviderMock.NodeTypeProviderStub{},
		NodesCoordinatorRegistryFactory: &shardingMocks.NodesCoordinatorRegistryFactoryMock{},
		ManagedPeersHolder:              &testscommon.ManagedPeersHolderStub{},
		SnapshotsEnabled:                false,
		NodeProcessingMode:              common.Normal,
		StateStatsHandler:               disabled.NewStateStatistics(),
		RepopulateTokensSupplies:        false,
		ProofsPool:                      &dataRetrieverMocks.ProofsPoolMock{},
		EnableEpochsHandler:             &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

func TestNewMetaStorageHandler_InvalidConfigErr(t *testing.T) {
	args := createStorageHandlerArgs()
	args.GeneralConfig = config.Config{}

	mtStrHandler, err := NewMetaStorageHandler(args)
	assert.True(t, check.IfNil(mtStrHandler))
	assert.NotNil(t, err)
}

func TestNewMetaStorageHandler_CreateForMetaErr(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createStorageHandlerArgs()
	mtStrHandler, err := NewMetaStorageHandler(args)
	assert.False(t, check.IfNil(mtStrHandler))
	assert.Nil(t, err)
}

func TestMetaStorageHandler_saveLastHeader(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createStorageHandlerArgs()
	mtStrHandler, _ := NewMetaStorageHandler(args)
	header := &block.MetaBlock{Nonce: 0}

	headerHash, _ := core.CalculateHash(args.Marshaller, args.Hasher, header)
	expectedBootInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MetachainShardId, Hash: headerHash,
	}

	bootHeaderInfo, err := mtStrHandler.saveLastHeader(header)
	assert.Nil(t, err)
	assert.Equal(t, expectedBootInfo, bootHeaderInfo)
}

func TestMetaStorageHandler_saveLastCrossNotarizedHeaders(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createStorageHandlerArgs()
	mtStrHandler, _ := NewMetaStorageHandler(args)

	hdr1 := &block.Header{Nonce: 1}
	hdr2 := &block.Header{Nonce: 2}
	hdrHash1, _ := core.CalculateHash(args.Marshaller, args.Hasher, hdr1)
	hdrHash2, _ := core.CalculateHash(args.Marshaller, args.Hasher, hdr2)

	hdr3 := &block.MetaBlock{
		Nonce: 3,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{
			{HeaderHash: hdrHash1}, {HeaderHash: hdrHash2},
		}},
	}

	hdrs := map[string]data.HeaderHandler{string(hdrHash1): hdr1, string(hdrHash2): hdr2}
	crossNotarizedHdrs, err := mtStrHandler.saveLastCrossNotarizedHeaders(hdr3, hdrs)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(crossNotarizedHdrs))
}

func TestMetaStorageHandler_saveTriggerRegistry(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createStorageHandlerArgs()
	mtStrHandler, _ := NewMetaStorageHandler(args)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Nonce: 3},
		PreviousEpochStart:  &block.MetaBlock{Nonce: 2},
	}

	_, err := mtStrHandler.saveTriggerRegistry(components)
	assert.Nil(t, err)
}

func TestMetaStorageHandler_saveDataToStorage(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createStorageHandlerArgs()
	mtStrHandler, _ := NewMetaStorageHandler(args)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Nonce: 3},
		PreviousEpochStart:  &block.MetaBlock{Nonce: 2},
	}

	err := mtStrHandler.SaveDataToStorage(components)
	assert.Nil(t, err)
}

func TestMetaStorageHandler_SaveDataToStorageMissingStorer(t *testing.T) {
	t.Parallel()

	t.Run("missing BootstrapUnit", testMetaWithMissingStorer(dataRetriever.BootstrapUnit, 1))
	t.Run("missing MetaBlockUnit", testMetaWithMissingStorer(dataRetriever.MetaBlockUnit, 1))
	t.Run("missing MetaHdrNonceHashDataUnit", testMetaWithMissingStorer(dataRetriever.MetaHdrNonceHashDataUnit, 1))
	t.Run("missing MetaBlockUnit", testMetaWithMissingStorer(dataRetriever.MetaBlockUnit, 2))                       // saveMetaHdrForEpochTrigger(components.EpochStartMetaBlock)
	t.Run("missing BootstrapUnit", testMetaWithMissingStorer(dataRetriever.BootstrapUnit, 2))                       // saveMetaHdrForEpochTrigger(components.EpochStartMetaBlock)
	t.Run("missing MetaBlockUnit", testMetaWithMissingStorer(dataRetriever.MetaBlockUnit, 3))                       // saveMetaHdrForEpochTrigger(components.PreviousEpochStart)
	t.Run("missing BootstrapUnit", testMetaWithMissingStorer(dataRetriever.BootstrapUnit, 3))                       // saveMetaHdrForEpochTrigger(components.PreviousEpochStart)
	t.Run("missing MetaBlockUnit", testMetaWithMissingStorer(dataRetriever.MetaBlockUnit, 4))                       // saveMetaHdrToStorage(components.PreviousEpochStart)
	t.Run("missing MetaHdrNonceHashDataUnit", testMetaWithMissingStorer(dataRetriever.MetaHdrNonceHashDataUnit, 2)) // saveMetaHdrToStorage(components.PreviousEpochStart)
}

func testMetaWithMissingStorer(missingUnit dataRetriever.UnitType, atCallNumber int) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		defer func() {
			_ = os.RemoveAll("./Epoch_0")
		}()

		args := createStorageHandlerArgs()
		mtStrHandler, _ := NewMetaStorageHandler(args)
		counter := 0
		mtStrHandler.storageService = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				counter++
				if counter < atCallNumber {
					return &storageStubs.StorerStub{}, nil
				}

				if unitType == missingUnit ||
					strings.Contains(unitType.String(), missingUnit.String()) {
					return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
				}

				return &storageStubs.StorerStub{}, nil
			},
		}
		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: &block.MetaBlock{Nonce: 3},
			PreviousEpochStart:  &block.MetaBlock{Nonce: 2},
		}

		err := mtStrHandler.SaveDataToStorage(components)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
		require.True(t, strings.Contains(err.Error(), missingUnit.String()))
	}
}

func TestGetSelfNotarizedMetaForShard_FirstPendingMetaAvailable(t *testing.T) {
	t.Parallel()

	metaHash := []byte("meta-hash")
	metaHdr := &block.MetaBlock{Nonce: 593}

	syncedHeaders := map[string]data.HeaderHandler{
		string(metaHash): metaHdr,
	}

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetFirstPendingMetaBlockCalled: func() []byte { return metaHash },
		GetShardIDCalled:               func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	hash, hdr, found := msh.getSelfNotarizedMetaForShard(epochStartData, syncedHeaders)
	require.True(t, found)
	assert.Equal(t, metaHash, hash)
	assert.Equal(t, uint64(593), hdr.GetNonce())
}

func TestGetSelfNotarizedMetaForShard_EmptyFirstPendingMeta(t *testing.T) {
	t.Parallel()

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetFirstPendingMetaBlockCalled: func() []byte { return nil },
		GetShardIDCalled:               func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	_, _, found := msh.getSelfNotarizedMetaForShard(epochStartData, map[string]data.HeaderHandler{})
	require.False(t, found)
}

func TestGetSelfNotarizedMetaForShard_FirstPendingMetaNotSynced(t *testing.T) {
	t.Parallel()

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetFirstPendingMetaBlockCalled: func() []byte { return []byte("missing-meta") },
		GetShardIDCalled:               func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	_, _, found := msh.getSelfNotarizedMetaForShard(epochStartData, map[string]data.HeaderHandler{})
	require.False(t, found)
}

func TestSaveIntermediateMetaBlocksToStorage(t *testing.T) {
	t.Parallel()

	t.Run("saves chain of meta blocks to storage", func(t *testing.T) {
		t.Parallel()

		hash597 := []byte("hash-597")
		hash596 := []byte("hash-596")
		hash595 := []byte("hash-595")
		meta597 := &block.MetaBlock{Nonce: 597, PrevHash: hash596}
		meta596 := &block.MetaBlock{Nonce: 596, PrevHash: hash595}
		meta595 := &block.MetaBlock{Nonce: 595}

		syncedHeaders := map[string]data.HeaderHandler{
			string(hash597): meta597,
			string(hash596): meta596,
			string(hash595): meta595,
		}

		epochStartMeta := &block.MetaBlock{Nonce: 598, PrevHash: hash597}

		putCount := 0
		msh := &metaStorageHandler{
			baseStorageHandler: &baseStorageHandler{
				marshalizer:         &mock.MarshalizerMock{},
				hasher:              &hashingMocks.HasherMock{},
				uint64Converter:     &mock.Uint64ByteSliceConverterMock{},
				enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
				proofsPool:          &dataRetrieverMocks.ProofsPoolMock{},
				storageService: &storageStubs.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &storageStubs.StorerStub{
							PutCalled: func(key, data []byte) error {
								putCount++
								return nil
							},
						}, nil
					},
				},
			},
		}

		err := msh.saveIntermediateMetaBlocksToStorage(epochStartMeta, syncedHeaders)
		require.Nil(t, err)
		assert.True(t, putCount > 0)
	})

	t.Run("stops at missing header", func(t *testing.T) {
		t.Parallel()

		epochStartMeta := &block.MetaBlock{Nonce: 598, PrevHash: []byte("missing")}

		msh := &metaStorageHandler{baseStorageHandler: &baseStorageHandler{}}
		err := msh.saveIntermediateMetaBlocksToStorage(epochStartMeta, make(map[string]data.HeaderHandler))
		require.Nil(t, err)
	})

	t.Run("stops at genesis nonce", func(t *testing.T) {
		t.Parallel()

		genesisHash := []byte("genesis-hash")
		genesisMeta := &block.MetaBlock{Nonce: 0}

		syncedHeaders := map[string]data.HeaderHandler{
			string(genesisHash): genesisMeta,
		}

		epochStartMeta := &block.MetaBlock{Nonce: 1, PrevHash: genesisHash}

		msh := &metaStorageHandler{
			baseStorageHandler: &baseStorageHandler{
				marshalizer:         &mock.MarshalizerMock{},
				hasher:              &hashingMocks.HasherMock{},
				uint64Converter:     &mock.Uint64ByteSliceConverterMock{},
				enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
				proofsPool:          &dataRetrieverMocks.ProofsPoolMock{},
				storageService: &storageStubs.ChainStorerStub{
					GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
						return &storageStubs.StorerStub{
							PutCalled: func(key, data []byte) error { return nil },
						}, nil
					},
				},
			},
		}

		err := msh.saveIntermediateMetaBlocksToStorage(epochStartMeta, syncedHeaders)
		require.Nil(t, err)
	})

	t.Run("empty prev hash returns nil", func(t *testing.T) {
		t.Parallel()

		epochStartMeta := &block.MetaBlock{Nonce: 1}

		msh := &metaStorageHandler{baseStorageHandler: &baseStorageHandler{}}
		err := msh.saveIntermediateMetaBlocksToStorage(epochStartMeta, make(map[string]data.HeaderHandler))
		require.Nil(t, err)
	})
}

func TestCollectIntermediateMetaBlocks(t *testing.T) {
	t.Parallel()

	t.Run("nil previous epoch start", func(t *testing.T) {
		t.Parallel()

		meta2Hash := []byte("meta2hash")
		meta2 := &block.MetaBlock{Nonce: 2}

		epochStart := &block.MetaBlock{Nonce: 3, PrevHash: meta2Hash}
		headers := map[string]data.HeaderHandler{
			string(meta2Hash): meta2,
		}

		blocks := collectIntermediateMetaBlocks(epochStart, nil, headers)
		require.Len(t, blocks, 1)
		assert.Equal(t, uint64(2), blocks[0].GetNonce())
	})

	t.Run("no intermediate blocks", func(t *testing.T) {
		t.Parallel()

		prevEpochStart := &block.MetaBlock{Nonce: 5}
		epochStart := &block.MetaBlock{Nonce: 6, PrevHash: []byte("missing")}

		blocks := collectIntermediateMetaBlocks(epochStart, prevEpochStart, map[string]data.HeaderHandler{})
		require.Empty(t, blocks)
	})

	t.Run("stops at previous epoch start nonce", func(t *testing.T) {
		t.Parallel()

		prevHash := []byte("prev")
		prevEpochStart := &block.MetaBlock{Nonce: 10}
		intermediateBlock := &block.MetaBlock{Nonce: 10, PrevHash: []byte("older")}
		epochStart := &block.MetaBlock{Nonce: 12, PrevHash: prevHash}

		headers := map[string]data.HeaderHandler{
			string(prevHash): intermediateBlock,
		}

		blocks := collectIntermediateMetaBlocks(epochStart, prevEpochStart, headers)
		require.Empty(t, blocks)
	})

	t.Run("collects chain of intermediate blocks", func(t *testing.T) {
		t.Parallel()

		hash11 := []byte("h11")
		hash12 := []byte("h12")
		hash13 := []byte("h13")

		prevEpochStart := &block.MetaBlock{Nonce: 10}
		meta11 := &block.MetaBlock{Nonce: 11, PrevHash: []byte("prevEpochHash")}
		meta12 := &block.MetaBlock{Nonce: 12, PrevHash: hash11}
		meta13 := &block.MetaBlock{Nonce: 13, PrevHash: hash12}
		epochStart := &block.MetaBlock{Nonce: 14, PrevHash: hash13}

		headers := map[string]data.HeaderHandler{
			string(hash11): meta11,
			string(hash12): meta12,
			string(hash13): meta13,
		}

		blocks := collectIntermediateMetaBlocks(epochStart, prevEpochStart, headers)
		require.Len(t, blocks, 3)
	})

	t.Run("stops when header not in map", func(t *testing.T) {
		t.Parallel()

		hash12 := []byte("h12")
		meta12 := &block.MetaBlock{Nonce: 12, PrevHash: []byte("missing")}
		epochStart := &block.MetaBlock{Nonce: 14, PrevHash: hash12}

		headers := map[string]data.HeaderHandler{
			string(hash12): meta12,
		}

		blocks := collectIntermediateMetaBlocks(epochStart, nil, headers)
		require.Len(t, blocks, 1)
		assert.Equal(t, uint64(12), blocks[0].GetNonce())
	})

	t.Run("empty prev hash on epoch start", func(t *testing.T) {
		t.Parallel()

		epochStart := &block.MetaBlock{Nonce: 5}
		blocks := collectIntermediateMetaBlocks(epochStart, nil, map[string]data.HeaderHandler{})
		require.Empty(t, blocks)
	})
}

func TestComputePendingMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("nil previous epoch start with empty epoch start", func(t *testing.T) {
		t.Parallel()

		epochStart := &block.MetaBlock{
			Nonce: 5,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0},
					{ShardID: 1},
				},
			},
		}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  nil,
			Headers:             map[string]data.HeaderHandler{},
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("epoch start with pending miniblock headers", func(t *testing.T) {
		t.Parallel()

		epochStart := &block.MetaBlock{
			Nonce: 5,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{
						ShardID: 0,
						PendingMiniBlockHeaders: []block.MiniBlockHeader{
							{Hash: []byte("mb1"), SenderShardID: 1, ReceiverShardID: 0},
							{Hash: []byte("mb2"), SenderShardID: 2, ReceiverShardID: 0},
						},
					},
					{
						ShardID: 1,
					},
				},
			},
		}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  nil,
			Headers:             map[string]data.HeaderHandler{},
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, uint32(0), result[0].ShardID)
		assert.Len(t, result[0].MiniBlocksHashes, 2)
	})

	t.Run("intermediate blocks build up pending state via toggle", func(t *testing.T) {
		t.Parallel()

		mbHash := []byte("crossShardMb")
		intermediateHash := []byte("interHash")

		// intermediate block creates a cross-shard miniblock (shard1 -> shard0)
		intermediate := &block.MetaBlock{
			Nonce: 11,
			ShardInfo: []block.ShardData{
				{
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: mbHash, SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		}

		epochStart := &block.MetaBlock{
			Nonce:    12,
			PrevHash: intermediateHash,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{
						ShardID: 0,
						PendingMiniBlockHeaders: []block.MiniBlockHeader{
							{Hash: mbHash, SenderShardID: 1, ReceiverShardID: 0},
						},
					},
					{ShardID: 1},
				},
			},
		}

		prevEpochStart := &block.MetaBlock{Nonce: 10}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  prevEpochStart,
			Headers: map[string]data.HeaderHandler{
				string(intermediateHash): intermediate,
			},
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, uint32(0), result[0].ShardID)
		assert.Len(t, result[0].MiniBlocksHashes, 1)
	})

	t.Run("toggle removes miniblock that appears twice", func(t *testing.T) {
		t.Parallel()

		mbHash := []byte("mb")
		hash11 := []byte("h11")
		hash12 := []byte("h12")

		prevEpochStart := &block.MetaBlock{Nonce: 10}

		// block 11: MB created (added to pending)
		meta11 := &block.MetaBlock{
			Nonce:    11,
			PrevHash: []byte("prevEpoch"),
			ShardInfo: []block.ShardData{
				{
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: mbHash, SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		}

		// block 12: same MB executed (removed from pending)
		meta12 := &block.MetaBlock{
			Nonce:    12,
			PrevHash: hash11,
			ShardInfo: []block.ShardData{
				{
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: mbHash, SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		}

		// epoch start: no pending left
		epochStart := &block.MetaBlock{
			Nonce:    13,
			PrevHash: hash12,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0},
					{ShardID: 1},
				},
			},
		}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  prevEpochStart,
			Headers: map[string]data.HeaderHandler{
				string(hash11): meta11,
				string(hash12): meta12,
			},
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("missing intermediate blocks does not panic", func(t *testing.T) {
		t.Parallel()

		epochStart := &block.MetaBlock{
			Nonce:    5,
			PrevHash: []byte("missing"),
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0},
				},
			},
		}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  &block.MetaBlock{Nonce: 3},
			Headers:             map[string]data.HeaderHandler{},
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("nil headers map does not panic", func(t *testing.T) {
		t.Parallel()

		epochStart := &block.MetaBlock{
			Nonce: 5,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0},
				},
			},
		}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  nil,
			Headers:             nil,
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Empty(t, result)
	})

	t.Run("multiple shards with pending", func(t *testing.T) {
		t.Parallel()

		epochStart := &block.MetaBlock{
			Nonce: 5,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{
						ShardID: 0,
						PendingMiniBlockHeaders: []block.MiniBlockHeader{
							{Hash: []byte("mb1"), SenderShardID: 1, ReceiverShardID: 0},
						},
					},
					{
						ShardID: 1,
						PendingMiniBlockHeaders: []block.MiniBlockHeader{
							{Hash: []byte("mb2"), SenderShardID: 0, ReceiverShardID: 1},
							{Hash: []byte("mb3"), SenderShardID: 2, ReceiverShardID: 1},
						},
					},
					{
						ShardID: 2,
					},
				},
			},
		}

		components := &ComponentsNeededForBootstrap{
			EpochStartMetaBlock: epochStart,
			PreviousEpochStart:  nil,
			Headers:             map[string]data.HeaderHandler{},
		}

		result, err := computePendingMiniBlocks(components)
		require.Nil(t, err)
		require.Len(t, result, 2)

		pendingByShardID := make(map[uint32]int)
		for _, info := range result {
			pendingByShardID[info.ShardID] = len(info.MiniBlocksHashes)
		}
		assert.Equal(t, 1, pendingByShardID[0])
		assert.Equal(t, 2, pendingByShardID[1])
	})
}
