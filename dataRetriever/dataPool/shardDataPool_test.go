package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/stretchr/testify/assert"
)

//------- NewDataPool

func TestNewShardedDataPool_NilTransactionsShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		nil,
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, data.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilUnsignedTransactionsShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		nil,
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilUnsignedTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilHeadersShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, data.ErrNilHeadersDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilHeaderNoncesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, data.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilTxBlocksShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, data.ErrNilTxBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, data.ErrNilPeerChangeBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilMetaBlocksShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		nil,
		&mock.Uint64CacherStub{},
	)

	assert.Equal(t, data.ErrNilMetaBlockPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_NilMetaHeaderNoncesShouldErr(t *testing.T) {
	tdp, err := dataPool.NewShardedDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.CacherStub{},
		&mock.Uint64CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilMetaBlockNoncesPool, err)
	assert.Nil(t, tdp)
}

func TestNewShardedDataPool_OkValsShouldWork(t *testing.T) {
	transactions := &mock.ShardedDataStub{}
	scResults := &mock.ShardedDataStub{}
	headers := &mock.CacherStub{}
	headerNonces := &mock.Uint64CacherStub{}
	txBlocks := &mock.CacherStub{}
	peersBlock := &mock.CacherStub{}
	metaChainBlocks := &mock.CacherStub{}
	metaHeaderNonces := &mock.Uint64CacherStub{}
	tdp, err := dataPool.NewShardedDataPool(
		transactions,
		scResults,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		metaChainBlocks,
		metaHeaderNonces,
	)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, transactions == tdp.Transactions())
	assert.True(t, headers == tdp.Headers())
	assert.True(t, headerNonces == tdp.HeadersNonces())
	assert.True(t, txBlocks == tdp.MiniBlocks())
	assert.True(t, peersBlock == tdp.PeerChangesBlocks())
	assert.True(t, metaChainBlocks == tdp.MetaBlocks())
	assert.True(t, scResults == tdp.UnsignedTransactions())
}
