package aotSelection

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/txcachestubs"
	"github.com/multiversx/mx-chain-go/txcache"
)

func createDefaultArgs() AOTSelectorArgs {
	return AOTSelectorArgs{
		NodesCoordinator: &shardingMocks.NodesCoordinatorStub{},
		ShardCoordinator: &mock.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 { return 0 },
		},
		KeysHandler:          &testscommon.KeysHandlerStub{},
		NodeRedundancy:       &consensusMock.NodeRedundancyHandlerStub{},
		TxCache:              &txcachestubs.TxCacheStub{},
		AccountsAdapter:      &stateMock.AccountsStub{},
		TransactionProcessor: &testscommon.TxProcessorStub{},
		TxVersionChecker:     &testscommon.TxVersionCheckerStub{},
		BlockChain:           &testscommon.ChainHandlerStub{},
		EconomicsDataHandler: &economicsmocks.EconomicsHandlerMock{
			BlockCapacityOverestimationFactorCalled: func() uint64 {
				return 200 // 2x overestimation
			},
		},
		SelectionTimeout: 100 * time.Millisecond,
		CacheSize:        5,
		MaxTxsPerBlock:   30000,
	}
}

func createLeaderArgs(leaderPubKey []byte) AOTSelectorArgs {
	args := createDefaultArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
			return shardingMocks.NewValidatorMock(leaderPubKey, 1, 0), nil, nil
		},
	}
	args.KeysHandler = &testscommon.KeysHandlerStub{
		IsOriginalPublicKeyOfTheNodeCalled: func(_ []byte) bool {
			return true
		},
		IsKeyManagedByCurrentNodeCalled: func(_ []byte) bool {
			return false
		},
	}
	args.NodeRedundancy = &consensusMock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool { return false },
	}
	return args
}

func createHeader(nonce uint64, randSeed []byte, epoch uint32) *testscommon.HeaderHandlerStub {
	return &testscommon.HeaderHandlerStub{
		GetNonceCalled:    func() uint64 { return nonce },
		GetRandSeedCalled: func() []byte { return randSeed },
		EpochField:        epoch,
	}
}

func TestNewAOTSelector(t *testing.T) {
	t.Parallel()

	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.NodesCoordinator = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilNodesCoordinator, err)
	})

	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.ShardCoordinator = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})

	t.Run("nil KeysHandler should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.KeysHandler = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilKeysHandler, err)
	})

	t.Run("nil TxCache should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.TxCache = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilTxCache, err)
	})

	t.Run("nil AccountsAdapter should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.AccountsAdapter = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilAccountsAdapter, err)
	})

	t.Run("nil TransactionProcessor should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.TransactionProcessor = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilTxProcessor, err)
	})

	t.Run("nil TxVersionChecker should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.TxVersionChecker = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilTransactionVersionChecker, err)
	})

	t.Run("nil BlockChain should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.BlockChain = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilBlockChain, err)
	})

	t.Run("nil EconomicsDataHandler should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.EconomicsDataHandler = nil
		sel, err := NewAOTSelector(args)
		require.Nil(t, sel)
		require.Equal(t, process.ErrNilEconomicsData, err)
	})

	t.Run("nil NodeRedundancy should error", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.NodeRedundancy = nil
		sel, err := NewAOTSelector(args)
		require.Equal(t, process.ErrNilNodeRedundancyHandler, err)
		require.Nil(t, sel)
	})

	t.Run("should use defaults for zero config values", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		args.SelectionTimeout = 0
		args.CacheSize = 0
		args.MaxTxsPerBlock = 0
		sel, err := NewAOTSelector(args)
		require.NoError(t, err)
		require.NotNil(t, sel)
		require.Equal(t, defaultSelectionTimeout, sel.selectionTimeout)
		require.Equal(t, defaultMaxTxsPerBlock, sel.maxTxsPerBlock)
	})

	t.Run("should work with all valid args", func(t *testing.T) {
		t.Parallel()
		args := createDefaultArgs()
		sel, err := NewAOTSelector(args)
		require.NoError(t, err)
		require.NotNil(t, sel)
		require.Equal(t, 100*time.Millisecond, sel.selectionTimeout)
		require.Equal(t, 30000, sel.maxTxsPerBlock)
	})
}

func TestAOTSelector_TriggerAOTSelectionNilHeader(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)
	require.NotPanics(t, func() {
		sel.TriggerAOTSelection(nil, 10)
	})
}

