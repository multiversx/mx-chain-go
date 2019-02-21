package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

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

	assert.Equal(t, hash, txBlockBody.Hash())
	assert.Equal(t, string(hash), txBlockBody.ID())
}

func TestInterceptedTxBlockBody_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	txBlockBody := block.NewInterceptedTxBlockBody()
	txBlockBody.ShardID = shard

	assert.Equal(t, shard, txBlockBody.Shard())
}

func TestInterceptedTxBlockBody_CreateShouldNotProduceNils(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()
	txBlockCreated := txBlockBody.Create()

	assert.NotNil(t, txBlockCreated)
	assert.NotNil(t, txBlockCreated.(*block.InterceptedTxBlockBody).TxBlockBody)
}

func TestInterceptedTxBlockBody_CreateShouldNotProduceSameObject(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()
	txBlockCreated := txBlockBody.Create()

	assert.False(t, txBlockBody == txBlockCreated)
	assert.False(t, txBlockCreated.(*block.InterceptedTxBlockBody).TxBlockBody ==
		txBlockBody.TxBlockBody)
}

func TestInterceptedTxBlockBody_IntegrityInvalidStateBlockShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = nil
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrNilRootHash, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityNilMiniBlocksShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}
	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0

	assert.Equal(t, process.ErrNilMiniBlocks, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityMiniblockWithNilTxHashesShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: nil},
	}

	assert.Equal(t, process.ErrNilTxHashes, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityMiniblockWithInvalidTxHashShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0), nil}},
	}

	assert.Equal(t, process.ErrNilTxHash, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityNilTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}
	txBlk.TxBlockBody = nil

	assert.Equal(t, process.ErrNilTxBlockBody, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrNilShardCoordinator, txBlk.Integrity(nil))
}

func TestInterceptedTxBlockBody_IntegrityMiniblockWithInvalidShardIdsShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 4, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrInvalidShardId, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Nil(t, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestTxBlockBodyWrapper_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 10
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Equal(t, process.ErrInvalidShardId, txBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: &block2.TxBlockBody{}}

	txBlk.RootHash = make([]byte, 0)
	txBlk.ShardID = 0
	txBlk.MiniBlocks = []block2.MiniBlock{
		{ShardID: 0, TxHashes: [][]byte{make([]byte, 0)}},
	}

	assert.Nil(t, txBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}
