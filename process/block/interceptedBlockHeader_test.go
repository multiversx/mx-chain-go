package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestInterceptedHeader_NewShouldNotCreateNilHeader(t *testing.T) {
	t.Parallel()

	multiSig := mock.NewMultiSigner()

	hdr := block.NewInterceptedHeader(multiSig)

	assert.NotNil(t, hdr.Header)
}

func TestInterceptedHeader_GetUnderlingObjectShouldReturnHeader(t *testing.T) {
	t.Parallel()

	multiSig := mock.NewMultiSigner()
	hdr := block.NewInterceptedHeader(multiSig)

	assert.True(t, hdr.GetUnderlyingObject() == hdr.Header)
}

func TestInterceptedHeader_GetHeaderShouldReturnHeader(t *testing.T) {
	t.Parallel()

	multiSig := mock.NewMultiSigner()
	hdr := block.NewInterceptedHeader(multiSig)

	assert.True(t, hdr.GetHeader() == hdr.Header)
}

func TestInterceptedHeader_GetterSetterHash(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	multiSig := mock.NewMultiSigner()
	hdr := block.NewInterceptedHeader(multiSig)
	hdr.SetHash(hash)

	assert.Equal(t, hash, hdr.Hash())
}

func TestInterceptedHeader_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	multiSig := mock.NewMultiSigner()
	hdr := block.NewInterceptedHeader(multiSig)
	hdr.ShardId = shard

	assert.Equal(t, shard, hdr.Shard())
}

func TestInterceptedHeader_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Equal(t, process.ErrNilShardCoordinator, hdr.Integrity(nil))
}

func TestInterceptedHeader_IntegrityNilBlockBodyHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = nil
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Equal(t, process.ErrNilBlockBodyHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilPubKeysBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)
	hdr.ShardId = 2

	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = nil
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Equal(t, process.ErrNilPreviousBlockHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = nil
	hdr.Commitment = make([]byte, 0)
	hdr.ShardId = 0

	assert.Equal(t, process.ErrNilSignature, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}
	hdr.Header = nil

	assert.Equal(t, process.ErrNilBlockHeader, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityInvalidBlockBodyTypeShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Equal(t, process.ErrInvalidBlockBodyType, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilCommitmentShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = nil
	hdr.ShardId = 0

	assert.Equal(t, process.ErrNilCommitment, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Nil(t, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Nil(t, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_VerifySigOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedHeader{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Nil(t, hdr.VerifySig())
}
