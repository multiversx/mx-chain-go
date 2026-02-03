package txcache

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
)

var expectedError = errors.New("expected error")

func createMockTxSelectionOptions(gasRequested uint64, maxNumTxs int) common.TxSelectionOptions {
	options, _ := holders.NewTxSelectionOptions(
		gasRequested,
		maxNumTxs,
		10,
		haveTimeTrueForSelection,
	)
	return options
}

func createMockTxSelectionOptionsWithTimeFunc(gasRequested uint64, maxNumTxs int, haveTimeFunc func() bool) common.TxSelectionOptions {
	options, _ := holders.NewTxSelectionOptions(
		gasRequested,
		maxNumTxs,
		10,
		haveTimeFunc,
	)
	return options
}

func createMockTxBoundsConfig() config.TxCacheBoundsConfig {
	return config.TxCacheBoundsConfig{
		MaxNumBytesPerSenderUpperBound: maxNumBytesPerSenderUpperBoundTest,
		MaxTrackedBlocks:               maxTrackedBlocks,
	}
}

func TestTxCache_SelectTransactions(t *testing.T) {
	t.Parallel()

	t.Run("should return errNilSelectionSession error", func(t *testing.T) {
		t.Parallel()

		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		_, _, err := cache.SelectTransactions(nil, options, 0)
		require.Equal(t, errNilSelectionSession, err)
	})

	t.Run("should return the error from GetRootHash", func(t *testing.T) {
		t.Parallel()

		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)
		session := &txcachemocks.SelectionSessionMock{
			GetRootHashCalled: func() ([]byte, error) {
				return nil, expectedError
			},
		}

		_, _, err := cache.SelectTransactions(session, options, 0)
		require.Equal(t, expectedError, err)
	})
}

func TestTxCache_SelectTransactions_Dummy(t *testing.T) {
	t.Run("all having same PPU", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)
		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 5)
		session.SetNonce([]byte("carol"), 1)

		cache.AddTx(createRelayedTx([]byte("hash-alice-4"), "alice", "relayer", 4))
		cache.AddTx(createRelayedTx([]byte("hash-alice-3"), "alice", "relayer", 3))
		cache.AddTx(createRelayedTx([]byte("hash-alice-2"), "alice", "relayer", 2))
		cache.AddTx(createRelayedTx([]byte("hash-alice-1"), "alice", "relayer", 1))
		cache.AddTx(createRelayedTx([]byte("hash-bob-7"), "bob", "relayer", 7))
		cache.AddTx(createRelayedTx([]byte("hash-bob-6"), "bob", "relayer", 6))
		cache.AddTx(createRelayedTx([]byte("hash-bob-5"), "bob", "relayer", 5))
		cache.AddTx(createRelayedTx([]byte("hash-carol-1"), "carol", "relayer", 1))

		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)
		require.Len(t, selected, 8)
		require.Equal(t, 400000, int(accumulatedGas))

		// Check order
		require.Equal(t, "hash-alice-1", string(selected[0].TxHash))
		require.Equal(t, "hash-alice-2", string(selected[1].TxHash))
		require.Equal(t, "hash-alice-3", string(selected[2].TxHash))
		require.Equal(t, "hash-alice-4", string(selected[3].TxHash))
		require.Equal(t, "hash-bob-5", string(selected[4].TxHash))
		require.Equal(t, "hash-bob-6", string(selected[5].TxHash))
		require.Equal(t, "hash-bob-7", string(selected[6].TxHash))
		require.Equal(t, "hash-carol-1", string(selected[7].TxHash))
	})

	t.Run("alice > carol > bob", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 5)
		session.SetNonce([]byte("carol"), 3)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withGasPrice(100).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5).withGasPrice(50).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-3"), "carol", 3).withGasPrice(75).withValue(big.NewInt(0)))

		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)
		require.Len(t, selected, 3)
		require.Equal(t, 150000, int(accumulatedGas))

		// Check order
		require.Equal(t, "hash-alice-1", string(selected[0].TxHash))
		require.Equal(t, "hash-carol-3", string(selected[1].TxHash))
		require.Equal(t, "hash-bob-5", string(selected[2].TxHash))
	})
}

