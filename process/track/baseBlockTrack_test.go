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
		genesisBlocks[ShardID] = createGenesisBlock(ShardID)
	}

	genesisBlocks[sharding.MetachainShardId] = createGenesisMetaBlock()

	return genesisBlocks
}

func createGenesisBlock(ShardID uint32) *block.Header {
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
	memDB, _ := memorydb.New()

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

func CreateShardMockArguments() track.ArgShardTracker {
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
			Rounder:          &mock.RounderMock{},
			ShardCoordinator: shardCoordinatorMock,
			Store:            initStore(),
			StartHeaders:     genesisBlocks,
		},
		PoolsHolder: mock.NewPoolsHolderMock(),
	}

	return arguments
}

func TestNewBlockTrack_ShouldErrNilHasher(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.Hasher = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilHeaderValidator(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.HeaderValidator = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilHeaderValidator, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilMarshalizer(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.Marshalizer = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilPoolsHolder(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.PoolsHolder = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilPoolsHolder, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilRounder(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.Rounder = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilRounder, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilShardCoordinator(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.ShardCoordinator = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilStorage(t *testing.T) {
	arguments := CreateShardMockArguments()
	arguments.Store = nil
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Equal(t, process.ErrNilStorage, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldWork(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, err := track.NewShardBlockTrack(arguments)

	assert.Nil(t, err)
	assert.NotNil(t, bt)
}

func TestNewBlockTrack_AddHeaderForShardShouldNotUpdateIfHeaderIsNil(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, _ := track.NewShardBlockTrack(arguments)

	bt.AddTrackedHeader(nil, nil)
	lastHeader := bt.LastHeaderForShard(0)
	assert.Nil(t, lastHeader)
}

func TestNewBlockTrack_AddHeaderShouldNotUpdateIfRoundIsLowerOrEqual(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, _ := track.NewShardBlockTrack(arguments)

	header1 := &block.Header{Round: 2}
	header2 := &block.Header{Round: 1}
	header3 := &block.Header{Round: 2}

	bt.AddTrackedHeader(header1, []byte("hash1"))
	lastHeader := bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)

	bt.AddTrackedHeader(header2, []byte("hash2"))
	lastHeader = bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)

	bt.AddTrackedHeader(header3, []byte("hash3"))
	lastHeader = bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)
}

func TestNewBlockTrack_AddHeaderShouldUpdate(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, _ := track.NewShardBlockTrack(arguments)

	header1 := &block.Header{Round: 2}
	header2 := &block.Header{Round: 3}
	header3 := &block.Header{Round: 4}

	header4 := &block.MetaBlock{Round: 2}
	header5 := &block.MetaBlock{Round: 3}
	header6 := &block.MetaBlock{Round: 4}

	bt.AddTrackedHeader(header1, []byte("hash1"))
	bt.AddTrackedHeader(header4, []byte("hash4"))
	lastHeader := bt.LastHeaderForShard(0)
	lastHeaderMeta := bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header1, lastHeader)
	assert.Equal(t, header4, lastHeaderMeta)

	bt.AddTrackedHeader(header2, []byte("hash2"))
	bt.AddTrackedHeader(header5, []byte("hash5"))
	lastHeader = bt.LastHeaderForShard(0)
	lastHeaderMeta = bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header2, lastHeader)
	assert.Equal(t, header5, lastHeaderMeta)

	bt.AddTrackedHeader(header3, []byte("hash3"))
	bt.AddTrackedHeader(header6, []byte("hash6"))
	lastHeader = bt.LastHeaderForShard(0)
	lastHeaderMeta = bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header3, lastHeader)
	assert.Equal(t, header6, lastHeaderMeta)
}

func TestNewBlockTrack_LastHeaderForShardShouldWork(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, _ := track.NewShardBlockTrack(arguments)

	header1 := &block.Header{ShardId: 0, Round: 2}
	header2 := &block.Header{ShardId: 1, Round: 2}
	header3 := &block.MetaBlock{Round: 2}

	bt.AddTrackedHeader(header1, []byte("hash1"))
	bt.AddTrackedHeader(header2, []byte("hash2"))
	bt.AddTrackedHeader(header3, []byte("hash3"))

	lastHeader := bt.LastHeaderForShard(header1.GetShardID())
	assert.Equal(t, header1, lastHeader)

	lastHeader = bt.LastHeaderForShard(header2.GetShardID())
	assert.Equal(t, header2, lastHeader)

	lastHeader = bt.LastHeaderForShard(header3.GetShardID())
	assert.Equal(t, header3, lastHeader)
}

func TestNewBlockTrack_IsShardStuckShoudReturnFalseWhenListIsEmpty(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, _ := track.NewShardBlockTrack(arguments)

	isShardStuck := bt.IsShardStuck(0)
	assert.False(t, isShardStuck)
}

func TestNewBlockTrack_IsShardStuckShoudReturnFalse(t *testing.T) {
	arguments := CreateShardMockArguments()
	bt, _ := track.NewShardBlockTrack(arguments)

	bt.AddTrackedHeader(&block.Header{Round: 1, Nonce: 1}, []byte("hash1"))
	isShardStuck := bt.IsShardStuck(0)
	assert.False(t, isShardStuck)
}

func TestNewBlockTrack_IsShardStuckShoudReturnTrue(t *testing.T) {
	arguments := CreateShardMockArguments()
	rounderMock := &mock.RounderMock{}
	arguments.Rounder = rounderMock
	bt, _ := track.NewShardBlockTrack(arguments)

	bt.AddTrackedHeader(&block.Header{Round: 1, Nonce: 1}, []byte("hash1"))
	rounderMock.RoundIndex = process.MaxRoundsWithoutCommittedBlock + 1
	isShardStuck := bt.IsShardStuck(0)
	assert.True(t, isShardStuck)
}
