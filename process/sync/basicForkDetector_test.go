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
	bfd.SetCheckpointNonce(3)
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

func TestBasicForkDetector_AddHeaderWithSignedAndProcessedBlockShouldSetCheckpoint(t *testing.T) {
	t.Parallel()
	hdr1 := &block.Header{Nonce: 69, PubKeysBitmap: []byte("X")}
	hash1 := []byte("hash1")
	bfd := sync.NewBasicForkDetector()
	_ = bfd.AddHeader(hdr1, hash1, true)
	assert.Equal(t, hdr1.Nonce, bfd.CheckpointNonce())
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

func TestBasicForkDetector_RemoveProcessedHeaderShouldNotRemoveWhenHeaderStoredIsNilOrEmpty(t *testing.T) {
	t.Parallel()
	pubKeysBitmap := make([]byte, 0)
	hdr1 := &block.Header{Nonce: 2, PubKeysBitmap: pubKeysBitmap}
	hash := []byte("hash1")
	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(hdr1, hash, true)
	err := bfd.RemoveProcessedHeader(3)
	assert.Equal(t, sync.ErrNilOrEmptyInfoStored, err)
	hInfos := bfd.GetHeaders(2)
	assert.NotNil(t, hInfos)
	assert.Equal(t, 1, len(hInfos))
}

func TestBasicForkDetector_RemoveProcessedHeaderShouldRemoveTheWholeNonce(t *testing.T) {
	t.Parallel()
	pubKeysBitmap := []byte("X")
	hdr1 := &block.Header{Nonce: 2, PubKeysBitmap: pubKeysBitmap}
	hash := []byte("hash1")
	bfd := sync.NewBasicForkDetector()
	_ = bfd.AddHeader(hdr1, hash, true)
	_ = bfd.RemoveProcessedHeader(2)
	hInfos := bfd.GetHeaders(2)
	assert.Nil(t, hInfos)
}

func TestBasicForkDetector_RemoveProcessedHeaderShouldRemoveOnlyProcessedHeaderStored(t *testing.T) {
	t.Parallel()
	pubKeysBitmap := []byte("X")
	hdr1 := &block.Header{Nonce: 2, PubKeysBitmap: pubKeysBitmap}
	hash1 := []byte("hash1")
	pubKeysBitmap = make([]byte, 0)
	hdr2 := &block.Header{Nonce: 2, PubKeysBitmap: pubKeysBitmap}
	hash2 := []byte("hash2")
	bfd := sync.NewBasicForkDetector()
	_ = bfd.AddHeader(hdr1, hash1, true)
	_ = bfd.AddHeader(hdr2, hash2, false)
	_ = bfd.RemoveProcessedHeader(2)
	hInfos := bfd.GetHeaders(2)
	assert.NotNil(t, hInfos)
	assert.Equal(t, 1, len(hInfos))
	assert.Equal(t, make([]byte, 0), hInfos[0].Header().PubKeysBitmap)
}

func TestBasicForkDetector_AppendShouldSetIsProcessed(t *testing.T) {
	t.Parallel()
	pubKeysBitmap := make([]byte, 0)
	hash := []byte("hash")
	hdr1 := &block.Header{Nonce: 2, PubKeysBitmap: pubKeysBitmap}
	hdr2 := &block.Header{Nonce: 2, PubKeysBitmap: pubKeysBitmap}
	bfd := sync.NewBasicForkDetector()
	_ = bfd.AddHeader(hdr1, hash, false)
	_ = bfd.AddHeader(hdr2, hash, true)
	hInfos := bfd.GetHeaders(2)
	assert.Equal(t, 1, len(hInfos))
	assert.True(t, hInfos[0].IsProcessed())
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

func TestBasicForkDetector_RemovePastHeadersShouldWork(t *testing.T) {
	t.Parallel()

	hdr1 := &block.Header{Nonce: 1}
	hash1 := []byte("hash1")
	hdr2 := &block.Header{Nonce: 2}
	hash2 := []byte("hash2")
	hdr3 := &block.Header{Nonce: 3}
	hash3 := []byte("hash3")

	bfd := sync.NewBasicForkDetector()

	_ = bfd.AddHeader(hdr1, hash1, false)
	_ = bfd.AddHeader(hdr2, hash2, false)
	_ = bfd.AddHeader(hdr3, hash3, false)

	bfd.RemovePastHeaders(3)

	hInfos := bfd.GetHeaders(3)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(2)
	assert.Nil(t, hInfos)

	hInfos = bfd.GetHeaders(1)
	assert.Nil(t, hInfos)
}
