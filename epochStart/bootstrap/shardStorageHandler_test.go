package bootstrap

import (
	"fmt"
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

}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksWithScheduled(t *testing.T) {

}

func Test_removeMbFromProcessedList(t *testing.T) {

}

func Test_addMbToPendingList(t *testing.T) {

}

func Test_removeMbsFromProcessed(t *testing.T) {

}

func Test_addMbsToPending(t *testing.T) {

}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocksErrorGettingEpochStartShardData(t *testing.T) {
	defer func() {
		_ = os.RemoveAll("./Epoch_0")
	}()

	args := createDefaultShardStorageArgs()
	shardStorage, _ := NewShardStorageHandler(args.generalConfig, args.prefsConfig, args.shardCoordinator, args.pathManagerHandler, args.marshalizer, args.hasher, 1, args.uint64Converter, args.nodeTypeProvider)
	meta := &block.MetaBlock{}
	headers := map[string]data.HeaderHandler{

	}

	miniBlocksInMeta, pendingMiniBlocksInfoList, err :=  shardStorage.getProcessedAndPendingMiniBlocks(meta, headers)
	require.Nil(t, miniBlocksInMeta)
	require.Nil(t, pendingMiniBlocksInfoList)
	require.Equal(t, epochStart.ErrEpochStartDataForShardNotFound, err)
}

func TestShardStorageHandler_getProcessedAndPendingMiniBlocks(t *testing.T) {

}

func TestShardStorageHandler_saveLastCrossNotarizedHeadersWithoutScheduledGetShardHeaderErr(t *testing.T) {
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

func createDefaultEpochStartShardData(lastFinishedMetaBlock []byte, shardHeaderHash []byte) []block.EpochStartShardData {
	return []block.EpochStartShardData{
		{
			HeaderHash:            shardHeaderHash,
			ShardID:               0,
			LastFinishedMetaBlock: lastFinishedMetaBlock,
			PendingMiniBlockHeaders: []block.MiniBlockHeader{
				{SenderShardID: 1, ReceiverShardID: 0},
			},
		},
	}
}
