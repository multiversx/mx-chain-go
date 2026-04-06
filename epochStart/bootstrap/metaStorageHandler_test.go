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

func TestGetSelfNotarizedMetaForShard_ShardHeaderReferencesMetaBlock(t *testing.T) {
	t.Parallel()

	metaHash := []byte("meta-hash")
	metaHdr := &block.MetaBlock{Nonce: 593}
	shardHdrHash := []byte("shard-hdr-hash")
	shardHdr := &block.Header{
		Nonce:           602,
		MetaBlockHashes: [][]byte{[]byte("older-meta"), metaHash},
	}

	syncedHeaders := map[string]data.HeaderHandler{
		string(shardHdrHash): shardHdr,
		string(metaHash):     metaHdr,
	}

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte { return shardHdrHash },
		GetShardIDCalled:    func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	hash, hdr, found := msh.getSelfNotarizedMetaForShard(epochStartData, syncedHeaders)
	require.True(t, found)
	assert.Equal(t, metaHash, hash)
	assert.Equal(t, uint64(593), hdr.GetNonce())
}

func TestGetSelfNotarizedMetaForShard_WalksBackToPreviousHeader(t *testing.T) {
	t.Parallel()

	metaHash := []byte("meta-hash")
	metaHdr := &block.MetaBlock{Nonce: 591}
	prevShardHdrHash := []byte("prev-shard-hash")
	prevShardHdr := &block.Header{
		Nonce:           601,
		MetaBlockHashes: [][]byte{metaHash},
	}
	shardHdrHash := []byte("shard-hdr-hash")
	shardHdr := &block.Header{
		Nonce:    602,
		PrevHash: prevShardHdrHash,
	}

	syncedHeaders := map[string]data.HeaderHandler{
		string(shardHdrHash):     shardHdr,
		string(prevShardHdrHash): prevShardHdr,
		string(metaHash):         metaHdr,
	}

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte { return shardHdrHash },
		GetShardIDCalled:    func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	hash, hdr, found := msh.getSelfNotarizedMetaForShard(epochStartData, syncedHeaders)
	require.True(t, found)
	assert.Equal(t, metaHash, hash)
	assert.Equal(t, uint64(591), hdr.GetNonce())
}

func TestGetSelfNotarizedMetaForShard_ShardHeaderNotFound(t *testing.T) {
	t.Parallel()

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte { return []byte("missing") },
		GetShardIDCalled:    func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	_, _, found := msh.getSelfNotarizedMetaForShard(epochStartData, map[string]data.HeaderHandler{})
	require.False(t, found)
}

func TestGetSelfNotarizedMetaForShard_GenesisNonceReturnsNotFound(t *testing.T) {
	t.Parallel()

	shardHdrHash := []byte("shard-hdr-hash")
	shardHdr := &block.Header{Nonce: 0}

	syncedHeaders := map[string]data.HeaderHandler{
		string(shardHdrHash): shardHdr,
	}

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte { return shardHdrHash },
		GetShardIDCalled:    func() uint32 { return 0 },
	}

	msh := &metaStorageHandler{}
	_, _, found := msh.getSelfNotarizedMetaForShard(epochStartData, syncedHeaders)
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
