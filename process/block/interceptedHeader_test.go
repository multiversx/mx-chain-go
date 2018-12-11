package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
)

//------- Check()

func TestInterceptedHeaderCheckNilHeaderShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()
	hdr.Header = nil

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckNilPrevHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = nil
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckNilPubKeysBitmapShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckNilBlockBodyHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = nil
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckInvalidBlockBodyPeerShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer + 1
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckNilSignatureShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = nil
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckNilCommitmentShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = nil
	hdr.RootHash = make([]byte, 0)

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckNilRootHashShouldRetFalse(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = nil

	assert.False(t, hdr.Check())
}

func TestInterceptedHeaderCheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)

	assert.True(t, hdr.Check())
}

//------- Getters and Setters

func TestInterceptedHeaderAllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()
	hdr := block.NewInterceptedHeader()

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.BlockBodyPeer
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.ShardId = 56

	hash := []byte("aaaa")
	hdr.SetHash(hash)
	assert.Equal(t, hash, hdr.Hash())
	assert.Equal(t, string(hash), hdr.ID())

	newHdr := hdr.New()
	assert.NotNil(t, newHdr)
	assert.NotNil(t, newHdr.(*block.InterceptedHeader).Header)

	assert.Equal(t, uint32(56), hdr.Shard())
}
