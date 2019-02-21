package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestInterceptedPeerBlockBody_NewShouldNotCreateNilBlock(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()

	assert.NotNil(t, peerBlockBody.PeerBlockBody)
	assert.NotNil(t, peerBlockBody.StateBlockBody)
}

func TestInterceptedPeerBlockBody_GetUnderlingObjectShouldReturnBlock(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()

	assert.True(t, peerBlockBody.GetUnderlyingObject() == peerBlockBody.PeerBlockBody)
}

func TestInterceptedPeerBlockBody_GetterSetterHashID(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	peerBlockBody.SetHash(hash)

	assert.Equal(t, hash, peerBlockBody.Hash())
	assert.Equal(t, string(hash), peerBlockBody.ID())
}

func TestInterceptedPeerBlockBody_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	peerBlockBody.ShardID = shard

	assert.Equal(t, shard, peerBlockBody.Shard())
}

func TestInterceptedPeerBlockBody_CreateShouldNotProduceNils(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	peerBlockCreated := peerBlockBody.Create()

	assert.NotNil(t, peerBlockCreated)
	assert.NotNil(t, peerBlockCreated.(*block.InterceptedPeerBlockBody).PeerBlockBody)
}

func TestInterceptedPeerBlockBody_CreateShouldNotProduceSameObject(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	peerBlockCreated := peerBlockBody.Create()

	assert.False(t, peerBlockBody == peerBlockCreated)
	assert.False(t, peerBlockCreated.(*block.InterceptedPeerBlockBody).PeerBlockBody == peerBlockBody.PeerBlockBody)
}

func TestInterceptedPeerBlockBody_IntegrityInvalidStateBlockShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = nil
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilRootHash, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityNilPeerChangesShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = nil

	assert.Equal(t, process.ErrNilPeerChanges, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityPeerChangeWithInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 1},
	}

	assert.Equal(t, process.ErrInvalidShardId, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityPeerChangeWithNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: nil, ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilPublicKey, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilShardCoordinator, peerBlk.Integrity(nil))
}

func TestInterceptedPeerBlockBody_IntegrityNilPeerBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.PeerBlockBody = nil

	assert.Equal(t, process.ErrNilPeerBlockBody, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Nil(t, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = nil
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilRootHash, peerBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Nil(t, peerBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}
