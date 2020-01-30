package track_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
	processBlock "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for ShardID := uint32(0); ShardID < shardCoordinator.NumberOfShards(); ShardID++ {
		genesisBlocks[ShardID] = createGenesisShardHeader(ShardID)
	}

	genesisBlocks[core.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisShardHeader(ShardID uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:         0,
		Round:         0,
		Signature:     rootHash,
		RandSeed:      rootHash,
		PrevRandSeed:  rootHash,
		ShardId:       ShardID,
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

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateStorageUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateStorageUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateStorageUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateStorageUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateStorageUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateStorageUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, generateStorageUnit())
	return store
}

func generateStorageUnit() storage.Storer {
	memDB := memorydb.New()

	storer, _ := storageUnit.NewStorageUnit(
		generateTestCache(),
		memDB,
	)

	return storer
}

func generateTestCache() storage.Cacher {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 1000, 1)
	return cache
}

func CreateShardTrackerMockArguments() track.ArgShardTracker {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	argsHeaderValidator := processBlock.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := processBlock.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgShardTracker{
		ArgBaseTracker: track.ArgBaseTracker{
			Hasher:           &mock.HasherMock{},
			HeaderValidator:  headerValidator,
			Marshalizer:      &mock.MarshalizerMock{},
			RequestHandler:   &mock.RequestHandlerStub{},
			Rounder:          &mock.RounderMock{},
			ShardCoordinator: shardCoordinatorMock,
			Store:            initStore(),
			StartHeaders:     genesisBlocks,
		},
		PoolsHolder: mock.NewPoolsHolderMock(),
	}

	return arguments
}

func CreateMetaTrackerMockArguments() track.ArgMetaTracker {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinatorMock.CurrentShard = core.MetachainShardId
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	argsHeaderValidator := processBlock.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := processBlock.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgMetaTracker{
		ArgBaseTracker: track.ArgBaseTracker{
			Hasher:           &mock.HasherMock{},
			HeaderValidator:  headerValidator,
			Marshalizer:      &mock.MarshalizerMock{},
			RequestHandler:   &mock.RequestHandlerStub{},
			Rounder:          &mock.RounderMock{},
			ShardCoordinator: shardCoordinatorMock,
			Store:            initStore(),
			StartHeaders:     genesisBlocks,
		},
		PoolsHolder: mock.NewPoolsHolderMock(),
	}

	return arguments
}

func CreateBaseTrackerMockArguments() track.ArgBaseTracker {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	argsHeaderValidator := processBlock.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := processBlock.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgBaseTracker{
		Hasher:           &mock.HasherMock{},
		HeaderValidator:  headerValidator,
		Marshalizer:      &mock.MarshalizerMock{},
		RequestHandler:   &mock.RequestHandlerStub{},
		Rounder:          &mock.RounderMock{},
		ShardCoordinator: shardCoordinatorMock,
		Store:            initStore(),
		StartHeaders:     genesisBlocks,
	}

	return arguments
}

func TestNewBlockTrack_ShouldErrCheckTrackerNilParameters(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.Hasher = nil
	sbt, err := track.NewShardBlockTrack(shardArguments)

	assert.NotNil(t, err)
	assert.Nil(t, sbt)

	metaArguments := CreateMetaTrackerMockArguments()
	metaArguments.Hasher = nil
	mbt, err := track.NewMetaBlockTrack(metaArguments)

	assert.NotNil(t, err)
	assert.Nil(t, mbt)
}

func TestNewBlockTrack_ShouldErrNilPoolsHolder(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.PoolsHolder = nil
	sbt, err := track.NewShardBlockTrack(shardArguments)

	assert.Equal(t, process.ErrNilPoolsHolder, err)
	assert.Nil(t, sbt)

	metaArguments := CreateMetaTrackerMockArguments()
	metaArguments.PoolsHolder = nil
	mbt, err := track.NewMetaBlockTrack(metaArguments)

	assert.Equal(t, process.ErrNilPoolsHolder, err)
	assert.Nil(t, mbt)
}

func TestNewBlockTrack_ShouldErrNilHeadersDataPool(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.PoolsHolder = &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return nil
		},
	}
	sbt, err := track.NewShardBlockTrack(shardArguments)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, sbt)

	metaArguments := CreateShardTrackerMockArguments()
	metaArguments.PoolsHolder = &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return nil
		},
	}
	mbt, err := track.NewShardBlockTrack(metaArguments)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, mbt)
}

func TestNewBlockTrack_ShouldErrNotarizedHeadersSliceIsNil(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.StartHeaders = nil
	sbt, err := track.NewShardBlockTrack(shardArguments)

	assert.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
	assert.Nil(t, sbt)

	metaArguments := CreateMetaTrackerMockArguments()
	metaArguments.StartHeaders = nil
	mbt, err := track.NewMetaBlockTrack(metaArguments)

	assert.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
	assert.Nil(t, mbt)
}

func TestNewBlockTrack_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, err := track.NewShardBlockTrack(shardArguments)

	assert.Nil(t, err)
	assert.NotNil(t, sbt)

	metaArguments := CreateShardTrackerMockArguments()
	mbt, err := track.NewShardBlockTrack(metaArguments)

	assert.Nil(t, err)
	assert.NotNil(t, mbt)
}

func TestGetSelfHeaders_ShouldReturnEmptySliceWhenErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)
	headerInfo := sbt.GetSelfHeaders(&block.Header{})

	assert.Equal(t, 0, len(headerInfo))

	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)
	headerInfo = mbt.GetSelfHeaders(&block.MetaBlock{})

	assert.Equal(t, 0, len(headerInfo))
}

