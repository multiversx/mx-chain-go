package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewDataPool

func TestNewMetaDataPool_NilMetaBlockShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		nil,
		&mock.CacherStub{},
		&mock.HeadersCacherStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMiniBlockHashesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilShardHeaderShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilShardHeaderPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilTrieNodesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.HeadersCacherStub{},
		nil,
		&mock.Uint64SyncMapCacherStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTrieNodesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.HeadersCacherStub{},
		nil,
		&mock.ShardedDataStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilUnsingedPoolNoncesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.HeadersCacherStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilUnsignedTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_ConfigOk(t *testing.T) {
	t.Parallel()

	headers := &mock.HeadersCacherStub{}
	miniBlocks := &mock.CacherStub{}
	transactions := &mock.ShardedDataStub{}
	unsigned := &mock.ShardedDataStub{}
	trieNodes := &mock.CacherStub{}

	tdp, err := dataPool.NewMetaDataPool(
		miniBlocks,
		trieNodes,
		headers,
		transactions,
		unsigned,
		&mock.TxForCurrentBlockStub{},
	)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, headers == tdp.Headers())
	assert.True(t, miniBlocks == tdp.MiniBlocks())
	assert.True(t, transactions == tdp.Transactions())
	assert.True(t, unsigned == tdp.UnsignedTransactions())
}
