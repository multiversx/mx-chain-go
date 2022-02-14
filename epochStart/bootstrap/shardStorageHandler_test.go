package bootstrap

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShardStorageHandler_ShouldWork(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, err := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)
	assert.False(t, check.IfNil(shardStrHandler))
	assert.Nil(t, err)
}

func TestShardStorageHandler_SaveDataToStorageShardDataNotFound(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, _ := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Epoch: 1},
		PreviousEpochStart:  &block.MetaBlock{Epoch: 1},
		ShardHeader:         &block.Header{Nonce: 1},
	}

	err := shardStrHandler.SaveDataToStorage(components)
	assert.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_SaveDataToStorageMissingHeader(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, _ := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{
			Epoch: 1,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0, Nonce: 1},
				},
			},
		},
		PreviousEpochStart: &block.MetaBlock{Epoch: 1},
		ShardHeader:        &block.Header{Nonce: 1},
	}

	err := shardStrHandler.SaveDataToStorage(components)
	assert.Equal(t, epochStart.ErrMissingHeader, err)
}

func TestShardStorageHandler_SaveDataToStorage(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	gCfg := testscommon.GetGeneralConfig()
	prefsConfig := config.PreferencesConfig{}
	coordinator := &mock.ShardCoordinatorStub{}
	pathManager := &testscommon.PathManagerStub{}
	marshalizer := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	uit64Cvt := &mock.Uint64ByteSliceConverterMock{}
	nodeTypeProvider := &nodeTypeProviderMock.NodeTypeProviderStub{}

	shardStrHandler, _ := NewShardStorageHandler(gCfg, prefsConfig, coordinator, pathManager, marshalizer, hasher, 1, uit64Cvt, nodeTypeProvider)

	hash1 := []byte("hash1")
	hdr1 := block.MetaBlock{
		Nonce: 1,
	}
	headers := map[string]data.HeaderHandler{
		string(hash1): &hdr1,
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{
			Epoch: 1,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{ShardID: 0, Nonce: 1, FirstPendingMetaBlock: hash1, LastFinishedMetaBlock: hash1},
				},
			},
		},
		PreviousEpochStart: &block.MetaBlock{Epoch: 1},
		ShardHeader:        &block.Header{Nonce: 1},
		Headers:            headers,
		NodesConfig:        &nodesCoordinator.NodesCoordinatorRegistry{},
	}

	err := shardStrHandler.SaveDataToStorage(components)
	assert.Nil(t, err)
}

