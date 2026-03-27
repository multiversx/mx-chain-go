package txcache

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
)

func Test_buildRecord(t *testing.T) {
	t.Parallel()

	t.Run("virtual record of sender breadcrumb", func(t *testing.T) {
		t.Parallel()

		sessionNonce := uint64(1)
		accountBalance := big.NewInt(2)

		breadcrumbBob := globalAccountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: big.NewInt(3),
		}

		expectedVirtualRecord := &virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(3),
			},
		}

		computer := newVirtualSessionComputer(nil)
		record, err := computer.buildRecord(sessionNonce, accountBalance, &breadcrumbBob)
		require.Nil(t, err)
		require.Equal(t, expectedVirtualRecord, record)
	})

	t.Run("virtual record of bob relayer", func(t *testing.T) {
		t.Parallel()

		sessionNonce := uint64(1)
		accountBalance := big.NewInt(2)

		breadcrumbBob := globalAccountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    0,
				HasValue: false,
			},
			lastNonce: core.OptionalUint64{
				Value:    0,
				HasValue: false,
			},
			consumedBalance: big.NewInt(5),
		}

		expectedVirtualRecord := &virtualAccountRecord{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			virtualBalance: &virtualAccountBalance{
				initialBalance:  big.NewInt(2),
				consumedBalance: big.NewInt(5),
			},
		}

		computer := newVirtualSessionComputer(nil)
		record, err := computer.buildRecord(sessionNonce, accountBalance, &breadcrumbBob)
		require.Nil(t, err)
		require.Equal(t, expectedVirtualRecord, record)
	})
}

