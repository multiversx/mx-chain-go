package metachain

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		genesisBlocks[shardId] = createGenesisBlock(shardId)
	}

	genesisBlocks[core.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(shardId uint32) data.HeaderHandler {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardID:       shardId,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createGenesisMetaBlock() *block.MetaBlock {
	rootHash := []byte("roothash")
	return &block.MetaBlock{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		PubKeysBitmap: rootHash,
		RootHash:      rootHash,
		PrevHash:      rootHash,
	}
}

func createMockEpochStartCreatorArguments() ArgsNewEpochStartData {
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	startHeaders := createGenesisBlocks(shardCoordinator)
	argsNewEpochStartData := ArgsNewEpochStartData{
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &mock.HasherStub{},
		Store:               createMetaStore(),
		DataPool:            dataRetrieverMock.NewPoolsHolderStub(),
		BlockTracker:        mock.NewBlockTrackerMock(shardCoordinator, startHeaders),
		ShardCoordinator:    shardCoordinator,
		EpochStartTrigger:   &mock.EpochStartTriggerStub{},
		RequestHandler:      &testscommon.RequestHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	return argsNewEpochStartData
}

func createMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})
	persist, _ := database.NewlruDB(100000)
	unit, _ := storageunit.NewStorageUnit(cache, persist)

	return unit
}

func createMetaStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())

	return store
}

func TestEpochStartData_NilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.Marshalizer = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilMarshalizer, err)
}

func TestEpochStartData_NilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.Hasher = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilHasher, err)
}

func TestEpochStartData_NilStore(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.Store = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilStorage, err)
}

func TestEpochStartData_NilDataPool(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.DataPool = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestEpochStartData_NilBlockTracker(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.BlockTracker = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilBlockTracker, err)
}

func TestEpochStartData_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.ShardCoordinator = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestEpochStartData_NilRequestHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.RequestHandler = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilRequestHandler, err)
}

func TestEpochStartData_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.EnableEpochsHandler = nil

	esd, err := NewEpochStartData(arguments)
	require.Nil(t, esd)
	require.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestVerifyEpochStartDataForMetablock_NotEpochStartBlock(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()

	esd, _ := NewEpochStartData(arguments)

	metaBlock := &block.MetaBlock{}

	err := esd.VerifyEpochStartDataForMetablock(metaBlock)
	require.NoError(t, err)
}

func TestVerifyEpochStartDataForMetablock_DataDoesNotMatch(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.Hasher = &mock.HasherStub{
		ComputeCalled: func(s string) []byte {
			token := make([]byte, 4)
			_, _ = rand.Read(token)
			return token
		},
	}

	esd, _ := NewEpochStartData(arguments)

	metaBlock := &block.MetaBlock{
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{HeaderHash: []byte("hash")},
			},
		},
	}

	err := esd.VerifyEpochStartDataForMetablock(metaBlock)
	require.Equal(t, process.ErrEpochStartDataDoesNotMatch, err)
}

func TestEpochStartCreator_getLastFinalizedMetaHashForShardMetaHashNotReturnsGenesis(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	epoch, err := NewEpochStartData(arguments)
	require.Nil(t, err)
	require.False(t, check.IfNil(epoch))

	round := uint64(10)

	shardHdr := &block.Header{Round: round}
	last, lastFinal, shardHdrs, err := epoch.lastFinalizedFirstPendingListHeadersForShard(shardHdr)
	assert.Nil(t, last)
	assert.True(t, bytes.Equal(lastFinal, []byte(core.EpochStartIdentifier(0))))
	assert.Equal(t, shardHdr, shardHdrs[0])
	assert.Nil(t, err)
}

func TestEpochStartCreator_getLastFinalizedMetaHashForShardShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return false
		},
	}

	dPool := dataRetrieverMock.NewPoolsHolderStub()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	metaHash1 := []byte("hash1")
	metaHash2 := []byte("hash2")
	mbHash1 := []byte("mb_hash1")
	dPool.HeadersCalled = func() dataRetriever.HeadersPool {
		cs := &mock.HeadersCacherStub{}
		cs.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
			return &block.Header{
				PrevHash:         []byte("hash1"),
				Nonce:            2,
				Round:            2,
				PrevRandSeed:     []byte("roothash"),
				MiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash1, SenderShardID: 1}},
				MetaBlockHashes:  [][]byte{metaHash1, metaHash2},
			}, nil
		}
		return cs
	}

	arguments.DataPool = dPool

	epoch, _ := NewEpochStartData(arguments)
	round := uint64(10)
	nonce := uint64(1)

	shardHdr := &block.Header{
		Round:           round,
		Nonce:           nonce,
		MetaBlockHashes: [][]byte{mbHash1},
	}
	last, lastFinal, shardHdrs, err := epoch.lastFinalizedFirstPendingListHeadersForShard(shardHdr)
	assert.NotNil(t, last)
	assert.NotNil(t, lastFinal)
	assert.NotNil(t, shardHdrs)
	assert.Nil(t, err)
}

