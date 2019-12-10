package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

func TestNewBlockTrack_ShouldErrNilHasher(t *testing.T) {
	bt, err := track.NewShardBlockTrack(
		nil,
		&mock.MarshalizerMock{},
		&mock.PoolsHolderMock{},
		&mock.RounderMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilMarshalizer(t *testing.T) {
	bt, err := track.NewShardBlockTrack(
		&mock.HasherMock{},
		nil,
		&mock.PoolsHolderMock{},
		&mock.RounderMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilPoolsHolder(t *testing.T) {
	bt, err := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		&mock.RounderMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilPoolsHolder, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilHeadersDataPool(t *testing.T) {
	bt, err := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.PoolsHolderMock{},
		&mock.RounderMock{},
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilRounder(t *testing.T) {
	bt, err := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		nil,
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilRounder, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldErrNilShardCoordinator(t *testing.T) {
	bt, err := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		nil,
		nil,
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, bt)
}

func TestNewBlockTrack_ShouldWork(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, err := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)

	assert.Nil(t, err)
	assert.NotNil(t, bt)
}

func TestNewBlockTrack_AddHeaderForShardShouldNotUpdateIfHeaderIsNil(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)
	bt.AddHeader(nil, nil)
	lastHeader := bt.LastHeaderForShard(0)
	assert.Nil(t, lastHeader)
}

func TestNewBlockTrack_AddHeaderShouldNotUpdateIfRoundIsLowerOrEqual(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)

	header1 := &block.Header{Round: 2}
	header2 := &block.Header{Round: 1}
	header3 := &block.Header{Round: 2}

	bt.AddHeader(header1, []byte("hash1"))
	lastHeader := bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)

	bt.AddHeader(header2, []byte("hash2"))
	lastHeader = bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)

	bt.AddHeader(header3, []byte("hash3"))
	lastHeader = bt.LastHeaderForShard(0)
	assert.Equal(t, header1, lastHeader)
}

func TestNewBlockTrack_AddHeaderShouldUpdate(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)

	header1 := &block.Header{Round: 2}
	header2 := &block.Header{Round: 3}
	header3 := &block.Header{Round: 4}

	header4 := &block.MetaBlock{Round: 2}
	header5 := &block.MetaBlock{Round: 3}
	header6 := &block.MetaBlock{Round: 4}

	bt.AddHeader(header1, []byte("hash1"))
	bt.AddHeader(header4, []byte("hash4"))
	lastHeader := bt.LastHeaderForShard(0)
	lastHeaderMeta := bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header1, lastHeader)
	assert.Equal(t, header4, lastHeaderMeta)

	bt.AddHeader(header2, []byte("hash2"))
	bt.AddHeader(header5, []byte("hash5"))
	lastHeader = bt.LastHeaderForShard(0)
	lastHeaderMeta = bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header2, lastHeader)
	assert.Equal(t, header5, lastHeaderMeta)

	bt.AddHeader(header3, []byte("hash3"))
	bt.AddHeader(header6, []byte("hash6"))
	lastHeader = bt.LastHeaderForShard(0)
	lastHeaderMeta = bt.LastHeaderForShard(sharding.MetachainShardId)
	assert.Equal(t, header3, lastHeader)
	assert.Equal(t, header6, lastHeaderMeta)
}

func TestNewBlockTrack_LastHeaderForShardShouldWork(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)

	header1 := &block.Header{ShardId: 0, Round: 2}
	header2 := &block.Header{ShardId: 1, Round: 2}
	header3 := &block.MetaBlock{Round: 2}

	bt.AddHeader(header1, []byte("hash1"))
	bt.AddHeader(header2, []byte("hash2"))
	bt.AddHeader(header3, []byte("hash3"))

	lastHeader := bt.LastHeaderForShard(header1.GetShardID())
	assert.Equal(t, header1, lastHeader)

	lastHeader = bt.LastHeaderForShard(header2.GetShardID())
	assert.Equal(t, header2, lastHeader)

	lastHeader = bt.LastHeaderForShard(header3.GetShardID())
	assert.Equal(t, header3, lastHeader)
}

func TestNewBlockTrack_IsShardStuckShoudReturnFalseWhenListIsEmpty(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)

	isShardStuck := bt.IsShardStuck(0)
	assert.False(t, isShardStuck)
}

func TestNewBlockTrack_IsShardStuckShoudReturnFalse(t *testing.T) {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		&mock.RounderMock{},
		shardCoordinatorMock,
		genesisBlocks,
	)

	bt.AddHeader(&block.Header{Round: 1, Nonce: 1}, []byte("hash1"))
	isShardStuck := bt.IsShardStuck(0)
	assert.False(t, isShardStuck)
}

func TestNewBlockTrack_IsShardStuckShoudReturnTrue(t *testing.T) {
	rounderMock := &mock.RounderMock{}
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	genesisBlocks := createGenesisBlocks(shardCoordinatorMock)
	bt, _ := track.NewShardBlockTrack(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		mock.NewPoolsHolderMock(),
		rounderMock,
		shardCoordinatorMock,
		genesisBlocks,
	)

	bt.AddHeader(&block.Header{Round: 1, Nonce: 1}, []byte("hash1"))
	rounderMock.RoundIndex = process.MaxRoundsWithoutCommittedBlock + 1
	isShardStuck := bt.IsShardStuck(0)
	assert.True(t, isShardStuck)
}
