package txcache

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_newVirtualSelectionSession(t *testing.T) {
	t.Parallel()

	session := txcachemocks.NewSelectionSessionMock()
	virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))
	require.NotNil(t, virtualSession)
}

func Test_getVirtualRecord(t *testing.T) {
	t.Parallel()

	t.Run("should return virtual record", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(3),
			},
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &expectedRecord,
		}

		actualRecord, err := virtualSession.getRecord([]byte("alice"))
		require.NoError(t, err)
		require.Equal(t, &expectedRecord, actualRecord)
	})

	t.Run("should return account from real session", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(0),
			},
		}
		actualRecord, err := virtualSession.getRecord([]byte("alice"))

		require.NoError(t, err)
		require.Equal(t, expectedRecord.initialNonce, actualRecord.initialNonce)
		require.Equal(t, expectedRecord.getInitialBalance(), actualRecord.getInitialBalance())
		require.Equal(t, expectedRecord.getConsumedBalance(), actualRecord.getConsumedBalance())
	})

	t.Run("should create empty record when account does not exist", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(0), false, nil
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		actualRecord, err := virtualSession.getRecord([]byte("alice"))
		require.NoError(t, err)
		require.Equal(t, core.OptionalUint64{Value: 0, HasValue: true}, actualRecord.initialNonce)
		require.Equal(t, big.NewInt(0), actualRecord.getInitialBalance())
		require.Equal(t, big.NewInt(0), actualRecord.getConsumedBalance())
	})

	t.Run("should err", func(t *testing.T) {
		t.Parallel()

		expErr := errors.New("error")
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, nil, false, expErr
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		actualRecord, err := virtualSession.getRecord([]byte("alice"))
		require.Nil(t, actualRecord)
		require.Equal(t, expErr, err)
	})
}

func Test_getNonce(t *testing.T) {
	t.Parallel()

	t.Run("should return nonce from real session", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
			},
		}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		virtualRecord, err := virtualSession.getRecord([]byte("alice"))
		require.NoError(t, err)

		actualNonce, err := virtualSession.getNonceForAccountRecord(virtualRecord)

		require.NoError(t, err)
		require.Equal(t, uint64(2), actualNonce)
	})

	t.Run("should return nonce from account record", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
			},
		}

		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(3),
			},
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &expectedRecord,
		}

		aliceRecord, err := virtualSession.getRecord([]byte("alice"))
		require.NoError(t, err)

		actualNonce, err := virtualSession.getNonceForAccountRecord(aliceRecord)

		require.NoError(t, err)
		require.Equal(t, uint64(3), actualNonce)
	})

	t.Run("should err", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		aliceRecord, err := newVirtualAccountRecord(core.OptionalUint64{Value: 0, HasValue: false}, big.NewInt(1))
		require.NoError(t, err)

		_, err = virtualSession.getNonceForAccountRecord(aliceRecord)
		require.Equal(t, errNonceNotSet, err)
	})

	t.Run("should return errNonceNotSet", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
			},
		}

		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		expectedRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    0,
				HasValue: false,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(3),
			},
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &expectedRecord,
		}

		aliceRecord, err := virtualSession.getRecord([]byte("alice"))
		require.NoError(t, err)

		_, err = virtualSession.getNonceForAccountRecord(aliceRecord)
		require.Equal(t, errNonceNotSet, err)
	})
}