func TestEpochStartCreator_CreateEpochStartFromMetaBlockEpochIsNotStarted(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return false
		},
	}

	epoch, _ := NewEpochStartData(arguments)

	epStart, err := epoch.CreateEpochStartData()
	assert.Nil(t, err)

	emptyEpochStart := block.EpochStart{}
	assert.Equal(t, emptyEpochStart, *epStart)
}

func TestMetaProcessor_CreateEpochStartFromMetaBlockShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
	}

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")

	startHeaders := createGenesisBlocks(arguments.ShardCoordinator)
	arguments.BlockTracker = mock.NewBlockTrackerMock(arguments.ShardCoordinator, startHeaders)

	hdr := startHeaders[0].(*block.Header)
	hdr.MetaBlockHashes = [][]byte{hash1, hash2}
	hdr.Nonce = 1
	startHeaders[0] = hdr

	dPool := dataRetrieverMock.NewPoolsHolderStub()
	dPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	metaHash1 := []byte("hash1")
	metaHash2 := []byte("hash2")
	mbHash1 := []byte("mb_hash1")
	dPool.HeadersCalled = func() dataRetriever.HeadersPool {
		cs := &mock.HeadersCacherStub{}
		cs.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
			return &block.Header{
				PrevHash:         []byte("hash1"),
				Nonce:            1,
				Round:            1,
				PrevRandSeed:     []byte("roothash"),
				MiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash1, SenderShardID: 1}},
				MetaBlockHashes:  [][]byte{metaHash1, metaHash2},
			}, nil
		}

		return cs
	}
	arguments.DataPool = dPool
	metaHdrStorage, _ := arguments.Store.GetStorer(dataRetriever.MetaBlockUnit)
	meta1 := &block.MetaBlock{Nonce: 100}

	var hdrs []block.MiniBlockHeader
	hdrs = append(hdrs, block.MiniBlockHeader{
		Hash:            hash1,
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         2,
	})
	hdrs = append(hdrs, block.MiniBlockHeader{
		Hash:            mbHash1,
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         2,
	})
	shardData := block.ShardData{ShardID: 1, ShardMiniBlockHeaders: hdrs}
	meta2 := &block.MetaBlock{Nonce: 101, PrevHash: metaHash1, ShardInfo: []block.ShardData{shardData}}

	marshaledData, _ := arguments.Marshalizer.Marshal(meta1)
	_ = metaHdrStorage.Put(metaHash1, marshaledData)

	marshaledData, _ = arguments.Marshalizer.Marshal(meta2)
	_ = metaHdrStorage.Put(metaHash2, marshaledData)

	epoch, _ := NewEpochStartData(arguments)

	epStart, err := epoch.CreateEpochStartData()
	assert.Nil(t, err)
	assert.NotNil(t, epStart)
	assert.Equal(t, hash1, epStart.LastFinalizedHeaders[0].LastFinishedMetaBlock)
	assert.Equal(t, hash2, epStart.LastFinalizedHeaders[0].FirstPendingMetaBlock)
	assert.Equal(t, 1, len(epStart.LastFinalizedHeaders[0].PendingMiniBlockHeaders))

	err = epoch.VerifyEpochStartDataForMetablock(&block.MetaBlock{EpochStart: *epStart})
	assert.Nil(t, err)
}

