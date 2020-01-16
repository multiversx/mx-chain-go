package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	block2 "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func createGenesisBlocks(shardCoordinator sharding.Coordinator) map[uint32]data.HeaderHandler {
	genesisBlocks := make(map[uint32]data.HeaderHandler)
	for ShardID := uint32(0); ShardID < shardCoordinator.NumberOfShards(); ShardID++ {
		genesisBlocks[ShardID] = createGenesisShardHeader(ShardID)
	}

	genesisBlocks[sharding.MetachainShardId] = createGenesisMetaBlock()

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
	argsHeaderValidator := block2.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := block2.NewHeaderValidator(argsHeaderValidator)

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
	shardCoordinatorMock.CurrentShard = sharding.MetachainShardId
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	argsHeaderValidator := block2.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := block2.NewHeaderValidator(argsHeaderValidator)

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
	argsHeaderValidator := block2.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := block2.NewHeaderValidator(argsHeaderValidator)

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
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{block.ShardData{ShardID: 1}},
	}

	headerInfo := sbt.GetSelfHeaders(metaBlock)
	assert.Equal(t, 0, len(headerInfo))
}

func TestShardGetSelfHeaders_ShouldReturnEmptySliceWhenErrGetShardHeader(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{block.ShardData{ShardID: 0}},
	}

	headerInfo := sbt.GetSelfHeaders(metaBlock)
	assert.Equal(t, 0, len(headerInfo))
}

func TestShardGetSelfHeaders_ShouldWork(t *testing.T) {
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
		ShardInfo: []block.ShardData{block.ShardData{ShardID: 0}},
	}

	headerInfo := sbt.GetSelfHeaders(metaBlock)
	assert.Equal(t, 1, len(headerInfo))
}

func TestMetaGetSelfHeaders_ShouldReturnEmptySliceWhenErrGetMetaHeader(t *testing.T) {
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	header := &block.Header{
		MetaBlockHashes: [][]byte{[]byte("hash")},
	}

	headerInfo := mbt.GetSelfHeaders(header)
	assert.Equal(t, 0, len(headerInfo))
}

func TestMetaGetSelfHeaders_ShouldWork(t *testing.T) {
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
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)
	sbt.CleanupHeadersBehindNonce(sharding.MetachainShardId, 1, 1)

	_, _, headers, _ := sbt.ComputeLongestSelfChain()
	assert.Nil(t, headers)
}

func TestShardComputeLongestSelfChain_ShouldReturnEmptySliceWhenComputeLongestChainReturnNil(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	_, _, headers, _ := sbt.ComputeLongestSelfChain()
	assert.Equal(t, 0, len(headers))
}