func Test_accumulateConsumedBalance(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	t.Run("when sender is fee payer", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))

		a := createTx([]byte("a-7"), "a", 7)
		b := createTx([]byte("a-8"), "a", 8).withValue(oneQuintillionBig)

		a.precomputeFields(host)
		b.precomputeFields(host)

		virtualRecord1, err := virtualSession.getRecord([]byte("a"))
		require.NoError(t, err)

		err = virtualSession.accumulateConsumedBalance(a, virtualRecord1)
		require.NoError(t, err)

		virtualRecord1, err = virtualSession.getRecord([]byte("a"))
		require.NoError(t, err)
		require.Equal(t, "50000000000000", virtualRecord1.getConsumedBalance().String())

		err = virtualSession.accumulateConsumedBalance(b, virtualRecord1)
		require.NoError(t, err)

		virtualRecord2, err := virtualSession.getRecord([]byte("a"))
		require.NoError(t, err)
		require.Equal(t, "1000100000000000000", virtualRecord2.getConsumedBalance().String())

	})

	t.Run("when relayer is fee payer", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))

		a := createTx([]byte("a-7"), "a", 7).withRelayer([]byte("b")).withGasLimit(100_000)
		b := createTx([]byte("a-8"), "a", 8).withValue(oneQuintillionBig).withRelayer([]byte("b")).withGasLimit(100_000)

		a.precomputeFields(host)
		b.precomputeFields(host)

		virtualRecord1, err := virtualSession.getRecord([]byte("a"))
		require.NoError(t, err)

		err = virtualSession.accumulateConsumedBalance(a, virtualRecord1)
		require.NoError(t, err)

		virtualRecord1, err = virtualSession.getRecord([]byte("a"))
		require.NoError(t, err)
		require.Equal(t, "0", virtualRecord1.getConsumedBalance().String())

		virtualRecord2, err := virtualSession.getRecord([]byte("b"))
		require.NoError(t, err)
		require.Equal(t, "100000000000000", virtualRecord2.getConsumedBalance().String())

		err = virtualSession.accumulateConsumedBalance(b, virtualRecord1)
		require.NoError(t, err)

		virtualRecord1, err = virtualSession.getRecord([]byte("a"))
		require.NoError(t, err)
		require.Equal(t, "1000000000000000000", virtualRecord1.getConsumedBalance().String())

		virtualRecord2, err = virtualSession.getRecord([]byte("b"))
		require.NoError(t, err)
		require.Equal(t, "200000000000000", virtualRecord2.getConsumedBalance().String())
	})
}

func Test_detectWillFeeExceedBalance(t *testing.T) {
	t.Parallel()

	t.Run("should exceed balance", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		aliceRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(1),
			},
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &aliceRecord,
		}

		tx := WrappedTransaction{
			Fee:      big.NewInt(2),
			FeePayer: []byte("alice"),
		}

		actualRes := virtualSession.detectWillFeeExceedBalance(&tx)
		require.True(t, actualRes)
	})

	t.Run("should not exceed balance", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{}
		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		aliceRecord := virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(5),
				consumedBalance: big.NewInt(1),
			},
		}
		virtualSession.virtualAccountsByAddress = map[string]*virtualAccountRecord{
			"alice": &aliceRecord,
		}

		tx := WrappedTransaction{
			Fee:      big.NewInt(2),
			FeePayer: []byte("alice"),
		}

		actualRes := virtualSession.detectWillFeeExceedBalance(&tx)
		require.False(t, actualRes)
	})
}

func Test_isIncorrectlyGuarded(t *testing.T) {
	t.Parallel()

	t.Run("should return not correctly guarded", func(t *testing.T) {
		t.Parallel()

		sessionMock := txcachemocks.SelectionSessionMock{
			IsIncorrectlyGuardedCalled: func(tx data.TransactionHandler) bool {
				return true
			},
		}

		virtualSession := newVirtualSelectionSession(&sessionMock, make(map[string]*virtualAccountRecord))

		actualRes := virtualSession.isIncorrectlyGuarded(nil)
		require.True(t, actualRes)
	})
}

func TestBenchmarkVirtualSelectionSession_getRecord(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numAccounts = 300, numTransactionsPerAccount = 100", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		virtualSession := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))

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
				_, err := virtualSession.getRecord(randomAddresses.getItem(i))
				require.NoError(t, err)
			}
		}

		sw.Stop(t.Name())

		require.Equal(t, numAccounts, session.NumCallsGetAccountNonceAndBalance)
	})

	t.Run("numAccounts = 10_000, numTransactionsPerAccount = 3", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))

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
				_, err := sessionWrapper.getRecord(randomAddresses.getItem(i))
				require.NoError(t, err)
			}
		}

		sw.Stop(t.Name())

		require.Equal(t, numAccounts, session.NumCallsGetAccountNonceAndBalance)
	})

	t.Run("numAccounts = 30_000, numTransactionsPerAccount = 1", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newVirtualSelectionSession(session, make(map[string]*virtualAccountRecord))

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
				_, err := sessionWrapper.getRecord(randomAddresses.getItem(i))
				require.NoError(t, err)
			}
		}

		sw.Stop(t.Name())

		require.Equal(t, numAccounts, session.NumCallsGetAccountNonceAndBalance)
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
	// VirtualSelectionSession operations should have a negligible (or small) impact on the performance!
	// 0.011677s (TestBenchmarkVirtualSelectionSession_getNonce/_numAccounts_=_300,_numTransactionsPerAccount=_100)
	// 0.016253s (TestBenchmarkVirtualSelectionSession_getNonce/_numAccounts_=_10_000,_numTransactionsPerAccount=_3)
	// 0.028861s (TestBenchmarkVirtualSelectionSession_getNonce/_numAccounts_=_30_000,_numTransactionsPerAccount=_1)
}
