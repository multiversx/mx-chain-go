package dataPool_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------- NewDataPool

func createMockDataPoolArgs() dataPool.DataPoolArgs {
	return dataPool.DataPoolArgs{
		Transactions:              testscommon.NewShardedDataStub(),
		UnsignedTransactions:      testscommon.NewShardedDataStub(),
		RewardTransactions:        testscommon.NewShardedDataStub(),
		Headers:                   &mock.HeadersCacherStub{},
		MiniBlocks:                cache.NewCacherStub(),
		PeerChangesBlocks:         cache.NewCacherStub(),
		TrieNodes:                 cache.NewCacherStub(),
		TrieNodesChunks:           cache.NewCacherStub(),
		CurrentBlockTransactions:  &mock.TxForCurrentBlockStub{},
		CurrentEpochValidatorInfo: &mock.ValidatorInfoForCurrentEpochStub{},
		SmartContracts:            cache.NewCacherStub(),
		PeerAuthentications:       cache.NewCacherStub(),
		Heartbeats:                cache.NewCacherStub(),
		ValidatorsInfo:            testscommon.NewShardedDataStub(),
		Proofs:                    &dataRetrieverMocks.ProofsPoolMock{},
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

func TestNewDataPool_NilPeerAuthenticationsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.PeerAuthentications = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.True(t, errors.Is(err, dataRetriever.ErrNilPeerAuthenticationPool))
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilHeartbeatsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.Heartbeats = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.True(t, errors.Is(err, dataRetriever.ErrNilHeartbeatPool))
	assert.Nil(t, tdp)
}

func TestNewDataPool_NilValidatorsInfoShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.ValidatorsInfo = nil
	tdp, err := dataPool.NewDataPool(args)

	assert.Equal(t, dataRetriever.ErrNilValidatorInfoPool, err)
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

func TestNewDataPool_NilCurrBlockTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.CurrentBlockTransactions = nil
	tdp, err := dataPool.NewDataPool(args)

	require.Nil(t, tdp)
	require.Equal(t, dataRetriever.ErrNilCurrBlockTxs, err)
}

func TestNewDataPool_NilCurrEpochValidatorInfoShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	args.CurrentEpochValidatorInfo = nil
	tdp, err := dataPool.NewDataPool(args)

	require.Nil(t, tdp)
	require.Equal(t, dataRetriever.ErrNilCurrentEpochValidatorInfo, err)
}

func TestNewDataPool_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockDataPoolArgs()
	tdp, err := dataPool.NewDataPool(args)

	assert.Nil(t, err)
	require.False(t, tdp.IsInterfaceNil())
	// pointer checking
	assert.True(t, args.Transactions == tdp.Transactions())
	assert.True(t, args.UnsignedTransactions == tdp.UnsignedTransactions())
	assert.True(t, args.RewardTransactions == tdp.RewardTransactions())
	assert.True(t, args.Headers == tdp.Headers())
	assert.True(t, args.MiniBlocks == tdp.MiniBlocks())
	assert.True(t, args.PeerChangesBlocks == tdp.PeerChangesBlocks())
	assert.True(t, args.CurrentBlockTransactions == tdp.CurrentBlockTxs())
	assert.True(t, args.CurrentEpochValidatorInfo == tdp.CurrentEpochValidatorInfo())
	assert.True(t, args.TrieNodes == tdp.TrieNodes())
	assert.True(t, args.TrieNodesChunks == tdp.TrieNodesChunks())
	assert.True(t, args.SmartContracts == tdp.SmartContracts())
	assert.True(t, args.PeerAuthentications == tdp.PeerAuthentications())
	assert.True(t, args.Heartbeats == tdp.Heartbeats())
	assert.True(t, args.ValidatorsInfo == tdp.ValidatorsInfo())
}

func TestNewDataPool_Close(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("trie nodes close returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockDataPoolArgs()
		args.TrieNodes = &cache.CacherStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		tdp, _ := dataPool.NewDataPool(args)
		assert.NotNil(t, tdp)
		err := tdp.Close()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("peer authentications close returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockDataPoolArgs()
		args.PeerAuthentications = &cache.CacherStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		tdp, _ := dataPool.NewDataPool(args)
		assert.NotNil(t, tdp)
		err := tdp.Close()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("both fail", func(t *testing.T) {
		t.Parallel()

		tnExpectedErr := errors.New("tn expected error")
		paExpectedErr := errors.New("pa expected error")
		args := createMockDataPoolArgs()
		tnCalled, paCalled := false, false
		args.TrieNodes = &cache.CacherStub{
			CloseCalled: func() error {
				tnCalled = true
				return tnExpectedErr
			},
		}
		args.PeerAuthentications = &cache.CacherStub{
			CloseCalled: func() error {
				paCalled = true
				return paExpectedErr
			},
		}
		tdp, _ := dataPool.NewDataPool(args)
		assert.NotNil(t, tdp)
		err := tdp.Close()
		assert.Equal(t, paExpectedErr, err)
		assert.True(t, tnCalled)
		assert.True(t, paCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockDataPoolArgs()
		tnCalled, paCalled := false, false
		args.TrieNodes = &cache.CacherStub{
			CloseCalled: func() error {
				tnCalled = true
				return nil
			},
		}
		args.PeerAuthentications = &cache.CacherStub{
			CloseCalled: func() error {
				paCalled = true
				return nil
			},
		}
		tdp, _ := dataPool.NewDataPool(args)
		assert.NotNil(t, tdp)
		err := tdp.Close()
		assert.Nil(t, err)
		assert.True(t, tnCalled)
		assert.True(t, paCalled)
	})
}