func TestShardGetSelfHeaders_ShouldReturnEmptySliceWhenNoHeadersForSelfShard(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{{ShardID: 1}},
	}
	headerInfo := sbt.GetSelfHeaders(metaBlock)

	assert.Equal(t, 0, len(headerInfo))
}

func TestShardGetSelfHeaders_ShouldReturnEmptySliceWhenErrGetShardHeader(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{{ShardID: 0}},
	}
	headerInfo := sbt.GetSelfHeaders(metaBlock)

	assert.Equal(t, 0, len(headerInfo))
}

func TestShardGetSelfHeaders_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.PoolsHolder = &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.Header{}, nil
				},
			}
		},
	}
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{{ShardID: 0}},
	}
	headerInfo := sbt.GetSelfHeaders(metaBlock)

	assert.Equal(t, 1, len(headerInfo))
}

func TestMetaGetSelfHeaders_ShouldReturnEmptySliceWhenErrGetMetaHeader(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	header := &block.Header{
		MetaBlockHashes: [][]byte{[]byte("hash")},
	}
	headerInfo := mbt.GetSelfHeaders(header)

	assert.Equal(t, 0, len(headerInfo))
}

func TestMetaGetSelfHeaders_ShouldWork(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	metaArguments.PoolsHolder = &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.MetaBlock{}, nil
				},
			}
		},
	}
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	header := &block.Header{
		MetaBlockHashes: [][]byte{[]byte("hash")},
	}
	headerInfo := mbt.GetSelfHeaders(header)

	assert.Equal(t, 1, len(headerInfo))
}

func TestShardComputeLongestSelfChain_ShouldReturnNilWhenErrGetLastNotarizedHeader(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.CleanupHeadersBehindNonce(core.MetachainShardId, 1, 1)
	_, _, headers, _ := sbt.ComputeLongestSelfChain()

	assert.Nil(t, headers)
}

func TestShardComputeLongestSelfChain_ShouldReturnEmptySliceWhenComputeLongestChainReturnNil(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	_, _, headers, _ := sbt.ComputeLongestSelfChain()

	assert.Equal(t, 0, len(headers))
}

func TestShardComputeLongestSelfChain_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeader := shardArguments.StartHeaders[shardArguments.ShardCoordinator.SelfId()]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	hdr1 := &block.Header{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr1)

	hdr2 := &block.Header{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr2)

	hdr3 := &block.Header{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr3)

	sbt.AddTrackedHeader(hdr1, hdr1Hash)
	sbt.AddTrackedHeader(hdr2, hdr2Hash)
	sbt.AddTrackedHeader(hdr3, hdr3Hash)

	lastNotarizedHeader, lastNotarizedHeaderHash, headers, hashes := sbt.ComputeLongestSelfChain()

	require.Equal(t, 2, len(headers))
	assert.Equal(t, startHeaderHash, lastNotarizedHeaderHash)
	assert.Equal(t, hashes[0], hdr1Hash)
	assert.Equal(t, hashes[1], hdr2Hash)
	assert.Equal(t, startHeader, lastNotarizedHeader)
	assert.Equal(t, headers[0], hdr1)
	assert.Equal(t, headers[1], hdr2)
}

func TestMetaComputeLongestSelfChain_ShouldReturnNilWhenErrGetLastNotarizedHeader(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	mbt.CleanupHeadersBehindNonce(core.MetachainShardId, 1, 1)
	_, _, headers, _ := mbt.ComputeLongestSelfChain()

	assert.Nil(t, headers)
}

func TestMetaComputeLongestSelfChain_ShouldReturnEmptySliceWhenComputeLongestChainReturnNil(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	_, _, headers, _ := mbt.ComputeLongestSelfChain()

	assert.Equal(t, 0, len(headers))
}

func TestMetaComputeLongestSelfChain_ShouldWork(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	startHeader := metaArguments.StartHeaders[metaArguments.ShardCoordinator.SelfId()]
	startHeaderHash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, startHeader)

	hdr1 := &block.MetaBlock{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr1)

	hdr2 := &block.MetaBlock{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr2)

	hdr3 := &block.MetaBlock{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr3)

	mbt.AddTrackedHeader(hdr1, hdr1Hash)
	mbt.AddTrackedHeader(hdr2, hdr2Hash)
	mbt.AddTrackedHeader(hdr3, hdr3Hash)

	lastNotarizedHeader, lastNotarizedHeaderHash, headers, hashes := mbt.ComputeLongestSelfChain()

	require.Equal(t, 2, len(headers))
	assert.Equal(t, startHeaderHash, lastNotarizedHeaderHash)
	assert.Equal(t, hashes[0], hdr1Hash)
	assert.Equal(t, hashes[1], hdr2Hash)
	assert.Equal(t, startHeader, lastNotarizedHeader)
	assert.Equal(t, headers[0], hdr1)
	assert.Equal(t, headers[1], hdr2)
}

func TestComputePendingMiniBlockHeaders_ShouldReturnZeroWhenHeadersSliceIsEmpty(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()

	sbt, _ := track.NewShardBlockTrack(shardArguments)
	sbt.ComputeNumPendingMiniBlocks([]data.HeaderHandler{})

	assert.Equal(t, uint32(0), sbt.GetNumPendingMiniBlocks(shardArguments.ShardCoordinator.SelfId()))
}

func TestComputePendingMiniBlockHeaders_ShouldReturnZeroWhenErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.ComputeNumPendingMiniBlocks([]data.HeaderHandler{&block.Header{}})

	assert.Equal(t, uint32(0), sbt.GetNumPendingMiniBlocks(shardArguments.ShardCoordinator.SelfId()))
}

func TestComputePendingMiniBlockHeaders_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.ComputeNumPendingMiniBlocks([]data.HeaderHandler{&block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID:              0,
				NumPendingMiniBlocks: 2,
			},
		}}})

	assert.Equal(t, uint32(2), sbt.GetNumPendingMiniBlocks(0))
}

