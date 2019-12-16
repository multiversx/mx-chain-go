package genesis_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

var val1 = big.NewInt(10)
var val2 = big.NewInt(20)
var rootHash = []byte("root hash")
var validatorStatsRootHash = []byte("validator stats root hash")

func createAccountStub(sndAddr, rcvAddr []byte,
	acntSrc, acntDst *state.Account,
) *mock.AccountsStub {
	accounts := mock.AccountsStub{}

	accounts.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
		if bytes.Equal(addressContainer.Bytes(), sndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), rcvAddr) {
			return acntDst, nil
		}

		return nil, errors.New("failure")
	}

	return &accounts
}

func prepareAccountsAndBalancesMap() (*mock.AccountsStub, map[string]*big.Int, *state.Account, *state.Account) {
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}

	adr1 := []byte("accnt1")
	adr2 := []byte("accnt2")

	accnt1, _ := state.NewAccount(mock.NewAddressMock(adr1), tracker)
	accnt2, _ := state.NewAccount(mock.NewAddressMock(adr2), tracker)

	accounts := createAccountStub(adr1, adr2, accnt1, accnt2)
	accounts.JournalLenCalled = func() int {
		return 0
	}
	accounts.CommitCalled = func() (i []byte, e error) {
		return rootHash, nil
	}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}

	m := make(map[string]*big.Int)
	m[string(adr1)] = val1
	m[string(adr2)] = val2

	return accounts, m, accnt1, accnt2
}

//------- CreateGenesisBlockFromInitialBalances

func TestCreateGenesisBlockFromInitialBalances_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.AddressConverterMock{},
		make(map[string]*big.Int),
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestCreateGenesisBlockFromInitialBalances_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		&mock.AccountsStub{},
		nil,
		&mock.AddressConverterMock{},
		make(map[string]*big.Int),
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCreateGenesisBlockFromInitialBalances_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		make(map[string]*big.Int),
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestCreateGenesisBlockFromInitialBalances_NilBalanceMapShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.AddressConverterMock{},
		nil,
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestCreateGenesisBlockFromInitialBalances_AccountStateDirtyShouldErr(t *testing.T) {
	t.Parallel()

	accounts := &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 1
		},
	}

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.AddressConverterMock{},
		make(map[string]*big.Int),
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestCreateGenesisBlockFromInitialBalances_TrieCommitFailsShouldRevert(t *testing.T) {
	t.Parallel()

	revertCalled := false
	errCommit := errors.New("should err")

	accounts, balances, _, _ := prepareAccountsAndBalancesMap()
	accounts.CommitCalled = func() (i []byte, e error) {
		return nil, errCommit
	}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		revertCalled = true
		return nil
	}

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.AddressConverterMock{},
		balances,
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, errCommit, err)
	assert.True(t, revertCalled)
}

func TestCreateGenesisBlockFromInitialBalances_AccountsFailShouldErr(t *testing.T) {
	t.Parallel()

	accounts, balances, _, _ := prepareAccountsAndBalancesMap()
	errAccounts := errors.New("accounts error")

	accounts.GetAccountWithJournalCalled =
		func(addressContainer state.AddressContainer) (wrapper state.AccountHandler, e error) {
			return nil, errAccounts
		}

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.AddressConverterMock{},
		balances,
		0,
		validatorStatsRootHash,
	)

	assert.Nil(t, header)
	assert.Equal(t, errAccounts, err)
}

func TestTxProcessor_SetBalancesToTrieOkValsShouldWork(t *testing.T) {
	t.Parallel()

	accounts, balances, accnt1, accnt2 := prepareAccountsAndBalancesMap()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.AddressConverterMock{},
		balances,
		0,
		validatorStatsRootHash,
	)

	assert.Equal(t,
		&dataBlock.Header{
			Nonce:                  0,
			ShardId:                mock.NewOneShardCoordinatorMock().SelfId(),
			BlockBodyType:          dataBlock.StateBlock,
			Signature:              rootHash,
			RootHash:               rootHash,
			PrevRandSeed:           rootHash,
			RandSeed:               rootHash,
			ValidatorStatsRootHash: validatorStatsRootHash,
		},
		header,
	)
	assert.Nil(t, err)
	assert.Equal(t, val1, accnt1.Balance)
	assert.Equal(t, val2, accnt2.Balance)
}
