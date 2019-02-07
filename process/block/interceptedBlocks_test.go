package block_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
)

//------- InterceptedHeader

func TestInterceptedHeader_NewShouldNotCreateNilHeader(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	assert.NotNil(t, hdr.Header)
}

func TestInterceptedHeader_GetUnderlingObjectShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	assert.True(t, hdr.GetUnderlyingObject() == hdr.Header)
}

func TestInterceptedHeader_GetHeaderShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	assert.True(t, hdr.GetHeader() == hdr.Header)
}

func TestInterceptedHeader_GetterSetterHashID(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	hdr := block.NewInterceptedHeader()
	hdr.SetHash(hash)

	assert.Equal(t, hash, hdr.Hash())
}

func TestInterceptedHeader_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	hdr := block.NewInterceptedHeader()
	hdr.ShardId = shard

	assert.Equal(t, shard, hdr.Shard())
}

//------- InterceptedPeerBlockBody

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
}

func TestInterceptedPeerBlockBody_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	peerBlockBody.ShardID = shard

	assert.Equal(t, shard, peerBlockBody.Shard())
}

//------- InterceptedStateBlockBody

func TestInterceptedStateBlockBody_NewShouldNotCreateNilBlock(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	assert.NotNil(t, stateBlockBody.StateBlockBody)
}

func TestInterceptedStateBlockBody_GetUnderlingObjectShouldReturnBlock(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	assert.True(t, stateBlockBody.GetUnderlyingObject() == stateBlockBody.StateBlockBody)
}

func TestInterceptedStateBlockBody_GetterSetterHashID(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	stateBlockBody := block.NewInterceptedStateBlockBody()
	stateBlockBody.SetHash(hash)

	assert.Equal(t, hash, stateBlockBody.Hash())
}

func TestInterceptedStateBlockBody_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	stateBlockBody := block.NewInterceptedStateBlockBody()
	stateBlockBody.ShardID = shard

	assert.Equal(t, shard, stateBlockBody.Shard())
}

//------- InterceptedTxBlockBody

func TestInterceptedTxBlockBody_NewShouldNotCreateNilBlock(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	assert.NotNil(t, txBlockBody.TxBlockBody)
	assert.NotNil(t, txBlockBody.StateBlockBody)
}

func TestInterceptedTxBlockBody_GetUnderlingObjectShouldReturnBlock(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	assert.True(t, txBlockBody.GetUnderlyingObject() == txBlockBody.TxBlockBody)
}

func TestInterceptedTxBlockBody_GetterSetterHashID(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	txBlockBody := block.NewInterceptedTxBlockBody()
	txBlockBody.SetHash(hash)

	assert.Equal(t, string(hash), txBlockBody.ID())
}

func TestInterceptedTxBlockBody_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	txBlockBody := block.NewInterceptedTxBlockBody()
	txBlockBody.ShardID = shard

	assert.Equal(t, shard, txBlockBody.Shard())
}