func TestReceivedHeader_ShouldAddMetaBlockToTrackedHeaders(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{Nonce: 1}
	metaBlockHash := []byte("hash")
	sbt.ReceivedHeader(metaBlock, metaBlockHash)
	headers, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	require.Equal(t, 1, len(headers))
	assert.Equal(t, metaBlock, headers[0])
}

func TestReceivedHeader_ShouldAddShardHeaderToTrackedHeaders(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{Nonce: 1}
	headerHash := []byte("hash")
	sbt.ReceivedHeader(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	require.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestReceivedShardHeader_ShouldReturnWhenErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{Nonce: 1}
	metaBlockHash := []byte("hash")
	sbt.ReceivedShardHeader(metaBlock, metaBlockHash)
	headers, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	assert.Equal(t, 0, len(headers))
}

func TestReceivedShardHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{Nonce: 1}
	headerHash := []byte("hash")
	sbt.ReceivedShardHeader(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	require.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestReceivedMetaBlock_ShouldReturnWhenErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{Nonce: 1}
	headerHash := []byte("hash")
	sbt.ReceivedMetaBlock(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 0, len(headers))
}

func TestReceivedMetaBlock_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{Nonce: 1}
	metaBlockHash := []byte("hash")
	sbt.ReceivedMetaBlock(metaBlock, metaBlockHash)
	headers, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	require.Equal(t, 1, len(headers))
	assert.Equal(t, metaBlock, headers[0])
}

func TestAddHeader_ShouldNotAddIfItAlreadyExist(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")

	sbt.AddHeader(header, headerHash)
	sbt.AddHeader(header, headerHash)

	headers, _ := sbt.GetTrackedHeaders(shardArguments.ShardCoordinator.SelfId())

	require.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestAddHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	hdr1 := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	hdr1Hash := []byte("hash1")

	hdr2 := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   2,
	}
	hdr2Hash := []byte("hash2")

	sbt.AddHeader(hdr1, hdr1Hash)
	sbt.AddHeader(hdr2, hdr2Hash)

	headers, _ := sbt.GetTrackedHeaders(shardArguments.ShardCoordinator.SelfId())

	require.Equal(t, 2, len(headers))
	assert.Equal(t, hdr1, headers[0])
	assert.Equal(t, hdr2, headers[1])
}

func TestAddCrossNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash := []byte("hash")

	sbt.AddCrossNotarizedHeader(metaBlock.GetShardID(), metaBlock, metaBlockHash)
	lastCrossNotarizedHeader, _, _ := sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())

	assert.Equal(t, metaBlock, lastCrossNotarizedHeader)
}

func TestAddSelfNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")

	sbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	lastSelfNotarizedHeader, _, _ := sbt.GetLastSelfNotarizedHeader(header.GetShardID())

	assert.Equal(t, header, lastSelfNotarizedHeader)
}

func TestAddTrackedHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		Nonce: 1,
	}
	headerHash := []byte("hash")

	sbt.AddTrackedHeader(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	require.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestCleanupHeadersBehindNonce_ShouldCleanSelfNotarizedHeaders(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")

	sbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	sbt.AddTrackedHeader(header, headerHash)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash := []byte("hash")

	sbt.AddCrossNotarizedHeader(metaBlock.GetShardID(), metaBlock, metaBlockHash)
	sbt.AddTrackedHeader(metaBlock, metaBlockHash)

	sbt.CleanupHeadersBehindNonce(shardArguments.ShardCoordinator.SelfId(), 2, 2)

	lastSelfNotarizedHeader, _, _ := sbt.GetLastSelfNotarizedHeader(header.GetShardID())
	lastCrossNotarizedHeader, _, _ := sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	trackedHeadersForSelfShard, _ := sbt.GetTrackedHeaders(header.GetShardID())
	trackedHeadersForCrossShard, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	assert.Nil(t, lastSelfNotarizedHeader)
	assert.Equal(t, metaBlock, lastCrossNotarizedHeader)
	assert.Equal(t, 0, len(trackedHeadersForSelfShard))
	assert.Equal(t, 1, len(trackedHeadersForCrossShard))
}

func TestCleanupHeadersBehindNonce_ShouldCleanCrossNotarizedHeaders(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	sbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	sbt.AddTrackedHeader(header, headerHash)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash := []byte("hash")
	sbt.AddCrossNotarizedHeader(metaBlock.GetShardID(), metaBlock, metaBlockHash)
	sbt.AddTrackedHeader(metaBlock, metaBlockHash)

	sbt.CleanupHeadersBehindNonce(core.MetachainShardId, 2, 2)

	lastSelfNotarizedHeader, _, _ := sbt.GetLastSelfNotarizedHeader(header.GetShardID())
	lastCrossNotarizedHeader, _, _ := sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	trackedHeadersForSelfShard, _ := sbt.GetTrackedHeaders(header.GetShardID())
	trackedHeadersForCrossShard, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	assert.Equal(t, header, lastSelfNotarizedHeader)
	assert.Nil(t, lastCrossNotarizedHeader)
	assert.Equal(t, 1, len(trackedHeadersForSelfShard))
	assert.Equal(t, 0, len(trackedHeadersForCrossShard))
}

func TestCleanupTrackedHeadersBehindNonce_ShouldReturnWhenNonceIsZeroOrShardNotExist(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	sbt.AddTrackedHeader(header, headerHash)

	sbt.CleanupTrackedHeadersBehindNonce(header.GetShardID(), 0)
	trackedHeaders, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 1, len(trackedHeaders))

	sbt.CleanupTrackedHeadersBehindNonce(header.GetShardID()+1, 2)
	trackedHeaders, _ = sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 1, len(trackedHeaders))
}