func TestTxCache_SelectTransactionsWithBandwidth_Dummy(t *testing.T) {
	t.Run("transactions with no data field", func(t *testing.T) {
		options := createMockTxSelectionOptions(760000, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 5)
		session.SetNonce([]byte("carol"), 1)

		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4).withGasLimit(100000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withGasLimit(100000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withGasLimit(500000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withGasLimit(200000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-7"), "bob", 7).withGasLimit(400000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-6"), "bob", 6).withGasLimit(50000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-5"), "bob", 5).withGasLimit(50000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-1"), "carol", 1).withGasLimit(50000).withValue(big.NewInt(0)))

		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)
		require.Len(t, selected, 5)
		require.Equal(t, 750000, int(accumulatedGas))

		// Check order
		require.Equal(t, "hash-bob-5", string(selected[0].TxHash))
		require.Equal(t, "hash-bob-6", string(selected[1].TxHash))
		require.Equal(t, "hash-carol-1", string(selected[2].TxHash))
		require.Equal(t, "hash-alice-1", string(selected[3].TxHash))
		require.Equal(t, "hash-bob-7", string(selected[4].TxHash))
	})
}

func TestTxCache_SelectTransactions_HandlesNotExecutableTransactions(t *testing.T) {
	t.Run("with middle gaps", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-5"), "alice", 5).withValue(big.NewInt(0))) // gap
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 42).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 44).withValue(big.NewInt(0))) // gap
		cache.AddTx(createTx([]byte("hash-bob-45"), "bob", 45).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-10"), "carol", 10).withValue(big.NewInt(0))) // gap
		cache.AddTx(createTx([]byte("hash-carol-11"), "carol", 11).withValue(big.NewInt(0)))

		sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)

		expectedNumSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 300000, int(accumulatedGas))
	})

	t.Run("with initial gaps", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		// Good
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withValue(big.NewInt(0)))

		// Initial gap
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 44).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 45).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 46).withValue(big.NewInt(0)))

		// Good
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8).withValue(big.NewInt(0)))

		sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)

		expectedNumSelected := 3 + 0 + 2 // 3 alice + 0 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 250000, int(accumulatedGas))
	})

	t.Run("with lower nonces", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)
		session.SetNonce([]byte("carol"), 7)

		// Good
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withValue(big.NewInt(0)))

		// A few with lower nonce
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 40).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 41).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 42).withValue(big.NewInt(0)))

		// Good
		cache.AddTx(createTx([]byte("hash-carol-7"), "carol", 7).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-carol-8"), "carol", 8).withValue(big.NewInt(0)))

		sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)

		expectedNumSelected := 3 + 1 + 2 // 3 alice + 1 bob + 2 carol
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 300000, int(accumulatedGas))
	})

	t.Run("with duplicated nonces", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3a"), "alice", 3).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3b"), "alice", 3).withGasPrice(oneBillion * 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3c"), "alice", 3).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4).withValue(big.NewInt(0)))

		sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)
		require.Len(t, sorted, 4)
		require.Equal(t, 200000, int(accumulatedGas))

		require.Equal(t, "hash-alice-1", string(sorted[0].TxHash))
		require.Equal(t, "hash-alice-2", string(sorted[1].TxHash))
		require.Equal(t, "hash-alice-3b", string(sorted[2].TxHash))
		require.Equal(t, "hash-alice-4", string(sorted[3].TxHash))
	})

	t.Run("with fee exceeding balance", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetBalance([]byte("alice"), big.NewInt(150000000000000))
		session.SetNonce([]byte("bob"), 42)
		session.SetBalance([]byte("bob"), big.NewInt(70000000000000))

		// Enough balance
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withValue(big.NewInt(0)))

		// Not enough balance
		cache.AddTx(createTx([]byte("hash-bob-42"), "bob", 40).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-43"), "bob", 41).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-44"), "bob", 42).withValue(big.NewInt(0)))

		sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)

		expectedNumSelected := 3 + 1 // 3 alice + 1 bob
		require.Len(t, sorted, expectedNumSelected)
		require.Equal(t, 200000, int(accumulatedGas))
	})

	t.Run("with incorrectly guarded", func(t *testing.T) {
		options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		session := txcachemocks.NewSelectionSessionMock()
		session.SetNonce([]byte("alice"), 1)
		session.SetNonce([]byte("bob"), 42)

		session.IsIncorrectlyGuardedCalled = func(tx data.TransactionHandler) bool {
			return bytes.Equal(tx.GetData(), []byte("t"))
		}

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withData([]byte("x")).withGasLimit(100000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-42a"), "bob", 42).withData([]byte("y")).withGasLimit(100000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-43a"), "bob", 43).withData([]byte("z")).withGasLimit(100000).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-bob-43b"), "bob", 43).withData([]byte("t")).withGasLimit(100000).withValue(big.NewInt(0)))

		sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		require.NoError(t, err)
		require.Len(t, sorted, 3)
		require.Equal(t, 300000, int(accumulatedGas))

		require.Equal(t, "hash-alice-1", string(sorted[0].TxHash))
		require.Equal(t, "hash-bob-42a", string(sorted[1].TxHash))
		require.Equal(t, "hash-bob-43a", string(sorted[2].TxHash))
	})
}

