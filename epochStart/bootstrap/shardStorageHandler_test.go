package bootstrap

import (
	"bytes"
	"fmt"
	"os"
	"sort"
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
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochStartMocks "github.com/ElrondNetwork/elrond-go/testscommon/bootstrapMocks/epochStart"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShardStorageHandler_ShouldWork(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, err := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)

	assert.False(t, check.IfNil(shardStorage))
	assert.Nil(t, err)
}

func TestShardStorageHandler_SaveDataToStorageShardDataNotFound(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Epoch: 1},
		PreviousEpochStart:  &block.MetaBlock{Epoch: 1},
		ShardHeader:         &block.Header{Nonce: 1},
	}

	err := shardStorage.SaveDataToStorage(components, false)
	assert.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_SaveDataToStorageMissingHeader(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)

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

	err := shardStorage.SaveDataToStorage(components, false)
	assert.Equal(t, epochStart.ErrMissingHeader, err)
}

func TestShardStorageHandler_SaveDataToStorage(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)

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
		NodesConfig:        &sharding.NodesCoordinatorRegistry{},
	}

	err := shardStorage.SaveDataToStorage(components, false)
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

func TestShardStorageHandler_getCrossProcessedMbsDestMeByHeader(t *testing.T) {
	mb1From1To0 := block.MiniBlockHeader{
		Hash:            []byte("mb hash1"),
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	mb2From1To0 := block.MiniBlockHeader{
		Hash:            []byte("mb hash2"),
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	mb3From2To0 := block.MiniBlockHeader{
		Hash:            []byte("mb hash3"),
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	mb4Intra := block.MiniBlockHeader{
		Hash:            []byte("mb hash 4"),
		SenderShardID:   0,
		ReceiverShardID: 0,
	}
	mb5Intra := block.MiniBlockHeader{
		Hash:            []byte("mb hash 5"),
		SenderShardID:   0,
		ReceiverShardID: 0,
	}

	intraMbs := []block.MiniBlockHeader{
		mb4Intra, mb5Intra,
	}

	crossMbs := []block.MiniBlockHeader{
		mb1From1To0, mb2From1To0, mb3From2To0,
	}

	mbs := append(intraMbs, crossMbs...)

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shardHeader := &block.Header{
		Nonce: 100,
		MiniBlockHeaders: mbs,
	}

	expectedMbs := map[uint32][]data.MiniBlockHeaderHandler{
		1: {&mb1From1To0, &mb2From1To0, &mb3From2To0},
	}

	processedMbs := shardStorage.getCrossProcessedMbsDestMeByHeader(shardHeader)
	require.Equal(t, processedMbs, expectedMbs)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithScheduledErrorGettingProcessedAndPendingMbs(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	meta := &block.MetaBlock{
		Nonce:      100,
		EpochStart: block.EpochStart{},
	}
	headers := map[string]data.HeaderHandler{}
	header := &block.Header{Nonce: 100}

	miniBlocksInMeta, pendingMiniBlocks, err := shardStorage.getProcessedAndPendingMiniBlocksWithScheduled(meta, headers, header, false)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocks)
	require.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithScheduledNoScheduled(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	scenario := createPendingAndProcessedMiniBlocksScenario()

	processedMiniBlocks, pendingMiniBlocks, err := shardStorage.getProcessedAndPendingMiniBlocksWithScheduled(scenario.metaBlock, scenario.headers, scenario.shardHeader, false)

	require.Nil(t, err)
	sortBootstrapMbsInfo(pendingMiniBlocks)
	require.Equal(t, scenario.expectedPendingMbs, pendingMiniBlocks)
	require.Equal(t, scenario.expectedProcessedMbs, processedMiniBlocks)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithScheduledWrongHeaderType(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	scenario := createPendingAndProcessedMiniBlocksScenario()

	wrongShardHeader := &block.MetaBlock{}
	for k, header := range scenario.headers {
		if bytes.Equal(header.GetPrevHash(), scenario.shardHeader.GetPrevHash()) {
			// replace the header with a wrong headerType
			scenario.headers[k] = wrongShardHeader
			break
		}
	}

	processedMiniBlocks, pendingMiniBlocks, err := shardStorage.getProcessedAndPendingMiniBlocksWithScheduled(scenario.metaBlock, scenario.headers, wrongShardHeader, true)
	require.Nil(t, processedMiniBlocks)
	require.Nil(t, pendingMiniBlocks)
	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithScheduled(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	scenario := createPendingAndProcessedMiniBlocksScenario()
	processedMiniBlocks, pendingMiniBlocks, err := shardStorage.getProcessedAndPendingMiniBlocksWithScheduled(scenario.metaBlock, scenario.headers, scenario.shardHeader, true)

	require.Nil(t, err)
	sortBootstrapMbsInfo(pendingMiniBlocks)
	require.Equal(t, scenario.expectedPendingMbsWithScheduled, pendingMiniBlocks)
	require.Equal(t, scenario.expectedProcessedMbsWithScheduled, processedMiniBlocks)
}

func Test_removeMbFromProcessedList(t *testing.T) {
	t.Parallel()

	mbHash := []byte("hash1")
	mbHash2 := []byte("hash2")
	mbHash3 := []byte("hash3")
	metaHash := []byte("metaHash")
	metaHash2 := []byte("metaHash2")
	processedMbs := []bootstrapStorage.MiniBlocksInMeta{
		{MetaHash: metaHash, MiniBlocksHashes: [][]byte{mbHash}},
		{MetaHash: metaHash2, MiniBlocksHashes: [][]byte{mbHash2, mbHash3}},
	}

	mbHandler := &block.MiniBlockHeader{
		Hash:            mbHash3,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	expectedProcessedMbs := []bootstrapStorage.MiniBlocksInMeta{
		{MetaHash: metaHash, MiniBlocksHashes: [][]byte{mbHash}},
		{MetaHash: metaHash2, MiniBlocksHashes: [][]byte{mbHash2}},
	}

	resultingProcessedMbs := removeMbFromProcessedList(mbHandler, processedMbs)
	require.Equal(t, expectedProcessedMbs, resultingProcessedMbs)
}

func Test_removeMbFromProcessedListLastOne(t *testing.T) {
	t.Parallel()

	mbHash := []byte("hash1")
	metaHash := []byte("metaHash")
	processedMbs := []bootstrapStorage.MiniBlocksInMeta{
		{MetaHash: metaHash, MiniBlocksHashes: [][]byte{mbHash}},
	}

	mbHandler := &block.MiniBlockHeader{
		Hash:            mbHash,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	expectedProcessedMbs := []bootstrapStorage.MiniBlocksInMeta{}

	resultingProcessedMbs := removeMbFromProcessedList(mbHandler, processedMbs)
	require.Equal(t, expectedProcessedMbs, resultingProcessedMbs)
}

func Test_addMbToPendingListNoPreviousEntryForShard(t *testing.T) {
	t.Parallel()

	mbHash := []byte("hash1")
	mbHandler := &block.MiniBlockHeader{
		Hash:            mbHash,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	mbHash1 := []byte("existing hash")
	pendingMbsInfo := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 1, MiniBlocksHashes: [][]byte{mbHash1}},
	}
	expectedPendingMbsInfo := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 1, MiniBlocksHashes: [][]byte{mbHash1}},
		{ShardID: 0, MiniBlocksHashes: [][]byte{mbHash}},
	}

	resultingMbsInfo := addMbToPendingList(mbHandler, pendingMbsInfo)
	require.Equal(t, expectedPendingMbsInfo, resultingMbsInfo)
}

func Test_addMbToPendingListWithPreviousEntryForShard(t *testing.T) {
	t.Parallel()

	mbHash := []byte("hash1")
	mbHandler := &block.MiniBlockHeader{
		Hash:            mbHash,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	mbHash1 := []byte("existing hash")
	mbHash2 := []byte("existing hash 2")
	pendingMbsInfo := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 1, MiniBlocksHashes: [][]byte{mbHash1}},
		{ShardID: 0, MiniBlocksHashes: [][]byte{mbHash2}},
	}
	expectedPendingMbsInfo := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 1, MiniBlocksHashes: [][]byte{mbHash1}},
		{ShardID: 0, MiniBlocksHashes: [][]byte{mbHash2, mbHash}},
	}

	resultingMbsInfo := addMbToPendingList(mbHandler, pendingMbsInfo)
	require.Equal(t, expectedPendingMbsInfo, resultingMbsInfo)
}

func Test_removeMbsFromProcessed(t *testing.T) {
	t.Parallel()

	mb1Hash := []byte("hash1")
	mb2Hash := []byte("hash2")
	mb3Hash := []byte("hash3")
	mb4Hash := []byte("hash4")
	mb5Hash := []byte("hash5")
	mb6Hash := []byte("hash6")

	metaHash1 := []byte("metaHash1")
	metaHash2 := []byte("metaHash2")
	metaHash3 := []byte("metaHash3")
	processedMbs := []bootstrapStorage.MiniBlocksInMeta{
		{MetaHash: metaHash1, MiniBlocksHashes: [][]byte{mb1Hash, mb2Hash}},
		{MetaHash: metaHash2, MiniBlocksHashes: [][]byte{mb3Hash, mb4Hash, mb5Hash}},
		{MetaHash: metaHash3, MiniBlocksHashes: [][]byte{mb6Hash}},
	}

	mb0HashIntra := []byte("mb1 hash")
	mb1HashIntra := []byte("mb2 hash")
	mb2HashIntra := []byte("mb1 hash")
	shard0Mbs := []data.MiniBlockHeaderHandler{
		&block.MiniBlockHeader{
			Hash:            mb5Hash,
			SenderShardID:   1,
			ReceiverShardID: 0,
		},
		&block.MiniBlockHeader{
			Hash:            mb0HashIntra,
			SenderShardID:   0,
			ReceiverShardID: 0,
		},
		&block.MiniBlockHeader{
			Hash:            mb1HashIntra,
			SenderShardID:   0,
			ReceiverShardID: 0,
		},
		&block.MiniBlockHeader{
			Hash:            mb6Hash,
			SenderShardID:   2,
			ReceiverShardID: 0,
		},
	}

	shard1Mbs := []data.MiniBlockHeaderHandler{
		&block.MiniBlockHeader{
			Hash:            mb4Hash,
			SenderShardID:   2,
			ReceiverShardID: 1,
		},
		&block.MiniBlockHeader{
			Hash:            mb2HashIntra,
			SenderShardID:   1,
			ReceiverShardID: 1,
		},
	}

	mapMbHeaderHandlers := map[uint32][]data.MiniBlockHeaderHandler{
		0: shard0Mbs,
		1: shard1Mbs,
	}

	expectedProcessed := []bootstrapStorage.MiniBlocksInMeta{
		{MetaHash: metaHash1, MiniBlocksHashes: [][]byte{mb1Hash, mb2Hash}},
		{MetaHash: metaHash2, MiniBlocksHashes: [][]byte{mb3Hash}},
	}
	processedMbsInfo := removeMbsFromProcessed(processedMbs, mapMbHeaderHandlers)
	require.Equal(t, expectedProcessed, processedMbsInfo)
}

func Test_addMbsToPending(t *testing.T) {
	t.Parallel()

	mb1Sh1To0Hash := []byte("hash1 1 to 0")
	mb2Sh1To0Hash := []byte("hash2 1 to 0")
	mb3MetaTo0Hash := []byte("hash3 meta to 0")
	mb4Sh2To1Hash := []byte("hash4 2 to 1")
	mb5Sh0To1Hash := []byte("hash5 0 to 1")
	mb6Sh1ToMetaHash := []byte("hash6 1 to meta")

	mb1PendingHash := []byte("hashPending1")
	mb2PendingHash := []byte("hashPending2")
	mb3PendingHash := []byte("hashPending3")

	pendingMbs := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 0, MiniBlocksHashes: [][]byte{mb1PendingHash, mb2PendingHash}},
		{ShardID: 1, MiniBlocksHashes: [][]byte{mb3PendingHash}},
	}

	mb1Header1To0 := &block.MiniBlockHeader{
		Hash:            mb1Sh1To0Hash,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}
	mb2Header1To0 := &block.MiniBlockHeader{
		Hash:            mb2Sh1To0Hash,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}
	mb3HeaderMetaTo0 := &block.MiniBlockHeader{
		Hash:            mb3MetaTo0Hash,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
	}
	mb4Header2To1 := &block.MiniBlockHeader{
		Hash:            mb4Sh2To1Hash,
		SenderShardID:   2,
		ReceiverShardID: 1,
	}
	mb5Header0To1 := &block.MiniBlockHeader{
		Hash:            mb5Sh0To1Hash,
		SenderShardID:   0,
		ReceiverShardID: 1,
	}
	mb6Header1ToMeta := &block.MiniBlockHeader{
		Hash:            mb6Sh1ToMetaHash,
		SenderShardID:   1,
		ReceiverShardID: core.MetachainShardId,
	}

	mbsToShard0 := []data.MiniBlockHeaderHandler{mb1Header1To0, mb2Header1To0, mb3HeaderMetaTo0}
	mbsToShard1 := []data.MiniBlockHeaderHandler{mb4Header2To1, mb5Header0To1}
	mbsToMeta := []data.MiniBlockHeaderHandler{mb6Header1ToMeta}

	mapMbHeaderHandlers := map[uint32][]data.MiniBlockHeaderHandler{
		0:                     mbsToShard0,
		1:                     mbsToShard1,
		core.MetachainShardId: mbsToMeta,
	}

	expectedPendingMbs := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 0, MiniBlocksHashes: [][]byte{mb1PendingHash, mb2PendingHash, mb1Sh1To0Hash, mb2Sh1To0Hash, mb3MetaTo0Hash}},
		{ShardID: 1, MiniBlocksHashes: [][]byte{mb3PendingHash, mb4Sh2To1Hash, mb5Sh0To1Hash}},
		{ShardID: core.MetachainShardId, MiniBlocksHashes: [][]byte{mb6Sh1ToMetaHash}},
	}

	pendingMbsInfo := addMbsToPending(pendingMbs, mapMbHeaderHandlers)

	require.Equal(t, expectedPendingMbs, pendingMbsInfo)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksErrorGettingEpochStartShardData(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

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

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

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

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

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

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

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

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

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

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledGetShardHeaderErr(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)

	headers := map[string]data.HeaderHandler{}
	meta := &block.MetaBlock{
		Nonce:      100,
		EpochStart: block.EpochStart{},
	}

	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, false)
	require.Nil(t, bootstrapHeaderInfo)
	require.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledMissingLastCrossMetaHdrHash(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shard0HeaderHash := "shard0 header hash"
	lastFinishedMetaBlock := "last finished meta block"

	headers := map[string]data.HeaderHandler{shard0HeaderHash: &block.Header{Nonce: 100}}
	shardInfo := []block.ShardData{{HeaderHash: []byte(shard0HeaderHash), ShardMiniBlockHeaders: nil, ShardID: 0}}
	epochStartShardData := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte(""))

	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: epochStartShardData,
		},
		ShardInfo: shardInfo,
	}

	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, false)
	require.Nil(t, bootstrapHeaderInfo)
	require.Equal(t, epochStart.ErrMissingHeader, err)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shard0HeaderHash := "shard0 header hash"
	lastFinishedMetaBlock := "last finished meta block"

	headers := map[string]data.HeaderHandler{
		shard0HeaderHash: &block.Header{Nonce: 100}, lastFinishedMetaBlock: &block.Header{Nonce: 99}, // wrong header type
	}
	shardInfo := []block.ShardData{{HeaderHash: []byte(shard0HeaderHash), ShardMiniBlockHeaders: nil, ShardID: 0}}
	epochStartShardData := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte(""))

	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: epochStartShardData,
		},
		ShardInfo: shardInfo,
	}

	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, false)
	require.Nil(t, bootstrapHeaderInfo)
	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledErrorWritingToStorage(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	expectedErr := fmt.Errorf("expected error")
	// Simulate an error when writing to storage with a mock marshaller
	args.marshalizer = &testscommon.MarshalizerStub{MarshalCalled: func(obj interface{}) ([]byte, error) {
		return nil, expectedErr
	}}
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shard0HeaderHash := "shard0 header hash"
	lastFinishedMetaBlock := "last finished meta block"

	headers := map[string]data.HeaderHandler{
		shard0HeaderHash: &block.Header{Nonce: 100}, lastFinishedMetaBlock: &block.MetaBlock{Nonce: 99},
	}
	shardInfo := []block.ShardData{{HeaderHash: []byte(shard0HeaderHash), ShardMiniBlockHeaders: nil, ShardID: 0}}
	epochStartShardData := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte(""))

	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: epochStartShardData,
		},
		ShardInfo: shardInfo,
	}

	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, false)
	require.Nil(t, bootstrapHeaderInfo)
	require.Equal(t, expectedErr, err)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduled(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shard0HeaderHash := "shard0 header hash"
	lastFinishedMetaBlock := "last finished meta block"

	headers := map[string]data.HeaderHandler{
		shard0HeaderHash: &block.Header{Nonce: 100}, lastFinishedMetaBlock: &block.MetaBlock{Nonce: 99},
	}
	shardInfo := []block.ShardData{{HeaderHash: []byte(shard0HeaderHash), ShardMiniBlockHeaders: nil, ShardID: 0}}
	epochStartShardData := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte(""))

	meta := &block.MetaBlock{
		Nonce: 100,
		Epoch: 10,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: epochStartShardData,
		},
		ShardInfo: shardInfo,
	}

	// TODO: Check if Epoch field should also be filled by saveLastCrossNotarizedHeaders, as currently it is not
	expectedBootstrapHeaderInfo := []bootstrapStorage.BootstrapHeaderInfo{
		{ShardId: core.MetachainShardId, Nonce: headers[lastFinishedMetaBlock].GetNonce(), Hash: []byte(lastFinishedMetaBlock)},
	}
	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, false)
	require.Nil(t, err)
	require.Equal(t, expectedBootstrapHeaderInfo, bootstrapHeaderInfo)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithScheduledErrorUpdatingLastCrossMetaHeaders(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shard0HeaderHash := "shard0 header hash"
	lastFinishedMetaBlock := "last finished meta block"

	headers := map[string]data.HeaderHandler{shard0HeaderHash: &block.Header{Nonce: 100}}
	shardInfo := []block.ShardData{{HeaderHash: []byte(shard0HeaderHash), ShardMiniBlockHeaders: nil, ShardID: 0}}
	epochStartShardData := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte(""))

	meta := &block.MetaBlock{
		Nonce: 100,
		Epoch: 10,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: epochStartShardData,
		},
		ShardInfo: shardInfo,
	}

	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, true)
	require.Nil(t, bootstrapHeaderInfo)
	require.Equal(t, epochStart.ErrMissingHeader, err)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithScheduled(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	shard0HeaderHash := "shard0 header hash"
	lastFinishedMetaBlock := "last finished meta block"
	prevMetaHash := "prev metaHlock hash"

	headers := map[string]data.HeaderHandler{
		shard0HeaderHash:      &block.Header{Nonce: 100},
		lastFinishedMetaBlock: &block.MetaBlock{Nonce: 99, PrevHash: []byte(prevMetaHash)},
		prevMetaHash:          &block.MetaBlock{Nonce: 98},
	}
	shardInfo := []block.ShardData{{HeaderHash: []byte(shard0HeaderHash), ShardMiniBlockHeaders: nil, ShardID: 0}}
	epochStartShardData := createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte(shard0HeaderHash))

	meta := &block.MetaBlock{
		Nonce: 100,
		Epoch: 10,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: epochStartShardData,
		},
		ShardInfo: shardInfo,
	}

	expectedBootstrapHeaderInfo := []bootstrapStorage.BootstrapHeaderInfo{
		{ShardId: core.MetachainShardId, Nonce: headers[lastFinishedMetaBlock].GetNonce(), Hash: []byte(lastFinishedMetaBlock)},
	}
	bootstrapHeaderInfo, err := shardStorage.saveLastCrossNotarizedHeaders(meta, headers, true)
	require.Nil(t, err)
	require.Equal(t, expectedBootstrapHeaderInfo, bootstrapHeaderInfo)
}

