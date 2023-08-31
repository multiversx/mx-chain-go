package bootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	epochStartMocks "github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks/epochStart"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShardStorageHandler_ShouldWork(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, err := NewShardStorageHandler(args)

	assert.False(t, check.IfNil(shardStorage))
	assert.Nil(t, err)
}

func TestShardStorageHandler_SaveDataToStorageShardDataNotFound(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: &block.MetaBlock{Epoch: 1},
		PreviousEpochStart:  &block.MetaBlock{Epoch: 1},
		ShardHeader:         &block.Header{Nonce: 1},
	}

	err := shardStorage.SaveDataToStorage(components, components.ShardHeader, false, nil)
	assert.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_SaveDataToStorageMissingHeader(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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

	err := shardStorage.SaveDataToStorage(components, components.ShardHeader, false, nil)
	assert.True(t, errors.Is(err, epochStart.ErrMissingHeader))
}

func TestShardStorageHandler_SaveDataToStorageMissingStorer(t *testing.T) {
	t.Parallel()

	t.Run("missing BootstrapUnit", testShardWithMissingStorer(dataRetriever.BootstrapUnit, 1))
	t.Run("missing BlockHeaderUnit", testShardWithMissingStorer(dataRetriever.BlockHeaderUnit, 1))
	t.Run("missing ShardHdrNonceHashDataUnit", testShardWithMissingStorer(dataRetriever.ShardHdrNonceHashDataUnit, 1))
	t.Run("missing MetaBlockUnit", testShardWithMissingStorer(dataRetriever.MetaBlockUnit, 1)) // saveMetaHdrForEpochTrigger(components.EpochStartMetaBlock)
	t.Run("missing BootstrapUnit", testShardWithMissingStorer(dataRetriever.BootstrapUnit, 2)) // saveMetaHdrForEpochTrigger(components.EpochStartMetaBlock)
	t.Run("missing MetaBlockUnit", testShardWithMissingStorer(dataRetriever.MetaBlockUnit, 2)) // saveMetaHdrForEpochTrigger(components.PreviousEpochStart)
	t.Run("missing BootstrapUnit", testShardWithMissingStorer(dataRetriever.BootstrapUnit, 3)) // saveMetaHdrForEpochTrigger(components.PreviousEpochStart)
}

func testShardWithMissingStorer(missingUnit dataRetriever.UnitType, atCallNumber int) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		defer func() {
			_ = os.RemoveAll("./Epoch_0")
		}()

		counter := 0
		args := createDefaultShardStorageArgs()
		args.CurrentEpoch = 1
		shardStorage, _ := NewShardStorageHandler(args)

		shardStorage.storageService = &storageStubs.ChainStorerStub{
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

		err := shardStorage.SaveDataToStorage(components, components.ShardHeader, false, nil)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), storage.ErrKeyNotFound.Error()))
		require.True(t, strings.Contains(err.Error(), missingUnit.String()))
	}
}

func TestShardStorageHandler_SaveDataToStorage(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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

	err := shardStorage.SaveDataToStorage(components, components.ShardHeader, false, nil)
	assert.Nil(t, err)
}

func Test_getMiniBlockHeadersForDest(t *testing.T) {
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

	shardMbHeaders := getMiniBlockHeadersForDest(metablock, 0)
	assert.Equal(t, shardMbHeaders[string(hash1)], shardMiniBlockHeader)
	assert.NotNil(t, shardMbHeaders[string(hash2)])
}