func TestTxCache_SelectTransactions_WhenTransactionsAddedInReversedNonceOrder(t *testing.T) {
	options := createMockTxSelectionOptions(math.MaxUint64, math.MaxInt)
	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)

	session := txcachemocks.NewSelectionSessionMock()

	// Add "nSenders" * "nTransactionsPerSender" transactions in the cache (in reversed nonce order)
	nSenders := 1000
	nTransactionsPerSender := 100
	nTotalTransactions := nSenders * nTransactionsPerSender

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := fmt.Sprintf("sender:%d", senderTag)

		for txNonce := nTransactionsPerSender - 1; txNonce >= 0; txNonce-- {
			txHash := fmt.Sprintf("hash:%d:%d", senderTag, txNonce)
			tx := createTx([]byte(txHash), sender, uint64(txNonce)).withValue(big.NewInt(0))
			cache.AddTx(tx)
		}
	}

	require.Equal(t, uint64(nTotalTransactions), cache.CountTx())

	sorted, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
	require.NoError(t, err)
	require.Len(t, sorted, nTotalTransactions)
	require.Equal(t, 5_000_000_000, int(accumulatedGas))

	// Check order
	nonces := make(map[string]uint64, nSenders)

	for _, tx := range sorted {
		nonce := tx.Tx.GetNonce()
		sender := string(tx.Tx.GetSndAddr())
		previousNonce := nonces[sender]

		require.LessOrEqual(t, previousNonce, nonce)
		nonces[sender] = nonce
	}
}

func TestTxCache_selectTransactionsFromBunches(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))
		options := createMockTxSelectionOptions(10_000_000_000, math.MaxInt)
		selected, accumulatedGas := selectTransactionsFromBunches(virtualSession, []bunchOfTransactions{}, options)

		require.Equal(t, 0, len(selected))
		require.Equal(t, uint64(0), accumulatedGas)
	})
}

