package mempool

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-go/txcache"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
)

var (
	transferredValue = 1
	configSourceMe   = txcache.ConfigSourceMe{
		Name:                        "test",
		NumChunks:                   16,
		NumBytesThreshold:           maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:  maxNumBytesPerSenderUpperBoundTest,
		CountThreshold:              math.MaxUint32,
		CountPerSenderThreshold:     math.MaxUint32,
		EvictionEnabled:             false,
		NumItemsToPreemptivelyEvict: 1,
		TxCacheBoundsConfig: config.TxCacheBoundsConfig{
			MaxNumBytesPerSenderUpperBound: maxNumBytesPerSenderUpperBoundTest,
		},
	}
)

// benchmark for the creation of breadcrumbs (which are created whit each proposed block)
func TestBenchmark_OnProposedWithManyTxsAndSenders(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	sw := core.NewStopWatch()

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)
		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("30_000 txs with 1000 addresses", func(t *testing.T) {
		numTxs := 30_000
		// create some fake address for each account
		accounts := createFakeAddresses(1000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("30_000 txs with 100 addresses", func(t *testing.T) {
		numTxs := 30_000
		// create some fake address for each account
		accounts := createFakeAddresses(100)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("30_000 txs with 10 addresses", func(t *testing.T) {
		numTxs := 30_000
		// create some fake address for each account
		accounts := createFakeAddresses(10)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("60_000 txs with 1000 addresses", func(t *testing.T) {
		numTxs := 60_000
		// create some fake address for each account
		accounts := createFakeAddresses(1000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("60_000 txs with 100 addresses", func(t *testing.T) {
		numTxs := 60_000
		// create some fake address for each account
		accounts := createFakeAddresses(100)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	t.Run("60_000 txs with 10 addresses", func(t *testing.T) {
		numTxs := 60_000
		// create some fake address for each account
		accounts := createFakeAddresses(10)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxs,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxs, len(selectedTransactions))

		proposedBlock1 := createProposedBlock(selectedTransactions)

		sw.Start(t.Name())

		// measure the time spent
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock1, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		sw.Stop(t.Name())

		require.Nil(t, err)
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             13th Gen Intel(R) Core(TM) i7-13700H
	//     CPU family:           6
	//     Model:                186
	//     Thread(s) per core:   2
	//     Core(s) per socket:   14
	//

	// 0.006654s (TestBenchmark_OnProposedWithManyTxsAndSenders/30_000_txs_with_10_addresses)
	// 0.008395s (TestBenchmark_OnProposedWithManyTxsAndSenders/30_000_txs_with_100_addresses)
	// 0.007358s (TestBenchmark_OnProposedWithManyTxsAndSenders/30_000_txs_with_1000_addresses)
	// 0.017164s (TestBenchmark_OnProposedWithManyTxsAndSenders/30_000_txs_with_10_000_addresses)

	// 0.016959s (TestBenchmark_OnProposedWithManyTxsAndSenders/60_000_txs_with_10_addresses)
	// 0.019727s (TestBenchmark_OnProposedWithManyTxsAndSenders/60_000_txs_with_100_addresses)
	// 0.025110s (TestBenchmark_OnProposedWithManyTxsAndSenders/60_000_txs_with_1000_addresses)
	// 0.026893s (TestBenchmark_OnProposedWithManyTxsAndSenders/60_000_txs_with_10_000_addresses)
}

// benchmark for the selection of txs
func TestBenchmark_FirstSelectionWithManyTxsAndSenders(t *testing.T) {
	sw := core.NewStopWatch()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("15_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

		sw.Start(t.Name())
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
	})

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

		sw.Start(t.Name())
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 120_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

		sw.Start(t.Name())
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
	})

	t.Run("90_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 180_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

		sw.Start(t.Name())
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
	})

	t.Run("100_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 200_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := big.NewInt(0)
		numTxsAsBigInt := big.NewInt(int64(numTxs))

		_ = initialAmount.Mul(numTxsAsBigInt, big.NewInt(int64(50_000*1_000_000_000)))
		_ = initialAmount.Add(initialAmount, big.NewInt(int64(numTxs*transferredValue)))

		selectionSession := createDefaultSelectionSessionMockWithBigInt(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration*2,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

		sw.Start(t.Name())
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
	})

	t.Run("1_000_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 2_000_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := big.NewInt(0)
		numTxsAsBigInt := big.NewInt(int64(numTxs))

		_ = initialAmount.Mul(numTxsAsBigInt, big.NewInt(int64(50_000*1_000_000_000)))
		_ = initialAmount.Add(initialAmount, big.NewInt(int64(numTxs*1)))

		selectionSession := createDefaultSelectionSessionMockWithBigInt(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000*10,
			numTxsToBeSelected,
			selectionLoopMaximumDuration*3,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)

		sw.Start(t.Name())
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             13th Gen Intel(R) Core(TM) i7-13700H
	//     CPU family:           6
	//     Model:                186
	//     Thread(s) per core:   2
	//     Core(s) per socket:   14
	//
	// 0.024594s (TestBenchmark_SelectionWithManyTxsAndSenders/15_000_txs_with_10_000_addresses)
	// 0.043228s (TestBenchmark_SelectionWithManyTxsAndSenders/30_000_txs_with_10_000_addresses)
	// 0.081093s (TestBenchmark_SelectionWithManyTxsAndSenders/60_000_txs_with_10_000_addresses)
	// 0.107394s (TestBenchmark_SelectionWithManyTxsAndSenders/90_000_txs_with_10_000_addresses)
	// 0.122278s (TestBenchmark_SelectionWithManyTxsAndSenders/100_000_txs_with_10_000_addresses)
	// 1.177650s (TestBenchmark_SelectionWithManyTxsAndSenders/1_000_000_txs_with_10_000_addresses)
}

// benchmark for the selection of txs
func TestBenchmark_SecondSelectionWithManyTxsAndSenders(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	sw := core.NewStopWatch()

	t.Run("15_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 30_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		proposedBlock := createProposedBlock(selectedTransactions)
		// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		require.Nil(t, err)

		// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
		sw.Start(t.Name())
		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		// propose the block and make sure the selection works well
		proposedBlock = createProposedBlock(selectedTransactions)
		err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, 0, len(selectedTransactions))
	})

	t.Run("30_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 60_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		proposedBlock := createProposedBlock(selectedTransactions)
		// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		require.Nil(t, err)

		// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
		sw.Start(t.Name())
		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		// propose the block and make sure the selection works well
		proposedBlock = createProposedBlock(selectedTransactions)
		err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, 0, len(selectedTransactions))
	})

	t.Run("60_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 120_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		proposedBlock := createProposedBlock(selectedTransactions)
		// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		require.Nil(t, err)

		// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
		sw.Start(t.Name())
		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		// propose the block and make sure the selection works well
		proposedBlock = createProposedBlock(selectedTransactions)
		err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, 0, len(selectedTransactions))
	})

	t.Run("90_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 180_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := int64(numTxs*50_000*1_000_000_000 + numTxs*transferredValue)

		selectionSession := createDefaultSelectionSessionMock(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		proposedBlock := createProposedBlock(selectedTransactions)
		// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		require.Nil(t, err)

		// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
		sw.Start(t.Name())
		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		// propose the block and make sure the selection works well
		proposedBlock = createProposedBlock(selectedTransactions)
		err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, 0, len(selectedTransactions))
	})

	t.Run("100_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 200_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := big.NewInt(0)
		numTxsAsBigInt := big.NewInt(int64(numTxs))

		_ = initialAmount.Mul(numTxsAsBigInt, big.NewInt(int64(50_000*1_000_000_000)))
		_ = initialAmount.Add(initialAmount, big.NewInt(int64(numTxs*transferredValue)))

		selectionSession := createDefaultSelectionSessionMockWithBigInt(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000,
			numTxsToBeSelected,
			selectionLoopMaximumDuration,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		proposedBlock := createProposedBlock(selectedTransactions)
		// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		require.Nil(t, err)

		// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
		sw.Start(t.Name())
		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		// propose the block and make sure the selection works well
		proposedBlock = createProposedBlock(selectedTransactions)
		err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, 0, len(selectedTransactions))
	})

	t.Run("1_000_000 txs with 10_000 addresses", func(t *testing.T) {
		numTxs := 2_000_000
		numTxsToBeSelected := numTxs / 2
		// create some fake address for each account
		accounts := createFakeAddresses(10_000)

		host := txcachemocks.NewMempoolHostMock()
		txpool, err := txcache.NewTxCache(configSourceMe, host)

		require.Nil(t, err)
		require.NotNil(t, txpool)

		// assuming the scenario when we always have the same sender, assure we have enough balance for fees and transfers
		initialAmount := big.NewInt(0)
		numTxsAsBigInt := big.NewInt(int64(numTxs))

		_ = initialAmount.Mul(numTxsAsBigInt, big.NewInt(int64(50_000*1_000_000_000)))
		_ = initialAmount.Add(initialAmount, big.NewInt(int64(numTxs*transferredValue)))

		selectionSession := createDefaultSelectionSessionMockWithBigInt(initialAmount)
		options := holders.NewTxSelectionOptions(
			10_000_000_000*10,
			numTxsToBeSelected,
			selectionLoopMaximumDuration*10,
			10,
		)

		nonceTracker := newNoncesTracker()
		// create numTxs random transactions
		createRandomTxs(txpool, numTxs, nonceTracker, accounts)

		require.Equal(t, numTxs, int(txpool.CountTx()))

		blockchainInfo := holders.NewBlockchainInfo([]byte("blockHash0"), 1)
		selectedTransactions, _ := txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		proposedBlock := createProposedBlock(selectedTransactions)
		// propose those txs in order to track them (create the breadcrumbs used for the virtual records)
		err = txpool.OnProposedBlock([]byte("blockHash1"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash0"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		require.Nil(t, err)

		// measure the time for the second selection (now we use the breadcrumbs to create the virtual records)
		sw.Start(t.Name())
		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		sw.Stop(t.Name())

		require.Equal(t, numTxsToBeSelected, len(selectedTransactions))

		// propose the block and make sure the selection works well
		proposedBlock = createProposedBlock(selectedTransactions)
		err = txpool.OnProposedBlock([]byte("blockHash2"), proposedBlock, &block.Header{
			Nonce:    0,
			PrevHash: []byte("blockHash1"),
			RootHash: []byte(fmt.Sprintf("rootHash%d", 0)),
		})

		selectedTransactions, _ = txpool.SelectTransactions(selectionSession, options, blockchainInfo)
		require.Equal(t, 0, len(selectedTransactions))
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}

	// (1)
	// Vendor ID:                GenuineIntel
	//   Model name:             13th Gen Intel(R) Core(TM) i7-13700H
	//     CPU family:           6
	//     Model:                186
	//     Thread(s) per core:   2
	//     Core(s) per socket:   14
	//
	// 0.040347s (TestBenchmark_SecondSelectionWithManyTxsAndSenders/15_000_txs_with_10_000_addresses)
	// 0.071081s (TestBenchmark_SecondSelectionWithManyTxsAndSenders/30_000_txs_with_10_000_addresses)
	// 0.148189s (TestBenchmark_SecondSelectionWithManyTxsAndSenders/60_000_txs_with_10_000_addresses)
	// 0.211526s (TestBenchmark_SecondSelectionWithManyTxsAndSenders/90_000_txs_with_10_000_addresses)
	// 0.227855s (TestBenchmark_SecondSelectionWithManyTxsAndSenders/100_000_txs_with_10_000_addresses)
	// 2.250175s (TestBenchmark_SecondSelectionWithManyTxsAndSenders/1_000_000_txs_with_10_000_addresses)
}