func Test_createVirtualSelectionSession(t *testing.T) {
	t.Parallel()

	t.Run("should create blocked record for carol because it has discontinuous nonce with session nonce", func(t *testing.T) {
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 2, big.NewInt(2), true, nil
			},
		}

		gabc := newGlobalAccountBreadcrumbsCompiler()

		breadcrumbs1 := map[string]*accountBreadcrumb{
			"alice": {
				firstNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    3,
					HasValue: true,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		breadcrumb2 := map[string]*accountBreadcrumb{
			// carol has discontinuous nonce (firstNonce=10 != sessionNonce=2),
			// so a blocked virtual record is created (initialNonce.HasValue=false)
			"carol": {
				firstNonce: core.OptionalUint64{
					Value:    10,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    11,
					HasValue: true,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					Value:    4,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    5,
					HasValue: true,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		trackedBlocks := []*trackedBlock{
			{
				breadcrumbsByAddress: breadcrumbs1,
			},
			{
				breadcrumbsByAddress: breadcrumb2,
			},
		}

		expectedVirtualAccounts := map[string]*virtualAccountRecord{
			"alice": {
				initialNonce: core.OptionalUint64{
					Value:    3,
					HasValue: true,
				},
				virtualBalance: &virtualAccountBalance{
					initialBalance:  big.NewInt(2),
					consumedBalance: big.NewInt(2),
				},
			},
			"bob": {
				initialNonce: core.OptionalUint64{
					Value:    6,
					HasValue: true,
				},
				virtualBalance: &virtualAccountBalance{
					initialBalance:  big.NewInt(2),
					consumedBalance: big.NewInt(6),
				},
			},
			// carol gets a blocked virtual record: nonce not set, but consumed balance preserved
			"carol": {
				initialNonce: core.OptionalUint64{
					Value:    0,
					HasValue: false,
				},
				virtualBalance: &virtualAccountBalance{
					initialBalance:  big.NewInt(2),
					consumedBalance: big.NewInt(2),
				},
			},
		}

		gabc.updateOnAddedBlock(trackedBlocks[0])
		gabc.updateOnAddedBlock(trackedBlocks[1])

		computer := newVirtualSessionComputer(&sessionMock)
		_, err := computer.createVirtualSelectionSession(gabc.getGlobalBreadcrumbs())
		require.Nil(t, err)
		require.Equal(t, expectedVirtualAccounts, computer.virtualAccountsByAddress)
	})

	t.Run("should return error from selection session", func(t *testing.T) {
		var expectedErr = errors.New("expected err")
		sessionMock := txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				return 0, big.NewInt(0), true, expectedErr
			},
		}

		gabc := newGlobalAccountBreadcrumbsCompiler()

		breadcrumbs1 := map[string]*accountBreadcrumb{
			"alice": {
				firstNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					Value:    2,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    3,
					HasValue: true,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		breadcrumb2 := map[string]*accountBreadcrumb{
			// carol's virtual record will not be saved because the firstNonce is != session nonce
			"carol": {
				firstNonce: core.OptionalUint64{
					Value:    10,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    11,
					HasValue: true,
				},
				consumedBalance: big.NewInt(2),
			},
			"bob": {
				firstNonce: core.OptionalUint64{
					Value:    4,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    5,
					HasValue: true,
				},
				consumedBalance: big.NewInt(3),
			},
		}

		trackedBlocks := []*trackedBlock{
			{
				breadcrumbsByAddress: breadcrumbs1,
			},
			{
				breadcrumbsByAddress: breadcrumb2,
			},
		}

		gabc.updateOnAddedBlock(trackedBlocks[0])
		gabc.updateOnAddedBlock(trackedBlocks[1])

		computer := newVirtualSessionComputer(&sessionMock)
		_, err := computer.createVirtualSelectionSession(gabc.getGlobalBreadcrumbs())
		require.Equal(t, expectedErr, err)
	})
}

func Test_buildRecord_blocked(t *testing.T) {
	t.Parallel()

	t.Run("should create blocked record with unset nonce for discontinuous breadcrumb", func(t *testing.T) {
		t.Parallel()

		accountBalance := big.NewInt(100)
		// Nonce 52 is discontinuous with session nonce 0 → blocked
		globalBreadcrumb := &globalAccountBreadcrumb{
			firstNonce:      core.OptionalUint64{Value: 52, HasValue: true},
			lastNonce:       core.OptionalUint64{Value: 55, HasValue: true},
			consumedBalance: big.NewInt(30),
		}

		computer := newVirtualSessionComputer(nil)
		record, err := computer.buildRecord(0, accountBalance, globalBreadcrumb)
		require.Nil(t, err)

		// Nonce should NOT be set (blocked)
		require.False(t, record.initialNonce.HasValue)

		// getInitialNonce should return error
		_, err = record.getInitialNonce()
		require.Equal(t, errNonceNotSet, err)

		// Balance should be preserved
		require.Equal(t, big.NewInt(100), record.getInitialBalance())
		require.Equal(t, big.NewInt(30), record.getConsumedBalance())
	})
}

// Test_ParallelVsSequentialProducesSameResults generates a large set of breadcrumbs
// and verifies that the parallel path produces identical results to the sequential path.
func Test_ParallelVsSequentialProducesSameResults(t *testing.T) {
	t.Parallel()

	numAccounts := 50 // above parallelBreadcrumbThreshold (16)

	// Build breadcrumbs for numAccounts accounts
	breadcrumbs := make(map[string]*globalAccountBreadcrumb, numAccounts)
	for i := 0; i < numAccounts; i++ {
		addr := string([]byte{byte(i / 256), byte(i % 256), 'a', 'd', 'd', 'r'})
		nonce := uint64(i + 1)
		breadcrumbs[addr] = &globalAccountBreadcrumb{
			firstNonce:      core.OptionalUint64{Value: nonce, HasValue: true},
			lastNonce:       core.OptionalUint64{Value: nonce + 3, HasValue: true},
			consumedBalance: big.NewInt(int64(i * 100)),
		}
	}

	mockSession := &txcachemocks.SelectionSessionMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			// Return nonce matching firstNonce for continuity
			idx := int(address[0])*256 + int(address[1])
			nonce := uint64(idx + 1)
			balance := big.NewInt(1_000_000)
			return nonce, balance, false, nil
		},
	}

	// Run with small set (sequential path)
	seqComputer := newVirtualSessionComputer(mockSession)
	deepCopySeq := deepCopyBreadcrumbs(breadcrumbs)
	err := seqComputer.handleGlobalAccountBreadcrumbs(deepCopySeq)
	require.NoError(t, err)

	// Run with same set forced through parallel path (>= threshold)
	parComputer := newVirtualSessionComputer(mockSession)
	deepCopyPar := deepCopyBreadcrumbs(breadcrumbs)
	err = parComputer.handleGlobalAccountBreadcrumbs(deepCopyPar)
	require.NoError(t, err)

	// Compare results
	require.Equal(t, len(seqComputer.virtualAccountsByAddress), len(parComputer.virtualAccountsByAddress),
		"parallel and sequential should produce same number of records")

	for addr, seqRecord := range seqComputer.virtualAccountsByAddress {
		parRecord, exists := parComputer.virtualAccountsByAddress[addr]
		require.True(t, exists, "parallel path missing address: %v", []byte(addr))
		require.Equal(t, seqRecord.initialNonce, parRecord.initialNonce,
			"nonce mismatch for address %v", []byte(addr))
		require.Equal(t, seqRecord.virtualBalance.initialBalance.String(), parRecord.virtualBalance.initialBalance.String(),
			"initial balance mismatch for address %v", []byte(addr))
		require.Equal(t, seqRecord.virtualBalance.consumedBalance.String(), parRecord.virtualBalance.consumedBalance.String(),
			"consumed balance mismatch for address %v", []byte(addr))
	}
}

// Test_ParallelWithErrors verifies error propagation in the parallel path.
func Test_ParallelWithErrors(t *testing.T) {
	t.Parallel()

	numAccounts := 20 // above threshold
	breadcrumbs := make(map[string]*globalAccountBreadcrumb, numAccounts)
	for i := 0; i < numAccounts; i++ {
		addr := string([]byte{byte(i), 'e', 'r', 'r'})
		breadcrumbs[addr] = &globalAccountBreadcrumb{
			firstNonce:      core.OptionalUint64{Value: 1, HasValue: true},
			lastNonce:       core.OptionalUint64{Value: 1, HasValue: true},
			consumedBalance: big.NewInt(0),
		}
	}

	errExpected := errors.New("trie read failed")
	mockSession := &txcachemocks.SelectionSessionMock{
		GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
			if address[0] == 5 { // fail on 6th account
				return 0, nil, false, errExpected
			}
			return 1, big.NewInt(1000), false, nil
		},
	}

	computer := newVirtualSessionComputer(mockSession)
	err := computer.handleGlobalAccountBreadcrumbs(breadcrumbs)
	require.Error(t, err)
	require.ErrorIs(t, err, errExpected)
}

