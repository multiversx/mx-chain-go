package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
)

//------- InterceptedHeader

func TestInterceptedHeader_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.ShardId = 56

	hash := []byte("aaaa")
	hdr.SetHash(hash)
	assert.Equal(t, hash, hdr.Hash())
	assert.Equal(t, string(hash), hdr.ID())

	newHdr := hdr.Create()
	assert.NotNil(t, newHdr)
	assert.NotNil(t, newHdr.(*block.InterceptedHeader).Header)

	assert.Equal(t, uint32(56), hdr.Shard())
}

//------- InterceptedPeerBlockBody

func TestInterceptedPeerBlockBody_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	peerBlockBody := block.NewInterceptedPeerBlockBody()

	peerBlockBody.PeerBlockBody.ShardID = 45
	peerBlockBody.PeerBlockBody.Changes = make([]block2.PeerChange, 0)

	assert.Equal(t, uint32(45), peerBlockBody.Shard())
	assert.Equal(t, 0, len(peerBlockBody.Changes))

	hash := []byte("aaaa")
	peerBlockBody.SetHash(hash)
	assert.Equal(t, hash, peerBlockBody.Hash())
	assert.Equal(t, string(hash), peerBlockBody.ID())

	newPeerBB := peerBlockBody.Create()
	assert.NotNil(t, newPeerBB)
	assert.NotNil(t, newPeerBB.(*block.InterceptedPeerBlockBody).PeerBlockBody)

	assert.Equal(t, uint32(45), peerBlockBody.Shard())
}

//------- InterceptedStateBlockBody

func TestInterceptedStateBlockBody_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	stateBlockBody.ShardID = 45
	stateBlockBody.RootHash = []byte("aaa")

	assert.Equal(t, uint32(45), stateBlockBody.Shard())

	hash := []byte("aaaa")
	stateBlockBody.SetHash(hash)
	assert.Equal(t, hash, stateBlockBody.Hash())
	assert.Equal(t, string(hash), stateBlockBody.ID())

	newBB := stateBlockBody.Create()
	assert.NotNil(t, newBB)
	assert.NotNil(t, newBB.(*block.InterceptedStateBlockBody).StateBlockBody)

	assert.Equal(t, uint32(45), stateBlockBody.Shard())
}

//------- InterceptedTxBlockBody

func TestInterceptedTxBlockBody_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	txBlockBody.TxBlockBody.ShardID = 45
	txBlockBody.TxBlockBody.MiniBlocks = make([]block2.MiniBlock, 0)

	assert.Equal(t, uint32(45), txBlockBody.Shard())
	assert.Equal(t, 0, len(txBlockBody.MiniBlocks))

	hash := []byte("aaaa")
	txBlockBody.SetHash(hash)
	assert.Equal(t, hash, txBlockBody.Hash())
	assert.Equal(t, string(hash), txBlockBody.ID())

	newTxBB := txBlockBody.Create()
	assert.NotNil(t, newTxBB)
	assert.NotNil(t, newTxBB.(*block.InterceptedTxBlockBody).TxBlockBody)

	assert.Equal(t, uint32(45), txBlockBody.Shard())
}
