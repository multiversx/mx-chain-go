package block_test

import (
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
	"testing"
)

//------- Check

func TestInterceptedPeerBlockBodyCheckNilPeerBlodyShouldRetTrue(t *testing.T) {
	inBB := block.NewInterceptedPeerBlockBody()
	inBB.PeerBlockBody = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBodyCheckNilChangesShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedPeerBlockBody()
	inBB.Changes = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBodyCheckEmptyChangessShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedPeerBlockBody()
	inBB.Changes = make([]block2.PeerChange, 0)
	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBodyCheckEmptyPubKeyListShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedPeerBlockBody()

	change := block2.PeerChange{}

	inBB.Changes = make([]block2.PeerChange, 0)
	inBB.Changes = append(inBB.Changes, change)

	assert.False(t, inBB.Check())
}

func TestInterceptedPeerBlockBodyCheckOkValsShouldRetTrue(t *testing.T) {
	inBB := block.NewInterceptedPeerBlockBody()

	change := block2.PeerChange{}
	change.PubKey = []byte{65}

	inBB.Changes = make([]block2.PeerChange, 0)
	inBB.Changes = append(inBB.Changes, change)

	assert.True(t, inBB.Check())
}

//------- Getters and Setters

func TestInterceptedPeerBlockBodyAllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()

	peerBlockBody.PeerBlockBody.ShardId = 45
	peerBlockBody.PeerBlockBody.Changes = make([]block2.PeerChange, 0)

	assert.Equal(t, uint32(45), peerBlockBody.Shard())
	assert.Equal(t, 0, len(peerBlockBody.Changes))

	hash := []byte("aaaa")
	peerBlockBody.SetHash(hash)
	assert.Equal(t, hash, peerBlockBody.Hash())
	assert.Equal(t, string(hash), peerBlockBody.ID())

	newPeerBB := peerBlockBody.New()
	assert.NotNil(t, newPeerBB)
	assert.NotNil(t, newPeerBB.(*block.InterceptedPeerBlockBody).PeerBlockBody)

	assert.Equal(t, uint32(45), peerBlockBody.Shard())
}