func TestShardStorageHandler_getCrossProcessedMiniBlockHeadersDestMe(t *testing.T) {
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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

	shardHeader := &block.Header{
		Nonce:            100,
		MiniBlockHeaders: mbs,
	}

	expectedMbs := map[string]data.MiniBlockHeaderHandler{
		string(mb1From1To0.Hash): &mb1From1To0,
		string(mb2From1To0.Hash): &mb2From1To0,
		string(mb3From2To0.Hash): &mb3From2To0,
	}

	processedMbs := shardStorage.getCrossProcessedMiniBlockHeadersDestMe(shardHeader)
	require.Equal(t, processedMbs, expectedMbs)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithScheduledErrorGettingProcessedAndPendingMbs(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

	scenario := createPendingAndProcessedMiniBlocksScenario()
	processedMiniBlocks, pendingMiniBlocks, err := shardStorage.getProcessedAndPendingMiniBlocksWithScheduled(scenario.metaBlock, scenario.headers, scenario.shardHeader, true)

	require.Nil(t, err)
	sortBootstrapMbsInfo(pendingMiniBlocks)
	require.Equal(t, scenario.expectedPendingMbsWithScheduled, pendingMiniBlocks)
	require.Equal(t, scenario.expectedProcessedMbsWithScheduled, processedMiniBlocks)
}

func Test_addMiniBlockToPendingListNoPreviousEntryForShard(t *testing.T) {
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

	resultingMbsInfo := addMiniBlockToPendingList(mbHandler, pendingMbsInfo)
	require.Equal(t, expectedPendingMbsInfo, resultingMbsInfo)
}

func Test_addMiniBlockToPendingListWithPreviousEntryForShard(t *testing.T) {
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

	resultingMbsInfo := addMiniBlockToPendingList(mbHandler, pendingMbsInfo)
	require.Equal(t, expectedPendingMbsInfo, resultingMbsInfo)
}

func Test_addMiniBlocksToPending(t *testing.T) {
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

	mapMbHeaderHandlers := map[string]data.MiniBlockHeaderHandler{
		string(mbsToShard0[0].GetHash()): mbsToShard0[0],
		string(mbsToShard0[1].GetHash()): mbsToShard0[1],
		string(mbsToShard0[2].GetHash()): mbsToShard0[2],
		string(mbsToShard1[0].GetHash()): mbsToShard1[0],
		string(mbsToShard1[1].GetHash()): mbsToShard1[1],
		string(mbsToMeta[0].GetHash()):   mbsToMeta[0],
	}

	expectedPendingMbs := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 0, MiniBlocksHashes: [][]byte{mb1PendingHash, mb2PendingHash, mb1Sh1To0Hash, mb2Sh1To0Hash, mb3MetaTo0Hash}},
		{ShardID: 1, MiniBlocksHashes: [][]byte{mb3PendingHash, mb4Sh2To1Hash, mb5Sh0To1Hash}},
		{ShardID: core.MetachainShardId, MiniBlocksHashes: [][]byte{mb6Sh1ToMetaHash}},
	}

	pendingMbsInfo := addMiniBlocksToPending(pendingMbs, mapMbHeaderHandlers)

	mbFound := 0
	for _, pendingMbInfo := range pendingMbsInfo {
		for _, mbHash := range pendingMbInfo.MiniBlocksHashes {
			mbFound += getExpectedMbHashes(expectedPendingMbs, pendingMbInfo, mbHash)
		}
	}

	require.Equal(t, 9, mbFound)
}

