package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewShardBlockTracker_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, err := track.NewShardBlockTracker(nil, marshalizer, shardCoordinator, store)
	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewShardBlockTracker_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, err := track.NewShardBlockTracker(pools, nil, shardCoordinator, store)
	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewShardBlockTracker_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	store := &mock.ChainStorerMock{}

	mbt, err := track.NewShardBlockTracker(pools, marshalizer, nil, store)
	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewShardBlockTracker_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	mbt, err := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, nil)
	assert.Nil(t, mbt)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewShardBlockTracker_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, err := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	assert.Nil(t, err)
	assert.NotNil(t, mbt)
}

func TestShardBlockTracker_AddBlockShouldWork(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	hdr1 := &block.Header{Nonce: 2}
	mbt.AddBlock(hdr1)
	hdr2 := &block.Header{Nonce: 3}
	mbt.AddBlock(hdr2)
	headers := mbt.UnnotarisedBlocks()
	assert.Equal(t, 2, len(headers))
}

func TestShardBlockTracker_SetBlockBroadcastRoundShoudNotSetRoundWhenNonceDoesNotExist(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	hdr := &block.Header{Nonce: 2}
	mbt.AddBlock(hdr)
	mbt.SetBlockBroadcastRound(1, 10)
	assert.Equal(t, int32(0), mbt.BlockBroadcastRound(1))
}

func TestShardBlockTracker_SetBlockBroadcastRoundShoudSetRound(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	hdr := &block.Header{Nonce: 2}
	mbt.AddBlock(hdr)
	mbt.SetBlockBroadcastRound(2, 10)
	assert.Equal(t, int32(10), mbt.BlockBroadcastRound(2))
}

func TestShardBlockTracker_RemoveNotarisedBlocksShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	err := mbt.RemoveNotarisedBlocks(nil)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestShardBlockTracker_RemoveNotarisedBlocksShouldNotRemoveIfShardIdIsNotSelf(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	metaBlock := &block.MetaBlock{}
	shardInfo := make([]block.ShardData, 0)
	sd := block.ShardData{ShardId: 1, HeaderHash: []byte("1")}
	shardInfo = append(shardInfo, sd)
	metaBlock.ShardInfo = shardInfo
	header := &block.Header{Nonce: 1}
	mbt.AddBlock(header)
	_ = mbt.RemoveNotarisedBlocks(metaBlock)
	assert.Equal(t, 1, len(mbt.UnnotarisedBlocks()))
}

func TestShardBlockTracker_RemoveNotarisedBlocksShouldNotRemoveIfGetShardHeaderErr(t *testing.T) {
	t.Parallel()

	pools := &mock.PoolsHolderStub{
		HeadersCalled: func() storage.Cacher {
			return nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	metaBlock := &block.MetaBlock{}
	shardInfo := make([]block.ShardData, 0)
	sd := block.ShardData{ShardId: 0, HeaderHash: []byte("1")}
	shardInfo = append(shardInfo, sd)
	metaBlock.ShardInfo = shardInfo
	header := &block.Header{Nonce: 1}
	mbt.AddBlock(header)
	_ = mbt.RemoveNotarisedBlocks(metaBlock)
	assert.Equal(t, 1, len(mbt.UnnotarisedBlocks()))
}

func TestShardBlockTracker_RemoveNotarisedBlocksShouldWork(t *testing.T) {
	t.Parallel()

	header := &block.Header{Nonce: 1}

	pools := &mock.PoolsHolderStub{
		HeadersCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return header, true
				},
			}
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	store := &mock.ChainStorerMock{}

	mbt, _ := track.NewShardBlockTracker(pools, marshalizer, shardCoordinator, store)
	metaBlock := &block.MetaBlock{}
	shardInfo := make([]block.ShardData, 0)
	sd := block.ShardData{ShardId: 0, HeaderHash: []byte("1")}
	shardInfo = append(shardInfo, sd)
	metaBlock.ShardInfo = shardInfo
	mbt.AddBlock(header)
	_ = mbt.RemoveNotarisedBlocks(metaBlock)
	assert.Equal(t, 0, len(mbt.UnnotarisedBlocks()))
}
