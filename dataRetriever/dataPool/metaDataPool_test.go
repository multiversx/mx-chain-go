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
		&mock.CacherStub{},
		&mock.Uint64SyncMapCacherStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMetaBlockPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilMiniBlockHeaderHashesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.Uint64SyncMapCacherStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMiniBlockHashesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilShardHeaderShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.CacherStub{},
		nil,
		&mock.Uint64SyncMapCacherStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilShardHeaderPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilHeaderNoncesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		nil,
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMetaBlockNoncesPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64SyncMapCacherStub{},
		nil,
		&mock.ShardedDataStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_NilUnsingedPoolNoncesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewMetaDataPool(
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.Uint64SyncMapCacherStub{},
		&mock.ShardedDataStub{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilUnsignedTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewMetaDataPool_ConfigOk(t *testing.T) {
	t.Parallel()

	metaBlocks := &mock.CacherStub{}
	shardHeaders := &mock.CacherStub{}
	miniBlocks := &mock.CacherStub{}
	hdrsNonces := &mock.Uint64SyncMapCacherStub{}
	transactions := &mock.ShardedDataStub{}
	unsigned := &mock.ShardedDataStub{}

	tdp, err := dataPool.NewMetaDataPool(
		metaBlocks,
		miniBlocks,
		shardHeaders,
		hdrsNonces,
		transactions,
		unsigned,
	)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, metaBlocks == tdp.MetaBlocks())
	assert.True(t, shardHeaders == tdp.ShardHeaders())
	assert.True(t, miniBlocks == tdp.MiniBlocks())
	assert.True(t, hdrsNonces == tdp.HeadersNonces())
	assert.True(t, transactions == tdp.Transactions())
	assert.True(t, unsigned == tdp.UnsignedTransactions())
}
