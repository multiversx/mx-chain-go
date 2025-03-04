package txcache

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestSelectionSessionWrapper_accumulateConsumedBalance(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	t.Run("when sender is fee payer", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		a := createTx([]byte("a-7"), "a", 7)
		b := createTx([]byte("a-8"), "a", 8).withValue(oneQuintillionBig)

		a.precomputeFields(host)
		b.precomputeFields(host)

		sessionWrapper.accumulateConsumedBalance(a)
		require.Equal(t, "50000000000000", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())

		sessionWrapper.accumulateConsumedBalance(b)
		require.Equal(t, "1000100000000000000", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())
	})

	t.Run("when relayer is fee payer", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		a := createTx([]byte("a-7"), "a", 7).withRelayer([]byte("b")).withGasLimit(100_000)
		b := createTx([]byte("a-8"), "a", 8).withValue(oneQuintillionBig).withRelayer([]byte("b")).withGasLimit(100_000)

		a.precomputeFields(host)
		b.precomputeFields(host)

		sessionWrapper.accumulateConsumedBalance(a)
		require.Equal(t, "0", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())
		require.Equal(t, "100000000000000", sessionWrapper.getAccountRecord([]byte("b")).consumedBalance.String())

		sessionWrapper.accumulateConsumedBalance(b)
		require.Equal(t, "1000000000000000000", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())
		require.Equal(t, "200000000000000", sessionWrapper.getAccountRecord([]byte("b")).consumedBalance.String())
	})
}

func TestSelectionSessionWrapper_detectWillFeeExceedBalance(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	t.Run("unknown", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		a := createTx([]byte("tx-1"), "alice", 42)
		a.precomputeFields(host)

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(a))
	})

	t.Run("will not exceed for (a) and (b), but will exceed for (c)", func(t *testing.T) {
		a := createTx([]byte("tx-1"), "alice", 42)
		b := createTx([]byte("tx-2"), "alice", 43).withValue(oneQuintillionBig)
		c := createTx([]byte("tx-3"), "alice", 44).withValue(oneQuintillionBig)

		a.precomputeFields(host)
		b.precomputeFields(host)
		c.precomputeFields(host)

		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		session.SetBalance([]byte("alice"), big.NewInt(oneQuintillion+50000000000000+1))
		recordAlice := sessionWrapper.getAccountRecord([]byte("alice"))
		require.Equal(t, "0", recordAlice.consumedBalance.String())
		require.Equal(t, "1000050000000000001", recordAlice.initialBalance.String())

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(a))

		sessionWrapper.accumulateConsumedBalance(a)

		require.Equal(t, "50000000000000", recordAlice.consumedBalance.String())

		// Even though, in reality, that will be an invalid (but executable) transaction (insufficient balance).
		require.False(t, sessionWrapper.detectWillFeeExceedBalance(b))

		sessionWrapper.accumulateConsumedBalance(b)

		require.Equal(t, "1000100000000000000", recordAlice.consumedBalance.String())
		require.True(t, sessionWrapper.detectWillFeeExceedBalance(c))
	})

	t.Run("will not exceed for (a) and (b), but will exceed for (c) (with relayed)", func(t *testing.T) {
		a := createTx([]byte("tx-1"), "alice", 42).withRelayer([]byte("carol")).withGasLimit(100_000)
		b := createTx([]byte("tx-2"), "alice", 43).withValue(oneQuintillionBig).withRelayer([]byte("carol")).withGasLimit(100_000)
		c := createTx([]byte("tx-3"), "alice", 44).withValue(oneQuintillionBig).withRelayer([]byte("carol")).withGasLimit(100_000)

		a.precomputeFields(host)
		b.precomputeFields(host)
		c.precomputeFields(host)

		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		session.SetBalance([]byte("alice"), big.NewInt(oneQuintillion))
		session.SetBalance([]byte("carol"), big.NewInt(100000000000000+100000000000000+1))
		recordAlice := sessionWrapper.getAccountRecord([]byte("alice"))
		recordCarol := sessionWrapper.getAccountRecord([]byte("carol"))
		require.Equal(t, "0", recordAlice.consumedBalance.String())
		require.Equal(t, "1000000000000000000", recordAlice.initialBalance.String())
		require.Equal(t, "0", recordCarol.consumedBalance.String())
		require.Equal(t, "200000000000001", recordCarol.initialBalance.String())

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(a))

		sessionWrapper.accumulateConsumedBalance(a)

		require.Equal(t, "0", recordAlice.consumedBalance.String())
		require.Equal(t, "100000000000000", recordCarol.consumedBalance.String())

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(b))

		sessionWrapper.accumulateConsumedBalance(b)

		require.Equal(t, "1000000000000000000", recordAlice.consumedBalance.String())
		require.Equal(t, "200000000000000", recordCarol.consumedBalance.String())
		require.True(t, sessionWrapper.detectWillFeeExceedBalance(c))
	})
}