func getExpectedMbHashes(
	expectedPendingMbs []bootstrapStorage.PendingMiniBlocksInfo,
	pendingMbInfo bootstrapStorage.PendingMiniBlocksInfo,
	mbHash []byte,
) int {
	mbFound := 0
	for _, expectedPendingMb := range expectedPendingMbs {
		if expectedPendingMb.ShardID != pendingMbInfo.ShardID {
			continue
		}

		for _, expectedMbHash := range expectedPendingMb.MiniBlocksHashes {
			if bytes.Equal(mbHash, expectedMbHash) {
				mbFound++
			}
		}
	}

	return mbFound
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksErrorGettingEpochStartShardData(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

	meta := &block.MetaBlock{
		Nonce:      100,
		EpochStart: block.EpochStart{},
	}
	headers := map[string]data.HeaderHandler{}

	miniBlocksInMeta, pendingMiniBlocksInfoList, firstPendingMetaBlockHash, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Nil(t, firstPendingMetaBlockHash)
	require.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksMissingHeader(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	lastFinishedMetaBlock := "last finished meta block"
	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

	meta := &block.MetaBlock{
		Nonce: 100,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: createDefaultEpochStartShardData([]byte(lastFinishedMetaBlock), []byte("headerHash")),
		},
	}
	headers := map[string]data.HeaderHandler{}

	miniBlocksInMeta, pendingMiniBlocksInfoList, firstPendingMetaBlockHash, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Nil(t, firstPendingMetaBlockHash)
	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWrongHeader(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	lastFinishedMetaBlockHash := "last finished meta block"
	firstPendingMeta := "first pending meta"
	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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

	miniBlocksInMeta, pendingMiniBlocksInfoList, firstPendingMetaBlockHash, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Nil(t, firstPendingMetaBlockHash)
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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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

	miniBlocksInMeta, pendingMiniBlocksInfoList, firstPendingMetaBlockHash, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Nil(t, firstPendingMetaBlockHash)
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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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

	miniBlocksInMeta, pendingMiniBlocksInfoList, firstPendingMetaBlockHash, err := shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, err)
	require.Len(t, pendingMiniBlocksInfoList, 0)
	require.Len(t, miniBlocksInMeta, 0)
	require.Equal(t, lastFinishedHeaders[0].FirstPendingMetaBlock, firstPendingMetaBlockHash)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithProcessedAndPendingMbs(t *testing.T) {
	t.Parallel()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

	scenario := createPendingAndProcessedMiniBlocksScenario()
	processedMiniBlocks, pendingMiniBlocks, firstPendingMetaBlockHash, err := shardStorage.getProcessedAndPendingMiniBlocks(scenario.metaBlock, scenario.headers)

	require.Nil(t, err)
	require.Equal(t, scenario.expectedPendingMbs, pendingMiniBlocks)
	require.Equal(t, scenario.expectedProcessedMbs, processedMiniBlocks)
	require.Equal(t, scenario.expectedProcessedMbs[0].MetaHash, firstPendingMetaBlockHash)
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledGetShardHeaderErr(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.Marshalizer = &marshallerMock.MarshalizerStub{MarshalCalled: func(obj interface{}) ([]byte, error) {
		return nil, expectedErr
	}}
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithScheduled(t *testing.T) {
	t.Parallel()

	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	args.CurrentEpoch = 1
	shardStorage, _ := NewShardStorageHandler(args)

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
	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
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
	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
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
	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
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

func createDefaultShardStorageArgs() StorageHandlerArgs {
	return StorageHandlerArgs{
		GeneralConfig:                   testscommon.GetGeneralConfig(),
		PrefsConfig:                     config.PreferencesConfig{},
		ShardCoordinator:                &mock.ShardCoordinatorStub{},
		PathManagerHandler:              &testscommon.PathManagerStub{},
		Marshalizer:                     &mock.MarshalizerMock{},
		Hasher:                          &hashingMocks.HasherMock{},
		CurrentEpoch:                    0,
		Uint64Converter:                 &mock.Uint64ByteSliceConverterMock{},
		NodeTypeProvider:                &nodeTypeProviderMock.NodeTypeProviderStub{},
		NodeProcessingMode:              common.Normal,
		ManagedPeersHolder:              &testscommon.ManagedPeersHolderStub{},
		AdditionalStorageServiceCreator: &testscommon.AdditionalStorageServiceFactoryMock{},
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

	txCount := uint32(100)
	crossMbHeaders := []block.MiniBlockHeader{
		{Hash: []byte("mb_1_0_0"), SenderShardID: 1, ReceiverShardID: 0, TxCount: txCount},
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
		{MetaHash: []byte(firstPendingMetaHash), MiniBlocksHashes: [][]byte{crossMbHeaders[0].Hash}, FullyProcessed: []bool{true}, IndexOfLastTxProcessed: []int32{int32(txCount - 1)}},
	}

	expectedPendingMbsWithScheduled := []bootstrapStorage.PendingMiniBlocksInfo{
		{ShardID: 0, MiniBlocksHashes: [][]byte{crossMbHeaders[1].Hash, crossMbHeaders[2].Hash, crossMbHeaders[3].Hash, crossMbHeaders[4].Hash, crossMbHeaders[0].Hash}},
	}
	expectedProcessedMbsWithScheduled := make([]bootstrapStorage.MiniBlocksInMeta, 0)

	headers := map[string]data.HeaderHandler{
		lastFinishedMetaBlockHash: &block.MetaBlock{
			Nonce:    94,
			PrevHash: []byte("before last finished meta"),
		},
		firstPendingMetaHash: &block.MetaBlock{
			Nonce:    95,
			PrevHash: []byte(lastFinishedMetaBlockHash),
			ShardInfo: []block.ShardData{
				{ShardID: 0, HeaderHash: []byte("header hash 1"), ShardMiniBlockHeaders: intraMbHeaders},
				{ShardID: 1, HeaderHash: []byte(prevShardHeaderHash), ShardMiniBlockHeaders: processedMbsHeaders},
				{ShardID: 1, HeaderHash: []byte("header hash 2"), ShardMiniBlockHeaders: []block.MiniBlockHeader{crossMbHeaders[4]}},
				{ShardID: 2, HeaderHash: []byte("header hash 3 "), ShardMiniBlockHeaders: []block.MiniBlockHeader{crossMbHeaders[1], crossMbHeaders[3]}},
				{ShardID: core.MetachainShardId, HeaderHash: []byte("header hash 4"), ShardMiniBlockHeaders: []block.MiniBlockHeader{crossMbHeaders[2]}},
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

func Test_updatePendingMiniBlocksForScheduled(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	shardMiniBlockHeader := block.MiniBlockHeader{SenderShardID: 1, Hash: hash1}
	metablock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					shardMiniBlockHeader,
				},
			},
		},
	}

	referencedMetaBlockHashes := [][]byte{[]byte("meta_hash1"), []byte("meta_hash2")}
	pendingMiniBlocks := make([]bootstrapStorage.PendingMiniBlocksInfo, 0)
	pendingMiniBlocks = append(pendingMiniBlocks, bootstrapStorage.PendingMiniBlocksInfo{
		ShardID:          0,
		MiniBlocksHashes: [][]byte{hash1, hash2},
	})
	headers := make(map[string]data.HeaderHandler)
	headers["meta_hash2"] = metablock

	remainingPendingMiniBlocks, err := updatePendingMiniBlocksForScheduled(referencedMetaBlockHashes, pendingMiniBlocks, headers, 0)
	assert.Nil(t, err)
	require.Equal(t, 1, len(remainingPendingMiniBlocks))
	require.Equal(t, 1, len(remainingPendingMiniBlocks[0].MiniBlocksHashes))
	assert.Equal(t, hash2, remainingPendingMiniBlocks[0].MiniBlocksHashes[0])
}

func Test_getProcessedMiniBlocksForFinishedMeta(t *testing.T) {
	t.Parallel()

	metaHash1 := []byte("metaBlock_hash1")
	metaHash2 := []byte("metaBlock_hash2")
	miniBlockHash := []byte("miniBlock_hash1")
	referencedMetaBlockHashes := [][]byte{metaHash1, metaHash2}

	headers := make(map[string]data.HeaderHandler)

	_, err := getProcessedMiniBlocksForFinishedMeta(referencedMetaBlockHashes, headers, 0)
	assert.True(t, errors.Is(err, epochStart.ErrMissingHeader))

	headers[string(metaHash1)] = &block.Header{}

	_, err = getProcessedMiniBlocksForFinishedMeta(referencedMetaBlockHashes, headers, 0)
	assert.Equal(t, epochStart.ErrWrongTypeAssertion, err)

	headers[string(metaHash1)] = &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						TxCount:         100,
						SenderShardID:   1,
						ReceiverShardID: 0,
						Hash:            miniBlockHash,
					},
				},
			},
		},
	}

	miniBlocksInMeta, err := getProcessedMiniBlocksForFinishedMeta(referencedMetaBlockHashes, headers, 0)
	assert.Nil(t, err)

	require.Equal(t, 1, len(miniBlocksInMeta))
	assert.Equal(t, metaHash1, miniBlocksInMeta[0].MetaHash)

	require.Equal(t, 1, len(miniBlocksInMeta[0].MiniBlocksHashes))
	assert.Equal(t, miniBlockHash, miniBlocksInMeta[0].MiniBlocksHashes[0])

	require.Equal(t, 1, len(miniBlocksInMeta[0].IndexOfLastTxProcessed))
	assert.Equal(t, int32(99), miniBlocksInMeta[0].IndexOfLastTxProcessed[0])

	require.Equal(t, 1, len(miniBlocksInMeta[0].FullyProcessed))
	assert.True(t, miniBlocksInMeta[0].FullyProcessed[0])
}

func Test_updateProcessedMiniBlocksForScheduled(t *testing.T) {
	t.Parallel()

	mbHash1 := []byte("miniBlock_hash1")

	mbHash2 := []byte("miniBlock_hash2")
	mbHeader2 := &block.MiniBlockHeader{}
	_ = mbHeader2.SetIndexOfFirstTxProcessed(10)

	metaBlockHash := []byte("metaBlock_hash1")
	processedMiniBlocks := []bootstrapStorage.MiniBlocksInMeta{
		{
			MetaHash:               metaBlockHash,
			MiniBlocksHashes:       [][]byte{mbHash1, mbHash2},
			FullyProcessed:         []bool{true, false},
			IndexOfLastTxProcessed: []int32{100, 50},
		},
	}

	mapHashMiniBlockHeaders := make(map[string]data.MiniBlockHeaderHandler)
	mapHashMiniBlockHeaders[string(mbHash2)] = mbHeader2

	miniBlocksInMeta, err := updateProcessedMiniBlocksForScheduled(processedMiniBlocks, mapHashMiniBlockHeaders)
	assert.Nil(t, err)

	require.Equal(t, 1, len(miniBlocksInMeta))
	assert.Equal(t, metaBlockHash, miniBlocksInMeta[0].MetaHash)

	require.Equal(t, 2, len(miniBlocksInMeta[0].MiniBlocksHashes))
	assert.Equal(t, mbHash1, miniBlocksInMeta[0].MiniBlocksHashes[0])
	assert.Equal(t, mbHash2, miniBlocksInMeta[0].MiniBlocksHashes[1])

	require.Equal(t, 2, len(miniBlocksInMeta[0].FullyProcessed))
	assert.True(t, miniBlocksInMeta[0].FullyProcessed[0])
	assert.False(t, miniBlocksInMeta[0].FullyProcessed[1])

	require.Equal(t, 2, len(miniBlocksInMeta[0].IndexOfLastTxProcessed))
	assert.Equal(t, int32(100), miniBlocksInMeta[0].IndexOfLastTxProcessed[0])
	assert.Equal(t, int32(9), miniBlocksInMeta[0].IndexOfLastTxProcessed[1])
}

func Test_getPendingMiniBlocksHashes(t *testing.T) {
	mbHash1 := []byte("mbHash1")
	mbHash2 := []byte("mbHash2")
	mbHash3 := []byte("mbHash3")
	mbHash4 := []byte("mbHash4")
	mbHash5 := []byte("mbHash5")
	mbHash6 := []byte("mbHash6")

	pendingMbsInfo := []bootstrapStorage.PendingMiniBlocksInfo{
		{
			ShardID:          0,
			MiniBlocksHashes: [][]byte{mbHash1, mbHash2},
		},
		{
			ShardID:          0,
			MiniBlocksHashes: [][]byte{mbHash3},
		},
		{
			ShardID:          0,
			MiniBlocksHashes: [][]byte{mbHash4, mbHash5, mbHash6},
		},
	}
	expectedPendingMbHashes := [][]byte{mbHash1, mbHash2, mbHash3, mbHash4, mbHash5, mbHash6}

	pendingMbHashes := getPendingMiniBlocksHashes(pendingMbsInfo)
	require.Equal(t, expectedPendingMbHashes, pendingMbHashes)
}

func Test_getProcessedMiniBlockHashesForMetaBlockHash(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hashMeta := []byte("metaHash1")

	shardMiniBlockHeaders := []block.MiniBlockHeader{
		{SenderShardID: 0, ReceiverShardID: 1, Hash: hash1},
		{SenderShardID: 1, ReceiverShardID: 0, Hash: hash2},
	}

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID:               0,
				ShardMiniBlockHeaders: shardMiniBlockHeaders,
			},
		},
	}

	headers := make(map[string]data.HeaderHandler)
	headers[string(hashMeta)] = metaBlock
	expectedProcessedMbs := [][]byte{hash1}

	processedMbs, err := getProcessedMiniBlockHashesForMetaBlockHash(1, hashMeta, headers)

	require.Nil(t, err)
	require.Equal(t, expectedProcessedMbs, processedMbs)
}