func Test_updateLastCrossMetaHdrHashIfNeededGetShardHeaderErr(t *testing.T) {
	t.Parallel()

	metaHdrKey := "key2"
	lastCrossMetaHdrHash := ""

	headers := map[string]data.HeaderHandler{
		metaHdrKey: &block.MetaBlock{},
	}
	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte {
			return []byte("unknown hash")
		},
	}

	lastCrossMetaHash, err := updateLastCrossMetaHdrHashIfNeeded(headers, epochStartData, []byte(lastCrossMetaHdrHash))
	require.Equal(t, epochStart.ErrMissingHeader, err)
	require.Nil(t, lastCrossMetaHash)
}

func Test_updateLastCrossMetaHdrHashIfNeededNoFinalizedMetaHashesInShardHeader(t *testing.T) {
	t.Parallel()

	metaHdrKey := "key2"
	shardHdrKey := "key1"
	lastCrossMetaHdrHash := "originalLastCrossMetaHdrHash"

	headers := map[string]data.HeaderHandler{
		lastCrossMetaHdrHash: &block.MetaBlock{
			Nonce: 98,
		},
		metaHdrKey: &block.MetaBlock{
			Nonce: 99,
		},
		shardHdrKey: &block.Header{
			Nonce: 100,
		},
	}
	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte {
			return []byte(shardHdrKey)
		},
	}

	updatedCrossMetaHash, err := updateLastCrossMetaHdrHashIfNeeded(headers, epochStartData, []byte(lastCrossMetaHdrHash))
	require.Nil(t, err)
	require.Equal(t, []byte(lastCrossMetaHdrHash), updatedCrossMetaHash)
}