func TestCleanupTrackedHeadersBehindNonce_ShouldNotCleanupWhenNonceIsGreaterOrEqual(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")

	sbt.AddTrackedHeader(header, headerHash)

	sbt.CleanupTrackedHeadersBehindNonce(header.GetShardID(), 1)
	trackedHeaders, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 1, len(trackedHeaders))
}

func TestCleanupTrackedHeadersBehindNonce_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")

	sbt.AddTrackedHeader(header, headerHash)

	sbt.CleanupTrackedHeadersBehindNonce(header.GetShardID(), 2)
	trackedHeaders, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 0, len(trackedHeaders))
}

func TestComputeLongestChain_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeader := shardArguments.StartHeaders[shardArguments.ShardCoordinator.SelfId()]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	hdr1 := &block.Header{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr1)

	hdr2 := &block.Header{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr2)

	hdr3 := &block.Header{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr3)

	sbt.AddTrackedHeader(hdr1, hdr1Hash)
	sbt.AddTrackedHeader(hdr2, hdr2Hash)
	sbt.AddTrackedHeader(hdr3, hdr3Hash)

	headers, _ := sbt.ComputeLongestChain(shardArguments.ShardCoordinator.SelfId(), hdr1)

	require.Equal(t, 1, len(headers))
	assert.Equal(t, hdr2, headers[0])
}

func TestComputeLongestMetaChainFromLastNotarized_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.CleanupHeadersBehindNonce(core.MetachainShardId, 0, 1)
	_, _, err := sbt.ComputeLongestMetaChainFromLastNotarized()

	assert.Equal(t, err, process.ErrNotarizedHeadersSliceForShardIsNil)
}

func TestComputeLongestMetaChainFromLastNotarized_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeader := shardArguments.StartHeaders[core.MetachainShardId]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	hdr1 := &block.MetaBlock{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr1)

	hdr2 := &block.MetaBlock{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr2)

	hdr3 := &block.MetaBlock{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr3)

	sbt.AddTrackedHeader(hdr1, hdr1Hash)
	sbt.AddTrackedHeader(hdr2, hdr2Hash)
	sbt.AddTrackedHeader(hdr3, hdr3Hash)

	headers, _, _ := sbt.ComputeLongestMetaChainFromLastNotarized()

	require.Equal(t, 2, len(headers))
	assert.Equal(t, hdr1, headers[0])
	assert.Equal(t, hdr2, headers[1])
}

func TestComputeLongestShardsChainsFromLastNotarized_ShouldErrNotarizedHeadersSliceForShardIsNil(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	for shardID := uint32(0); shardID < metaArguments.ShardCoordinator.NumberOfShards(); shardID++ {
		mbt.CleanupHeadersBehindNonce(shardID, 0, 1)
	}
	_, _, _, err := mbt.ComputeLongestShardsChainsFromLastNotarized()

	assert.Equal(t, err, process.ErrNotarizedHeadersSliceForShardIsNil)
}

func TestComputeLongestShardsChainsFromLastNotarized_ShouldWork(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	startHeaderShard0 := metaArguments.StartHeaders[0]
	startHeaderShard0Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, startHeaderShard0)

	hdr1Shard0 := &block.Header{
		ShardId:      0,
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderShard0Hash,
		PrevRandSeed: startHeaderShard0.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr1Shard0)

	hdr2Shard0 := &block.Header{
		ShardId:      0,
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1Shard0.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr2Shard0)

	hdr3Shard0 := &block.Header{
		ShardId:      0,
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2Shard0.GetRandSeed(),
	}
	hdr3Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr3Shard0)

	mbt.AddTrackedHeader(hdr1Shard0, hdr1Hash)
	mbt.AddTrackedHeader(hdr2Shard0, hdr2Hash)
	mbt.AddTrackedHeader(hdr3Shard0, hdr3Hash)

	startHeaderShard1 := metaArguments.StartHeaders[1]
	startHeaderShard1Hash, _ := core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, startHeaderShard1)

	hdr1Shard1 := &block.Header{
		ShardId:      1,
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderShard1Hash,
		PrevRandSeed: startHeaderShard1.GetRandSeed(),
	}
	hdr1Hash, _ = core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr1Shard1)

	hdr2Shard1 := &block.Header{
		ShardId:      1,
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1Shard1.GetRandSeed(),
	}
	hdr2Hash, _ = core.CalculateHash(metaArguments.Marshalizer, metaArguments.Hasher, hdr2Shard1)

	mbt.AddTrackedHeader(hdr1Shard1, hdr1Hash)
	mbt.AddTrackedHeader(hdr2Shard1, hdr2Hash)

	_, _, mapShardHeaders, _ := mbt.ComputeLongestShardsChainsFromLastNotarized()

	require.Equal(t, 2, len(mapShardHeaders[0]))
	require.Equal(t, 1, len(mapShardHeaders[1]))
	assert.Equal(t, hdr1Shard0, mapShardHeaders[0][0])
	assert.Equal(t, hdr2Shard0, mapShardHeaders[0][1])
	assert.Equal(t, hdr1Shard1, mapShardHeaders[1][0])
}

func TestDisplayTrackedHeaders_ShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	sbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	sbt.AddTrackedHeader(header, headerHash)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash := []byte("hash")
	sbt.AddCrossNotarizedHeader(metaBlock.GetShardID(), metaBlock, metaBlockHash)
	sbt.AddTrackedHeader(metaBlock, metaBlockHash)

	_ = logger.SetLogLevel("track:DEBUG")
	sbt.DisplayTrackedHeaders()
}

func TestDisplayTrackedHeadersForShard_ShouldNotPanicWhenTrackedHeadersSliceIsEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	_ = logger.SetLogLevel("track:DEBUG")
	sbt.DisplayTrackedHeadersForShard(0, "test")
}

