package coordinator

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewTxTypeHandler_NilAddrConv(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTxTypeHandler_NilShardCoord(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		&mock.AddressConverterMock{},
		nil,
		&mock.AccountsStub{},
	)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxTypeHandler_NilAccounts(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
	)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewTxTypeHandler_ValsOk(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)
}

func generateRandomByteSlice(size int) []byte {
	buff := make([]byte, size)
	_, _ = rand.Reader.Read(buff)

	return buff
}

func createAccounts(tx *transaction.Transaction) (state.AccountHandler, state.AccountHandler) {
	journalizeCalled := 0
	saveAccountCalled := 0
	tracker := &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {
			journalizeCalled++
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
	}

	acntSrc, _ := state.NewAccount(mock.NewAddressMock(tx.SndAddr), tracker)
	acntSrc.Balance = acntSrc.Balance.Add(acntSrc.Balance, tx.GetValue())
	totalFee := big.NewInt(0)
	totalFee = totalFee.Mul(big.NewInt(int64(tx.GasLimit)), big.NewInt(int64(tx.GasPrice)))
	acntSrc.Balance = acntSrc.Balance.Add(acntSrc.Balance, totalFee)

	acntDst, _ := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), tracker)

	return acntSrc, acntDst
}

func TestTxTypeHandler_ComputeTransactionTypeNil(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	_, err = tth.ComputeTransactionType(nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestTxTypeHandler_ComputeTransactionTypeNilTx(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = "45"

	tx = nil
	_, err = tth.ComputeTransactionType(tx)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestTxTypeHandler_ComputeTransactionTypeErrWrongTransaction(t *testing.T) {
	t.Parallel()

	tth, err := NewTxTypeHandler(
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Value = "45"

	_, err = tth.ComputeTransactionType(tx)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestTxTypeHandler_ComputeTransactionTypeScDeployment(t *testing.T) {
	t.Parallel()

	addressConverter := &mock.AddressConverterMock{}
	tth, err := NewTxTypeHandler(
		addressConverter,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, addressConverter.AddressLen())
	tx.Data = "data"
	tx.Value = "45"

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.SCDeployment, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeScInvoking(t *testing.T) {
	t.Parallel()

	addrConverter := &mock.AddressConverterMock{}
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Data = "data"
	tx.Value = "45"

	_, acntDst := createAccounts(tx)
	acntDst.SetCode([]byte("code"))

	addressConverter := &mock.AddressConverterMock{}
	tth, err := NewTxTypeHandler(
		addressConverter,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return acntDst, nil
		}},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.SCInvoking, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeMoveBalance(t *testing.T) {
	t.Parallel()

	addrConverter := &mock.AddressConverterMock{}
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Data = "data"
	tx.Value = "45"

	_, acntDst := createAccounts(tx)

	addressConverter := &mock.AddressConverterMock{}
	tth, err := NewTxTypeHandler(
		addressConverter,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return acntDst, nil
		}},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.MoveBalance, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeRewardTx(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}
	tth, err := NewTxTypeHandler(
		addrConv,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.AccountsStub{},
	)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &rewardTx.RewardTx{RcvAddr: []byte("leader")}
	txType, err := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.ErrWrongTransaction, err)
	assert.Equal(t, process.InvalidTransaction, txType)

	tx = &rewardTx.RewardTx{RcvAddr: generateRandomByteSlice(addrConv.AddressLen())}
	txType, err = tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.RewardTx, txType)
}
