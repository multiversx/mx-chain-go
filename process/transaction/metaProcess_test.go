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
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func createMetaTxProcessor() process.TransactionProcessor {
	txProc, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	return txProc
}

//------- NewMetaTxProcessor

func TestNewMetaTxProcessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		nil,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilPubkeyConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilShardCoordinatorMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		nil,
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilSCProcessorShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilTxTypeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		nil,
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		nil,
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		nil,
	)

	assert.Equal(t, process.ErrNilEpochNotifier, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

//------- ProcessTransaction

func TestMetaTxProcessor_ProcessTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	execTx := createMetaTxProcessor()

	_, err := execTx.ProcessTransaction(nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestMetaTxProcessor_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		adb,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	_, err := execTx.ProcessTransaction(&tx)
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		adb,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestMetaTxProcessor_ProcessMoveBalancesShouldCallProcessIfError(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	called := false
	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		adb,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{
			ProcessIfErrorCalled: func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int) error {
				called = true
				return nil
			},
		},
		&mock.TxTypeHandlerMock{},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, nil, err)
	assert.True(t, called)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubkeyConverter().Len())
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)

	acntDst, err := state.NewUserAccount(tx.RcvAddr)
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
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return 0, nil
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		adb,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
				return process.SCInvoking
			},
		},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldReturnErrWhenExecutionFails(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubkeyConverter().Len())
	tx.Value = big.NewInt(45)

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &mock.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return 0, process.ErrNoVM
	}

	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		adb,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
			return process.SCInvoking
		}},
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrNoVM, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestMetaTxProcessor_ProcessTransactionScTxShouldNotBeCalledWhenAdrDstIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubkeyConverter().Len())
	tx.Value = big.NewInt(45)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, tx.RcvAddr) {
			return 1
		}

		return 0
	}

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &mock.SCProcessorMock{}
	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return 0, process.ErrNoVM
	}
	calledIfError := false
	scProcessorMock.ProcessIfErrorCalled = func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int) error {
		calledIfError = true
		return nil
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  createMockPubkeyConverter(),
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: make(map[string]struct{}),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	computeType, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	execTx, _ := txproc.NewMetaTxProcessor(
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		adb,
		createMockPubkeyConverter(),
		shardCoordinator,
		scProcessorMock,
		computeType,
		createFreeTxFeeHandler(),
		0,
		0,
		&mock.EpochNotifierStub{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, nil, err)
	assert.False(t, wasCalled)
	assert.True(t, calledIfError)
}