func TestBenchmarkTxCache_acquireBunchesOfTransactions(t *testing.T) {
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           1000000000,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              300001,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig:         createMockTxBoundsConfig(),
	}

	host := txcachemocks.NewMempoolHostMock()

	sw := core.NewStopWatch()

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 10000, 100)

		require.Equal(t, 1000000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 10000)
		require.Len(t, bunches[0], 100)
		require.Len(t, bunches[len(bunches)-1], 100)
	})

	t.Run("numSenders = 50000, numTransactions = 2", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 50000, 2)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 50000)
		require.Len(t, bunches[0], 2)
		require.Len(t, bunches[len(bunches)-1], 2)
	})

	t.Run("numSenders = 100000, numTransactions = 1", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 100000, 1)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 100000)
		require.Len(t, bunches[0], 1)
		require.Len(t, bunches[len(bunches)-1], 1)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 300000, 1)

		require.Equal(t, 300000, int(cache.CountTx()))

		sw.Start(t.Name())
		bunches := cache.acquireBunchesOfTransactions()
		sw.Stop(t.Name())

		require.Len(t, bunches, 300000)
		require.Len(t, bunches[0], 1)
		require.Len(t, bunches[len(bunches)-1], 1)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
	//     CPU family:           6
	//     Model:                140
	//     Thread(s) per core:   2
	//     Core(s) per socket:   4
	//
	// 0.014468s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_10000,_numTransactions_=_100)
	// 0.019183s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_50000,_numTransactions_=_2)
	// 0.013876s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_100000,_numTransactions_=_1)
	// 0.056631s (TestBenchmarkTxCache_acquireBunchesOfTransactions/numSenders_=_300000,_numTransactions_=_1)
}

func TestBenchmarkTxCache_selectTransactionsFromBunches(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numSenders = 1000, numTransactions = 1000", func(t *testing.T) {
		options := createMockTxSelectionOptions(10_000_000_000, math.MaxInt)
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(virtualSession, bunches, options)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 10000, numTransactions = 100", func(t *testing.T) {
		options := createMockTxSelectionOptions(10_000_000_000, math.MaxInt)
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(virtualSession, bunches, options)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 3", func(t *testing.T) {
		options := createMockTxSelectionOptions(10_000_000_000, math.MaxInt)
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))
		bunches := createBunchesOfTransactionsWithUniformDistribution(100000, 3)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(virtualSession, bunches, options)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		options := createMockTxSelectionOptions(10_000_000_000, math.MaxInt)
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))

		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)

		sw.Start(t.Name())
		selected, accumulatedGas := selectTransactionsFromBunches(virtualSession, bunches, options)
		sw.Stop(t.Name())

		require.Equal(t, 200000, len(selected))
		require.Equal(t, uint64(10_000_000_000), accumulatedGas)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
	//     CPU family:           6
	//     Model:                140
	//     Thread(s) per core:   2
	//     Core(s) per socket:   4
	//
	// 0.074999s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_1000,_numTransactions_=_1000)
	// 0.059256s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_10000,_numTransactions_=_100)
	// 0.389317s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_100000,_numTransactions_=_3)
	// 0.498457s (TestBenchmarkTxCache_selectTransactionsFromBunches/numSenders_=_300000,_numTransactions_=_1)
}

func TestTxCache_selectTransactionsFromBunches_loopBreaks_whenTakesTooLong(t *testing.T) {
	t.Run("numSenders = 300000, numTransactions = 1", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))
		options := createMockTxSelectionOptionsWithTimeFunc(10_000_000_000, 50_000, haveTimeFalseForSelection)
		bunches := createBunchesOfTransactionsWithUniformDistribution(300000, 1)
		selected, accumulatedGas := selectTransactionsFromBunches(virtualSession, bunches, options)

		require.Less(t, len(selected), 50_000)
		require.Less(t, int(accumulatedGas), 10_000_000_000)
	})
}