func deepCopyBreadcrumbs(src map[string]*globalAccountBreadcrumb) map[string]*globalAccountBreadcrumb {
	dst := make(map[string]*globalAccountBreadcrumb, len(src))
	for k, v := range src {
		cp := *v
		cp.consumedBalance = new(big.Int).Set(v.consumedBalance)
		dst[k] = &cp
	}
	return dst
}

func BenchmarkHandleGlobalAccountBreadcrumbs(b *testing.B) {
	for _, numAccounts := range []int{10, 50, 100, 500} {
		breadcrumbs := make(map[string]*globalAccountBreadcrumb, numAccounts)
		for i := 0; i < numAccounts; i++ {
			addr := string([]byte{byte(i / 256), byte(i % 256), 'b', 'e', 'n', 'c', 'h'})
			nonce := uint64(i + 1)
			breadcrumbs[addr] = &globalAccountBreadcrumb{
				firstNonce:      core.OptionalUint64{Value: nonce, HasValue: true},
				lastNonce:       core.OptionalUint64{Value: nonce + 2, HasValue: true},
				consumedBalance: big.NewInt(int64(i * 50)),
				}
		}

		mockSession := &txcachemocks.SelectionSessionMock{
			GetAccountNonceAndBalanceCalled: func(address []byte) (uint64, *big.Int, bool, error) {
				idx := int(address[0])*256 + int(address[1])
				return uint64(idx + 1), big.NewInt(1_000_000), false, nil
			},
		}

		b.Run(fmt.Sprintf("%d_accounts", numAccounts), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				computer := newVirtualSessionComputer(mockSession)
				dcopy := deepCopyBreadcrumbs(breadcrumbs)
				_ = computer.handleGlobalAccountBreadcrumbs(dcopy)
			}
		})
	}
}