func TestMetaProcessor_CreateEpochStartFromMetaBlockEdgeCaseChecking(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
	}
	arguments.Hasher = &hashingMocks.HasherMock{}
	arguments.ShardCoordinator, _ = sharding.NewMultiShardCoordinator(3, core.MetachainShardId)
	arguments.DataPool = dataRetrieverMock.CreatePoolsHolder(1, core.MetachainShardId)

	startHeaders := createGenesisBlocks(arguments.ShardCoordinator)
	blockTracker := mock.NewBlockTrackerMock(arguments.ShardCoordinator, startHeaders)

	mbHashes := [][]byte{[]byte("mb_hash1"), []byte("mb_hash2"), []byte("mb_hash3"), []byte("mb_hash4")}
	metaHdrStorage, _ := arguments.Store.GetStorer(dataRetriever.MetaBlockUnit)
	var mbHdrs1 []block.MiniBlockHeader
	mbHdrs1 = append(mbHdrs1, block.MiniBlockHeader{
		Hash:            mbHashes[0],
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxCount:         2,
	})
	mbHdrs1 = append(mbHdrs1, block.MiniBlockHeader{
		Hash:            mbHashes[1],
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxCount:         2,
	})
	shardData := block.ShardData{ShardID: 0, ShardMiniBlockHeaders: mbHdrs1}

	meta1 := &block.MetaBlock{Nonce: 100, ShardInfo: []block.ShardData{shardData}}
	metaHash1, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, meta1)

	var mbHdrs2 []block.MiniBlockHeader
	mbHdrs2 = append(mbHdrs2, block.MiniBlockHeader{
		Hash:            mbHashes[2],
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxCount:         2,
	})
	mbHdrs2 = append(mbHdrs2, block.MiniBlockHeader{
		Hash:            mbHashes[3],
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxCount:         2,
	})
	shardData = block.ShardData{ShardID: 0, ShardMiniBlockHeaders: mbHdrs2}
	meta2 := &block.MetaBlock{Nonce: 101, PrevHash: metaHash1, ShardInfo: []block.ShardData{shardData}}
	metaHash2, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, meta2)

	marshaledData, _ := arguments.Marshalizer.Marshal(meta1)
	_ = metaHdrStorage.Put(metaHash1, marshaledData)

	marshaledData, _ = arguments.Marshalizer.Marshal(meta2)
	_ = metaHdrStorage.Put(metaHash2, marshaledData)

	shardID := uint32(1)
	prevPrevShardHdr := &block.Header{
		Nonce:     8,
		ShardID:   shardID,
		TimeStamp: 0,
		Round:     8,
		Epoch:     0,
	}
	prevPrevHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, prevPrevShardHdr)
	arguments.DataPool.Headers().AddHeader(prevPrevHash, prevPrevShardHdr)

	shardMBHdrs := append(mbHdrs1, mbHdrs2...)
	prevShardHdr := &block.Header{
		Nonce:            9,
		PrevHash:         prevPrevHash,
		ShardID:          shardID,
		Round:            9,
		Epoch:            0,
		MiniBlockHeaders: shardMBHdrs,
		MetaBlockHashes:  [][]byte{metaHash1, metaHash2},
	}
	prevShardHdrHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, prevShardHdr)
	arguments.DataPool.Headers().AddHeader(prevShardHdrHash, prevShardHdr)

	shardHdr := &block.Header{
		Nonce:    10,
		ShardID:  shardID,
		Round:    10,
		Epoch:    0,
		PrevHash: prevShardHdrHash,
	}
	shardHdrHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, shardHdr)
	arguments.DataPool.Headers().AddHeader(shardHdrHash, shardHdr)

	blockTracker.AddCrossNotarizedHeader(shardID, shardHdr, shardHdrHash)
	arguments.BlockTracker = blockTracker

	epoch, _ := NewEpochStartData(arguments)

	epStart, err := epoch.CreateEpochStartData()
	assert.Nil(t, err)
	assert.NotNil(t, epStart)
	assert.Equal(t, metaHash1, epStart.LastFinalizedHeaders[shardID].LastFinishedMetaBlock)
	assert.Equal(t, metaHash2, epStart.LastFinalizedHeaders[shardID].FirstPendingMetaBlock)
	assert.Equal(t, 0, len(epStart.LastFinalizedHeaders[shardID].PendingMiniBlockHeaders))

	err = epoch.VerifyEpochStartDataForMetablock(&block.MetaBlock{EpochStart: *epStart})
	assert.Nil(t, err)
}

