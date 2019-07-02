package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createTestInterceptedMetaHeader() *block.InterceptedMetaHeader {
	return block.NewInterceptedMetaHeader(
		mock.NewMultiSigner(),
		&mock.ChronologyValidatorStub{
			ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
				return nil
			},
		},
	)
}

func TestInterceptedMetaHeader_NewShouldNotCreateNilHeader(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()

	assert.NotNil(t, hdr.MetaBlock)
}

func TestInterceptedMetaHeader_GetHeaderShouldReturnHeader(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()

	assert.True(t, hdr.GetMetaHeader() == hdr.MetaBlock)
}

func TestInterceptedMetaHeader_GetterSetterHash(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	hdr := createTestInterceptedMetaHeader()
	hdr.SetHash(hash)

	assert.Equal(t, hash, hdr.Hash())
}

func TestInterceptedMetaHeader_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilShardCoordinator, hdr.Integrity(nil))
}

func TestInterceptedMetaHeader_IntegrityNilPubKeysBitmapShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = nil
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPreviousBlockHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = nil
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilSignature, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilRootHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = nil
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilRootHash, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilPrevRandHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = nil
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPrevRandSeed, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilRandHashShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = nil

	assert.Equal(t, process.ErrNilRandSeed, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilInvalidShardIdOnShardedDataShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.ShardInfo = []block2.ShardData{
		{
			ShardId: 1,
			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
				{
					ReceiverShardId: 0,
					SenderShardId:   0,
				},
			},
		},
	}

	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilInvalidRecvShardIdOnShardedDataShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.ShardInfo = []block2.ShardData{
		{
			ShardId: 0,
			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
				{
					ReceiverShardId: 1,
					SenderShardId:   0,
				},
			},
		},
	}

	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilInvalidSenderShardIdOnShardedDataShouldErr(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.ShardInfo = []block2.ShardData{
		{
			ShardId: 0,
			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
				{
					ReceiverShardId: 0,
					SenderShardId:   1,
				},
			},
		},
	}

	assert.Equal(t, process.ErrInvalidShardId, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedMetaHeader{MetaBlock: nil}

	assert.Equal(t, process.ErrNilMetaBlockHeader, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.ShardInfo = []block2.ShardData{
		{
			ShardId: 0,
			ShardMiniBlockHeaders: []block2.ShardMiniBlockHeader{
				{
					ReceiverShardId: 0,
					SenderShardId:   0,
				},
			},
		},
	}

	assert.Nil(t, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityOkValsWithEmptyShardDataShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Nil(t, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.InterceptedMetaHeader{MetaBlock: &block2.MetaBlock{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = nil
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilPubKeysBitmap, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityAndValidityNilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	hdr := block.NewInterceptedMetaHeader(
		mock.NewMultiSigner(),
		nil,
	)
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Equal(t, process.ErrNilChronologyValidator, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Nil(t, hdr.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedMetaHeader_VerifySigOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdr := createTestInterceptedMetaHeader()
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)

	assert.Nil(t, hdr.VerifySig())
}
