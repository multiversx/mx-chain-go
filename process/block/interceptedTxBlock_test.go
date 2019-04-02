package block_test

import (
	"testing"

	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestInterceptedTxBlockBody_NewShouldNotCreateNilBlock(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	assert.NotNil(t, txBlockBody.TxBlockBody)
}

func TestInterceptedTxBlockBody_GetUnderlingObjectShouldReturnBlock(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	_, ok := txBlockBody.GetUnderlyingObject().(dataBlock.Body)
	assert.True(t, ok)
}

func TestInterceptedTxBlockBody_GetterSetterHash(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	txBlockBody := block.NewInterceptedTxBlockBody()
	txBlockBody.SetHash(hash)

	assert.Equal(t, hash, txBlockBody.Hash())
}

func TestInterceptedTxBlockBody_IntegrityNilMiniBlocksShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: nil}

	assert.Equal(t, process.ErrNilTxBlockBody, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityMiniblockWithNilTxHashesShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: dataBlock.Body{
		{
			ReceiverShardID: 0,
			SenderShardID: 0,
			TxHashes: nil,
		},
	}}

	assert.Equal(t, process.ErrNilTxHashes, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityMiniblockWithInvalidTxHashShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: dataBlock.Body{
		{
			ReceiverShardID: 0,
			SenderShardID: 0,
			TxHashes: [][]byte{make([]byte, 0), nil},
		},
	}}

	assert.Equal(t, process.ErrNilTxHash, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityNilTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: nil}

	assert.Equal(t, process.ErrNilTxBlockBody, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: dataBlock.Body{
		{
			ReceiverShardID: 0,
			SenderShardID: 0,
			TxHashes: [][]byte{make([]byte, 0)},
		},
	}}

	assert.Equal(t, process.ErrNilShardCoordinator, txBlk.Integrity(nil))
}

func TestInterceptedTxBlockBody_IntegrityMiniblockWithInvalidShardIdsShouldErr(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: dataBlock.Body{
		{
			ReceiverShardID: 4,
			SenderShardID: 0,
			TxHashes: [][]byte{make([]byte, 0)},
		},
	}}
	assert.Equal(t, process.ErrInvalidShardId, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: dataBlock.Body{
		{
			ReceiverShardID: 0,
			SenderShardID: 0,
			TxHashes: [][]byte{make([]byte, 0)},
		},
	}}

	assert.Nil(t, txBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedTxBlockBody_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	txBlk := &block.InterceptedTxBlockBody{TxBlockBody: dataBlock.Body{
		{
			ReceiverShardID: 0,
			SenderShardID: 0,
			TxHashes: [][]byte{make([]byte, 0)},
		},
	}}

	assert.Nil(t, txBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}