func Test_getProcessedMiniBlockHashesForMetaBlockHashMissingHeaderShouldErr(t *testing.T) {
	hashMeta := []byte("hashMeta")
	headers := make(map[string]data.HeaderHandler)

	processedMbs, err := getProcessedMiniBlockHashesForMetaBlockHash(1, hashMeta, headers)

	require.True(t, errors.Is(err, epochStart.ErrMissingHeader))
	require.Nil(t, processedMbs)
}

func Test_getProcessedMiniBlockHashesForMetaBlockHashInvalidHeaderShouldErr(t *testing.T) {
	hashMeta := []byte("hashMeta")
	headers := make(map[string]data.HeaderHandler)
	headers[string(hashMeta)] = &block.Header{}

	processedMbs, err := getProcessedMiniBlockHashesForMetaBlockHash(1, hashMeta, headers)

	require.Equal(t, epochStart.ErrWrongTypeAssertion, err)
	require.Nil(t, processedMbs)
}

func Test_removeHash(t *testing.T) {
	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mbHash3 := []byte("hash3")
	mbHash4 := []byte("hash4")
	mbHash5 := []byte("hash5")
	mbHash6 := []byte("hash6")
	hashes := [][]byte{mbHash1, mbHash2, mbHash3, mbHash4, mbHash5, mbHash6}

	expectedRemoveMiddle := [][]byte{mbHash1, mbHash2, mbHash4, mbHash5, mbHash6}
	hashes = removeHash(hashes, mbHash3)
	require.Equal(t, expectedRemoveMiddle, hashes)

	expectedRemoveFirst := [][]byte{mbHash2, mbHash4, mbHash5, mbHash6}
	hashes = removeHash(hashes, mbHash1)
	require.Equal(t, expectedRemoveFirst, hashes)

	expectedRemoveLast := [][]byte{mbHash2, mbHash4, mbHash5}
	hashes = removeHash(hashes, mbHash6)
	require.Equal(t, expectedRemoveLast, hashes)
}