func TestShardComputeLongestSelfChain_ShouldWork(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeader := shardArguments.StartHeaders[shardArguments.ShardCoordinator.SelfId()]
	startHeaderMarshalized, _ := shardArguments.Marshalizer.Marshal(startHeader)
	startHeaderHash := shardArguments.Hasher.Compute(string(startHeaderMarshalized))

	hdr1 := &block.Header{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Marshalized, _ := shardArguments.Marshalizer.Marshal(hdr1)
	hdr1Hash := shardArguments.Hasher.Compute(string(hdr1Marshalized))

	hdr2 := &block.Header{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Marshalized, _ := shardArguments.Marshalizer.Marshal(hdr2)
	hdr2Hash := shardArguments.Hasher.Compute(string(hdr2Marshalized))

	hdr3 := &block.Header{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Marshalized, _ := shardArguments.Marshalizer.Marshal(hdr3)
	hdr3Hash := shardArguments.Hasher.Compute(string(hdr3Marshalized))

	sbt.AddTrackedHeader(hdr1, hdr1Hash)
	sbt.AddTrackedHeader(hdr2, hdr2Hash)
	sbt.AddTrackedHeader(hdr3, hdr3Hash)

	lastNotarizedHeader, lastNotarizedHeaderHash, headers, hashes := sbt.ComputeLongestSelfChain()
	assert.Equal(t, 2, len(headers))
	assert.Equal(t, startHeaderHash, lastNotarizedHeaderHash)
	assert.Equal(t, hashes[0], hdr1Hash)
	assert.Equal(t, hashes[1], hdr2Hash)
	assert.Equal(t, startHeader, lastNotarizedHeader)
	assert.Equal(t, headers[0], hdr1)
	assert.Equal(t, headers[1], hdr2)
}

func TestMetaComputeLongestSelfChain_ShouldReturnNilWhenErrGetLastNotarizedHeader(t *testing.T) {
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)
	mbt.CleanupHeadersBehindNonce(sharding.MetachainShardId, 1, 1)

	_, _, headers, _ := mbt.ComputeLongestSelfChain()
	assert.Nil(t, headers)
}

func TestMetaComputeLongestSelfChain_ShouldReturnEmptySliceWhenComputeLongestChainReturnNil(t *testing.T) {
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	_, _, headers, _ := mbt.ComputeLongestSelfChain()
	assert.Equal(t, 0, len(headers))
}

func TestMetaComputeLongestSelfChain_ShouldWork(t *testing.T) {
	metaArguments := CreateMetaTrackerMockArguments()
	mbt, _ := track.NewMetaBlockTrack(metaArguments)

	startHeader := metaArguments.StartHeaders[metaArguments.ShardCoordinator.SelfId()]
	startHeaderMarshalized, _ := metaArguments.Marshalizer.Marshal(startHeader)
	startHeaderHash := metaArguments.Hasher.Compute(string(startHeaderMarshalized))

	hdr1 := &block.MetaBlock{
		Round:        1,
		Nonce:        1,
		PrevHash:     startHeaderHash,
		PrevRandSeed: startHeader.GetRandSeed(),
	}
	hdr1Marshalized, _ := metaArguments.Marshalizer.Marshal(hdr1)
	hdr1Hash := metaArguments.Hasher.Compute(string(hdr1Marshalized))

	hdr2 := &block.MetaBlock{
		Round:        2,
		Nonce:        2,
		PrevHash:     hdr1Hash,
		PrevRandSeed: hdr1.GetRandSeed(),
	}
	hdr2Marshalized, _ := metaArguments.Marshalizer.Marshal(hdr2)
	hdr2Hash := metaArguments.Hasher.Compute(string(hdr2Marshalized))

	hdr3 := &block.MetaBlock{
		Round:        3,
		Nonce:        3,
		PrevHash:     hdr2Hash,
		PrevRandSeed: hdr2.GetRandSeed(),
	}
	hdr3Marshalized, _ := metaArguments.Marshalizer.Marshal(hdr3)
	hdr3Hash := metaArguments.Hasher.Compute(string(hdr3Marshalized))

	mbt.AddTrackedHeader(hdr1, hdr1Hash)
	mbt.AddTrackedHeader(hdr2, hdr2Hash)
	mbt.AddTrackedHeader(hdr3, hdr3Hash)

	lastNotarizedHeader, lastNotarizedHeaderHash, headers, hashes := mbt.ComputeLongestSelfChain()
	assert.Equal(t, 2, len(headers))
	assert.Equal(t, startHeaderHash, lastNotarizedHeaderHash)
	assert.Equal(t, hashes[0], hdr1Hash)
	assert.Equal(t, hashes[1], hdr2Hash)
	assert.Equal(t, startHeader, lastNotarizedHeader)
	assert.Equal(t, headers[0], hdr1)
	assert.Equal(t, headers[1], hdr2)
}

func TestComputePendingMiniBlockHeaders_ShouldReturnZeroWhenHeadersSliceIsEmpty(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.ComputeNumPendingMiniBlocks([]data.HeaderHandler{})
	assert.Equal(t, uint32(0), sbt.GetNumPendingMiniBlocks(shardArguments.ShardCoordinator.SelfId()))
}

func TestComputePendingMiniBlockHeaders_ShouldReturnZeroWhenErrWrongTypeAssertion(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.ComputeNumPendingMiniBlocks([]data.HeaderHandler{&block.Header{}})
	assert.Equal(t, uint32(0), sbt.GetNumPendingMiniBlocks(shardArguments.ShardCoordinator.SelfId()))
}

func TestComputePendingMiniBlockHeaders_ShouldWork(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	sbt.ComputeNumPendingMiniBlocks([]data.HeaderHandler{&block.MetaBlock{
		ShardInfo: []block.ShardData{
			block.ShardData{
				ShardID:              0,
				NumPendingMiniBlocks: 2,
			},
		}}})

	assert.Equal(t, uint32(2), sbt.GetNumPendingMiniBlocks(0))
}

func TestReceivedHeader_ShouldAddMetaBlockToTrackedHeaders(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{Nonce: 1}
	metaBlockHash := []byte("hash")
	sbt.ReceivedHeader(metaBlock, metaBlockHash)
	headers, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	assert.Equal(t, 1, len(headers))
	assert.Equal(t, metaBlock, headers[0])
}

func TestReceivedHeader_ShouldAddShardHeaderToTrackedHeaders(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{Nonce: 1}
	headerHash := []byte("hash")
	sbt.ReceivedHeader(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestReceivedShardHeader_ShouldReturnWhenErrWrongTypeAssertion(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{Nonce: 1}
	metaBlockHash := []byte("hash")
	sbt.ReceivedShardHeader(metaBlock, metaBlockHash)
	headers, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	assert.Equal(t, 0, len(headers))
}

func TestReceivedShardHeader_ShouldWork(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{Nonce: 1}
	headerHash := []byte("hash")
	sbt.ReceivedShardHeader(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestReceivedMetaBlock_ShouldReturnWhenErrWrongTypeAssertion(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{Nonce: 1}
	headerHash := []byte("hash")
	sbt.ReceivedMetaBlock(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 0, len(headers))
}

func TestReceivedMetaBlock_ShouldWork(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	metaBlock := &block.MetaBlock{Nonce: 1}
	metaBlockHash := []byte("hash")
	sbt.ReceivedMetaBlock(metaBlock, metaBlockHash)
	headers, _ := sbt.GetTrackedHeaders(metaBlock.GetShardID())

	assert.Equal(t, 1, len(headers))
	assert.Equal(t, metaBlock, headers[0])
}

func TestAddHeader_ShouldNotAddIfItAlreadyExist(t *testing.T) {
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

	assert.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

func TestAddHeader_ShouldWork(t *testing.T) {
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

	assert.Equal(t, 2, len(headers))
	assert.Equal(t, hdr1, headers[0])
	assert.Equal(t, hdr2, headers[1])
}

func TestAddCrossNotarizedHeader_ShouldWork(t *testing.T) {
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
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	header := &block.Header{
		Nonce: 1,
	}
	headerHash := []byte("hash")
	sbt.AddTrackedHeader(header, headerHash)
	headers, _ := sbt.GetTrackedHeaders(header.GetShardID())

	assert.Equal(t, 1, len(headers))
	assert.Equal(t, header, headers[0])
}

//func TestNewBlockTrack_AddHeaderForShardShouldNotUpdateIfHeaderIsNil(t *testing.T) {
//	arguments := CreateShardMockArguments()
//	bt, _ := track.NewShardBlockTrack(arguments)
//
//	bt.AddTrackedHeader(nil, nil)
//	lastHeader := bt.GetLastHeader(0)
//	assert.Nil(t, lastHeader)
//}
//
//func TestNewBlockTrack_AddHeaderShouldNotUpdateIfRoundIsLowerOrEqual(t *testing.T) {
//	arguments := CreateShardMockArguments()
//	bt, _ := track.NewShardBlockTrack(arguments)
//
//	header1 := &block.Header{Round: 2}
//	header2 := &block.Header{Round: 1}
//	header3 := &block.Header{Round: 2}
//
//	bt.AddTrackedHeader(header1, []byte("hash1"))
//	lastHeader := bt.GetLastHeader(0)
//	assert.Equal(t, header1, lastHeader)
//
//	bt.AddTrackedHeader(header2, []byte("hash2"))
//	lastHeader = bt.GetLastHeader(0)
//	assert.Equal(t, header1, lastHeader)
//
//	bt.AddTrackedHeader(header3, []byte("hash3"))
//	lastHeader = bt.GetLastHeader(0)
//	assert.Equal(t, header1, lastHeader)
//}
//
//func TestNewBlockTrack_AddHeaderShouldUpdate(t *testing.T) {
//	arguments := CreateShardMockArguments()
//	bt, _ := track.NewShardBlockTrack(arguments)
//
//	header1 := &block.Header{Round: 2}
//	header2 := &block.Header{Round: 3}
//	header3 := &block.Header{Round: 4}
//
//	header4 := &block.MetaBlock{Round: 2}
//	header5 := &block.MetaBlock{Round: 3}
//	header6 := &block.MetaBlock{Round: 4}
//
//	bt.AddTrackedHeader(header1, []byte("hash1"))
//	bt.AddTrackedHeader(header4, []byte("hash4"))
//	lastHeader := bt.GetLastHeader(0)
//	lastHeaderMeta := bt.GetLastHeader(sharding.MetachainShardId)
//	assert.Equal(t, header1, lastHeader)
//	assert.Equal(t, header4, lastHeaderMeta)
//
//	bt.AddTrackedHeader(header2, []byte("hash2"))
//	bt.AddTrackedHeader(header5, []byte("hash5"))
//	lastHeader = bt.GetLastHeader(0)
//	lastHeaderMeta = bt.GetLastHeader(sharding.MetachainShardId)
//	assert.Equal(t, header2, lastHeader)
//	assert.Equal(t, header5, lastHeaderMeta)
//
//	bt.AddTrackedHeader(header3, []byte("hash3"))
//	bt.AddTrackedHeader(header6, []byte("hash6"))
//	lastHeader = bt.GetLastHeader(0)
//	lastHeaderMeta = bt.GetLastHeader(sharding.MetachainShardId)
//	assert.Equal(t, header3, lastHeader)
//	assert.Equal(t, header6, lastHeaderMeta)
//}
//
//func TestNewBlockTrack_IsShardStuckShoudReturnFalseWhenListIsEmpty(t *testing.T) {
//	arguments := CreateShardTrackerMockArguments()
//	bt, _ := track.NewShardBlockTrack(arguments)
//
//	isShardStuck := bt.IsShardStuck(0)
//	assert.False(t, isShardStuck)
//}
//
//func TestNewBlockTrack_IsShardStuckShoudReturnFalse(t *testing.T) {
//	arguments := CreateShardTrackerMockArguments()
//	bt, _ := track.NewShardBlockTrack(arguments)
//
//	bt.AddTrackedHeader(&block.Header{Round: 1, Nonce: 1}, []byte("hash1"))
//	isShardStuck := bt.IsShardStuck(0)
//	assert.False(t, isShardStuck)
//}
//
//func TestNewBlockTrack_IsShardStuckShoudReturnTrue(t *testing.T) {
//	arguments := CreateShardTrackerMockArguments()
//	rounderMock := &mock.RounderMock{}
//	arguments.Rounder = rounderMock
//	bt, _ := track.NewShardBlockTrack(arguments)
//
//	bt.AddTrackedHeader(&block.Header{Round: 1, Nonce: 1}, []byte("hash1"))
//	rounderMock.RoundIndex = process.MaxRoundsWithoutCommittedBlock + 1
//	isShardStuck := bt.IsShardStuck(0)
//	assert.True(t, isShardStuck)
//}

func TestCheckTrackerNilParameters_ShouldErrNilHasher(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.Hasher = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilHasher, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilHeaderValidator(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.HeaderValidator = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilHeaderValidator, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilMarshalizer(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.Marshalizer = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilRequestHandler(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.RequestHandler = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilRounder(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.Rounder = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilRounder, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilShardCoordinator(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()

	baseArguments.ShardCoordinator = nil
	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCheckTrackerNilParameters_ShouldErrNilStorage(t *testing.T) {
	baseArguments := CreateBaseTrackerMockArguments()
	baseArguments.Store = nil

	err := track.CheckTrackerNilParameters(baseArguments)

	assert.Equal(t, process.ErrNilStorage, err)
}

func TestInitNotarizedHeaders_ShouldErrNotarizedHeadersSliceIsNil(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	err := sbt.InitNotarizedHeaders(nil)

	assert.Equal(t, process.ErrNotarizedHeadersSliceIsNil, err)
}

func TestInitNotarizedHeaders_ShouldWork(t *testing.T) {
	shardArguments := CreateShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)

	startHeaders := make(map[uint32]data.HeaderHandler)
	selfStartHeader := &block.Header{Nonce: 1}
	metachainStartHeader := &block.MetaBlock{Nonce: 1}
	startHeaders[shardArguments.ShardCoordinator.SelfId()] = selfStartHeader
	startHeaders[sharding.MetachainShardId] = metachainStartHeader
	err := sbt.InitNotarizedHeaders(startHeaders)
	lastCrossNotarizedHeaderForSelfShard, _, _ := sbt.GetLastCrossNotarizedHeader(shardArguments.ShardCoordinator.SelfId())
	lastCrossNotarizedHeaderForMetachain, _, _ := sbt.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
	lastSelfNotarizedHeaderForSelfShard, _, _ := sbt.GetLastSelfNotarizedHeader(shardArguments.ShardCoordinator.SelfId())
	lastSelfNotarizedHeaderForMetachain, _, _ := sbt.GetLastSelfNotarizedHeader(sharding.MetachainShardId)

	assert.Nil(t, err)
	assert.Equal(t, selfStartHeader, lastCrossNotarizedHeaderForSelfShard)
	assert.Equal(t, metachainStartHeader, lastCrossNotarizedHeaderForMetachain)
	assert.Equal(t, selfStartHeader, lastSelfNotarizedHeaderForSelfShard)
	assert.Equal(t, selfStartHeader, lastSelfNotarizedHeaderForMetachain)
}