func TestGetAllMiniBlocksWithDst(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	shardMiniBlockHeader := block.MiniBlockHeader{SenderShardID: 1, Hash: hash1}
	miniBlockHeader := block.MiniBlockHeader{SenderShardID: 1, Hash: hash2}
	metablock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					shardMiniBlockHeader,
					{SenderShardID: 0},
				},
			},
			{ShardID: 0},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			{SenderShardID: 0},
			miniBlockHeader,
		},
	}

	shardMbHeaders := getAllMiniBlocksWithDst(metablock, 0)
	assert.Equal(t, shardMbHeaders[string(hash1)], shardMiniBlockHeader)
	assert.NotNil(t, shardMbHeaders[string(hash2)])
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksErrorGettingEpochStartShardData(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	meta := &block.MetaBlock{
		Nonce:      100,
		EpochStart: block.EpochStart{},
	}
	headers := map[string]data.HeaderHandler{}

	miniBlocksInMeta, pendingMiniBlocksInfoList, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksMissingHeader(t *testing.T) {
	t.Parallel()

	lastFinishedMetaBlock := "last finished meta block"
	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte("headerHash")),
		},
	}
	headers := map[string]data.HeaderHandler{}

	miniBlocksInMeta, pendingMiniBlocksInfoList, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Equal(t, epochStart.ErrMissingHeader, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWrongHeader(t *testing.T) {
	t.Parallel()

	lastFinishedMetaBlockHash := "last finished meta block"
	firstPendingMeta := "first pending meta"
	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	lastFinishedHeaders := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlockHash), []byte("headerHash"))
	lastFinishedHeaders[0].FirstPendingMetaBlock = []byte(firstPendingMeta)
	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: lastFinishedHeaders,
		},
	}
	headers := map[string]data.HeaderHandler{
		lastFinishedMetaBlockHash: &block.MetaBlock{},
		firstPendingMeta:          &block.Header{},
	}

	miniBlocksInMeta, pendingMiniBlocksInfoList, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksNilMetaBlock(t *testing.T) {
	t.Parallel()

	lastFinishedMetaBlockHash := "last finished meta block"
	firstPendingMeta := "first pending meta"
	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	lastFinishedHeaders := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlockHash), []byte("headerHash"))
	lastFinishedHeaders[0].FirstPendingMetaBlock = []byte(firstPendingMeta)
	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: lastFinishedHeaders,
		},
	}

	var nilMetaBlock *block.MetaBlock
	headers := map[string]data.HeaderHandler{
		lastFinishedMetaBlockHash: &block.MetaBlock{},
		firstPendingMeta:          nilMetaBlock,
	}

	miniBlocksInMeta, pendingMiniBlocksInfoList, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Equal(t, epochStart.ErrNilMetaBlock, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksNoProcessedNoPendingMbs(t *testing.T) {
	t.Parallel()

	lastFinishedMetaBlockHash := "last finished meta block"
	firstPendingMeta := "first pending meta"
	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	lastFinishedHeaders := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlockHash), []byte("headerHash"))
	lastFinishedHeaders[0].FirstPendingMetaBlock = []byte(firstPendingMeta)
	lastFinishedHeaders[0].PendingMiniBlockHeaders = nil
	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: lastFinishedHeaders,
		},
	}

	neededMeta := &block.MetaBlock{Nonce: 98}

	headers := map[string]data.HeaderHandler{
		lastFinishedMetaBlockHash: &block.MetaBlock{},
		firstPendingMeta:          neededMeta,
	}

	miniBlocksInMeta, pendingMiniBlocksInfoList, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, err)
	require.Len(t, pendingMiniBlocksInfoList, 0)
	require.Len(t, miniBlocksInMeta, 0)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithProcessedAndPendingMbs(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	scenario := createPendingAndProcessedMiniBlocksScenario()
	processedMiniBlocks, pendingMiniBlocks, err := shardStorage.getProcessedAndPendingMiniBlocks(scenario.metaBlock, scenario.headers)

	require.Nil(t, err)
	require.Equal(t, scenario.expectedPendingMbs, pendingMiniBlocks)
	require.Equal(t, scenario.expectedProcessedMbs, processedMiniBlocks)
}

type shardStorageArgs struct {
	generalConfig      config.Config
	prefsConfig        config.PreferencesConfig
	shardCoordinator   sharding.Coordinator
	pathManagerHandler storage.PathManagerHandler
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	currentEpoch       uint32
	uint64Converter    typeConverters.Uint64ByteSliceConverter
	nodeTypeProvider   core.NodeTypeProviderHandler
}

func createDefaultShardStorageArgs() shardStorageArgs {
	return shardStorageArgs{
		generalConfig:      testscommon.GetGeneralConfig(),
		prefsConfig:        config.PreferencesConfig{},
		shardCoordinator:   &mock.ShardCoordinatorStub{},
		pathManagerHandler: &testscommon.PathManagerStub{},
		marshalizer:        &mock.MarshalizerMock{},
		hasher:             &testscommon.HasherMock{},
		currentEpoch:       0,
		uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		nodeTypeProvider:   &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
}

func createDefaultEpochStartShardData(lastFinishedMetaBlockHash []byte, shardHeaderHash []byte) []block.EpochStartShardData {
	return []block.EpochStartShardData{
		{
			HeaderHash:            shardHeaderHash,
			ShardID:               0,
			LastFinishedMetaBlock: lastFinishedMetaBlockHash,
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				{SenderShardID: 1, ReceiverShardID: 0},
			},
		},
	}
}

type scenarioData struct {
	shardHeader          *block.Header
	headers              map[string]data.HeaderHandler
	metaBlock            *block.MetaBlock
	expectedPendingMbs   []bootstrapStorage.PendingMiniBlocksInfo
	expectedProcessedMbs []bootstrapStorage.MiniBlocksInMeta
}

