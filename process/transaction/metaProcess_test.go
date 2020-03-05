package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state/accounts"

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
		createFreeTxFeeHandler(),
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
		createFreeTxFeeHandler(),
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
		createFreeTxFeeHandler(),
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
		createFreeTxFeeHandler(),
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
		createFreeTxFeeHandler(),
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
		createFreeTxFeeHandler(),
	)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.AccountsStub{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
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
		createFreeTxFeeHandler(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

//------- ProcessTransaction

func TestMetaTxProcessor_ProcessTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	execTx := createMetaTxProcessor()

	err := execTx.ProcessTransaction(nil)
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
		createFreeTxFeeHandler(),
	)

	addressConv.Fail = true

	err := execTx.ProcessTransaction(&transaction.Transaction{})
	assert.NotNil(t, err)
}

func TestMetaTxProcessor_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewMetaTxProcessor(
		adb,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
	)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	err := execTx.ProcessTransaction(&tx)
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

	acntSrc, err := accounts.NewUserAccount(mock.NewAddressMock(tx.SndAddr))
	assert.Nil(t, err)
	acntDst, err := accounts.NewUserAccount(mock.NewAddressMock(tx.RcvAddr))
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewMetaTxProcessor(
		adb,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
	)

	err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestMetaTxProcessor_ProcessMoveBalancesShouldFail(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	acntSrc, err := accounts.NewUserAccount(mock.NewAddressMock(tx.SndAddr))
	assert.Nil(t, err)
	acntDst, err := accounts.NewUserAccount(mock.NewAddressMock(tx.RcvAddr))
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		adb,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
	)

	err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	addrConverter := &mock.AddressConverterMock{}

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc, err := accounts.NewUserAccount(mock.NewAddressMock(tx.SndAddr))
	assert.Nil(t, err)

	acntDst, err := accounts.NewUserAccount(mock.NewAddressMock(tx.RcvAddr))
	assert.Nil(t, err)

	acntSrc.Balance = big.NewInt(46)
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}
	scProcessorMock := &mock.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) error {
		wasCalled = true
		return nil
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		adb,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, e error) {
				return process.SCInvoking, nil
			},
		},
		createFreeTxFeeHandler(),
	)

	err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldReturnErrWhenExecutionFails(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	addrConverter := &mock.AddressConverterMock{}

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(addrConverter.AddressLen())
	tx.Value = big.NewInt(45)

	acntSrc, err := accounts.NewUserAccount(mock.NewAddressMock(tx.SndAddr))
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)
	acntDst, err := accounts.NewUserAccount(mock.NewAddressMock(tx.RcvAddr))
	assert.Nil(t, err)
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &mock.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) error {
		wasCalled = true
		return process.ErrNoVM
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		adb,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, e error) {
			return process.SCInvoking, nil
		}},
		createFreeTxFeeHandler(),
	)

	err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrNoVM, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldNotBeCalledWhenAdrDstIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	saveAccountCalled := 0
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

	acntSrc, err := accounts.NewUserAccount(mock.NewAddressMock(tx.SndAddr))
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)
	acntDst, err := accounts.NewUserAccount(mock.NewAddressMock(tx.RcvAddr))
	assert.Nil(t, err)
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &mock.SCProcessorMock{}
	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) error {
		wasCalled = true
		return process.ErrNoVM
	}

	computeType, _ := coordinator.NewTxTypeHandler(
		&mock.AddressConverterMock{},
		shardCoordinator,
		adb,
	)

	execTx, _ := txproc.NewMetaTxProcessor(
		adb,
		&mock.AddressConverterMock{},
		shardCoordinator,
		scProcessorMock,
		computeType,
		createFreeTxFeeHandler(),
	)

	err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrWrongTransaction, err)
	assert.False(t, wasCalled)
}