func TestEpochStartCreator_computeStillPending(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartCreatorArguments()
	epoch, _ := NewEpochStartData(arguments)

	shardHdrs := make([]data.HeaderHandler, 0)
	miniBlockHeaders := make(map[string]block.MiniBlockHeader)
	mbHash1 := []byte("miniBlock_hash1")
	mbHash2 := []byte("miniBlock_hash2")
	mbHash3 := []byte("miniBlock_hash3")
	mbHeader1 := block.MiniBlockHeader{Hash: mbHash1, TxCount: 3}
	mbHeader2 := block.MiniBlockHeader{Hash: mbHash2}
	mbHeader3 := block.MiniBlockHeader{Hash: mbHash3, TxCount: 10}

	_ = mbHeader1.SetConstructionState(int32(block.Final))
	_ = mbHeader1.SetIndexOfFirstTxProcessed(0)
	_ = mbHeader1.SetIndexOfLastTxProcessed(2)

	_ = mbHeader3.SetConstructionState(int32(block.PartialExecuted))
	_ = mbHeader3.SetIndexOfFirstTxProcessed(1)
	_ = mbHeader3.SetIndexOfLastTxProcessed(3)

	miniBlockHeaders[string(mbHash1)] = mbHeader1
	miniBlockHeaders[string(mbHash2)] = mbHeader2
	miniBlockHeaders[string(mbHash3)] = mbHeader3

	mbh1 := block.MiniBlockHeader{
		Hash: mbHash1,
	}
	mbh2 := block.MiniBlockHeader{
		Hash: []byte("miniBlock_hash_missing"),
	}
	mbh3 := block.MiniBlockHeader{
		Hash: mbHash3,
	}

	_ = mbh3.SetConstructionState(int32(block.PartialExecuted))
	_ = mbh3.SetIndexOfFirstTxProcessed(4)
	_ = mbh3.SetIndexOfLastTxProcessed(8)

	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbh1, mbh2, mbh3},
	}

	shardHdrs = append(shardHdrs, header)

	stillPending := epoch.computeStillPending(0, shardHdrs, miniBlockHeaders)
	require.Equal(t, 2, len(stillPending))

	assert.Equal(t, mbHash2, stillPending[0].Hash)
	assert.Equal(t, mbHash3, stillPending[1].Hash)

	assert.Equal(t, int32(-1), stillPending[0].GetIndexOfFirstTxProcessed())
	assert.Equal(t, int32(-1), stillPending[0].GetIndexOfLastTxProcessed())

	assert.Equal(t, int32(4), stillPending[1].GetIndexOfFirstTxProcessed())
	assert.Equal(t, int32(8), stillPending[1].GetIndexOfLastTxProcessed())
}

func Test_initIndexesOfProcessedTxs(t *testing.T) {
	t.Parallel()

	miniBlockHeaders := make(map[string]block.MiniBlockHeader)
	mbh1 := block.MiniBlockHeader{
		TxCount: 5,
	}
	_ = mbh1.SetIndexOfFirstTxProcessed(1)
	_ = mbh1.SetIndexOfLastTxProcessed(2)

	mbh2 := block.MiniBlockHeader{
		TxCount: 5,
	}

	miniBlockHeaders["mbHash1"] = mbh1
	miniBlockHeaders["mbHash2"] = mbh2

	arguments := createMockEpochStartCreatorArguments()
	epoch, _ := NewEpochStartData(arguments)
	epoch.initIndexesOfProcessedTxs(miniBlockHeaders, 0)

	mbh := miniBlockHeaders["mbHash1"]
	assert.Equal(t, int32(1), mbh.GetIndexOfFirstTxProcessed())
	assert.Equal(t, int32(2), mbh.GetIndexOfLastTxProcessed())

	mbh = miniBlockHeaders["mbHash2"]
	assert.Equal(t, int32(-1), mbh.GetIndexOfFirstTxProcessed())
	assert.Equal(t, int32(-1), mbh.GetIndexOfLastTxProcessed())
}

func Test_computeStillPendingInShardHeader(t *testing.T) {
	t.Parallel()

	mbHash1 := []byte("mbHash1")
	mbHash2 := []byte("mbHash2")
	mbHash3 := []byte("mbHash3")

	mbh1 := block.MiniBlockHeader{
		TxCount: 6,
		Hash:    mbHash1,
	}

	mbh2 := block.MiniBlockHeader{
		TxCount: 6,
		Hash:    mbHash2,
	}
	_ = mbh2.SetConstructionState(int32(block.Final))

	mbh3 := block.MiniBlockHeader{
		TxCount: 6,
		Hash:    mbHash3,
	}
	oldIndexOfFirstTxProcessed := int32(1)
	oldIndexOfLastTxProcessed := int32(2)
	_ = mbh3.SetConstructionState(int32(block.PartialExecuted))
	_ = mbh3.SetIndexOfFirstTxProcessed(oldIndexOfFirstTxProcessed)
	_ = mbh3.SetIndexOfLastTxProcessed(oldIndexOfLastTxProcessed)

	shardHdr := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbh1, mbh2, mbh3},
	}

	newIndexOfFirstTxProcessed := int32(3)
	newIndexOfLastTxProcessed := int32(4)
	_ = shardHdr.MiniBlockHeaders[2].SetIndexOfFirstTxProcessed(newIndexOfFirstTxProcessed)
	_ = shardHdr.MiniBlockHeaders[2].SetIndexOfLastTxProcessed(newIndexOfLastTxProcessed)

	miniBlockHeaders := make(map[string]block.MiniBlockHeader)
	miniBlockHeaders[string(mbHash2)] = mbh2
	miniBlockHeaders[string(mbHash3)] = mbh3

	assert.Equal(t, 2, len(miniBlockHeaders))

	arguments := createMockEpochStartCreatorArguments()
	epoch, _ := NewEpochStartData(arguments)
	epoch.computeStillPendingInShardHeader(shardHdr, miniBlockHeaders, 0)
	assert.Equal(t, 1, len(miniBlockHeaders))

	_, ok := miniBlockHeaders[string(mbHash2)]
	assert.False(t, ok)

	mbh, ok := miniBlockHeaders[string(mbHash3)]
	require.True(t, ok)

	assert.Equal(t, newIndexOfFirstTxProcessed, mbh.GetIndexOfFirstTxProcessed())
	assert.Equal(t, newIndexOfLastTxProcessed, mbh.GetIndexOfLastTxProcessed())
}

