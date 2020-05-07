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

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func createAccountStub(sndAddr, rcvAddr []byte,
	acntSrc, acntDst state.UserAccountHandler,
) *mock.AccountsStub {
	adb := mock.AccountsStub{}

	adb.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
		if bytes.Equal(address, sndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(address, rcvAddr) {
			return acntDst, nil
		}

		return nil, errors.New("failure")
	}

	return &adb
}

func prepareAccountsAndBalancesMap() (*mock.AccountsStub, map[string]*big.Int, state.UserAccountHandler, state.UserAccountHandler) {
	adr1 := []byte("accnt1")
	adr2 := []byte("accnt2")

	accnt1, _ := state.NewUserAccount(adr1)
	accnt2, _ := state.NewUserAccount(adr2)

	adb := createAccountStub(adr1, adr2, accnt1, accnt2)
	adb.JournalLenCalled = func() int {
		return 0
	}
	adb.CommitCalled = func() (i []byte, e error) {
		return rootHash, nil
	}
	adb.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}

	m := make(map[string]*big.Int)
	m[string(adr1)] = val1
	m[string(adr2)] = val2

	return adb, m, accnt1, accnt2
}

//------- CreateGenesisBlockFromInitialBalances

func TestCreateGenesisBlockFromInitialBalances_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		nil,
		mock.NewOneShardCoordinatorMock(),
		createMockPubkeyConverter(),
		make(map[string]*big.Int),
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestCreateGenesisBlockFromInitialBalances_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		&mock.AccountsStub{},
		nil,
		createMockPubkeyConverter(),
		make(map[string]*big.Int),
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestCreateGenesisBlockFromInitialBalances_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		make(map[string]*big.Int),
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestCreateGenesisBlockFromInitialBalances_NilBalanceMapShouldErr(t *testing.T) {
	t.Parallel()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		createMockPubkeyConverter(),
		nil,
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestCreateGenesisBlockFromInitialBalances_AccountStateDirtyShouldErr(t *testing.T) {
	t.Parallel()

	adb := &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 1
		},
	}

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		adb,
		mock.NewOneShardCoordinatorMock(),
		createMockPubkeyConverter(),
		make(map[string]*big.Int),
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestCreateGenesisBlockFromInitialBalances_TrieCommitFailsShouldRevert(t *testing.T) {
	t.Parallel()

	revertCalled := false
	errCommit := errors.New("should err")

	adb, balances, _, _ := prepareAccountsAndBalancesMap()
	adb.CommitCalled = func() (i []byte, e error) {
		return nil, errCommit
	}
	adb.RevertToSnapshotCalled = func(snapshot int) error {
		revertCalled = true
		return nil
	}
	adb.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, err error) {
		return state.NewUserAccount(address)
	}

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		adb,
		mock.NewOneShardCoordinatorMock(),
		createMockPubkeyConverter(),
		balances,
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, errCommit, err)
	assert.True(t, revertCalled)
}

func TestCreateGenesisBlockFromInitialBalances_AccountsFailShouldErr(t *testing.T) {
	t.Parallel()

	adb, balances, _, _ := prepareAccountsAndBalancesMap()
	errAccounts := errors.New("accounts error")

	adb.LoadAccountCalled =
		func(address []byte) (wrapper state.AccountHandler, e error) {
			return nil, errAccounts
		}

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		adb,
		mock.NewOneShardCoordinatorMock(),
		createMockPubkeyConverter(),
		balances,
		0,
	)

	assert.Nil(t, header)
	assert.Equal(t, errAccounts, err)
}

func TestTxProcessor_SetBalancesToTrieOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adb, balances, accnt1, accnt2 := prepareAccountsAndBalancesMap()

	header, err := genesis.CreateShardGenesisBlockFromInitialBalances(
		adb,
		mock.NewOneShardCoordinatorMock(),
		createMockPubkeyConverter(),
		balances,
		0,
	)

	assert.Equal(t,
		&dataBlock.Header{
			Nonce:           0,
			ShardID:         mock.NewOneShardCoordinatorMock().SelfId(),
			BlockBodyType:   dataBlock.StateBlock,
			PubKeysBitmap:   []byte{1},
			Signature:       rootHash,
			RootHash:        rootHash,
			PrevRandSeed:    rootHash,
			RandSeed:        rootHash,
			AccumulatedFees: big.NewInt(0),
			DeveloperFees:   big.NewInt(0),
		},
		header,
	)
	assert.Nil(t, err)
	assert.Equal(t, val1, accnt1.GetBalance())
	assert.Equal(t, val2, accnt2.GetBalance())
}