func Test_removeHashes(t *testing.T) {
	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")
	mbHash3 := []byte("hash3")
	mbHash4 := []byte("hash4")
	mbHash5 := []byte("hash5")
	mbHash6 := []byte("hash6")
	hashes := [][]byte{mbHash1, mbHash2, mbHash3, mbHash4, mbHash5, mbHash6}

	expectedRemoveMiddle := [][]byte{mbHash1, mbHash2, mbHash5, mbHash6}
	middleHashes := [][]byte{mbHash3, mbHash4}
	updatedHashes := removeHashes(hashes, middleHashes)
	require.Equal(t, expectedRemoveMiddle, updatedHashes)

	expectedRemoveFirst := [][]byte{mbHash3, mbHash4, mbHash5, mbHash6}
	firstHashes := [][]byte{mbHash1, mbHash2}
	updatedHashes = removeHashes(hashes, firstHashes)
	require.Equal(t, expectedRemoveFirst, updatedHashes)

	expectedRemoveLast := [][]byte{mbHash1, mbHash2, mbHash3, mbHash4}
	lastHashes := [][]byte{mbHash5, mbHash6}
	updatedHashes = removeHashes(hashes, lastHashes)
	require.Equal(t, expectedRemoveLast, updatedHashes)

	expectedRemoveDifferent := [][]byte{mbHash1, mbHash2, mbHash3, mbHash4, mbHash5, mbHash6}
	different := [][]byte{[]byte("different")}
	updatedHashes = removeHashes(hashes, different)
	require.Equal(t, expectedRemoveDifferent, updatedHashes)
}