func Test_updateLastCrossMetaHdrHashIfNeededMissingOneReferencedMetaHeader(t *testing.T) {
	t.Parallel()

	shardHdrKey := "key1"
	metaHdrKey := "key2"
	lastCrossMetaHdrHash := "originalLastCrossMetaHdrHash"
	metaBlockHashes := [][]byte{
		[]byte("metaHdrHash1"),
		[]byte("metaHdrHash2"),
		[]byte("metaHdrHash3"),
	}

	headers := map[string]data.HeaderHandler{
		metaHdrKey: &block.MetaBlock{
			Nonce: 100,
		},
		shardHdrKey: &block.Header{
			Nonce:           100,
			MetaBlockHashes: metaBlockHashes,
		},
		string(metaBlockHashes[1]): &block.MetaBlock{
			Nonce: 99,
		},
		string(metaBlockHashes[2]): &block.MetaBlock{
			Nonce: 98,
		},
		// Missing metaBlockHashes[0]
	}

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte {
			return []byte(shardHdrKey)
		},
	}

	lastCrossMetaHash, err := updateLastCrossMetaHdrHashIfNeeded(headers, epochStartData, []byte(lastCrossMetaHdrHash))
	require.Nil(t, lastCrossMetaHash)
	require.Equal(t, epochStart.ErrMissingHeader, err)
}