func createPendingAndProcessedMiniBlocksScenario() scenarioData {
	lastFinishedMetaBlockHash := "lastMetaBlockHash"
	firstPendingMetaHash := "firstPendingMetaHash"
	prevShardHeaderHash := "prevShardHeaderHash"
	shardHeaderHash := "shardHeaderHash"

	crossMbHeaders := []block.MiniBlockHeader{
		{Hash: []byte("mb_1_0_0"), SenderShardID: 1, ReceiverShardID: 0},
		{Hash: []byte("mb_2_0_1"), SenderShardID: 2, ReceiverShardID: 0},
		{Hash: []byte("mb_meta_0_2"), SenderShardID: core.MetachainShardId, ReceiverShardID: 0},
		{Hash: []byte("mb_2_0_3"), SenderShardID: 2, ReceiverShardID: 0},
		{Hash: []byte("mb_1_0_4"), SenderShardID: 1, ReceiverShardID: 0},
	}

	intraMbHeaders := []block.MiniBlockHeader{
		{Hash: []byte("mb_0_0_0"), SenderShardID: 0, ReceiverShardID: 0},
		{Hash: []byte("mb_0_0_1"), SenderShardID: 0, ReceiverShardID: 0},
		{Hash: []byte("mb_0_0_2"), SenderShardID: 0, ReceiverShardID: 0},
	}

	processedMbsHeaders := []block.MiniBlockHeader{crossMbHeaders[0]}
	processedMbsHeaders = append(processedMbsHeaders, intraMbHeaders...)
	pendingMbsHeaders := crossMbHeaders[1:]

	shardHeader := &block.Header{
		ShardID:          0,
		Nonce:            96,
		PrevHash:         []byte(prevShardHeaderHash),
		MiniBlockHeaders: processedMbsHeaders,
	}

	expectedPendingMiniBlocks := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 0, MiniBlocksHashes: [][]byte{crossMbHeaders[1].Hash, crossMbHeaders[2].Hash, crossMbHeaders[3].Hash, crossMbHeaders[4].Hash}},
	}
	expectedProcessedMiniBlocks := []bootstrapStorage.MiniBlocksInMeta{
		{MetaHash: []byte(firstPendingMetaHash), MiniBlocksHashes: [][]byte{crossMbHeaders[0].Hash}},
	}

	headers := map[string]data.HeaderHandler{
		lastFinishedMetaBlockHash: &block.MetaBlock{
			Nonce:    94,
			PrevHash: []byte("before last finished meta"),
		},
		firstPendingMetaHash: &block.MetaBlock{
			Nonce:    95,
			PrevHash: []byte(lastFinishedMetaBlockHash),
			ShardInfo: []block.ShardData{
				{ShardID: 0, HeaderHash: []byte(prevShardHeaderHash), ShardMiniBlockHeaders: processedMbsHeaders},
				{ShardID: 1, HeaderHash: []byte("header hash "), ShardMiniBlockHeaders: []block.MiniBlockHeader{crossMbHeaders[4]}},
				{ShardID: 2, HeaderHash: []byte("header hash 2 "), ShardMiniBlockHeaders: []block.MiniBlockHeader{crossMbHeaders[1], crossMbHeaders[3]}},
				{ShardID: core.MetachainShardId, HeaderHash: []byte("header hash 3"), ShardMiniBlockHeaders: []block.MiniBlockHeader{crossMbHeaders[2]}},
			},
		},
		shardHeaderHash: shardHeader,
		prevShardHeaderHash: &block.Header{
			Nonce:    95,
			PrevHash: []byte("prevPrevShardHeaderHash"),
		},
	}

	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, Nonce: 96, FirstPendingMetaBlock: []byte(firstPendingMetaHash), LastFinishedMetaBlock: []byte(lastFinishedMetaBlockHash), PendingMiniBlockHeaders: pendingMbsHeaders},
			},
		},
	}

	return scenarioData{
		shardHeader:          shardHeader,
		headers:              headers,
		metaBlock:            meta,
		expectedPendingMbs:   expectedPendingMiniBlocks,
		expectedProcessedMbs: expectedProcessedMiniBlocks,
	}
}
