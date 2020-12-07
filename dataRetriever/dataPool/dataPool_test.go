package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//------- NewDataPool

func TestNewDataPool_NilTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		nil,
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		&mock.HeadersCacherStub{},
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilUnsignedTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		nil,
		testscommon.NewShardedDataStub(),
		&mock.HeadersCacherStub{},
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilUnsignedTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilRewardTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		nil,
		&mock.HeadersCacherStub{},
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilRewardTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		nil,
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilHeadersDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTxBlocksShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		&mock.HeadersCacherStub{},
		nil,
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilTxBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTrieNodesShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		&mock.HeadersCacherStub{},
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		nil,
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilTrieNodesPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilSmartContractsShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		&mock.HeadersCacherStub{},
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilSmartContractsPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	t.Parallel()

	tdp, err := dataPool.NewDataPool(
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		testscommon.NewShardedDataStub(),
		&mock.HeadersCacherStub{},
		testscommon.NewCacherStub(),
		nil,
		testscommon.NewCacherStub(),
		&mock.TxForCurrentBlockStub{},
		testscommon.NewCacherStub(),
	)

	assert.Equal(t, dataRetriever.ErrNilPeerChangeBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilCurrBlockShouldErr(t *testing.T) {
	transactions := testscommon.NewShardedDataStub()
	scResults := testscommon.NewShardedDataStub()
	rewardTransactions := testscommon.NewShardedDataStub()
	headers := &mock.HeadersCacherStub{}
	txBlocks := testscommon.NewCacherStub()
	peersBlock := testscommon.NewCacherStub()
	trieNodes := testscommon.NewCacherStub()
	smartContracts := testscommon.NewCacherStub()

	tdp, err := dataPool.NewDataPool(
		transactions,
		scResults,
		rewardTransactions,
		headers,
		txBlocks,
		peersBlock,
		trieNodes,
		nil,
		smartContracts,
	)

	require.Nil(t, tdp)
	require.Equal(t, dataRetriever.ErrNilCurrBlockTxs, err)
}

func TestNewDataPool_OkValsShouldWork(t *testing.T) {
	transactions := testscommon.NewShardedDataStub()
	scResults := testscommon.NewShardedDataStub()
	rewardTransactions := testscommon.NewShardedDataStub()
	headers := &mock.HeadersCacherStub{}
	txBlocks := testscommon.NewCacherStub()
	peersBlock := testscommon.NewCacherStub()
	trieNodes := testscommon.NewCacherStub()
	currBlock := &mock.TxForCurrentBlockStub{}
	smartContracts := testscommon.NewCacherStub()

	tdp, err := dataPool.NewDataPool(
		transactions,
		scResults,
		rewardTransactions,
		headers,
		txBlocks,
		peersBlock,
		trieNodes,
		currBlock,
		smartContracts,
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