func TestDisplayTrackedHeadersForShard_ShouldNotPanicWhenTheOnlyTrackedHeaderHasNonceZero(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   0,
	}
	headerHash := []byte("hash")
	sbt.AddTrackedHeader(header, headerHash)

	_ = logger.SetLogLevel("track:DEBUG")
	sbt.DisplayTrackedHeadersForShard(0, "test")
}

func TestDisplayTrackedHeadersForShard_ShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have paniced %v", r))
		}
	}()

	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	sbt.AddTrackedHeader(header, headerHash)

	_ = logger.SetLogLevel("track:DEBUG")
	sbt.DisplayTrackedHeadersForShard(0, "test")
}

func TestGetCrossNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	shardID := core.MetachainShardId
	metaBlock1 := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash1 := []byte("hash")
	sbt.AddCrossNotarizedHeader(shardID, metaBlock1, metaBlockHash1)

	metaBlock2 := &block.MetaBlock{
		Nonce: 2,
	}
	metaBlockHash2 := []byte("hash")
	sbt.AddCrossNotarizedHeader(shardID, metaBlock2, metaBlockHash2)

	crossNotarizedHeader, _, _ := sbt.GetCrossNotarizedHeader(shardID, 0)
	assert.Equal(t, metaBlock2, crossNotarizedHeader)

	crossNotarizedHeader, _, _ = sbt.GetCrossNotarizedHeader(shardID, 1)
	assert.Equal(t, metaBlock1, crossNotarizedHeader)
}

func TestGetLastCrossNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	shardID := core.MetachainShardId
	metaBlock1 := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash1 := []byte("hash")
	sbt.AddCrossNotarizedHeader(shardID, metaBlock1, metaBlockHash1)

	metaBlock2 := &block.MetaBlock{
		Nonce: 2,
	}
	metaBlockHash2 := []byte("hash")

	sbt.AddCrossNotarizedHeader(shardID, metaBlock2, metaBlockHash2)
	lastCrossNotarizedHeader, _, _ := sbt.GetLastCrossNotarizedHeader(shardID)

	assert.Equal(t, metaBlock2, lastCrossNotarizedHeader)
}

func TestGetLastCrossNotarizedHeadersForAllShards_ShouldWork(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	shardHeader1Shard0 := &block.Header{
		ShardId: 0,
		Nonce:   1,
	}
	shardHeaderHash1Shard0 := []byte("hash")
	mbt.AddCrossNotarizedHeader(0, shardHeader1Shard0, shardHeaderHash1Shard0)

	shardHeader1Shard1 := &block.Header{
		ShardId: 1,
		Nonce:   1,
	}
	shardHeaderHash1Shard1 := []byte("hash")
	mbt.AddCrossNotarizedHeader(1, shardHeader1Shard1, shardHeaderHash1Shard1)

	lastCrossNotarizedHeaders, _ := mbt.GetLastCrossNotarizedHeadersForAllShards()

	assert.Equal(t, shardHeader1Shard0, lastCrossNotarizedHeaders[0])
	assert.Equal(t, shardHeader1Shard1, lastCrossNotarizedHeaders[1])
}

func TestGetTrackedHeaders_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header1 := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash1 := []byte("hash")

	header2 := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   2,
	}
	headerHash2 := []byte("hash")

	sbt.AddTrackedHeader(header2, headerHash2)
	sbt.AddTrackedHeader(header1, headerHash1)

	trackedHeaders, _ := sbt.GetTrackedHeaders(shardArguments.ShardCoordinator.SelfId())

	require.Equal(t, 2, len(trackedHeaders))
	assert.Equal(t, header1, trackedHeaders[0])
	assert.Equal(t, header2, trackedHeaders[1])
}

func TestGetTrackedHeadersForAllShards_ShouldWork(t *testing.T) {
	t.Parallel()
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	shardHeader1Shard0 := &block.Header{
		ShardId: 0,
		Nonce:   1,
	}
	shardHeaderHash1Shard0 := []byte("hash")

	shardHeader1Shard1 := &block.Header{
		ShardId: 1,
		Nonce:   1,
	}
	shardHeaderHash1Shard1 := []byte("hash")

	mbt.AddTrackedHeader(shardHeader1Shard0, shardHeaderHash1Shard0)
	mbt.AddTrackedHeader(shardHeader1Shard1, shardHeaderHash1Shard1)

	trackedHeaders := mbt.GetTrackedHeadersForAllShards()

	assert.Equal(t, shardHeader1Shard0, trackedHeaders[0][0])
	assert.Equal(t, shardHeader1Shard1, trackedHeaders[1][0])
}

func TestSortHeadersFromNonce_ShouldNotSortWhenTrackedHeadersSliceForShardIsEmpty(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	headers, _ := sbt.SortHeadersFromNonce(0, 0)

	assert.Equal(t, 0, len(headers))
}

func TestSortHeadersFromNonce_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	shardHeader1 := &block.Header{
		Nonce: 1,
	}
	shardHeaderHash1 := []byte("hash")

	shardHeader2 := &block.Header{
		Nonce: 2,
	}
	shardHeaderHash2 := []byte("hash")

	sbt.AddTrackedHeader(shardHeader2, shardHeaderHash2)
	sbt.AddTrackedHeader(shardHeader1, shardHeaderHash1)

	headers, _ := sbt.SortHeadersFromNonce(0, 1)

	require.Equal(t, 2, len(headers))
	assert.Equal(t, headers[0], shardHeader1)
	assert.Equal(t, headers[1], shardHeader2)

	headers, _ = sbt.SortHeadersFromNonce(0, 2)

	require.Equal(t, 1, len(headers))
	assert.Equal(t, headers[0], shardHeader2)
}