func TestAOTSelector_TriggerAOTSelectionMetachainSkips(t *testing.T) {
	t.Parallel()

	selectionStarted := false
	args := createDefaultArgs()
	args.ShardCoordinator = &mock.ShardCoordinatorStub{
		SelfIdCalled: func() uint32 { return core.MetachainShardId },
	}
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			selectionStarted = true
			return nil, 0, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)
	time.Sleep(50 * time.Millisecond)
	require.False(t, selectionStarted)
}

func TestAOTSelector_TriggerAOTSelectionNotLeaderSkips(t *testing.T) {
	t.Parallel()

	selectionStarted := false
	args := createDefaultArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
			return shardingMocks.NewValidatorMock([]byte("other-leader-key"), 1, 0), nil, nil
		},
	}
	args.KeysHandler = &testscommon.KeysHandlerStub{
		IsOriginalPublicKeyOfTheNodeCalled: func(_ []byte) bool {
			return false
		},
		IsKeyManagedByCurrentNodeCalled: func(_ []byte) bool {
			return false
		},
	}
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			selectionStarted = true
			return nil, 0, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)
	time.Sleep(50 * time.Millisecond)
	require.False(t, selectionStarted)
}

func TestAOTSelector_TriggerAOTSelectionLeaderNextRound(t *testing.T) {
	t.Parallel()

	selectionDone := make(chan struct{})
	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			close(selectionDone)
			return []*txcache.WrappedTransaction{
				{TxHash: []byte("tx1")},
				{TxHash: []byte("tx2")},
			}, 50000, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)

	select {
	case <-selectionDone:
	case <-time.After(time.Second):
		t.Fatal("selection goroutine was not started")
	}
}

func TestAOTSelector_TriggerAOTSelectionLeaderRoundN2Only(t *testing.T) {
	t.Parallel()

	selectionDone := make(chan struct{})
	leaderPubKey := []byte("self-leader-key")
	otherPubKey := []byte("other-leader-key")

	args := createDefaultArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, round uint64, _ uint32, _ uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
			// round 101 (N+1): other leader; round 102 (N+2): self leader
			if round == 101 {
				return shardingMocks.NewValidatorMock(otherPubKey, 1, 0), nil, nil
			}
			return shardingMocks.NewValidatorMock(leaderPubKey, 1, 0), nil, nil
		},
	}
	args.KeysHandler = &testscommon.KeysHandlerStub{
		IsOriginalPublicKeyOfTheNodeCalled: func(pk []byte) bool {
			return string(pk) == string(leaderPubKey)
		},
		IsKeyManagedByCurrentNodeCalled: func(_ []byte) bool {
			return false
		},
	}
	args.NodeRedundancy = &consensusMock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool { return false },
	}
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			close(selectionDone)
			return nil, 0, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)

	select {
	case <-selectionDone:
	case <-time.After(time.Second):
		t.Fatal("selection goroutine was not started for N+2 leader")
	}
}

func TestAOTSelector_IsSelfLeaderForRoundComputeError(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
			return nil, nil, errors.New("compute error")
		},
	}
	sel, _ := NewAOTSelector(args)
	require.False(t, sel.isSelfLeaderForRound([]byte("rand"), 10, 1))
}

func TestAOTSelector_IsSelfLeaderForRoundSingleKeyMatch(t *testing.T) {
	t.Parallel()

	args := createLeaderArgs([]byte("my-key"))
	sel, _ := NewAOTSelector(args)
	require.True(t, sel.isSelfLeaderForRound([]byte("rand"), 10, 1))
}

func TestAOTSelector_IsSelfLeaderForRoundMultiKeyMatch(t *testing.T) {
	t.Parallel()

	leaderPubKey := []byte("managed-key")
	args := createDefaultArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
			return shardingMocks.NewValidatorMock(leaderPubKey, 1, 0), nil, nil
		},
	}
	args.KeysHandler = &testscommon.KeysHandlerStub{
		IsOriginalPublicKeyOfTheNodeCalled: func(_ []byte) bool {
			return false
		},
		IsKeyManagedByCurrentNodeCalled: func(_ []byte) bool {
			return true // but is managed multi-key
		},
	}
	sel, _ := NewAOTSelector(args)
	require.True(t, sel.isSelfLeaderForRound([]byte("rand"), 10, 1))
}