func Test_updateLastCrossMetaHdrHashIfNeeded(t *testing.T) {
	t.Parallel()

	shardHdrKey := "key1"
	metaHdrKey := "key2"
	lastCrossMetaHdrHash := "originalLastCrossMetaHdrHash"
	metaBlockHashes := [][]byte{
		[]byte("metaHdrHash1"),
		[]byte("metaHdrHash2"),
	}
	expectedLastCrossMetaHash := []byte("expectedLastCrossMetaHash")

	headers := map[string]data.HeaderHandler{
		metaHdrKey: &block.MetaBlock{
			Nonce: 100,
		},
		shardHdrKey: &block.Header{
			Nonce:           100,
			MetaBlockHashes: metaBlockHashes,
		},
		string(metaBlockHashes[0]): &block.MetaBlock{
			Nonce:    99,
			PrevHash: expectedLastCrossMetaHash,
		},
		string(metaBlockHashes[1]): &block.MetaBlock{
			Nonce: 98,
		},
	}

	epochStartData := &epochStartMocks.EpochStartShardDataStub{
		GetHeaderHashCalled: func() []byte {
			return []byte(shardHdrKey)
		},
	}

	lastCrossMetaHash, err := updateLastCrossMetaHdrHashIfNeeded(headers, epochStartData, []byte(lastCrossMetaHdrHash))
	require.Nil(t, err)
	require.Equal(t, expectedLastCrossMetaHash, lastCrossMetaHash)
}

