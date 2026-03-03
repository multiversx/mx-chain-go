package txcache

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
)

func Test_fromBreadcrumbToVirtualRecord(t *testing.T) {
	t.Parallel()

	t.Run("virtual record of sender breadcrumb", func(t *testing.T) {
		t.Parallel()

		address := "bob"
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
		err := computer.fromGlobalBreadcrumbToVirtualRecord(address, sessionNonce, accountBalance, &breadcrumbBob)
		require.Nil(t, err)

		actualVirtualRecord, ok := computer.virtualAccountsByAddress[address]
		require.True(t, ok)
		require.Equal(t, expectedVirtualRecord, actualVirtualRecord)
	})

	t.Run("virtual record of bob relayer", func(t *testing.T) {
		t.Parallel()

		address := "bob"
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
		err := computer.fromGlobalBreadcrumbToVirtualRecord(address, sessionNonce, accountBalance, &breadcrumbBob)
		require.Nil(t, err)

		actualVirtualRecord, ok := computer.virtualAccountsByAddress[address]
		require.True(t, ok)
		require.Equal(t, expectedVirtualRecord, actualVirtualRecord)
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

func Test_createBlockedVirtualRecord(t *testing.T) {
	t.Parallel()

	t.Run("should create blocked record with unset nonce", func(t *testing.T) {
		t.Parallel()

		address := "alice"
		accountBalance := big.NewInt(100)
		globalBreadcrumb := &globalAccountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    52,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    55,
				HasValue: true,
			},
			consumedBalance: big.NewInt(30),
		}

		computer := newVirtualSessionComputer(nil)
		err := computer.createBlockedVirtualRecord(address, accountBalance, globalBreadcrumb)
		require.Nil(t, err)

		record, ok := computer.virtualAccountsByAddress[address]
		require.True(t, ok)

		// Nonce should NOT be set (blocked)
		require.False(t, record.initialNonce.HasValue)

		// getInitialNonce should return error
		_, err = record.getInitialNonce()
		require.Equal(t, errNonceNotSet, err)

		// Balance should be preserved
		require.Equal(t, big.NewInt(100), record.getInitialBalance())
		require.Equal(t, big.NewInt(30), record.getConsumedBalance())
	})

	t.Run("should not overwrite existing record", func(t *testing.T) {
		t.Parallel()

		address := "alice"
		accountBalance := big.NewInt(100)
		globalBreadcrumb := &globalAccountBreadcrumb{
			firstNonce:      core.OptionalUint64{Value: 52, HasValue: true},
			lastNonce:       core.OptionalUint64{Value: 55, HasValue: true},
			consumedBalance: big.NewInt(30),
		}

		computer := newVirtualSessionComputer(nil)

		// Add an existing record first
		existingRecord, err := newVirtualAccountRecord(core.OptionalUint64{Value: 10, HasValue: true}, big.NewInt(50))
		require.Nil(t, err)
		computer.virtualAccountsByAddress[address] = existingRecord

		// Try to create a blocked record - should not overwrite
		err = computer.createBlockedVirtualRecord(address, accountBalance, globalBreadcrumb)
		require.Nil(t, err)

		record := computer.virtualAccountsByAddress[address]
		require.True(t, record.initialNonce.HasValue)
		require.Equal(t, uint64(10), record.initialNonce.Value)
	})
}
