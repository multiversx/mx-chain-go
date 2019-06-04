package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func createTestInterceptedHeader() *block.InterceptedHeader {
	return block.NewInterceptedHeader(
		mock.NewMultiSigner(),
		&mock.ChronologyValidatorStub{
			ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
				return nil
			},
		},
	)
}

func TestInterceptedHeader_NewShouldNotCreateNilHeader(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()

	assert.NotNil(t, hdr.Header)
}

func TestInterceptedHeader_GetUnderlingObjectShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()

	assert.True(t, hdr.GetUnderlyingObject() == hdr.Header)
}

func TestInterceptedHeader_GetHeaderShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()

	assert.True(t, hdr.GetHeader() == hdr.Header)
}

func TestInterceptedHeader_GetterSetterHash(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	hdr := createTestInterceptedHeader()
	hdr.SetHash(hash)

	assert.Equal(t, hash, hdr.Hash())
}

func TestInterceptedHeader_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	hdr := createTestInterceptedHeader()
	hdr.ShardId = shard

	assert.Equal(t, shard, hdr.Shard())
}

func TestInterceptedHeader_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilShardCoordinator, hdr.Integrity(nil))
}

func TestInterceptedHeader_IntegrityNilPubKeysBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = make([]byte, 0)
	hdr.ShardId = 2
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = nil
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPreviousBlockHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = nil
	hdr.ShardId = 0
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilSignature, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilRootHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.ShardId = 0
	hdr.RootHash = nil
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilRootHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilPrevRandHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.ShardId = 0
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = nil
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPrevRandSeed, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilRandHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.ShardId = 0
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = nil

	assert.Equal(t, process.ErrNilRandSeed, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.Header = nil

	assert.Equal(t, process.ErrNilBlockHeader, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityInvalidBlockBodyTypeShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = 254
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrInvalidBlockBodyType, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Nil(t, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityAndValidityNilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedHeader(
		mock.NewMultiSigner(),
		nil,
	)
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilChronologyValidator, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Nil(t, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_VerifySigOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Nil(t, hdr.VerifySig())
}