func TestAOTSelector_IsSelfLeaderForRoundNoKeyMatch(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (nodesCoordinator.Validator, []nodesCoordinator.Validator, error) {
			return shardingMocks.NewValidatorMock([]byte("other-key"), 1, 0), nil, nil
		},
	}
	args.KeysHandler = &testscommon.KeysHandlerStub{
		IsOriginalPublicKeyOfTheNodeCalled: func(_ []byte) bool {
			return false
		},
		IsKeyManagedByCurrentNodeCalled: func(_ []byte) bool {
			return false
		},
	}
	sel, _ := NewAOTSelector(args)
	require.False(t, sel.isSelfLeaderForRound([]byte("rand"), 10, 1))
}

func TestAOTSelector_ShouldConsiderSelfKeyInConsensusMainMachine(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.NodeRedundancy = &consensusMock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool { return false },
	}
	sel, _ := NewAOTSelector(args)
	require.True(t, sel.shouldConsiderSelfKeyInConsensus())
}

func TestAOTSelector_ShouldConsiderSelfKeyInConsensusRedundancyNodeMainActive(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.NodeRedundancy = &consensusMock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled:    func() bool { return true },
		IsMainMachineActiveCalled: func() bool { return true },
	}
	sel, _ := NewAOTSelector(args)
	require.False(t, sel.shouldConsiderSelfKeyInConsensus())
}

func TestAOTSelector_ShouldConsiderSelfKeyInConsensusRedundancyNodeMainInactive(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.NodeRedundancy = &consensusMock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled:    func() bool { return true },
		IsMainMachineActiveCalled: func() bool { return false },
	}
	sel, _ := NewAOTSelector(args)
	require.True(t, sel.shouldConsiderSelfKeyInConsensus())
}

func TestAOTSelector_GetPreSelectedTransactionsCacheMiss(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)
	result, found := sel.GetPreSelectedTransactions(42)
	require.Nil(t, result)
	require.False(t, found)
}

func TestAOTSelector_GetPreSelectedTransactionsCacheHit(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)

	expected := &process.AOTSelectionResult{
		TxHashes:            [][]byte{[]byte("tx1"), []byte("tx2")},
		GasProvided:         50000,
		PredictedBlockNonce: 42,
		Randomness:          []byte("rand"),
		SelectionTimestamp:  time.Now(),
	}
	cacheKey := []byte(fmt.Sprintf("%d", 42))
	sel.cache.Put(cacheKey, expected, 1)

	result, found := sel.GetPreSelectedTransactions(42)
	require.True(t, found)
	require.NotNil(t, result)
	require.Equal(t, expected.TxHashes, result.TxHashes)
	require.Equal(t, expected.GasProvided, result.GasProvided)
	require.Equal(t, uint64(42), result.PredictedBlockNonce)
}

func TestAOTSelector_GetPreSelectedTransactionsWaitsForOngoing(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)

	expected := &process.AOTSelectionResult{
		TxHashes:            [][]byte{[]byte("tx1")},
		GasProvided:         1000,
		PredictedBlockNonce: 55,
	}

	// Simulate an ongoing selection
	sel.selectionMut.Lock()
	sel.ongoingNonce = 55
	sel.resultChan = make(chan *process.AOTSelectionResult, 1)
	resultChan := sel.resultChan
	sel.selectionMut.Unlock()

	var gotResult *process.AOTSelectionResult
	var gotFound bool
	done := make(chan struct{})
	go func() {
		gotResult, gotFound = sel.GetPreSelectedTransactions(55)
		close(done)
	}()

	// Give the goroutine time to start waiting
	time.Sleep(20 * time.Millisecond)

	// Deliver the result
	resultChan <- expected

	select {
	case <-done:
		require.True(t, gotFound)
		require.NotNil(t, gotResult)
		require.Equal(t, expected.TxHashes, gotResult.TxHashes)
	case <-time.After(time.Second):
		t.Fatal("GetPreSelectedTransactions did not return after result was delivered")
	}
}

func TestAOTSelector_GetPreSelectedTransactionsOngoingForDifferentNonce(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)

	// Setup ongoing selection for nonce 55
	sel.selectionMut.Lock()
	sel.ongoingNonce = 55
	sel.resultChan = make(chan *process.AOTSelectionResult, 1)
	sel.selectionMut.Unlock()

	// Query for different nonce should go to cache (miss)
	result, found := sel.GetPreSelectedTransactions(99)
	require.Nil(t, result)
	require.False(t, found)
}

