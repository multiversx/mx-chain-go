package sync_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewBasicForkDetector_ShouldWork(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	assert.NotNil(t, bfd)
}

func TestBasicForkDetector_AddHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	err := bfd.AddHeader(nil, make([]byte, 0), true)

	assert.Equal(t, sync.ErrNilHeader, err)
}

func TestBasicForkDetector_AddHeaderNilHashShouldErr(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	err := bfd.AddHeader(&block.Header{}, nil, true)

	assert.Equal(t, sync.ErrNilHash, err)
}

func TestBasicForkDetector_AddHeaderLowerNonceShouldErr(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()
	bfd.SetCheckPointNonce(3)
	err := bfd.AddHeader(&block.Header{Nonce: 2}, make([]byte, 0), true)

	assert.Equal(t, sync.ErrLowerNonceInBlock, err)
}

func TestBasicForkDetector_AddHeaderNotPresentShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{}
	hash := make([]byte, 0)

	bfd := sync.NewBasicForkDetector()

	err := bfd.AddHeader(hdr, hash, true)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(0)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, hdr, hInfos[0].Header())
	assert.Equal(t, hash, hInfos[0].Hash())
}

func TestBasicForkDetector_AddHeaderPresentShouldAppend(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{}
	hash1 := []byte("hash1")

	hdr2 := &block.Header{}
	hash2 := []byte("hash2")

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(hdr1, hash1, true)
	err := bfd.AddHeader(hdr2, hash2, true)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(0)
	assert.Equal(t, 2, len(hInfos))

	assert.Equal(t, hdr1, hInfos[0].Header())
	assert.Equal(t, hash1, hInfos[0].Hash())

	assert.Equal(t, hdr2, hInfos[1].Header())
	assert.Equal(t, hash2, hInfos[1].Hash())
}

func TestBasicForkDetector_AddHeaderPresentShouldNotRewriteWhenSameHash(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{}
	hash := []byte("hash1")
	hdr2 := &block.Header{}

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(hdr1, hash, true)
	err := bfd.AddHeader(hdr2, hash, true)
	assert.Nil(t, err)

	hInfos := bfd.GetHeaders(0)
	assert.Equal(t, 1, len(hInfos))

	assert.Equal(t, hdr1, hInfos[0].Header())
	assert.Equal(t, hash, hInfos[0].Hash())
}

func TestBasicForkDetector_RemoveHeadersShouldWork(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{}
	hash := []byte("hash1")

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(hdr1, hash, true)
	hInfos := bfd.GetHeaders(0)
	assert.Equal(t, 1, len(hInfos))

	bfd.RemoveHeaders(0)

	hInfos = bfd.GetHeaders(0)
	assert.Nil(t, hInfos)
}

func TestBasicForkDetector_CheckForkOnlyOneHeaderOnANonceShouldRettrue(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(&block.Header{Nonce: 0}, []byte("hash1"), true)
	_ = bfd.AddHeader(&block.Header{Nonce: 1}, []byte("hash2"), true)

	assert.False(t, bfd.CheckFork())
}

func TestBasicForkDetector_CheckForkNodeHasNonEmptyBlockShouldRettrue(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(&block.Header{Nonce: 0}, []byte("hash1"), true)
	_ = bfd.AddHeader(&block.Header{Nonce: 1}, []byte("hash2"), false)
	_ = bfd.AddHeader(&block.Header{Nonce: 1, PubKeysBitmap: []byte{1}}, []byte("hash3"), true)

	assert.False(t, bfd.CheckFork())
}

func TestBasicForkDetector_CheckForkNodeHasEmptyBlockShouldRetfalse(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(&block.Header{Nonce: 0}, []byte("hash1"), true)
	_ = bfd.AddHeader(&block.Header{Nonce: 1}, []byte("hash2"), true)
	_ = bfd.AddHeader(&block.Header{Nonce: 1, PubKeysBitmap: []byte{1}}, []byte("hash3"), false)

	assert.True(t, bfd.CheckFork())
}

func TestBasicForkDetector_CheckForkNodeHasOnlyReceivedShouldRettrue(t *testing.T) {
	t.Parallel()

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(&block.Header{Nonce: 0}, []byte("hash1"), true)
	_ = bfd.AddHeader(&block.Header{Nonce: 1}, []byte("hash2"), false)
	_ = bfd.AddHeader(&block.Header{Nonce: 1, PubKeysBitmap: []byte{1}}, []byte("hash3"), false)

	assert.False(t, bfd.CheckFork())
}