func Test_getShardHeaderAndMetaHashesHeaderNotFound(t *testing.T) {
	t.Parallel()

	headers := map[string]data.HeaderHandler{
		"k1": &block.Header{},
		"k2": &block.Header{},
	}

	shardHeader, metaHashes, err := getShardHeaderAndMetaHashes(headers, []byte("unknownHash"))
	require.Equal(t, epochStart.ErrMissingHeader, err)
	require.Nil(t, shardHeader)
	require.Nil(t, metaHashes)
}

func Test_getShardHeaderAndMetaHashesWrongHeaderType(t *testing.T) {
	t.Parallel()

	key := "key1"
	headers := map[string]data.HeaderHandler{
		key:    &block.MetaBlock{},
		"key2": &block.Header{},
	}

	shardHeader, metaHashes, err := getShardHeaderAndMetaHashes(headers, []byte(key))
	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)
	require.Nil(t, shardHeader)
	require.Nil(t, metaHashes)
}

func Test_getShardHeaderAndMetaHashes(t *testing.T) {
	t.Parallel()

	shardHdrKey := "key1"
	metaHdrKey := "key2"
	metaBlockHashes := [][]byte{
		[]byte("metaHdrHash1"),
		[]byte("metaHdrHash2"),
		[]byte("metaHdrHash3"),
	}

	headers := map[string]data.HeaderHandler{
		metaHdrKey: &block.MetaBlock{},
		shardHdrKey: &block.Header{
			MetaBlockHashes: metaBlockHashes,
		},
	}

	shardHeader, metaHashes, err := getShardHeaderAndMetaHashes(headers, []byte(shardHdrKey))
	require.Nil(t, err)
	require.Equal(t, shardHeader, headers[shardHdrKey])
	require.Equal(t, metaHashes, headers[shardHdrKey].(data.ShardHeaderHandler).GetMetaBlockHashes())
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
		hasher:             &hashingMocks.HasherMock{},
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

func sortBootstrapMbsInfo(bootstrapMbsInfo []bootstrapStorage.PendingMiniBlocksInfo) {
	sort.Slice(bootstrapMbsInfo, func(i, j int) bool {
		return bootstrapMbsInfo[i].ShardID < bootstrapMbsInfo[j].ShardID
	})
}

type scenarioData struct {
	shardHeader                       *block.Header
	headers                           map[string]data.HeaderHandler
	metaBlock                         *block.MetaBlock
	expectedPendingMbs                []bootstrapStorage.PendingMiniBlocksInfo
	expectedProcessedMbs              []bootstrapStorage.MiniBlocksInMeta
	expectedPendingMbsWithScheduled   []bootstrapStorage.PendingMiniBlocksInfo
	expectedProcessedMbsWithScheduled []bootstrapStorage.MiniBlocksInMeta
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

	expectedPendingMbsWithScheduled := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 0, MiniBlocksHashes: [][]byte{crossMbHeaders[1].Hash, crossMbHeaders[2].Hash, crossMbHeaders[3].Hash, crossMbHeaders[4].Hash, crossMbHeaders[0].Hash}},
	}
	expectedProcessedMbsWithScheduled := []bootstrapStorage.MiniBlocksInMeta{}

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
		shardHeader:                       shardHeader,
		headers:                           headers,
		metaBlock:                         meta,
		expectedPendingMbs:                expectedPendingMiniBlocks,
		expectedProcessedMbs:              expectedProcessedMiniBlocks,
		expectedPendingMbsWithScheduled:   expectedPendingMbsWithScheduled,
		expectedProcessedMbsWithScheduled: expectedProcessedMbsWithScheduled,
	}
}