func Test_updateIndexesOfProcessedTxsEdgeCaseWithIndex0(t *testing.T) {
	t.Parallel()

	reserved := block.MiniBlockHeaderReserved{
		ExecutionType:           block.Normal,
		IndexOfFirstTxProcessed: -1,
		IndexOfLastTxProcessed:  -1,
	}
	reservedBytes, err := reserved.Marshal()
	assert.Nil(t, err)

	mbHeader := block.MiniBlockHeader{
		TxCount:  132,
		Reserved: reservedBytes,
	}
	miniblockHeaders := make(map[string]block.MiniBlockHeader)
	mbHash := "mb hash"

	reserved = block.MiniBlockHeaderReserved{
		ExecutionType:           block.Normal,
		State:                   block.PartialExecuted,
		IndexOfFirstTxProcessed: 0,
		IndexOfLastTxProcessed:  0,
	}
	reservedBytes, err = reserved.Marshal()
	assert.Nil(t, err)

	shardMiniblockHeader := &block.MiniBlockHeader{
		Hash:            []byte(mbHash),
		SenderShardID:   1,
		ReceiverShardID: 2,
		TxCount:         132,
		Type:            block.TxBlock,
		Reserved:        reservedBytes,
	}

	arguments := createMockEpochStartCreatorArguments()
	epoch, _ := NewEpochStartData(arguments)
	epoch.updateIndexesOfProcessedTxs(mbHeader, shardMiniblockHeader, mbHash, 2, miniblockHeaders)

	require.Equal(t, 1, len(miniblockHeaders))
	processedMbHeader, found := miniblockHeaders[mbHash]
	require.True(t, found)
	assert.True(t, len(processedMbHeader.Reserved) > 0)
	assert.Equal(t, int32(0), processedMbHeader.GetIndexOfFirstTxProcessed())
	assert.Equal(t, int32(0), processedMbHeader.GetIndexOfLastTxProcessed())
}

func Test_setIndexOfFirstAndLastTxProcessedShouldNotSetReserved(t *testing.T) {
	t.Parallel()

	partialExecutionEnableEpoch := uint32(5)

	arguments := createMockEpochStartCreatorArguments()
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			if flag == common.MiniBlockPartialExecutionFlag {
				return partialExecutionEnableEpoch
			}
			return 0
		},
	}
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochCalled: func() uint32 {
			return partialExecutionEnableEpoch - 1
		},
	}

	mbHeader := &block.MiniBlockHeader{}

	epoch, _ := NewEpochStartData(arguments)
	epoch.setIndexOfFirstAndLastTxProcessed(mbHeader, 1, 2)

	require.Nil(t, mbHeader.GetReserved())
}

func Test_setIndexOfFirstAndLastTxProcessedShouldSetReserved(t *testing.T) {
	t.Parallel()

	partialExecutionEnableEpoch := uint32(5)

	arguments := createMockEpochStartCreatorArguments()
	arguments.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			if flag == common.MiniBlockPartialExecutionFlag {
				return partialExecutionEnableEpoch
			}
			return 0
		},
	}
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochCalled: func() uint32 {
			return partialExecutionEnableEpoch
		},
	}

	mbHeader := &block.MiniBlockHeader{}

	epoch, _ := NewEpochStartData(arguments)
	epoch.setIndexOfFirstAndLastTxProcessed(mbHeader, 1, 2)

	require.NotNil(t, mbHeader.GetReserved())
}