func TestAOTSelector_CancelOngoingSelectionNoOp(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)
	require.NotPanics(t, func() {
		sel.CancelOngoingSelection()
	})
}

func TestAOTSelector_CancelOngoingSelectionCancelsContext(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, _ := NewAOTSelector(args)

	// Create a context and set it as the cancelFunc
	ctx, cancelFunc := context.WithCancel(context.Background())
	sel.selectionMut.Lock()
	sel.cancelFunc = cancelFunc
	sel.selectionMut.Unlock()

	sel.CancelOngoingSelection()

	select {
	case <-ctx.Done():
		// context cancelled as expected
	default:
		t.Fatal("context was not cancelled")
	}

	// Second call should not panic (cancelFunc is nil now)
	require.NotPanics(t, func() {
		sel.CancelOngoingSelection()
	})
}

func TestAOTSelector_RunAOTSelectionSuccess(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			return []*txcache.WrappedTransaction{
				{TxHash: []byte("hash1")},
				{TxHash: []byte("hash2")},
				{TxHash: []byte("hash3")},
			}, 75000, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.runAOTSelection(0, []byte("randomness"))

	result, found := sel.GetPreSelectedTransactions(0)
	require.True(t, found)
	require.NotNil(t, result)
	require.Equal(t, 3, len(result.TxHashes))
	require.Equal(t, []byte("hash1"), result.TxHashes[0])
	require.Equal(t, []byte("hash2"), result.TxHashes[1])
	require.Equal(t, []byte("hash3"), result.TxHashes[2])
	require.Equal(t, uint64(75000), result.GasProvided)
	require.Equal(t, uint64(0), result.PredictedBlockNonce)
	require.Equal(t, []byte("randomness"), result.Randomness)
}

func TestAOTSelector_RunAOTSelectionError(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			return nil, 0, errors.New("selection failed")
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.runAOTSelection(0, []byte("randomness"))

	result, found := sel.GetPreSelectedTransactions(0)
	require.Nil(t, result)
	require.False(t, found)
}

func TestAOTSelector_RunAOTSelectionCancelled(t *testing.T) {
	t.Parallel()

	cancelAfterSelect := make(chan struct{})
	args := createDefaultArgs()
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			// Wait for the cancel signal before returning
			<-cancelAfterSelect
			return []*txcache.WrappedTransaction{
				{TxHash: []byte("hash1")},
			}, 1000, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	done := make(chan struct{})
	go func() {
		sel.runAOTSelection(0, []byte("randomness"))
		close(done)
	}()

	// Wait for the goroutine to initialize
	time.Sleep(20 * time.Millisecond)

	// Cancel and then unblock the selection
	sel.CancelOngoingSelection()
	close(cancelAfterSelect)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runAOTSelection did not finish after cancel")
	}

	result, found := sel.GetPreSelectedTransactions(0)
	require.Nil(t, result)
	require.False(t, found)
}

func TestAOTSelector_RunAOTSelectionUsesEconomicsGasWhenZero(t *testing.T) {
	t.Parallel()

	economicsGasCalled := false
	args := createDefaultArgs()
	args.EconomicsDataHandler = &economicsmocks.EconomicsHandlerMock{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			economicsGasCalled = true
			return 2000000000
		},
		BlockCapacityOverestimationFactorCalled: func() uint64 {
			return 100
		},
	}
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			return nil, 0, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.runAOTSelection(0, []byte("randomness"))
	require.True(t, economicsGasCalled)
}

func TestAOTSelector_RunAOTSelectionEmptyResult(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			return []*txcache.WrappedTransaction{}, 0, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.runAOTSelection(0, []byte("randomness"))

	result, found := sel.GetPreSelectedTransactions(0)
	require.True(t, found)
	require.NotNil(t, result)
	require.Equal(t, 0, len(result.TxHashes))
	require.Equal(t, uint64(0), result.GasProvided)
}

func TestAOTSelector_RunAOTSelectionCleansUpOngoingState(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			return nil, 0, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	sel.runAOTSelection(0, []byte("randomness"))

	sel.selectionMut.Lock()
	require.Equal(t, uint64(0), sel.ongoingNonce)
	require.Nil(t, sel.resultChan)
	sel.selectionMut.Unlock()
}