func TestBenchmarkTxCache_doSelectTransactions(t *testing.T) {
	options := createMockTxSelectionOptions(10_000_000_000, 30_000)
	config := ConfigSourceMe{
		Name:                        "untitled",
		NumChunks:                   16,
		NumBytesThreshold:           1000000000,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              300001,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig:         createMockTxBoundsConfig(),
	}

	host := txcachemocks.NewMempoolHostMock()
	session := txcachemocks.NewSelectionSessionMock()

	sw := core.NewStopWatch()

	t.Run("numSenders = 10000, numTransactions = 100, maxNum = 30_000", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 10000, 100)

		require.Equal(t, 1000000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		sw.Stop(t.Name())

		require.NoError(t, err)
		require.Equal(t, 30_000, len(selected))
		require.Equal(t, uint64(1_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 50000, numTransactions = 2, maxNum = 30_000", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 50000, 2)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		sw.Stop(t.Name())

		require.NoError(t, err)
		require.Equal(t, 30_000, len(selected))
		require.Equal(t, uint64(1_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 100000, numTransactions = 1, maxNum = 30_000", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 100000, 1)

		require.Equal(t, 100000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		sw.Stop(t.Name())

		require.NoError(t, err)
		require.Equal(t, 30_000, len(selected))
		require.Equal(t, uint64(1_500_000_000), accumulatedGas)
	})

	t.Run("numSenders = 300000, numTransactions = 1, maxNum = 30_000", func(t *testing.T) {
		cache, err := NewTxCache(config, host, 0)
		require.Nil(t, err)

		addManyTransactionsWithUniformDistribution(cache, 300000, 1)

		require.Equal(t, 300000, int(cache.CountTx()))

		sw.Start(t.Name())
		selected, accumulatedGas, err := cache.SelectTransactions(session, options, 0)
		sw.Stop(t.Name())

		require.NoError(t, err)
		require.Equal(t, 30_000, len(selected))
		require.Equal(t, uint64(1_500_000_000), accumulatedGas)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
	//     CPU family:           6
	//     Model:                140
	//     Thread(s) per core:   2
	//     Core(s) per socket:   4
	//
	// 0.048709s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_10000,_numTransactions_=_100,_maxNum_=_30_000)
	// 0.076177s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_50000,_numTransactions_=_2,_maxNum_=_30_000)
	// 0.104399s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_100000,_numTransactions_=_1,_maxNum_=_30_000)
	// 0.319060s (TestBenchmarkTxCache_doSelectTransactions/numSenders_=_300000,_numTransactions_=_1,_maxNum_=_30_000)
}

func TestTxCache_SelectionWithOffset(t *testing.T) {
	t.Run("selection skips transactions based on offset", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		// Add transactions for alice
		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-4"), "alice", 4).withValue(big.NewInt(0)))

		// Get the sender list and set offset to 2 (skip first 2 transactions)
		senderList := cache.getListForSender("alice")
		senderList.mutex.Lock()
		senderList.incrementSelectionOffset(2)
		senderList.mutex.Unlock()

		// Selection should only return transactions starting from offset
		bunches := cache.acquireBunchesOfTransactions()
		require.Len(t, bunches, 1)
		require.Len(t, bunches[0], 2)
		require.Equal(t, "hash-alice-3", string(bunches[0][0].TxHash))
		require.Equal(t, "hash-alice-4", string(bunches[0][1].TxHash))
	})

	t.Run("selection returns empty bunch when all transactions are skipped", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))

		// Set offset beyond all transactions
		senderList := cache.getListForSender("alice")
		senderList.mutex.Lock()
		senderList.incrementSelectionOffset(2)
		senderList.mutex.Unlock()

		// Selection should return no bunches since the only sender has empty selection
		bunches := cache.acquireBunchesOfTransactions()
		require.Len(t, bunches, 0)
	})

	t.Run("getTxs returns all transactions regardless of offset", func(t *testing.T) {
		boundsConfig := createMockTxBoundsConfig()
		cache := newUnconstrainedCacheToTest(boundsConfig)

		cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
		cache.AddTx(createTx([]byte("hash-alice-3"), "alice", 3).withValue(big.NewInt(0)))

		senderList := cache.getListForSender("alice")
		senderList.mutex.Lock()
		senderList.incrementSelectionOffset(2)
		senderList.mutex.Unlock()

		// getTxs should still return all transactions (for API compatibility)
		allTxs := senderList.getTxs()
		require.Len(t, allTxs, 3)

		// getTxsForSelection should return only unselected transactions
		selectableTxs := senderList.getTxsForSelection()
		require.Len(t, selectableTxs, 1)
	})
}

func TestTxCache_SetSelectionOffsetsByLastNonce(t *testing.T) {
	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)

	cache.AddTx(createTx([]byte("hash-alice-1"), "alice", 1).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-2"), "alice", 2).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-bob-1"), "bob", 1).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-bob-2"), "bob", 2).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-bob-3"), "bob", 3).withValue(big.NewInt(0)))

	// Simulate setting offsets by last nonce (as would happen after OnProposed)
	// alice's last nonce in block is 1, bob's last nonce is 2
	lastNoncePerSender := map[string]uint64{
		"alice": 1,
		"bob":   2,
	}
	cache.SetSelectionOffsetsByLastNonce(lastNoncePerSender)

	// Check offsets - should point to first tx with nonce > lastNonce
	aliceList := cache.getListForSender("alice")
	require.Equal(t, 1, aliceList.getSelectionOffset()) // points to nonce 2 (index 1)

	bobList := cache.getListForSender("bob")
	require.Equal(t, 2, bobList.getSelectionOffset()) // points to nonce 3 (index 2)

	// Selection should skip the proposed transactions
	bunches := cache.acquireBunchesOfTransactions()
	require.Len(t, bunches, 2)

	// Find alice's and bob's bunches
	var aliceBunch, bobBunch bunchOfTransactions
	for _, bunch := range bunches {
		if string(bunch[0].Tx.GetSndAddr()) == "alice" {
			aliceBunch = bunch
		} else {
			bobBunch = bunch
		}
	}

	require.Len(t, aliceBunch, 1)
	require.Equal(t, "hash-alice-2", string(aliceBunch[0].TxHash))

	require.Len(t, bobBunch, 1)
	require.Equal(t, "hash-bob-3", string(bobBunch[0].TxHash))
}