func TestGetTrackedHeadersWithNonce_ShouldReturnNilWhenTrackedHeadersSliceForShardIsEmpty(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	headers, _ := sbt.GetTrackedHeadersWithNonce(0, 0)

	assert.Equal(t, 0, len(headers))
}

func TestGetTrackedHeadersWithNonce_ShouldReturnNilWhenTrackedHeadersSliceForNonceIsEmpty(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	shardHeader1 := &block.Header{
		Nonce: 1,
	}
	shardHeaderHash1 := []byte("hash1")

	shardHeader2 := &block.Header{
		Nonce: 1,
	}
	shardHeaderHash2 := []byte("hash2")

	sbt.AddTrackedHeader(shardHeader1, shardHeaderHash1)
	sbt.AddTrackedHeader(shardHeader2, shardHeaderHash2)

	headers, _ := sbt.GetTrackedHeadersWithNonce(0, 0)

	assert.Equal(t, 0, len(headers))
}

func TestGetTrackedHeadersWithNonce_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	shardHeader1 := &block.Header{
		Nonce: 1,
	}
	shardHeaderHash1 := []byte("hash1")

	shardHeader2 := &block.Header{
		Nonce: 1,
	}
	shardHeaderHash2 := []byte("hash2")

	sbt.AddTrackedHeader(shardHeader2, shardHeaderHash2)
	sbt.AddTrackedHeader(shardHeader1, shardHeaderHash1)

	headers, _ := sbt.GetTrackedHeadersWithNonce(0, 1)

	require.Equal(t, 2, len(headers))
	assert.Equal(t, headers[0], shardHeader2)
	assert.Equal(t, headers[1], shardHeader1)
}

func TestIsShardStuck_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeader := shardArguments.StartHeaders[core.MetachainShardId]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	hdr1 := &block.MetaBlock{
		Round: 1,
		Nonce: 1,
		ShardInfo: []block.ShardData{
			{
				NumPendingMiniBlocks: 99,
			},
		},
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr1)

	hdr2 := &block.MetaBlock{
		Round: 2,
		Nonce: 2,
		ShardInfo: []block.ShardData{
			{
				NumPendingMiniBlocks: 100,
			},
		},
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr2)

	hdr3 := &block.MetaBlock{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr3)

	sbt.ReceivedMetaBlock(hdr1, hdr1Hash)
	sbt.ReceivedMetaBlock(hdr2, hdr2Hash)

	assert.False(t, sbt.IsShardStuck(0))

	sbt.ReceivedMetaBlock(hdr3, hdr3Hash)

	assert.True(t, sbt.IsShardStuck(0))
}

func TestRegisterCrossNotarizedHeadersHandler_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	wg := sync.WaitGroup{}
	wg.Add(1)

	called := false
	sbt.RegisterCrossNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
		called = true
		wg.Done()
	})

	startHeader := shardArguments.StartHeaders[core.MetachainShardId]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	hdr1 := &block.MetaBlock{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr1)

	hdr2 := &block.MetaBlock{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr2)

	sbt.ReceivedMetaBlock(hdr1, hdr1Hash)
	sbt.ReceivedMetaBlock(hdr2, hdr2Hash)

	wg.Wait()

	assert.True(t, called)
}

func TestRegisterSelfNotarizedHeadersHandler_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	wg := sync.WaitGroup{}
	wg.Add(1)

	called := false
	sbt.RegisterSelfNotarizedHeadersHandler(func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
		called = true
		wg.Done()
	})

	startHeader := shardArguments.StartHeaders[shardArguments.ShardCoordinator.SelfId()]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	hdr1 := &block.Header{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr1)

	hdr2 := &block.Header{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Hash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr2)

	sbt.ReceivedShardHeader(hdr1, hdr1Hash)
	sbt.ReceivedShardHeader(hdr2, hdr2Hash)

	wg.Wait()

	assert.True(t, called)
}

func TestRemoveLastNotarizedHeaders_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash := []byte("hash")
	sbt.AddCrossNotarizedHeader(metaBlock.GetShardID(), metaBlock, metaBlockHash)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	sbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)

	lastCrossNotarizedHeader, _, _ := sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	lastSelfNotarizedHeader, _, _ := sbt.GetLastSelfNotarizedHeader(header.GetShardID())

	assert.Equal(t, metaBlock, lastCrossNotarizedHeader)
	assert.Equal(t, header, lastSelfNotarizedHeader)

	sbt.RemoveLastNotarizedHeaders()

	lastCrossNotarizedHeader, _, _ = sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	lastSelfNotarizedHeader, _, _ = sbt.GetLastSelfNotarizedHeader(header.GetShardID())

	assert.Equal(t, shardArguments.StartHeaders[metaBlock.GetShardID()], lastCrossNotarizedHeader)
	assert.Equal(t, shardArguments.StartHeaders[header.GetShardID()], lastSelfNotarizedHeader)
}

