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
}

func TestInterceptedPeerBlockBody_GetUnderlingObjectShouldReturnBlock(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	_, ok := peerBlockBody.GetUnderlyingObject().([]*block2.PeerChange)
	assert.True(t, ok)
}

func TestInterceptedPeerBlockBody_GetterSetterHash(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	peerBlockBody.SetHash(hash)

	assert.Equal(t, hash, peerBlockBody.Hash())
}

func TestInterceptedPeerBlockBody_IntegrityNilPeerChangesShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: nil}

	assert.Equal(t, process.ErrNilPeerBlockBody, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityPeerChangeWithInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: []*block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 1},
	}}

	assert.Equal(t, process.ErrInvalidShardId, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityPeerChangeWithNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: []*block2.PeerChange{
		{PubKey: nil, ShardIdDest: 0},
	}}

	assert.Equal(t, process.ErrNilPublicKey, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: []*block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}}

	assert.Equal(t, process.ErrNilShardCoordinator, peerBlk.Integrity(nil))
}

func TestInterceptedPeerBlockBody_IntegrityNilPeerBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: nil}

	assert.Equal(t, process.ErrNilPeerBlockBody, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: []*block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}}

	assert.Nil(t, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedPeerBlockBody_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	peerBlk := &block.InterceptedPeerBlockBody{PeerBlockBody: []*block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}}

	assert.Nil(t, peerBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}