func TestAOTSelector_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var sel *aotSelector
	require.True(t, sel.IsInterfaceNil())

	args := createDefaultArgs()
	sel, _ = NewAOTSelector(args)
	require.False(t, sel.IsInterfaceNil())
}

func TestAOTSelector_ConcurrentTriggerCancelGet(t *testing.T) {
	t.Parallel()

	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			time.Sleep(10 * time.Millisecond)
			return []*txcache.WrappedTransaction{{TxHash: []byte("tx1")}}, 1000, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(3)

		go func(nonce uint64) {
			defer wg.Done()
			sel.TriggerAOTSelection(createHeader(nonce, []byte("rand"), 1), nonce+100)
		}(uint64(i))

		go func() {
			defer wg.Done()
			sel.CancelOngoingSelection()
		}()

		go func(nonce uint64) {
			defer wg.Done()
			sel.GetPreSelectedTransactions(nonce)
		}(uint64(i))
	}

	wg.Wait()
}

func TestAOTSelector_TriggerAOTSelectionEpochChangeClearsCache(t *testing.T) {
	t.Parallel()

	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			return []*txcache.WrappedTransaction{
				{TxHash: []byte("tx1")},
			}, 1000, nil
		},
	}
	sel, err := NewAOTSelector(args)
	require.NoError(t, err)

	// Trigger at epoch 1 to set lastEpoch
	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)
	time.Sleep(100 * time.Millisecond) // wait for goroutine to complete

	// Verify cache has an entry for nonce 11
	result, found := sel.GetPreSelectedTransactions(11)
	require.True(t, found)
	require.NotNil(t, result)

	// Trigger at epoch 2 - should clear cache
	sel.TriggerAOTSelection(createHeader(20, []byte("randomness2"), 2), 200)
	time.Sleep(50 * time.Millisecond)

	// The old entry at nonce 11 should be gone
	result, found = sel.GetPreSelectedTransactions(11)
	require.False(t, found)
	require.Nil(t, result)
}

func TestAOTSelector_TriggerAOTSelectionEpochChangeCancelsOngoing(t *testing.T) {
	t.Parallel()

	selectionStarted := make(chan struct{}, 1)
	selectionBlocked := make(chan struct{})
	firstSelectionDone := make(chan struct{})
	callCount := 0
	var callMut sync.Mutex
	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			callMut.Lock()
			callCount++
			isFirst := callCount == 1
			callMut.Unlock()
			if isFirst {
				selectionStarted <- struct{}{}
				<-selectionBlocked
				defer close(firstSelectionDone)
			}
			return []*txcache.WrappedTransaction{{TxHash: []byte("tx1")}}, 1000, nil
		},
	}
	sel, err := NewAOTSelector(args)
	require.NoError(t, err)

	// Set lastEpoch to 1 first (need a non-zero lastEpoch for detection)
	sel.selectionMut.Lock()
	sel.lastEpoch = 1
	sel.selectionMut.Unlock()

	// Trigger selection at epoch 1 (starts a goroutine that blocks)
	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)

	// Wait for selection goroutine to start
	select {
	case <-selectionStarted:
	case <-time.After(time.Second):
		t.Fatal("selection goroutine was not started")
	}

	// Now trigger at epoch 2 - should cancel the ongoing selection
	sel.TriggerAOTSelection(createHeader(20, []byte("randomness2"), 2), 200)

	// Unblock the first selection goroutine and wait for it to complete
	close(selectionBlocked)
	select {
	case <-firstSelectionDone:
	case <-time.After(time.Second):
		t.Fatal("first selection did not complete in time")
	}

	// The result from the cancelled selection at nonce 11 should not be cached
	// (because cancel was triggered before the selection finished storing)
	result, found := sel.GetPreSelectedTransactions(11)
	require.False(t, found)
	require.Nil(t, result)
}