func TestRestoreToGenesis_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		Nonce: 1,
	}
	metaBlockHash := []byte("hash")
	sbt.AddCrossNotarizedHeader(metaBlock.GetShardID(), metaBlock, metaBlockHash)
	sbt.AddTrackedHeader(metaBlock, metaBlockHash)

	header := &block.Header{
		ShardId: shardArguments.ShardCoordinator.SelfId(),
		Nonce:   1,
	}
	headerHash := []byte("hash")
	sbt.AddSelfNotarizedHeader(header.GetShardID(), header, headerHash)
	sbt.AddTrackedHeader(header, headerHash)

	trackedHeaders, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())
	require.Equal(t, 1, len(trackedHeaders))
	assert.Equal(t, metaBlock, trackedHeaders[0])

	trackedHeaders, _ = sbt.GetTrackedHeaders(header.GetShardID())
	require.Equal(t, 1, len(trackedHeaders))
	assert.Equal(t, header, trackedHeaders[0])

	lastCrossNotarizedHeader, _, _ := sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	assert.Equal(t, metaBlock, lastCrossNotarizedHeader)

	lastSelfNotarizedHeader, _, _ := sbt.GetLastSelfNotarizedHeader(header.GetShardID())
	assert.Equal(t, header, lastSelfNotarizedHeader)

	sbt.RestoreToGenesis()

	trackedHeaders, _ = sbt.GetTrackedHeaders(metaBlock.GetShardID())
	assert.Equal(t, 0, len(trackedHeaders))

	trackedHeaders, _ = sbt.GetTrackedHeaders(header.GetShardID())
	assert.Equal(t, 0, len(trackedHeaders))

	lastCrossNotarizedHeader, _, _ = sbt.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	assert.Equal(t, shardArguments.StartHeaders[metaBlock.GetShardID()], lastCrossNotarizedHeader)

	lastSelfNotarizedHeader, _, _ = sbt.GetLastSelfNotarizedHeader(header.GetShardID())
	assert.Equal(t, shardArguments.StartHeaders[header.GetShardID()], lastSelfNotarizedHeader)
}

