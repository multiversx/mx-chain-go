package rewardTransaction_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/stretchr/testify/assert"
)

func TestNewRewardTxProcessor_NilAccountsDbShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		nil,
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewRewardTxProcessor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewRewardTxProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		nil,
		&mock.IntermediateTransactionHandlerMock{})

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxProcessor_NilRewardTxForwarderShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilIntermediateTransactionHandler, err)
}

func TestNewRewardTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	assert.NotNil(t, rtp)
	assert.Nil(t, err)
	assert.False(t, rtp.IsInterfaceNil())
}

func TestRewardTxProcessor_ProcessRewardTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	err := rtp.ProcessRewardTransaction(nil)
	assert.Equal(t, process.ErrNilRewardTransaction, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionNilTxValueShouldErr(t *testing.T) {
	t.Parallel()

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	rwdTx := rewardTx.RewardTx{Value: nil}
	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, process.ErrNilValueFromRewardTransaction, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionCannotCreateAddressShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot create address")
	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (state.AddressContainer, error) {
				return nil, expectedErr
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("rcvr"),
		ShardId: 0,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, expectedErr, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionAddressNotInNodesShardShouldNotExecute(t *testing.T) {
	t.Parallel()

	getAccountWithJournalWasCalled := false
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return uint32(5)
	}
	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{
			GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
				getAccountWithJournalWasCalled = true
				return nil, nil
			},
		},
		&mock.AddressConverterMock{},
		shardCoord,
		&mock.IntermediateTransactionHandlerMock{})

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("rcvr"),
		ShardId: 0,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Nil(t, err)
	// account should not be requested as the address is not in node's shard
	assert.False(t, getAccountWithJournalWasCalled)
}

func TestRewardTxProcessor_ProcessRewardTransactionCannotGetAccountShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot get account")
	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{
			GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
				return nil, expectedErr
			},
		},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("rcvr"),
		ShardId: 0,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, expectedErr, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionCannotAddIntermediateTxsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot add intermediate transactions")
	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{
			AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
				return expectedErr
			},
		})

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("rcvr"),
		ShardId: 0,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, expectedErr, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionWrongTypeAssertionAccountHolderShouldErr(t *testing.T) {
	t.Parallel()

	accountsDb := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
			return mock.NewAccountWrapMock(addressContainer, &mock.AccountTrackerStub{}), nil
		},
	}

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		accountsDb,
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("rcvr"),
		ShardId: 0,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionShouldWork(t *testing.T) {
	t.Parallel()

	journalizeWasCalled := false
	saveAccountWasCalled := false

	accountsDb := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (state.AccountHandler, error) {
			ats := &mock.AccountTrackerStub{
				JournalizeCalled: func(entry state.JournalEntry) {
					journalizeWasCalled = true
				},
				SaveAccountCalled: func(accountHandler state.AccountHandler) error {
					saveAccountWasCalled = true
					return nil
				},
			}
			return state.NewAccount(addressContainer, ats)
		},
	}

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		accountsDb,
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.IntermediateTransactionHandlerMock{})

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("rcvr"),
		ShardId: 0,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Nil(t, err)
	assert.True(t, journalizeWasCalled)
	assert.True(t, saveAccountWasCalled)
}
