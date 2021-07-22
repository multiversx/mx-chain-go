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

func createMockDataPoolArgs() dataPool.DataPoolArgs {
	return dataPool.DataPoolArgs{
		Transactions:             testscommon.NewShardedDataStub(),
		UnsignedTransactions:     testscommon.NewShardedDataStub(),
		RewardTransactions:       testscommon.NewShardedDataStub(),
		Headers:                  &mock.HeadersCacherStub{},
		MiniBlocks:               testscommon.NewCacherStub(),
		PeerChangesBlocks:        testscommon.NewCacherStub(),
		TrieNodes:                testscommon.NewCacherStub(),
		TrieNodesChunks:          testscommon.NewCacherStub(),
		CurrentBlockTransactions: &mock.TxForCurrentBlockStub{},
		SmartContracts:           testscommon.NewCacherStub(),
	}
}

func TestNewDataPool_NilTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.Transactions = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilUnsignedTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.UnsignedTransactions = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilUnsignedTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilRewardTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.RewardTransactions = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilRewardTransactionPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.Headers = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilHeadersDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilMiniblockCacheShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.MiniBlocks = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilTxBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTrieNodesShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.TrieNodes = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilTrieNodesPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilTrieNodesChunksShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.TrieNodesChunks = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilTrieNodesChunksPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilSmartContractsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.SmartContracts = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilSmartContractsPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.PeerChangesBlocks = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilPeerChangeBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilCurrBlockShouldErr(t *testing.T) {

	args := createMockDataPoolArgs()
	args.CurrentBlockTransactions = nil
	tdp, err := dataPool.NewDataPool(args)

	require.Nil(t, tdp)
	require.Equal(t, dataRetriever.ErrNilCurrBlockTxs, err)
}

func TestNewDataPool_OkValsShouldWork(t *testing.T) {
	args := dataPool.DataPoolArgs{
		Transactions:             testscommon.NewShardedDataStub(),
		UnsignedTransactions:     testscommon.NewShardedDataStub(),
		RewardTransactions:       testscommon.NewShardedDataStub(),
		Headers:                  &mock.HeadersCacherStub{},
		MiniBlocks:               testscommon.NewCacherStub(),
		PeerChangesBlocks:        testscommon.NewCacherStub(),
		TrieNodes:                testscommon.NewCacherStub(),
		TrieNodesChunks:          testscommon.NewCacherStub(),
		CurrentBlockTransactions: &mock.TxForCurrentBlockStub{},
		SmartContracts:           testscommon.NewCacherStub(),
	}

	tdp, err := dataPool.NewDataPool(args)

	assert.Nil(t, err)
	require.False(t, tdp.IsInterfaceNil())
	//pointer checking
	assert.True(t, args.Transactions == tdp.Transactions())
	assert.True(t, args.UnsignedTransactions == tdp.UnsignedTransactions())
	assert.True(t, args.RewardTransactions == tdp.RewardTransactions())
	assert.True(t, args.Headers == tdp.Headers())
	assert.True(t, args.MiniBlocks == tdp.MiniBlocks())
	assert.True(t, args.PeerChangesBlocks == tdp.PeerChangesBlocks())
	assert.True(t, args.CurrentBlockTransactions == tdp.CurrentBlockTxs())
	assert.True(t, args.TrieNodes == tdp.TrieNodes())
	assert.True(t, args.TrieNodesChunks == tdp.TrieNodesChunks())
	assert.True(t, args.SmartContracts == tdp.SmartContracts())
}
