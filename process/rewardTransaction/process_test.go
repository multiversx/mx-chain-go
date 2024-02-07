package rewardTransaction_test

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/rewardTransaction"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/state/trackableDataTrie"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

func TestNewRewardTxProcessor_NilAccountsDbShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		nil,
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewRewardTxProcessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&stateMock.AccountsStub{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewRewardTxProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&stateMock.AccountsStub{},
		createMockPubkeyConverter(),
		nil,
	)

	assert.Nil(t, rtp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	rtp, err := rewardTransaction.NewRewardTxProcessor(
		&stateMock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	assert.NotNil(t, rtp)
	assert.Nil(t, err)
	assert.False(t, rtp.IsInterfaceNil())
}

func TestRewardTxProcessor_ProcessRewardTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&stateMock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	err := rtp.ProcessRewardTransaction(nil)
	assert.Equal(t, process.ErrNilRewardTransaction, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionNilTxValueShouldErr(t *testing.T) {
	t.Parallel()

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&stateMock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	rwdTx := rewardTx.RewardTx{Value: nil}
	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, process.ErrNilValueFromRewardTransaction, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionAddressNotInNodesShardShouldNotExecute(t *testing.T) {
	t.Parallel()

	getAccountWithJournalWasCalled := false
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.ComputeIdCalled = func(address []byte) uint32 {
		return uint32(5)
	}
	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		&stateMock.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				getAccountWithJournalWasCalled = true
				return nil, nil
			},
		},
		createMockPubkeyConverter(),
		shardCoord,
	)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte("rcvr"),
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
		&stateMock.AccountsStub{
			LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		},
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte("rcvr"),
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, expectedErr, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionWrongTypeAssertionAccountHolderShouldErr(t *testing.T) {
	t.Parallel()

	accountsDb := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.PeerAccountHandlerMock{}, nil
		},
	}

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		accountsDb,
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte("rcvr"),
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountWasCalled := false

	accountsDb := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return accounts.NewUserAccount(address, &trie.DataTrieTrackerStub{}, &trie.TrieLeafParserStub{})
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			saveAccountWasCalled = true
			return nil
		},
	}

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		accountsDb,
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte("rcvr"),
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Nil(t, err)
	assert.True(t, saveAccountWasCalled)
}

func TestRewardTxProcessor_ProcessRewardTransactionMissingTrieNode(t *testing.T) {
	t.Parallel()

	missingNodeErr := fmt.Errorf(core.GetNodeFromDBErrorString)
	accountsDb := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			acc := stateMock.NewAccountWrapMock(address)
			acc.SetDataTrie(&trie.TrieStub{
				GetCalled: func(key []byte) ([]byte, uint32, error) {
					return nil, 0, missingNodeErr
				},
			},
			)

			return acc, nil
		},
	}

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		accountsDb,
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   big.NewInt(100),
		RcvAddr: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6},
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Equal(t, missingNodeErr, err)
}

func TestRewardTxProcessor_ProcessRewardTransactionToASmartContractShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountWasCalled := false

	address := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6}

	dtt, _ := trackableDataTrie.NewTrackableDataTrie(address, &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{}, enableEpochsHandlerMock.NewEnableEpochsHandlerStub())
	userAccount, _ := accounts.NewUserAccount(address, dtt, &trie.TrieLeafParserStub{})
	accountsDb := &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return userAccount, nil
		},
		SaveAccountCalled: func(accountHandler vmcommon.AccountHandler) error {
			saveAccountWasCalled = true
			return nil
		},
	}

	rtp, _ := rewardTransaction.NewRewardTxProcessor(
		accountsDb,
		createMockPubkeyConverter(),
		mock.NewMultiShardsCoordinatorMock(3),
	)

	rwdTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   big.NewInt(100),
		RcvAddr: address,
	}

	err := rtp.ProcessRewardTransaction(&rwdTx)
	assert.Nil(t, err)
	assert.True(t, saveAccountWasCalled)
	val, _, err := userAccount.RetrieveValue([]byte(core.ProtectedKeyPrefix + rewardTransaction.RewardKey))
	assert.Nil(t, err)
	assert.True(t, rwdTx.Value.Cmp(big.NewInt(0).SetBytes(val)) == 0)

	err = rtp.ProcessRewardTransaction(&rwdTx)
	assert.Nil(t, err)
	assert.True(t, saveAccountWasCalled)
	val, _, err = userAccount.RetrieveValue([]byte(core.ProtectedKeyPrefix + rewardTransaction.RewardKey))
	assert.Nil(t, err)
	rwdTx.Value.Add(rwdTx.Value, rwdTx.Value)
	assert.True(t, rwdTx.Value.Cmp(big.NewInt(0).SetBytes(val)) == 0)
}