func TestAOTSelector_TriggerAOTSelectionSameEpochKeepsCache(t *testing.T) {
	t.Parallel()

	callCount := 0
	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			callCount++
			return []*txcache.WrappedTransaction{
				{TxHash: []byte(fmt.Sprintf("tx-%d", callCount))},
			}, 1000, nil
		},
	}
	sel, err := NewAOTSelector(args)
	require.NoError(t, err)

	// Trigger at epoch 1
	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)
	time.Sleep(100 * time.Millisecond)

	// Verify cache has entry for nonce 11
	result, found := sel.GetPreSelectedTransactions(11)
	require.True(t, found)
	require.NotNil(t, result)
	firstTxHash := result.TxHashes[0]

	// Trigger again at same epoch 1 with different nonce
	sel.TriggerAOTSelection(createHeader(11, []byte("randomness2"), 1), 101)
	time.Sleep(100 * time.Millisecond)

	// Old entry at nonce 11 should still be in cache (not cleared)
	result, found = sel.GetPreSelectedTransactions(11)
	require.True(t, found)
	require.NotNil(t, result)
	require.Equal(t, firstTxHash, result.TxHashes[0])
}

func TestAOTSelector_Close(t *testing.T) {
	t.Parallel()

	args := createDefaultArgs()
	sel, err := NewAOTSelector(args)
	require.NoError(t, err)

	// Close should not panic
	err = sel.Close()
	require.NoError(t, err)

	// Second close should be idempotent
	err = sel.Close()
	require.NoError(t, err)
}

func TestAOTSelector_CloseStopsOngoingSelection(t *testing.T) {
	t.Parallel()

	selectionStarted := make(chan struct{})
	selectionBlocked := make(chan struct{})

	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			close(selectionStarted)
			// Block until cancelled or signaled
			<-selectionBlocked
			return []*txcache.WrappedTransaction{{TxHash: []byte("tx1")}}, 1000, nil
		},
	}
	sel, err := NewAOTSelector(args)
	require.NoError(t, err)

	// Trigger selection
	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)

	// Wait for selection to start
	select {
	case <-selectionStarted:
	case <-time.After(time.Second):
		t.Fatal("selection did not start in time")
	}

	// Close should cancel ongoing selection
	err = sel.Close()
	require.NoError(t, err)

	// Unblock the selection
	close(selectionBlocked)

	// Operations after close should be no-ops
	sel.TriggerAOTSelection(createHeader(11, []byte("randomness2"), 1), 101)
	result, found := sel.GetPreSelectedTransactions(11)
	require.False(t, found)
	require.Nil(t, result)
}

func TestAOTSelector_ConcurrentCancelDoesNotPanic(t *testing.T) {
	t.Parallel()

	args := createLeaderArgs([]byte("self-leader-key"))
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			// Simulate some work
			time.Sleep(5 * time.Millisecond)
			return []*txcache.WrappedTransaction{{TxHash: []byte("tx1")}}, 1000, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	// Run many concurrent triggers and cancels - this used to cause double-close panic
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)

		go func(nonce uint64) {
			defer wg.Done()
			sel.TriggerAOTSelection(createHeader(nonce, []byte("rand"), 1), nonce+100)
		}(uint64(i))

		go func() {
			defer wg.Done()
			sel.CancelOngoingSelection()
		}()

		go func(nonce uint64) {
			defer wg.Done()
			sel.GetPreSelectedTransactions(nonce)
		}(uint64(i))
	}

	// This should not panic
	require.NotPanics(t, func() {
		wg.Wait()
	})

	// Close should also not panic
	require.NotPanics(t, func() {
		_ = sel.Close()
	})
}

func TestAOTSelector_GetPreSelectedTransactionsTimeout(t *testing.T) {
	t.Parallel()

	args := createLeaderArgs([]byte("self-leader-key"))
	args.SelectionTimeout = 50 * time.Millisecond
	args.TxCache = &txcachestubs.TxCacheStub{
		SimulateSelectTransactionsCalled: func(_ txcache.SelectionSession, _ common.TxSelectionOptions, _ uint64) ([]*txcache.WrappedTransaction, uint64, error) {
			// Block longer than timeout
			time.Sleep(200 * time.Millisecond)
			return []*txcache.WrappedTransaction{{TxHash: []byte("tx1")}}, 1000, nil
		},
	}
	sel, _ := NewAOTSelector(args)

	// Trigger selection
	sel.TriggerAOTSelection(createHeader(10, []byte("randomness"), 1), 100)

	// Small delay to ensure selection starts
	time.Sleep(10 * time.Millisecond)

	// GetPreSelectedTransactions should timeout and return false
	start := time.Now()
	result, found := sel.GetPreSelectedTransactions(11)
	elapsed := time.Since(start)

	require.False(t, found)
	require.Nil(t, result)
	// Should return within timeout + some margin
	require.Less(t, elapsed, 100*time.Millisecond)
}