func TestTxCache_SetSelectionOffsetsByLastNonce_WithGaps(t *testing.T) {
	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)

	// Sender has transactions with nonces 10, 20, 30, 40, 50
	cache.AddTx(createTx([]byte("hash-alice-10"), "alice", 10).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-20"), "alice", 20).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-30"), "alice", 30).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-40"), "alice", 40).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-50"), "alice", 50).withValue(big.NewInt(0)))

	// Block contains transactions with nonces 30 and 40 (account nonce was 30)
	// The lastNonce in the block is 40
	lastNoncePerSender := map[string]uint64{
		"alice": 40,
	}
	cache.SetSelectionOffsetsByLastNonce(lastNoncePerSender)

	// Offset should point to first tx with nonce > 40, which is nonce 50 at index 4
	aliceList := cache.getListForSender("alice")
	require.Equal(t, 4, aliceList.getSelectionOffset())

	// Selection should return only nonce 50
	bunches := cache.acquireBunchesOfTransactions()
	require.Len(t, bunches, 1)
	require.Len(t, bunches[0], 1)
	require.Equal(t, "hash-alice-50", string(bunches[0][0].TxHash))
}

func TestTxCache_ResetSelectionOffsetsToNonce(t *testing.T) {
	boundsConfig := createMockTxBoundsConfig()
	cache := newUnconstrainedCacheToTest(boundsConfig)

	cache.AddTx(createTx([]byte("hash-alice-10"), "alice", 10).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-20"), "alice", 20).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-30"), "alice", 30).withValue(big.NewInt(0)))
	cache.AddTx(createTx([]byte("hash-alice-40"), "alice", 40).withValue(big.NewInt(0)))

	// Set offset to skip all transactions
	aliceList := cache.getListForSender("alice")
	aliceList.mutex.Lock()
	aliceList.incrementSelectionOffset(4)
	aliceList.mutex.Unlock()
	require.Equal(t, 4, aliceList.getSelectionOffset())

	// Reset offset to nonce 20 (should point to index 1)
	sendersWithFirstNonce := map[string]uint64{
		"alice": 20,
	}
	cache.ResetSelectionOffsetsToNonce(sendersWithFirstNonce)

	require.Equal(t, 1, aliceList.getSelectionOffset())

	// Selection should now return transactions starting from nonce 20
	bunches := cache.acquireBunchesOfTransactions()
	require.Len(t, bunches, 1)
	require.Len(t, bunches[0], 3)
	require.Equal(t, "hash-alice-20", string(bunches[0][0].TxHash))
}
