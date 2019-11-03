package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	txproc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
)

func createMetaTxProcessor() process.TransactionProcessor {
	txProc, _ := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	return txProc
}

//------- NewMetaTxProcessor

func TestNewMetaTxProcessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		nil,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilAddressConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	assert.Equal(t, process.ErrNilAddressConverter, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilShardCoordinatorMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		nil,
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilSCProcessorShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.TxTypeHandlerMock{},
	)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilTxTypeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

//------- ProcessTransaction

func TestMetaTxProcessor_ProcessTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	execTx := createMetaTxProcessor()

	err := execTx.ProcessTransaction(nil, 4)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestMetaTxProcessor_ProcessTransactionErrAddressConvShouldErr(t *testing.T) {
	t.Parallel()

	addressConv := &mock.AddressConverterMock{}

	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		addressConv,
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	addressConv.Fail = true

	err := execTx.ProcessTransaction(&transaction.Transaction{}, 4)
	assert.NotNil(t, err)
}

func TestMetaTxProcessor_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	t.Parallel()

	accounts := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewMetaTxProcessor(
		accounts,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.NotNil(t, err)
}

func TestMetaTxProcessor_ProcessCheckNotPassShouldErr(t *testing.T) {
	t.Parallel()

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	acntSrc, err := state.NewAccount(mock.NewAddressMock(tx.SndAddr), &mock.AccountTrackerStub{})
	assert.Nil(t, err)
	acntDst, err := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), &mock.AccountTrackerStub{})
	assert.Nil(t, err)

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewMetaTxProcessor(
		accounts,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	err = execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestMetaTxProcessor_ProcessMoveBalancesShouldFail(t *testing.T) {
	t.Parallel()

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

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	acntSrc, err := state.NewAccount(mock.NewAddressMock(tx.SndAddr), tracker)
	assert.Nil(t, err)
	acntDst, err := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), tracker)
	assert.Nil(t, err)

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewMetaTxProcessor(
		accounts,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
	)

	err = execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldWork(t *testing.T) {
	t.Parallel()

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

	addrConverter := &mock.AddressConverterMock{}

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc, err := state.NewAccount(mock.NewAddressMock(tx.SndAddr), tracker)
	assert.Nil(t, err)

	acntDst, err := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), tracker)
	assert.Nil(t, err)

	acntSrc.Balance = big.NewInt(46)
	acntDst.SetCode([]byte{65})

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	scProcessorMock := &mock.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler, round uint64) error {
		wasCalled = true
		return nil
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		accounts,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, e error) {
				return process.SCInvoking, nil
			},
		},
	)

	err = execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, journalizeCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldReturnErrWhenExecutionFails(t *testing.T) {
	t.Parallel()

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

	addrConverter := &mock.AddressConverterMock{}

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Value = big.NewInt(45)

	acntSrc, err := state.NewAccount(mock.NewAddressMock(tx.SndAddr), tracker)
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)
	acntDst, err := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), tracker)
	assert.Nil(t, err)
	acntDst.SetCode([]byte{65})

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	scProcessorMock := &mock.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler, round uint64) error {
		wasCalled = true
		return process.ErrNoVM
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		accounts,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, e error) {
			return process.SCInvoking, nil
		}},
	)

	err = execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, process.ErrNoVM, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, journalizeCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldNotBeCalledWhenAdrDstIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

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

	addrConverter := &mock.AddressConverterMock{}

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Value = big.NewInt(45)

	shardCoordinator.ComputeIdCalled = func(container state.AddressContainer) uint32 {
		if bytes.Equal(container.Bytes(), tx.RcvAddr) {
			return 1
		}

		return 0
	}

	acntSrc, err := state.NewAccount(mock.NewAddressMock(tx.SndAddr), tracker)
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)
	acntDst, err := state.NewAccount(mock.NewAddressMock(tx.RcvAddr), tracker)
	assert.Nil(t, err)
	acntDst.SetCode([]byte{65})

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	scProcessorMock := &mock.SCProcessorMock{}
	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler, round uint64) error {
		wasCalled = true
		return process.ErrNoVM
	}

	computeType, _ := coordinator.NewTxTypeHandler(
		&mock.AddressConverterMock{},
		shardCoordinator,
		accounts)

	execTx, _ := txproc.NewMetaTxProcessor(
		accounts,
		&mock.AddressConverterMock{},
		shardCoordinator,
		scProcessorMock,
		computeType,
	)

	err = execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, process.ErrWrongTransaction, err)
	assert.False(t, wasCalled)
}
