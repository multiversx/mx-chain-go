package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	txproc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func createMockNewMetaTxArgs() txproc.ArgsNewMetaTxProcessor {
	args := txproc.ArgsNewMetaTxProcessor{
		Hasher:           &mock.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		Accounts:         &testscommon.AccountsStub{},
		PubkeyConv:       createMockPubkeyConverter(),
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		ScProcessor:      &testscommon.SCProcessorMock{},
		TxTypeHandler:    &testscommon.TxTypeHandlerMock{},
		EconomicsFee:     createFreeTxFeeHandler(),
		ESDTEnableEpoch:  0,
		EpochNotifier:    &mock.EpochNotifierStub{},
	}
	return args
}

// ------- NewMetaTxProcessor

func TestNewMetaTxProcessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.Accounts = nil
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilPubkeyConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.PubkeyConv = nil
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilShardCoordinatorMockShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.ShardCoordinator = nil
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilSCProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.ScProcessor = nil
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilTxTypeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.TxTypeHandler = nil
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.EconomicsFee = nil
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewMetaTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	txProc, err := txproc.NewMetaTxProcessor(args)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

// ------- ProcessTransaction

func TestMetaTxProcessor_ProcessTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	txProc, _ := txproc.NewMetaTxProcessor(args)

	_, err := txProc.ProcessTransaction(nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestMetaTxProcessor_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)
	args := createMockNewMetaTxArgs()
	args.Accounts = adb
	txProc, _ := txproc.NewMetaTxProcessor(args)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	_, err := txProc.ProcessTransaction(&tx)
	assert.NotNil(t, err)
}

func TestMetaTxProcessor_ProcessCheckNotPassShouldErr(t *testing.T) {
	t.Parallel()

	// these values will trigger ErrHigherNonceInTransaction
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

	args := createMockNewMetaTxArgs()
	args.Accounts = adb
	txProc, _ := txproc.NewMetaTxProcessor(args)

	_, err = txProc.ProcessTransaction(&tx)
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
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	called := false
	args := createMockNewMetaTxArgs()
	args.Accounts = adb
	args.ScProcessor = &testscommon.SCProcessorMock{
		ProcessIfErrorCalled: func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
			called = true
			return nil
		},
	}
	txProc, _ := txproc.NewMetaTxProcessor(args)

	_, err = txProc.ProcessTransaction(&tx)
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
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}
	scProcessorMock := &testscommon.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return 0, nil
	}

	args := createMockNewMetaTxArgs()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	txProc, _ := txproc.NewMetaTxProcessor(args)

	_, err = txProc.ProcessTransaction(&tx)
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
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &testscommon.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return 0, process.ErrNoVM
	}

	args := createMockNewMetaTxArgs()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	txProc, _ := txproc.NewMetaTxProcessor(args)

	_, err = txProc.ProcessTransaction(&tx)
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
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &testscommon.SCProcessorMock{}
	wasCalled := false
	scProcessorMock.ExecuteSmartContractTransactionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return 0, process.ErrNoVM
	}
	calledIfError := false
	scProcessorMock.ProcessIfErrorCalled = func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
		calledIfError = true
		return nil
	}

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    createMockPubkeyConverter(),
		ShardCoordinator:   shardCoordinator,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		EpochNotifier:      &mock.EpochNotifierStub{},
		ESDTTransferParser: esdtTransferParser,
	}
	computeType, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	args := createMockNewMetaTxArgs()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = computeType
	args.ShardCoordinator = shardCoordinator
	txProc, _ := txproc.NewMetaTxProcessor(args)

	_, err = txProc.ProcessTransaction(&tx)
	assert.Equal(t, nil, err)
	assert.False(t, wasCalled)
	assert.True(t, calledIfError)
}
