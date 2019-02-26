package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewDataPool

func TestNewDataPool_NilTransactionsShouldErr(t *testing.T) {
	tdp, err := dataPool.NewDataPool(
		nil,
		&mock.ShardedDataStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
	)

	assert.Equal(t, data.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilHeadersShouldErr(t *testing.T) {
	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		nil,
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
	)

	assert.Equal(t, data.ErrNilHeadersDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilHeaderNoncesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.CacherStub{},
		&mock.CacherStub{},
	)

	assert.Equal(t, data.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTxBlocksShouldErr(t *testing.T) {
	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.Uint64CacherStub{},
		nil,
		&mock.CacherStub{},
	)

	assert.Equal(t, data.ErrNilTxBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		nil,
	)

	assert.Equal(t, data.ErrNilPeerChangeBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_OkValsShouldWork(t *testing.T) {
	transactions := &mock.ShardedDataStub{}
	headers := &mock.ShardedDataStub{}
	headerNonces := &mock.Uint64CacherStub{}
	txBlocks := &mock.CacherStub{}
	peersBlock := &mock.CacherStub{}

	tdp, err := dataPool.NewDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
	)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, transactions == tdp.Transactions())
	assert.True(t, headers == tdp.Headers())
	assert.True(t, headerNonces == tdp.HeadersNonces())
	assert.True(t, txBlocks == tdp.MiniBlocks())
	assert.True(t, peersBlock == tdp.PeerChangesBlocks())
}
