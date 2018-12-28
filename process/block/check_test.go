package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

// StateBlockBody
func TestStateBlockBodyWrapper_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 0

	assert.Equal(t, process.ErrNilShardCoordinator, stateBlk.Integrity(nil))
}

func TestStateBlockBodyWrapper_IntegrityInvalidShardShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 6

	assert.Equal(t, process.ErrInvalidShardId, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestStateBlockBodyWrapper_IntegrityNilRootHashShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = nil
	stateBlk.ShardID = 0

	assert.Equal(t, process.ErrNilRootHash, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestStateBlockBodyWrapper_IntegrityNilStateBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}
	stateBlk.StateBlockBody = nil

	assert.Equal(t, process.ErrNilStateBlockBody, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestStateBlockBodyWrapper_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 0

	assert.Nil(t, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestStateBlockBodyWrapper_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 6

	assert.Equal(t, process.ErrInvalidShardId, stateBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestStateBlockBodyWrapper_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	stateBlk := &block.StateBlockBodyWrapper{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 0

	assert.Nil(t, stateBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

// TxBlockBody
func TestTxBlockBodyWrapper_IntegrityInvalidStateBlockShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = nil
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrNilRootHash, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityNilMiniBlocksShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}
	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0

	assert.Equal(t, process.ErrNilMiniBlocks, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityMiniblockWithNilTxHashesShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: nil},
	}

	assert.Equal(t, process.ErrNilTxHashes, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityMiniblockWithInvalidTxHashShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0), nil}},
	}

	assert.Equal(t, process.ErrNilTxHash, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityNilTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}
	txBlk.TxBlockBody = nil

	assert.Equal(t, process.ErrNilTxBlockBody, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrNilShardCoordinator, txBlk.Integrity(nil))
}

func TestTxBlockBodyWrapper_IntegrityMiniblockWithInvalidShardIdsShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 4, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrInvalidShardId, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Nil(t, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 10
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrInvalidShardId, txBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txBlk := &block.TxBlockBodyWrapper{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Nil(t, txBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

// PeerBlockBodyWrapper
func TestPeerBlockBodyWrapper_IntegrityInvalidStateBlockShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = nil
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilRootHash, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityNilPeerChangesShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = nil

	assert.Equal(t, process.ErrNilPeerChanges, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityPeerChangeWithInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 1},
	}

	assert.Equal(t, process.ErrInvalidShardId, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityPeerChangeWithNilPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: nil, ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilPublicKey, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilShardCoordinator, peerBlk.Integrity(nil))
}

func TestPeerBlockBodyWrapper_IntegrityNilPeerBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}
	peerBlk.PeerBlockBody = nil

	assert.Equal(t, process.ErrNilPeerBlockBody, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Nil(t, peerBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = nil
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Equal(t, process.ErrNilRootHash, peerBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestPeerBlockBodyWrapper_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	peerBlk := &block.PeerBlockBodyWrapper{PeerBlockBody: &block2.PeerBlockBody{}}

	peerBlk.ShardID = 0
	peerBlk.RootHash = make([]byte, 0)
	peerBlk.Changes = []block2.PeerChange{
		{PubKey: make([]byte, 0), ShardIdDest: 0},
	}

	assert.Nil(t, peerBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

// Header
func TestInterceptedHeader_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}
	hdr.Header = nil

	assert.Equal(t, process.ErrNilBlockHeader, hdr.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedHeader_IntegrityInvalidBlockBodyTypeShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

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

	hdr := &block.HeaderWrapper{Header: &block2.Header{}}

	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.BlockBodyHash = make([]byte, 0)
	hdr.BlockBodyType = block2.PeerBlock
	hdr.Signature = make([]byte, 0)
	hdr.Commitment = make([]byte, 0)

	assert.Nil(t, hdr.VerifySig())
}
