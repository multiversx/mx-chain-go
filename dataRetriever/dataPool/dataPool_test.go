package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//------- NewDataPool

func TestNewDataPool_NilTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		nil,
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.HeadersCacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilUnsignedTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		nil,
		&mock.ShardedDataStub{},
		&mock.HeadersCacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilUnsignedTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilRewardTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.HeadersCacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilRewardTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilHeadersDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTxBlocksShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.HeadersCacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.CacherStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTxBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTrieNodesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.HeadersCacherStub{},
		&mock.CacherStub{},
		&mock.CacherStub{},
		nil,
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTrieNodesPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.ShardedDataStub{},
		&mock.HeadersCacherStub{},
		&mock.CacherStub{},
		nil,
		&mock.CacherStub{},
		&mock.TxForCurrentBlockStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilPeerChangeBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilCurrBlockShouldErr(t *testing.T) {
	transactions := &mock.ShardedDataStub{}
	scResults := &mock.ShardedDataStub{}
	rewardTransactions := &mock.ShardedDataStub{}
	headers := &mock.HeadersCacherStub{}
	txBlocks := &mock.CacherStub{}
	peersBlock := &mock.CacherStub{}
	trieNodes := &mock.CacherStub{}

	tdp, err := dataPool.NewDataPool(
		transactions,
		scResults,
		rewardTransactions,
		headers,
		txBlocks,
		peersBlock,
		trieNodes,
		nil,
	)

	require.Nil(t, tdp)
	require.Equal(t, dataRetriever.ErrNilCurrBlockTxs, err)
}

func TestNewDataPool_OkValsShouldWork(t *testing.T) {
	transactions := &mock.ShardedDataStub{}
	scResults := &mock.ShardedDataStub{}
	rewardTransactions := &mock.ShardedDataStub{}
	headers := &mock.HeadersCacherStub{}
	txBlocks := &mock.CacherStub{}
	peersBlock := &mock.CacherStub{}
	trieNodes := &mock.CacherStub{}
	currBlock := &mock.TxForCurrentBlockStub{}

	tdp, err := dataPool.NewDataPool(
		transactions,
		scResults,
		rewardTransactions,
		headers,
		txBlocks,
		peersBlock,
		trieNodes,
		currBlock,
	)

	assert.Nil(t, err)
	require.False(t, tdp.IsInterfaceNil())
	//pointer checking
	assert.True(t, transactions == tdp.Transactions())
	assert.True(t, scResults == tdp.UnsignedTransactions())
	assert.True(t, rewardTransactions == tdp.RewardTransactions())
	assert.True(t, headers == tdp.Headers())
	assert.True(t, txBlocks == tdp.MiniBlocks())
	assert.True(t, peersBlock == tdp.PeerChangesBlocks())
	assert.True(t, scResults == tdp.UnsignedTransactions())
	assert.True(t, currBlock == tdp.CurrentBlockTxs())
	assert.True(t, trieNodes == tdp.TrieNodes())
}