func TestCheckTrackerNilParameters_ShouldErrNilHasher(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.Hasher = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilHasher, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilHeaderValidator(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.HeaderValidator = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilHeaderValidator, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilMarshalizer(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.Marshalizer = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilRequestHandler(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.RequestHandler = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilRounder(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.Rounder = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilRounder, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilShardCoordinator(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.ShardCoordinator = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilStorage(t *testing.T) {
	t.Parallel()
	baseArguments := CreateBaseTrackerMockArguments()
	baseArguments.Store = nil

	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilStorage, err)
}

func TestInitNotarizedHeaders_ShouldErrNotarizedHeadersSliceIsNil(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	err := sbt.InitNotarizedHeaders(nil)

	assert.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
}

func TestInitNotarizedHeaders_ShouldWork(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeaders := make(map[uint32]data.HeaderHandler)
	selfStartHeader := &block.Header{Nonce: 1}
	metachainStartHeader := &block.MetaBlock{Nonce: 1}
	startHeaders[shardArguments.ShardCoordinator.SelfId()] = selfStartHeader
	startHeaders[core.MetachainShardId] = metachainStartHeader
	err := sbt.InitNotarizedHeaders(startHeaders)
	lastCrossNotarizedHeaderForSelfShard, _, _ := sbt.GetLastCrossNotarizedHeader(shardArguments.ShardCoordinator.SelfId())
	lastCrossNotarizedHeaderForMetachain, _, _ := sbt.GetLastCrossNotarizedHeader(core.MetachainShardId)
	lastSelfNotarizedHeaderForSelfShard, _, _ := sbt.GetLastSelfNotarizedHeader(shardArguments.ShardCoordinator.SelfId())
	lastSelfNotarizedHeaderForMetachain, _, _ := sbt.GetLastSelfNotarizedHeader(core.MetachainShardId)

	assert.Nil(t, err)
	assert.Equal(t, selfStartHeader, lastCrossNotarizedHeaderForSelfShard)
	assert.Equal(t, metachainStartHeader, lastCrossNotarizedHeaderForMetachain)
	assert.Equal(t, selfStartHeader, lastSelfNotarizedHeaderForSelfShard)
	assert.Equal(t, selfStartHeader, lastSelfNotarizedHeaderForMetachain)
}

func TestComputeLongestChain_ShouldWorkWithLongestChain(t *testing.T) {
	t.Parallel()
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeader := shardArguments.StartHeaders[shardArguments.ShardCoordinator.SelfId()]
	startHeaderHash, _ := core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, startHeader)

	longestChain := uint64(1000)
	chains := uint64(10)

	for j := uint64(0); j < chains; j++ {
		prevHash := startHeaderHash
		prevSeed := startHeader.GetRandSeed()
		for i := uint64(1); i <= longestChain+1+j; i++ {
			randSeed := shardArguments.Hasher.Compute(string(prevSeed))
			round := i
			if i > j {
				round += j
			}
			hdr := &block.Header{
				Round:        round,
				Nonce:        i,
				PrevHash:     prevHash,
				PrevRandSeed: prevSeed,
				RandSeed:     randSeed,
			}
			prevHash, _ = core.CalculateHash(shardArguments.Marshalizer, shardArguments.Hasher, hdr)
			prevSeed = hdr.RandSeed
			if i > j {
				sbt.AddTrackedHeader(hdr, prevHash)
			}
		}
	}

	headers, _ := sbt.ComputeLongestChain(shardArguments.ShardCoordinator.SelfId(), startHeader)

	assert.Equal(t, longestChain+chains-1, uint64(len(headers)))
}

//------- CheckBlockAgainstRounder

func TestBaseBlockTrack_CheckBlockAgainstRounderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	bbt := track.NewBaseBlockTrack()
	err := bbt.CheckBlockAgainstRounder(nil)

	assert.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestBaseBlockTrack_CheckBlockAgainstRounderHigherRoundShouldErr(t *testing.T) {
	t.Parallel()

	bbt := track.NewBaseBlockTrack()
	currentRound := int64(50)
	bbt.SetRounder(
		&mock.RounderMock{
			RoundIndex: currentRound,
		},
	)

	hdr := &block.Header{
		Round: uint64(currentRound + 2),
	}
	err := bbt.CheckBlockAgainstRounder(hdr)

	assert.True(t, errors.Is(err, process.ErrHigherRoundInBlock))
}

func TestBaseBlockTrack_CheckBlockAgainstRounderShouldWork(t *testing.T) {
	t.Parallel()

	bbt := track.NewBaseBlockTrack()
	currentRound := int64(50)
	bbt.SetRounder(
		&mock.RounderMock{
			RoundIndex: currentRound,
		},
	)

	hdr := &block.Header{
		Round: uint64(currentRound + 1),
	}
	err := bbt.CheckBlockAgainstRounder(hdr)

	assert.Nil(t, err)
}

//------- CheckBlockAgainstFinal

func TestBaseBlockTrack_CheckBlockAgainstFinalNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	bbt := track.NewBaseBlockTrack()
	err := bbt.CheckBlockAgainstFinal(nil)

	assert.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestBaseBlockTrack_CheckBlockAgainstFinalCurrentShardGetFinalFailsShouldErr(t *testing.T) {
	t.Parallel()

	crtShard := uint32(0)
	bbt := track.NewBaseBlockTrack()
	bbt.SetShardCoordinator(mock.NewMultiShardsCoordinatorMock(crtShard))
	expectedErr := errors.New("expected err")
	bbt.SetSelfNotarizer(
		&mock.BlockNotarizerHandlerMock{
			GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
				return nil, nil, expectedErr
			},
		},
	)
	hdr := &block.Header{
		ShardId: crtShard,
	}
	err := bbt.CheckBlockAgainstFinal(hdr)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestBaseBlockTrack_CheckBlockAgainstFinalCrossShardShardGetFinalFailsShouldErr(t *testing.T) {
	t.Parallel()

	crtShard := uint32(0)
	bbt := track.NewBaseBlockTrack()
	bbt.SetShardCoordinator(mock.NewMultiShardsCoordinatorMock(crtShard))
	expectedErr := errors.New("expected err")
	bbt.SetCrossNotarizer(
		&mock.BlockNotarizerHandlerMock{
			GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
				return nil, nil, expectedErr
			},
		},
	)
	hdr := &block.Header{
		ShardId: crtShard + 1,
	}
	err := bbt.CheckBlockAgainstFinal(hdr)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestBaseBlockTrack_CheckBlockAgainstFinalLowerRoundInBlockShouldErr(t *testing.T) {
	t.Parallel()

	crtShard := uint32(0)
	finalRound := uint64(667)
	bbt := track.NewBaseBlockTrack()
	bbt.SetShardCoordinator(mock.NewMultiShardsCoordinatorMock(crtShard))
	bbt.SetSelfNotarizer(
		&mock.BlockNotarizerHandlerMock{
			GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
				hdr := &block.Header{
					ShardId: crtShard,
					Round:   finalRound,
				}

				return hdr, make([]byte, 0), nil
			},
		},
	)
	hdr := &block.Header{
		ShardId: crtShard,
		Round:   finalRound - 1,
	}
	err := bbt.CheckBlockAgainstFinal(hdr)

	assert.True(t, errors.Is(err, process.ErrLowerRoundInBlock))
}

func TestBaseBlockTrack_CheckBlockAgainstFinalLowerNonceInBlockShouldErr(t *testing.T) {
	t.Parallel()

	crtShard := uint32(0)
	finalRound := uint64(667)
	finalNonce := uint64(334)
	bbt := track.NewBaseBlockTrack()
	bbt.SetShardCoordinator(mock.NewMultiShardsCoordinatorMock(crtShard))
	bbt.SetSelfNotarizer(
		&mock.BlockNotarizerHandlerMock{
			GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
				hdr := &block.Header{
					ShardId: crtShard,
					Round:   finalRound,
					Nonce:   finalNonce,
				}

				return hdr, make([]byte, 0), nil
			},
		},
	)
	hdr := &block.Header{
		ShardId: crtShard,
		Round:   finalRound,
		Nonce:   finalNonce - 1,
	}
	err := bbt.CheckBlockAgainstFinal(hdr)

	assert.True(t, errors.Is(err, process.ErrLowerNonceInBlock))
}

func TestBaseBlockTrack_CheckBlockAgainstFinalHigherNonceInBlockShouldErr(t *testing.T) {
	t.Parallel()

	crtShard := uint32(0)
	finalRound := uint64(667)
	finalNonce := uint64(334)
	bbt := track.NewBaseBlockTrack()
	bbt.SetShardCoordinator(mock.NewMultiShardsCoordinatorMock(crtShard))
	bbt.SetSelfNotarizer(
		&mock.BlockNotarizerHandlerMock{
			GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
				hdr := &block.Header{
					ShardId: crtShard,
					Round:   finalRound,
					Nonce:   finalNonce,
				}

				return hdr, make([]byte, 0), nil
			},
		},
	)
	hdr := &block.Header{
		ShardId: crtShard,
		Round:   finalRound + 1,
		Nonce:   finalNonce + 2,
	}
	err := bbt.CheckBlockAgainstFinal(hdr)

	assert.True(t, errors.Is(err, process.ErrHigherNonceInBlock))
}

func TestBaseBlockTrack_CheckBlockAgainstFinalShouldWork(t *testing.T) {
	t.Parallel()

	crtShard := uint32(0)
	finalRound := uint64(667)
	finalNonce := uint64(334)
	bbt := track.NewBaseBlockTrack()
	bbt.SetShardCoordinator(mock.NewMultiShardsCoordinatorMock(crtShard))
	bbt.SetSelfNotarizer(
		&mock.BlockNotarizerHandlerMock{
			GetFirstNotarizedHeaderCalled: func(shardID uint32) (handler data.HeaderHandler, bytes []byte, err error) {
				hdr := &block.Header{
					ShardId: crtShard,
					Round:   finalRound,
					Nonce:   finalNonce,
				}

				return hdr, make([]byte, 0), nil
			},
		},
	)
	hdr := &block.Header{
		ShardId: crtShard,
		Round:   finalRound + 2,
		Nonce:   finalNonce + 2,
	}
	err := bbt.CheckBlockAgainstFinal(hdr)

	assert.Nil(t, err)
}