func TestBenchmarkSelectionSessionWrapper_getNonce(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numAccounts = 300, numTransactionsPerAccount = 100", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		numAccounts := 300
		numTransactionsPerAccount := 100
		// See "detectSkippableSender()" and "detectSkippableTransaction()".
		numCallsGetNoncePerTransaction := 2
		numCallsGetNoncePerAccount := numTransactionsPerAccount * numCallsGetNoncePerTransaction

		for i := 0; i < numAccounts; i++ {
			session.SetNonce(randomAddresses.getItem(i), uint64(i))
		}

		sw.Start(t.Name())

		for i := 0; i < numAccounts; i++ {
			for j := 0; j < numCallsGetNoncePerAccount; j++ {
				_ = sessionWrapper.getNonce(randomAddresses.getItem(i))
			}
		}

		sw.Stop(t.Name())

		require.Equal(t, numAccounts, session.NumCallsGetAccountState)
	})

	t.Run("numAccounts = 10_000, numTransactionsPerAccount = 3", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		numAccounts := 10_000
		numTransactionsPerAccount := 3
		// See "detectSkippableSender()" and "detectSkippableTransaction()".
		numCallsGetNoncePerTransaction := 2
		numCallsGetNoncePerAccount := numTransactionsPerAccount * numCallsGetNoncePerTransaction

		for i := 0; i < numAccounts; i++ {
			session.SetNonce(randomAddresses.getItem(i), uint64(i))
		}

		sw.Start(t.Name())

		for i := 0; i < numAccounts; i++ {
			for j := 0; j < numCallsGetNoncePerAccount; j++ {
				_ = sessionWrapper.getNonce(randomAddresses.getItem(i))
			}
		}

		sw.Stop(t.Name())

		require.Equal(t, numAccounts, session.NumCallsGetAccountState)
	})

	t.Run("numAccounts = 30_000, numTransactionsPerAccount = 1", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		numAccounts := 30_000
		numTransactionsPerAccount := 1
		// See "detectSkippableSender()" and "detectSkippableTransaction()".
		numCallsGetNoncePerTransaction := 2
		numCallsGetNoncePerAccount := numTransactionsPerAccount * numCallsGetNoncePerTransaction

		for i := 0; i < numAccounts; i++ {
			session.SetNonce(randomAddresses.getItem(i), uint64(i))
		}

		sw.Start(t.Name())

		for i := 0; i < numAccounts; i++ {
			for j := 0; j < numCallsGetNoncePerAccount; j++ {
				_ = sessionWrapper.getNonce(randomAddresses.getItem(i))
			}
		}

		sw.Stop(t.Name())

		require.Equal(t, numAccounts, session.NumCallsGetAccountState)
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
	// Session wrapper operations should have a negligible (or small) impact on the performance!
	// 0.000826s (TestBenchmarkSelectionSessionWrapper_getNonce/_numAccounts_=_300,_numTransactionsPerAccount=_100)
	// 0.003263s (TestBenchmarkSelectionSessionWrapper_getNonce/_numAccounts_=_10_000,_numTransactionsPerAccount=_3)
	// 0.010291s (TestBenchmarkSelectionSessionWrapper_getNonce/_numAccounts_=_30_000,_numTransactionsPerAccount=_1)
}

func TestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numSenders = 300, numTransactionsPerSender = 100", func(t *testing.T) {
		doTestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance(t, sw, 300, 100)
	})

	t.Run("numSenders = 10_000, numTransactionsPerSender = 3", func(t *testing.T) {
		doTestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance(t, sw, 10_000, 3)
	})

	t.Run("numSenders = 30_000, numTransactionsPerSender = 1", func(t *testing.T) {
		doTestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance(t, sw, 30_000, 1)
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
	// Session wrapper operations should have a negligible (or small) impact on the performance!
	// 0.006629s (TestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance/numSenders_=_300,_numTransactionsPerSender_=_100)
	// 0.010478s (TestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance/numSenders_=_10_000,_numTransactionsPerSender_=_3)
	// 0.030631s (TestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance/numSenders_=_30_000,_numTransactionsPerSender_=_1)
}

func doTestBenchmarkSelectionSessionWrapper_detectWillFeeExceedBalance(t *testing.T, sw *core.StopWatch, numSenders int, numTransactionsPerSender int) {
	fee := 100000000000000
	transferredValue := 42

	session := txcachemocks.NewSelectionSessionMock()
	sessionWrapper := newSelectionSessionWrapper(session)

	for i := 0; i < numSenders; i++ {
		session.SetBalance(randomAddresses.getItem(i), oneQuintillionBig)
	}

	transactions := make([]*WrappedTransaction, numSenders*numTransactionsPerSender)

	for i := 0; i < numTransactionsPerSender; i++ {
		for j := 0; j < numSenders; j++ {
			sender := randomAddresses.getItem(j)
			feePayer := randomAddresses.getTailItem(j)

			transactions[j*numTransactionsPerSender+i] = &WrappedTransaction{
				Tx:               &transaction.Transaction{SndAddr: sender},
				Fee:              big.NewInt(int64(fee)),
				TransferredValue: big.NewInt(int64(transferredValue)),
				FeePayer:         feePayer,
			}
		}
	}

	sw.Start(t.Name())

	for _, tx := range transactions {
		if sessionWrapper.detectWillFeeExceedBalance(tx) {
			require.Fail(t, "unexpected")
		}

		sessionWrapper.accumulateConsumedBalance(tx)
	}

	sw.Stop(t.Name())

	for i := 0; i < numSenders; i++ {
		senderRecord := sessionWrapper.getAccountRecord(randomAddresses.getItem(i))
		feePayerRecord := sessionWrapper.getAccountRecord(randomAddresses.getTailItem(i))

		require.Equal(t, transferredValue*numTransactionsPerSender, int(senderRecord.consumedBalance.Uint64()))
		require.Equal(t, fee*numTransactionsPerSender, int(feePayerRecord.consumedBalance.Uint64()))
	}
}