func Test_getNeededMetaBlock(t *testing.T) {
	t.Parallel()

	neededMetaBlock, err := getNeededMetaBlock(nil, nil)
	assert.Nil(t, neededMetaBlock)
	assert.True(t, errors.Is(err, epochStart.ErrMissingHeader))

	wrongHash := []byte("wrongHash")
	headers := make(map[string]data.HeaderHandler)
	neededMetaBlock, err = getNeededMetaBlock(wrongHash, headers)
	assert.Nil(t, neededMetaBlock)
	assert.True(t, errors.Is(err, epochStart.ErrMissingHeader))

	hash := []byte("good hash")
	header := &block.Header{}
	headers[string(hash)] = header
	neededMetaBlock, err = getNeededMetaBlock(hash, headers)
	assert.Nil(t, neededMetaBlock)
	assert.True(t, errors.Is(err, epochStart.ErrWrongTypeAssertion))

	metaBlock := &block.MetaBlock{}
	headers[string(hash)] = metaBlock
	neededMetaBlock, err = getNeededMetaBlock(hash, headers)
	assert.Nil(t, err)
	assert.Equal(t, metaBlock, neededMetaBlock)
}

func Test_getProcessedMiniBlocks(t *testing.T) {
	t.Parallel()

	mbHash1 := []byte("hash1")
	mbHash2 := []byte("hash2")

	mbh1 := block.MiniBlockHeader{
		Hash:            mbHash1,
		SenderShardID:   1,
		ReceiverShardID: 0,
		TxCount:         5,
	}
	_ = mbh1.SetIndexOfLastTxProcessed(int32(mbh1.TxCount - 2))
	_ = mbh1.SetConstructionState(int32(block.PartialExecuted))

	mbh2 := block.MiniBlockHeader{
		Hash:            mbHash2,
		SenderShardID:   2,
		ReceiverShardID: 0,
		TxCount:         5,
	}
	_ = mbh2.SetIndexOfLastTxProcessed(int32(mbh2.TxCount - 1))
	_ = mbh2.SetConstructionState(int32(block.Final))

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID:               1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{mbh1},
			},
			{
				ShardID:               2,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{mbh2},
			},
		},
	}

	processedMiniBlocks := make([]bootstrapStorage.MiniBlocksInMeta, 0)
	referencedMetaBlockHash := []byte("hash")

	processedMiniBlocks = getProcessedMiniBlocks(metaBlock, 0, processedMiniBlocks, referencedMetaBlockHash)

	idxOfMiniBlock0, idxOfMiniBlock1 := 0, 1
	if bytes.Equal(processedMiniBlocks[0].MiniBlocksHashes[0], mbHash2) {
		idxOfMiniBlock0, idxOfMiniBlock1 = 1, 0
	}

	require.Equal(t, 1, len(processedMiniBlocks))
	require.Equal(t, 2, len(processedMiniBlocks[0].MiniBlocksHashes))
	require.Equal(t, 2, len(processedMiniBlocks[0].IndexOfLastTxProcessed))
	require.Equal(t, 2, len(processedMiniBlocks[0].FullyProcessed))

	require.Equal(t, referencedMetaBlockHash, processedMiniBlocks[0].MetaHash)
	assert.Equal(t, int32(mbh1.TxCount-2), processedMiniBlocks[0].IndexOfLastTxProcessed[idxOfMiniBlock0])
	assert.Equal(t, int32(mbh1.TxCount-1), processedMiniBlocks[0].IndexOfLastTxProcessed[idxOfMiniBlock1])
	assert.False(t, processedMiniBlocks[0].FullyProcessed[idxOfMiniBlock0])
	assert.True(t, processedMiniBlocks[0].FullyProcessed[idxOfMiniBlock1])
	assert.Equal(t, mbHash1, processedMiniBlocks[0].MiniBlocksHashes[idxOfMiniBlock0])
	assert.Equal(t, mbHash2, processedMiniBlocks[0].MiniBlocksHashes[idxOfMiniBlock1])
}

func Test_setMiniBlocksInfoWithPendingMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("set mini blocks info with pending mini blocks, having reserved field nil", func(t *testing.T) {
		t.Parallel()

		mbsInfo := &miniBlocksInfo{
			miniBlockHashes:              make([][]byte, 0),
			fullyProcessed:               make([]bool, 0),
			indexOfLastTxProcessed:       make([]int32, 0),
			pendingMiniBlocksMap:         make(map[string]struct{}),
			pendingMiniBlocksPerShardMap: make(map[uint32][][]byte),
		}

		mbHash := []byte("x")
		txCount := uint32(100)

		mbh1 := block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: txCount,
		}

		epochStartShardData := &block.EpochStartShardData{
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
			},
		}
		setMiniBlocksInfoWithPendingMiniBlocks(epochStartShardData, mbsInfo)
		assert.Equal(t, 0, len(mbsInfo.miniBlockHashes))
		assert.Equal(t, 0, len(mbsInfo.fullyProcessed))
		assert.Equal(t, 0, len(mbsInfo.indexOfLastTxProcessed))
	})

	t.Run("set mini blocks info with pending mini blocks, having index of last tx processed set before first tx", func(t *testing.T) {
		t.Parallel()

		mbsInfo := &miniBlocksInfo{
			miniBlockHashes:              make([][]byte, 0),
			fullyProcessed:               make([]bool, 0),
			indexOfLastTxProcessed:       make([]int32, 0),
			pendingMiniBlocksMap:         make(map[string]struct{}),
			pendingMiniBlocksPerShardMap: make(map[uint32][][]byte),
		}

		mbHash := []byte("x")
		txCount := uint32(100)

		mbh1 := block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: txCount,
		}

		indexOfLastTxProcessed := int32(-1)
		_ = mbh1.SetIndexOfLastTxProcessed(indexOfLastTxProcessed)

		epochStartShardData := &block.EpochStartShardData{
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
			},
		}
		setMiniBlocksInfoWithPendingMiniBlocks(epochStartShardData, mbsInfo)
		assert.Equal(t, 0, len(mbsInfo.miniBlockHashes))
		assert.Equal(t, 0, len(mbsInfo.fullyProcessed))
		assert.Equal(t, 0, len(mbsInfo.indexOfLastTxProcessed))
	})

	t.Run("set mini blocks info with pending mini blocks, having index of last tx processed set to last tx", func(t *testing.T) {
		t.Parallel()

		mbsInfo := &miniBlocksInfo{
			miniBlockHashes:              make([][]byte, 0),
			fullyProcessed:               make([]bool, 0),
			indexOfLastTxProcessed:       make([]int32, 0),
			pendingMiniBlocksMap:         make(map[string]struct{}),
			pendingMiniBlocksPerShardMap: make(map[uint32][][]byte),
		}

		mbHash := []byte("x")
		txCount := uint32(100)

		mbh1 := block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: txCount,
		}

		indexOfLastTxProcessed := int32(txCount) - 1
		_ = mbh1.SetIndexOfLastTxProcessed(indexOfLastTxProcessed)

		epochStartShardData := &block.EpochStartShardData{
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
			},
		}
		setMiniBlocksInfoWithPendingMiniBlocks(epochStartShardData, mbsInfo)
		assert.Equal(t, 0, len(mbsInfo.miniBlockHashes))
		assert.Equal(t, 0, len(mbsInfo.fullyProcessed))
		assert.Equal(t, 0, len(mbsInfo.indexOfLastTxProcessed))
	})

	t.Run("set mini blocks info with pending mini blocks, having index of last tx processed set to first tx", func(t *testing.T) {
		t.Parallel()

		mbsInfo := &miniBlocksInfo{
			miniBlockHashes:              make([][]byte, 0),
			fullyProcessed:               make([]bool, 0),
			indexOfLastTxProcessed:       make([]int32, 0),
			pendingMiniBlocksMap:         make(map[string]struct{}),
			pendingMiniBlocksPerShardMap: make(map[uint32][][]byte),
		}

		mbHash := []byte("x")
		txCount := uint32(100)

		mbh1 := block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: txCount,
		}

		indexOfLastTxProcessed := int32(0)
		_ = mbh1.SetIndexOfLastTxProcessed(indexOfLastTxProcessed)
		_ = mbh1.SetConstructionState(int32(block.PartialExecuted))

		epochStartShardData := &block.EpochStartShardData{
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
			},
		}
		setMiniBlocksInfoWithPendingMiniBlocks(epochStartShardData, mbsInfo)
		require.Equal(t, 1, len(mbsInfo.miniBlockHashes))
		require.Equal(t, 1, len(mbsInfo.fullyProcessed))
		require.Equal(t, 1, len(mbsInfo.indexOfLastTxProcessed))
		assert.Equal(t, mbHash, mbsInfo.miniBlockHashes[0])
		assert.False(t, mbsInfo.fullyProcessed[0])
		assert.Equal(t, indexOfLastTxProcessed, mbsInfo.indexOfLastTxProcessed[0])
	})

	t.Run("set mini blocks info with pending mini blocks, having index of last tx processed set to penultimate tx", func(t *testing.T) {
		t.Parallel()

		mbsInfo := &miniBlocksInfo{
			miniBlockHashes:              make([][]byte, 0),
			fullyProcessed:               make([]bool, 0),
			indexOfLastTxProcessed:       make([]int32, 0),
			pendingMiniBlocksMap:         make(map[string]struct{}),
			pendingMiniBlocksPerShardMap: make(map[uint32][][]byte),
		}

		mbHash := []byte("x")
		txCount := uint32(100)

		mbh1 := block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: txCount,
		}

		indexOfLastTxProcessed := int32(txCount) - 2
		_ = mbh1.SetIndexOfLastTxProcessed(indexOfLastTxProcessed)
		_ = mbh1.SetConstructionState(int32(block.PartialExecuted))

		epochStartShardData := &block.EpochStartShardData{
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
			},
		}
		setMiniBlocksInfoWithPendingMiniBlocks(epochStartShardData, mbsInfo)
		require.Equal(t, 1, len(mbsInfo.miniBlockHashes))
		require.Equal(t, 1, len(mbsInfo.fullyProcessed))
		require.Equal(t, 1, len(mbsInfo.indexOfLastTxProcessed))
		assert.Equal(t, mbHash, mbsInfo.miniBlockHashes[0])
		assert.False(t, mbsInfo.fullyProcessed[0])
		assert.Equal(t, indexOfLastTxProcessed, mbsInfo.indexOfLastTxProcessed[0])
	})

	t.Run("set mini blocks info with pending mini blocks, having index of last tx processed set somewhere in the middle of the range", func(t *testing.T) {
		t.Parallel()

		mbsInfo := &miniBlocksInfo{
			miniBlockHashes:              make([][]byte, 0),
			fullyProcessed:               make([]bool, 0),
			indexOfLastTxProcessed:       make([]int32, 0),
			pendingMiniBlocksMap:         make(map[string]struct{}),
			pendingMiniBlocksPerShardMap: make(map[uint32][][]byte),
		}

		mbHash := []byte("x")
		txCount := uint32(100)

		mbh1 := block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: txCount,
		}

		indexOfLastTxProcessed := int32(txCount / 2)
		_ = mbh1.SetIndexOfLastTxProcessed(indexOfLastTxProcessed)
		_ = mbh1.SetConstructionState(int32(block.PartialExecuted))

		epochStartShardData := &block.EpochStartShardData{
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
			},
		}
		setMiniBlocksInfoWithPendingMiniBlocks(epochStartShardData, mbsInfo)
		require.Equal(t, 1, len(mbsInfo.miniBlockHashes))
		require.Equal(t, 1, len(mbsInfo.fullyProcessed))
		require.Equal(t, 1, len(mbsInfo.indexOfLastTxProcessed))
		assert.Equal(t, mbHash, mbsInfo.miniBlockHashes[0])
		assert.False(t, mbsInfo.fullyProcessed[0])
		assert.Equal(t, indexOfLastTxProcessed, mbsInfo.indexOfLastTxProcessed[0])
	})
}
