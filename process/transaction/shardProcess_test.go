package transaction_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	txproc "github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/stretchr/testify/assert"
)

var flagActiveTrueHandler = func(epoch uint32) bool { return true }
var flagActiveFalseHandler = func(epoch uint32) bool { return false }

func generateRandomByteSlice(size int) []byte {
	buff := make([]byte, size)
	_, _ = rand.Reader.Read(buff)

	return buff
}

func feeHandlerMock() *economicsmocks.EconomicsHandlerStub {
	return &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return nil
		},
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(0)
		},
	}
}

func createAccountStub(sndAddr, rcvAddr []byte,
	acntSrc, acntDst state.UserAccountHandler,
) *stateMock.AccountsStub {
	adb := stateMock.AccountsStub{}

	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
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

func createArgsForTxProcessor() txproc.ArgsNewTxProcessor {
	args := txproc.ArgsNewTxProcessor{
		Accounts:         &stateMock.AccountsStub{},
		Hasher:           &hashingMocks.HasherMock{},
		PubkeyConv:       createMockPubKeyConverter(),
		Marshalizer:      &mock.MarshalizerMock{},
		SignMarshalizer:  &mock.MarshalizerMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		ScProcessor:      &testscommon.SCProcessorMock{},
		TxFeeHandler:     &mock.FeeAccumulatorStub{},
		TxTypeHandler:    &testscommon.TxTypeHandlerMock{},
		EconomicsFee:     feeHandlerMock(),
		ReceiptForwarder: &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:       &mock.ArgumentParserMock{},
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsPenalizedTooMuchGasFlagEnabledInEpochCalled: func(epoch uint32) bool { return true },
		},
		GuardianChecker:     &guardianMocks.GuardedAccountHandlerStub{},
		TxVersionChecker:    &testscommon.TxVersionCheckerStub{},
		TxLogsProcessor:     &mock.TxLogsProcessorStub{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
	}
	return args
}

func createTxProcessor() txproc.TxProcessor {
	txProc, _ := txproc.NewTxProcessor(createArgsForTxProcessor())
	return txProc
}

//------- NewTxProcessor

func TestNewTxProcessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.Accounts = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.Hasher = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilPubkeyConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.PubkeyConv = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilMarshalizerMockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.Marshalizer = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilShardCoordinatorMockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.ShardCoordinator = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilSCProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.ScProcessor = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilTxTypeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.TxTypeHandler = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilTxTypeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilEconomicsFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.EconomicsFee = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilReceiptForwarderShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.ReceiptForwarder = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilReceiptHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilBadTxForwarderShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.BadTxForwarder = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilBadTxHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.TxFeeHandler = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilUnsignedTxHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilArgsParserShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.ArgsParser = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilArgumentParser, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilScrForwarderShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.ScrForwarder = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilIntermediateTransactionHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.SignMarshalizer = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.EnableEpochsHandler = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilTxLogsProcessorShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.TxLogsProcessor = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilTxLogsProcessor, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilEnableRoundsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.EnableRoundsHandler = nil
	txProc, err := txproc.NewTxProcessor(args)

	assert.Equal(t, process.ErrNilEnableRoundsHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	txProc, err := txproc.NewTxProcessor(args)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

//------- getAccounts

func TestTxProcessor_GetAccountsShouldErrNilAddressContainer(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)
	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	adr1 := []byte{65}
	adr2 := []byte{67}

	_, _, err := execTx.GetAccounts(nil, adr2)
	assert.Equal(t, process.ErrNilAddressContainer, err)

	_, _, err = execTx.GetAccounts(adr1, nil)
	assert.Equal(t, process.ErrNilAddressContainer, err)
}

func TestTxProcessor_GetAccountsMalfunctionAccountsShouldErr(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	adr1 := []byte{65}
	adr2 := []byte{67}

	_, _, err := execTx.GetAccounts(adr1, adr2)
	assert.NotNil(t, err)
}

func TestTxProcessor_GetAccountsOkValsSrcShouldWork(t *testing.T) {
	t.Parallel()

	adb := stateMock.AccountsStub{}

	adr1 := []byte{65}
	adr2 := []byte{67}

	acnt1 := createUserAcc(adr1)
	acnt2 := createUserAcc(adr2)

	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, adr1) {
			return acnt1, nil
		}

		if bytes.Equal(address, adr2) {
			return nil, errors.New("failure on destination")
		}

		return nil, errors.New("failure")
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	args := createArgsForTxProcessor()
	args.Accounts = &adb
	args.ShardCoordinator = shardCoordinator
	execTx, _ := txproc.NewTxProcessor(args)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, adr2) {
			return 1
		}

		return 0
	}

	a1, a2, err := execTx.GetAccounts(adr1, adr2)

	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.NotEqual(t, acnt2, a2)
	assert.Nil(t, a2)
}

func TestTxProcessor_GetAccountsOkValsDsthouldWork(t *testing.T) {
	t.Parallel()

	adb := stateMock.AccountsStub{}

	adr1 := []byte{65}
	adr2 := []byte{67}

	acnt1 := createUserAcc(adr1)
	acnt2 := createUserAcc(adr2)

	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, adr1) {
			return nil, errors.New("failure on source")
		}

		if bytes.Equal(address, adr2) {
			return acnt2, nil
		}

		return nil, errors.New("failure")
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	args := createArgsForTxProcessor()
	args.Accounts = &adb
	args.ShardCoordinator = shardCoordinator
	execTx, _ := txproc.NewTxProcessor(args)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, adr1) {
			return 1
		}

		return 0
	}

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.NotEqual(t, acnt1, a1)
	assert.Nil(t, a1)
	assert.Equal(t, acnt2, a2)
}

func TestTxProcessor_GetAccountsOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	adr2 := []byte{67}

	acnt1 := createUserAcc(adr1)
	acnt2 := createUserAcc(adr2)

	adb := createAccountStub(adr1, adr2, acnt1, acnt2)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.Equal(t, acnt2, a2)
}

func TestTxProcessor_GetSameAccountShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	adr2 := []byte{65}

	acnt1 := createUserAcc(adr1)
	acnt2 := createUserAcc(adr2)

	adb := createAccountStub(adr1, adr2, acnt1, acnt2)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	a1, a2, err := execTx.GetAccounts(adr1, adr1)
	assert.Nil(t, err)
	assert.True(t, a1 == a2)
}

//------- checkTxValues

func TestTxProcessor_CheckTxValuesHigherNonceShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1 := createUserAcc(adr1)

	execTx := *createTxProcessor()

	acnt1.IncreaseNonce(6)

	err := execTx.CheckTxValues(&transaction.Transaction{Nonce: 7}, acnt1, nil, false)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestTxProcessor_CheckTxValuesLowerNonceShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1 := createUserAcc(adr1)

	execTx := *createTxProcessor()

	acnt1.IncreaseNonce(6)

	err := execTx.CheckTxValues(&transaction.Transaction{Nonce: 5}, acnt1, nil, false)
	assert.Equal(t, process.ErrLowerNonceInTransaction, err)
}

func TestTxProcessor_CheckTxValuesInsufficientFundsShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1 := createUserAcc(adr1)

	execTx := *createTxProcessor()

	_ = acnt1.AddToBalance(big.NewInt(67))

	err := execTx.CheckTxValues(&transaction.Transaction{Value: big.NewInt(68)}, acnt1, nil, false)
	assert.Equal(t, process.ErrInsufficientFunds, err)
}

func TestTxProcessor_CheckTxValuesMismatchedSenderUsernamesShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	senderAcc := createUserAcc(adr1)

	execTx := *createTxProcessor()

	_ = senderAcc.AddToBalance(big.NewInt(67))
	senderAcc.SetUserName([]byte("SRC"))

	tx := &transaction.Transaction{
		Value:       big.NewInt(10),
		SndUserName: []byte("notCorrect"),
	}

	err := execTx.CheckTxValues(tx, senderAcc, nil, false)
	assert.Equal(t, process.ErrUserNameDoesNotMatch, err)
}

func TestTxProcessor_CheckTxValuesMismatchedReceiverUsernamesShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	receiverAcc := createUserAcc(adr1)

	execTx := *createTxProcessor()

	_ = receiverAcc.AddToBalance(big.NewInt(67))
	receiverAcc.SetUserName([]byte("RECV"))

	tx := &transaction.Transaction{
		Value:       big.NewInt(10),
		RcvUserName: []byte("notCorrect"),
	}

	err := execTx.CheckTxValues(tx, nil, receiverAcc, false)
	assert.Equal(t, process.ErrUserNameDoesNotMatchInCrossShardTx, err)
}

func TestTxProcessor_CheckTxValuesCorrectUserNamesShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	senderAcc := createUserAcc(adr1)

	adr2 := []byte{66}
	recvAcc := createUserAcc(adr2)

	execTx := *createTxProcessor()

	_ = senderAcc.AddToBalance(big.NewInt(67))
	senderAcc.SetUserName([]byte("SRC"))
	recvAcc.SetUserName([]byte("RECV"))

	tx := &transaction.Transaction{
		Value:       big.NewInt(10),
		SndUserName: senderAcc.GetUserName(),
		RcvUserName: recvAcc.GetUserName(),
	}

	err := execTx.CheckTxValues(tx, senderAcc, recvAcc, false)
	assert.Nil(t, err)
}

func TestTxProcessor_CheckTxValuesOkValsShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1 := createUserAcc(adr1)

	execTx := *createTxProcessor()

	_ = acnt1.AddToBalance(big.NewInt(67))

	err := execTx.CheckTxValues(&transaction.Transaction{Value: big.NewInt(67)}, acnt1, nil, false)
	assert.Nil(t, err)
}

//------- increaseNonce

func TestTxProcessor_IncreaseNonceOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adrSrc := []byte{65}
	acntSrc := createUserAcc(adrSrc)

	execTx := *createTxProcessor()

	acntSrc.IncreaseNonce(45)

	execTx.IncreaseNonce(acntSrc)
	assert.Equal(t, uint64(46), acntSrc.GetNonce())
}

//------- ProcessTransaction

func TestTxProcessor_ProcessTransactionNilTxShouldErr(t *testing.T) {
	t.Parallel()

	execTx := *createTxProcessor()

	_, err := execTx.ProcessTransaction(nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestTxProcessor_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	_, err := execTx.ProcessTransaction(&tx)
	assert.NotNil(t, err)
}

func TestTxProcessor_ProcessCheckNotPassShouldErr(t *testing.T) {
	t.Parallel()

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestTxProcessor_ProcessCheckShouldPassWhenAdrSrcIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	testProcessCheck(t, 1, big.NewInt(45))
}

func TestTxProcessor_ProcessMoveBalancesShouldPassWhenAdrSrcIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	testProcessCheck(t, 0, big.NewInt(0))
}

func TestTxProcessor_ProcessWithTxFeeHandlerCheckErrorShouldErr(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	args := createArgsForTxProcessor()
	args.Accounts = adb

	expectedError := errors.New("validatity check failed")
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return expectedError
		}}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, expectedError, err)
}

func TestTxProcessor_ProcessWithWrongAssertionShouldErr(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	args := createArgsForTxProcessor()
	args.Accounts = &stateMock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &stateMock.PeerAccountHandlerMock{}, nil
		},
	}

	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTxProcessor_ProcessWithTxFeeHandlerInsufficientFeeShouldErr(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(9))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	args := createArgsForTxProcessor()
	args.Accounts = adb

	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(0).Add(acntSrc.GetBalance(), big.NewInt(1))
		}}

	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.True(t, errors.Is(err, process.ErrInsufficientFee))
}

func TestTxProcessor_ProcessWithInsufficientFundsShouldCreateReceiptErr(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(9))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	args := createArgsForTxProcessor()
	args.Accounts = adb

	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return process.ErrInsufficientFunds
		}}

	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, uint64(1), acntSrc.GetNonce())
}

func TestTxProcessor_ProcessWithUsernameMismatchCreateReceiptErr(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(9))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	args := createArgsForTxProcessor()
	args.Accounts = adb

	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return process.ErrUserNameDoesNotMatchInCrossShardTx
		}}

	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
}

func TestTxProcessor_ProcessWithUsernameMismatchAndSCProcessErrorShouldError(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(9))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return process.ErrUserNameDoesNotMatchInCrossShardTx
		}}

	expectedError := errors.New("expected")
	scProcessor := &testscommon.SCProcessorMock{
		ProcessIfErrorCalled: func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
			return expectedError
		},
	}

	args.ScProcessor = scProcessor

	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, expectedError, err)
}

func TestTxProcessor_ProcessMoveBalanceToSmartPayableContract(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, 32)
	tx.Value = big.NewInt(0)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return 0
	}

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	acntDst.SetCodeMetadata([]byte{0, vmcommon.MetadataPayable})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ShardCoordinator = shardCoordinator
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, 2, saveAccountCalled)
}

func testProcessCheck(t *testing.T, nonce uint64, value *big.Int) {
	saveAccountCalled := 0
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	tx := transaction.Transaction{}
	tx.Nonce = nonce
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = value

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, tx.SndAddr) {
			return 1
		}

		return 0
	}

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ShardCoordinator = shardCoordinator
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestTxProcessor_ProcessMoveBalancesShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, 2, saveAccountCalled)
}

func TestTxProcessor_ProcessOkValsShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 4
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(61)

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	acntSrc.IncreaseNonce(4)
	_ = acntSrc.AddToBalance(big.NewInt(90))
	_ = acntDst.AddToBalance(big.NewInt(10))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.GetNonce())
	assert.Equal(t, big.NewInt(29), acntSrc.GetBalance())
	assert.Equal(t, big.NewInt(71), acntDst.GetBalance())
	assert.Equal(t, 2, saveAccountCalled)
}

func TestTxProcessor_MoveBalanceWithFeesShouldWork(t *testing.T) {
	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 4
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(61)
	tx.GasPrice = 2
	tx.GasLimit = 2

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	acntSrc.IncreaseNonce(4)
	_ = acntSrc.AddToBalance(big.NewInt(90))
	_ = acntDst.AddToBalance(big.NewInt(10))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	txCost := big.NewInt(16)
	feeHandler := &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return nil
		},
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return txCost
		},
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.EconomicsFee = feeHandler
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.GetNonce())
	assert.Equal(t, big.NewInt(13), acntSrc.GetBalance())
	assert.Equal(t, big.NewInt(71), acntDst.GetBalance())
	assert.Equal(t, 2, saveAccountCalled)
}

func TestTxProcessor_ProcessTransactionScDeployTxShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubKeyConverter().Len())
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(46))
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &testscommon.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.DeploySmartContractCalled = func(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return vmcommon.Ok, nil
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, destinationTransactionType process.TransactionType) {
			return process.SCDeployment, process.SCDeployment
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestTxProcessor_ProcessTransactionBuiltInFunctionCallShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubKeyConverter().Len())
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(46))
	acntDst.SetCode([]byte{65})

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	scProcessorMock := &testscommon.SCProcessorMock{}

	wasCalled := false
	scProcessorMock.ExecuteBuiltInFunctionCalled = func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
		wasCalled = true
		return vmcommon.Ok, nil
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestTxProcessor_ProcessTransactionScTxShouldWork(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubKeyConverter().Len())
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc := createUserAcc(tx.SndAddr)
	acntDst := createUserAcc(tx.RcvAddr)

	_ = acntSrc.AddToBalance(big.NewInt(46))
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

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestTxProcessor_ProcessTransactionScTxShouldReturnErrWhenExecutionFails(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubKeyConverter().Len())
	tx.Value = big.NewInt(45)

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(45))
	acntDst := createUserAcc(tx.RcvAddr)
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
		return vmcommon.UserError, process.ErrNoVM
	}

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrNoVM, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, saveAccountCalled)
}

func TestTxProcessor_ProcessTransactionScTxShouldNotBeCalledWhenAdrDstIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	saveAccountCalled := 0
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateRandomByteSlice(createMockPubKeyConverter().Len())
	tx.Value = big.NewInt(45)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, tx.RcvAddr) {
			return 1
		}

		return 0
	}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(45))
	acntDst := createUserAcc(tx.RcvAddr)
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
		return vmcommon.UserError, process.ErrNoVM
	}

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    testscommon.NewPubkeyConverterMock(32),
		ShardCoordinator:   shardCoordinator,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	computeType, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardCoordinator
	args.TxTypeHandler = computeType
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.False(t, wasCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestTxProcessor_ProcessTxFeeIntraShard(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	negMoveBalanceFee := big.NewInt(0).Neg(moveBalanceFee)
	totalGiven := big.NewInt(100)
	args := createArgsForTxProcessor()
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return moveBalanceFee
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return totalGiven
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	tx := &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	acntSnd := &stateMock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negMoveBalanceFee) == 0)
		return nil
	}}

	cost, totalCost, err := execTx.ProcessTxFee(tx, acntSnd, nil, process.MoveBalance, false)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
	assert.True(t, totalGiven.Cmp(totalCost) == 0)
}

func TestTxProcessor_ProcessTxFeeCrossShardMoveBalance(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	negMoveBalanceFee := big.NewInt(0).Neg(moveBalanceFee)
	totalGiven := big.NewInt(100)
	args := createArgsForTxProcessor()
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return moveBalanceFee
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return totalGiven
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	tx := &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	acntSnd := &stateMock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negMoveBalanceFee) == 0)
		return nil
	}}

	cost, totalCost, err := execTx.ProcessTxFee(tx, acntSnd, nil, process.MoveBalance, false)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
	assert.True(t, totalCost.Cmp(totalGiven) == 0)

	tx = &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
		Data:     []byte("data"),
	}

	cost, totalCost, err = execTx.ProcessTxFee(tx, acntSnd, nil, process.MoveBalance, false)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
	assert.True(t, totalCost.Cmp(totalGiven) == 0)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	tx = &transaction.Transaction{
		RcvAddr:  scAddress,
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	cost, totalCost, err = execTx.ProcessTxFee(tx, acntSnd, nil, process.MoveBalance, false)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
	assert.True(t, totalCost.Cmp(totalGiven) == 0)
}

func TestTxProcessor_ProcessTxFeeCrossShardSCCall(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	args := createArgsForTxProcessor()
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return moveBalanceFee
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(0).Mul(moveBalanceFee, moveBalanceFee)
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	tx := &transaction.Transaction{
		RcvAddr:  scAddress,
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
		Data:     []byte("data"),
	}

	totalCost := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GetGasPrice()), big.NewInt(0).SetUint64(tx.GetGasLimit()))
	negTotalCost := big.NewInt(0).Neg(totalCost)
	acntSnd := &stateMock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negTotalCost) == 0)
		return nil
	}}

	cost, totalReturnedCost, err := execTx.ProcessTxFee(tx, acntSnd, nil, process.SCInvoking, false)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
	assert.True(t, totalReturnedCost.Cmp(totalCost) == 0)
}

func TestTxProcessor_ProcessTxFeeMoveBalanceUserTx(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	processingFee := big.NewInt(5)
	negMoveBalanceFee := big.NewInt(0).Neg(moveBalanceFee)
	args := createArgsForTxProcessor()
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return moveBalanceFee
		},
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return processingFee
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	tx := &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	acntSnd := &stateMock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negMoveBalanceFee) == 0)
		return nil
	}}

	cost, totalCost, err := execTx.ProcessTxFee(tx, acntSnd, nil, process.MoveBalance, true)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(processingFee) == 0)
	assert.True(t, totalCost.Cmp(processingFee) == 0)
}

func TestTxProcessor_ProcessTxFeeSCInvokeUserTx(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	processingFee := big.NewInt(5)
	negMoveBalanceFee := big.NewInt(0).Neg(moveBalanceFee)
	gasPerByte := uint64(1)
	args := createArgsForTxProcessor()
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return moveBalanceFee
		},
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			decreasedPrice := int64(gasToUse * 10 / 100)
			return big.NewInt(decreasedPrice)
		},
		ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
			gasLimit := moveBalanceFee.Uint64()

			dataLen := uint64(len(tx.GetData()))
			gasLimit += dataLen * gasPerByte

			return gasLimit
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	tx := &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
		Data:     []byte("aaa"),
	}

	acntSnd := &stateMock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negMoveBalanceFee) == 0)
		return nil
	}}

	cost, totalCost, err := execTx.ProcessTxFee(tx, acntSnd, nil, process.SCInvoking, true)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(processingFee) == 0)
	assert.True(t, totalCost.Cmp(processingFee) == 0)
}

func TestTxProcessor_ProcessTransactionShouldReturnErrForInvalidMetaTx(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = vm.StakingSCAddress
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100000000))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, nil)
	scProcessorMock := &testscommon.SCProcessorMock{
		ProcessIfErrorCalled: func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
			return acntSnd.AddToBalance(tx.GetValue())
		},
	}
	shardC, _ := sharding.NewMultiShardCoordinator(5, 3)
	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeMoveBalanceFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(1)
		},
	}
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsMetaProtectionFlagEnabledInEpochCalled: flagActiveTrueHandler,
	}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, err, process.ErrFailedTransaction)
	assert.Equal(t, uint64(1), acntSrc.GetNonce())
	assert.Equal(t, uint64(99999999), acntSrc.GetBalance().Uint64())

	tx.Data = []byte("something")
	tx.Nonce = tx.Nonce + 1
	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, err, process.ErrFailedTransaction)

	tx.Nonce = tx.Nonce + 1
	tx.GasLimit = 10_000_000
	_, err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
}

func TestTxProcessor_ProcessTransactionShouldTreatAsInvalidTxIfTxTypeIsWrong(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = vm.StakingSCAddress
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(46))

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, nil)
	shardC, _ := sharding.NewMultiShardCoordinator(5, 3)
	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ShardCoordinator = shardC
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(1)
		},
	}
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.InvalidTransaction, process.InvalidTransaction
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	_, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, err, process.ErrFailedTransaction)
	assert.Equal(t, uint64(1), acntSrc.GetNonce())
	assert.Equal(t, uint64(45), acntSrc.GetBalance().Uint64())
}

func TestTxProcessor_ProcessRelayedTransactionV2NotActiveShouldErr(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(0)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTxDest := []byte("sDST")
	userNonce := "00"
	userDataString := "execute@param1"

	marshalizer := &mock.MarshalizerMock{}
	userDataMarshalled, _ := marshalizer.Marshal(userDataString)
	tx.Data = []byte(core.RelayedTransactionV2 +
		"@" +
		hex.EncodeToString(userTxDest) +
		"@" +
		userNonce +
		"@" +
		hex.EncodeToString(userDataMarshalled) +
		"@" +
		"01a2")

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))

	acntFinal := createUserAcc(userTxDest)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTxDest) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	args.ArgsParser = smartContract.NewArgumentParser()
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionV2WithValueShouldErr(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(1)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTxDest := []byte("sDST")
	userNonce := "00"
	userDataString := "execute@param1"

	marshalizer := &mock.MarshalizerMock{}
	userDataMarshalled, _ := marshalizer.Marshal(userDataString)
	tx.Data = []byte(core.RelayedTransactionV2 +
		"@" +
		hex.EncodeToString(userTxDest) +
		"@" +
		userNonce +
		"@" +
		hex.EncodeToString(userDataMarshalled) +
		"@" +
		"01a2")

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))

	acntFinal := createUserAcc(userTxDest)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTxDest) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	args.ArgsParser = smartContract.NewArgumentParser()
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionV2ArgsParserShouldErr(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(0)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTxDest := []byte("sDST")
	userNonce := "00"
	userDataString := "execute@param1"

	marshalizer := &mock.MarshalizerMock{}
	userDataMarshalled, _ := marshalizer.Marshal(userDataString)
	tx.Data = []byte(core.RelayedTransactionV2 +
		"@" +
		hex.EncodeToString(userTxDest) +
		"@" +
		userNonce +
		"@" +
		hex.EncodeToString(userDataMarshalled) +
		"@" +
		"01a2")

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))

	acntFinal := createUserAcc(userTxDest)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTxDest) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	parseError := errors.New("parse error")

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return "", nil, parseError
		}}
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionV2InvalidParamCountShouldErr(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(0)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTxDest := []byte("sDST")
	userNonce := "00"
	userDataString := "execute@param1"

	marshalizer := &mock.MarshalizerMock{}
	userDataMarshalled, _ := marshalizer.Marshal(userDataString)
	tx.Data = []byte(core.RelayedTransactionV2 +
		"@" +
		hex.EncodeToString(userTxDest) +
		"@" +
		userNonce +
		"@" +
		hex.EncodeToString(userDataMarshalled) +
		"@" +
		"01a2" +
		"@" +
		"1010")

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))

	acntFinal := createUserAcc(userTxDest)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTxDest) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	args.ArgsParser = smartContract.NewArgumentParser()
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionV2(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(0)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTxDest := []byte("sDST")
	userNonce := "00"
	userDataString := "execute@param1"

	marshalizer := &mock.MarshalizerMock{}
	userDataMarshalled, _ := marshalizer.Marshal(userDataString)
	tx.Data = []byte(core.RelayedTransactionV2 +
		"@" +
		hex.EncodeToString(userTxDest) +
		"@" +
		userNonce +
		"@" +
		hex.EncodeToString(userDataMarshalled) +
		"@" +
		"01a2")

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))

	acntFinal := createUserAcc(userTxDest)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTxDest) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	args.ArgsParser = smartContract.NewArgumentParser()
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsRelayedTransactionsV2FlagEnabledInEpochCalled: flagActiveTrueHandler,
	}
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestTxProcessor_ProcessRelayedTransaction(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	args.ArgsParser = smartContract.NewArgumentParser()
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsRelayedTransactionsFlagEnabledInEpochCalled: func(epoch uint32) bool { return true },
	}
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)

	tx.Nonce = tx.Nonce + 1
	userTx.Nonce = userTx.Nonce + 1
	userTx.Value = big.NewInt(200)
	userTxMarshalled, _ = marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	returnCode, err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionArgsParserErrorShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	parseError := errors.New("parse error")

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return "", nil, parseError
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, parseError, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionMultipleArgumentsShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{[]byte("0"), []byte("1")}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionFailUnMarshalInnerShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{[]byte("0")}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionDifferentSenderInInnerTxThanReceiverShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = []byte("otherReceiver")
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionSmallerValueInnerTxShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(60)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionGasPriceMismatchShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionGasLimitMismatchShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 1
	tx.GasLimit = 6

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 10,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessRelayedTransactionDisabled(t *testing.T) {
	t.Parallel()

	pubKeyConverter := testscommon.NewPubkeyConverterMock(4)

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(10))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(10))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	scProcessorMock := &testscommon.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubKeyConverter,
		ShardCoordinator:   shardC,
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsESDTMetadataContinuousCleanupFlagEnabledField: true,
		},
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	args := createArgsForTxProcessor()
	args.Accounts = adb
	args.ScProcessor = scProcessorMock
	args.ShardCoordinator = shardC
	args.TxTypeHandler = txTypeHandler
	args.PubkeyConv = pubKeyConverter
	args.ArgsParser = smartContract.NewArgumentParser()
	called := false
	args.BadTxForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			called = true
			return nil
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	returnCode, err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, err, process.ErrFailedTransaction)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, called)
}

func TestTxProcessor_ConsumeMoveBalanceWithUserTx(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.EconomicsFee = &economicsmocks.EconomicsHandlerStub{
		ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
			return big.NewInt(1)
		},
		ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(150)
		},
	}
	args.TxFeeHandler = &mock.FeeAccumulatorStub{
		ProcessTransactionFeeCalled: func(cost *big.Int, devFee *big.Int, hash []byte) {
			assert.Equal(t, cost, big.NewInt(1))
		},
	}
	execTx, _ := txproc.NewTxProcessor(args)

	acntSrc := createUserAcc([]byte("address"))
	_ = acntSrc.AddToBalance(big.NewInt(100))

	originalTxHash := []byte("originalTxHash")
	userTx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		GasPrice: 100,
		GasLimit: 100,
	}

	err := execTx.ProcessMoveBalanceCostRelayedUserTx(userTx, &smartContractResult.SmartContractResult{}, acntSrc, originalTxHash)
	assert.Nil(t, err)
	assert.Equal(t, acntSrc.GetBalance(), big.NewInt(99))
}

func TestTxProcessor_IsCrossTxFromMeShouldWork(t *testing.T) {
	t.Parallel()

	shardC, _ := sharding.NewMultiShardCoordinator(2, 0)
	args := createArgsForTxProcessor()
	args.ShardCoordinator = shardC
	execTx, _ := txproc.NewTxProcessor(args)

	assert.False(t, execTx.IsCrossTxFromMe([]byte("ADR0"), []byte("ADR0")))
	assert.False(t, execTx.IsCrossTxFromMe([]byte("ADR1"), []byte("ADR1")))
	assert.False(t, execTx.IsCrossTxFromMe([]byte("ADR1"), []byte("ADR0")))
	assert.True(t, execTx.IsCrossTxFromMe([]byte("ADR0"), []byte("ADR1")))
}

func TestTxProcessor_ProcessUserTxOfTypeRelayedShouldError(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.RelayedTx, process.RelayedTx
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessUserTxOfTypeMoveBalanceShouldWork(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.MoveBalance, process.MoveBalance
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestTxProcessor_ProcessUserTxOfTypeSCDeploymentShouldWork(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.SCDeployment, process.SCDeployment
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestTxProcessor_ProcessUserTxOfTypeSCInvokingShouldWork(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.SCInvoking, process.SCInvoking
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestTxProcessor_ProcessUserTxOfTypeBuiltInFunctionCallShouldWork(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.BuiltInFunctionCall, process.BuiltInFunctionCall
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestTxProcessor_ProcessUserTxErrNotPayableShouldFailRelayTx(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	args.ScProcessor = &testscommon.SCProcessorMock{IsPayableCalled: func(_, _ []byte) (bool, error) {
		return false, process.ErrAccountNotPayable
	}}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.MoveBalance, process.MoveBalance
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
}

func TestTxProcessor_ProcessUserTxFailedBuiltInFunctionCall(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()
	args.ArgsParser = &mock.ArgumentParserMock{
		ParseCallDataCalled: func(data string) (string, [][]byte, error) {
			return core.RelayedTransaction, [][]byte{userTxMarshalled}, nil
		}}

	args.ScProcessor = &testscommon.SCProcessorMock{
		ExecuteBuiltInFunctionCalled: func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
			return vmcommon.UserError, process.ErrFailedTransaction
		},
	}

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	args.TxTypeHandler = &testscommon.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType, destinationTransactionType process.TransactionType) {
		return process.BuiltInFunctionCall, process.BuiltInFunctionCall
	}}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)
	returnCode, err := execTx.ProcessUserTx(&tx, &userTx, tx.Value, tx.Nonce, txHash)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.ExecutionFailed, returnCode)
}

func TestTxProcessor_ExecuteFailingRelayedTxShouldNotHaveNegativeFee(t *testing.T) {
	t.Parallel()

	userAddr := []byte("user")
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("sSRC")
	tx.RcvAddr = userAddr
	tx.Value = big.NewInt(50)
	tx.GasPrice = 2
	tx.GasLimit = 1

	userTx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(50),
		RcvAddr:  []byte("sDST"),
		SndAddr:  userAddr,
		GasPrice: 1,
		GasLimit: 1,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxMarshalled, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshalled))

	args := createArgsForTxProcessor()

	acntSrc := createUserAcc(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(100))
	acntDst := createUserAcc(tx.RcvAddr)
	_ = acntDst.AddToBalance(big.NewInt(100))
	acntFinal := createUserAcc(userTx.RcvAddr)
	_ = acntFinal.AddToBalance(big.NewInt(100))

	adb := &stateMock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, tx.SndAddr) {
			return acntSrc, nil
		}
		if bytes.Equal(address, tx.RcvAddr) {
			return acntDst, nil
		}
		if bytes.Equal(address, userTx.RcvAddr) {
			return acntFinal, nil
		}

		return nil, errors.New("failure")
	}
	args.Accounts = adb
	shardCoordinator := mock.NewOneShardCoordinatorMock()
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return 1
	}
	args.ShardCoordinator = shardCoordinator

	economicsFee := createFreeTxFeeHandler()
	args.EconomicsFee = economicsFee

	negativeCost := false
	args.TxFeeHandler = &mock.FeeAccumulatorStub{
		ProcessTransactionFeeCalled: func(cost *big.Int, devFee *big.Int, hash []byte) {
			if cost.Cmp(big.NewInt(0)) < 0 {
				negativeCost = true
			}
		},
	}

	execTx, _ := txproc.NewTxProcessor(args)

	txHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, tx)

	err := execTx.ExecuteFailedRelayedTransaction(&userTx, tx.SndAddr, tx.Value, tx.Nonce, &tx, txHash, "")
	assert.Nil(t, err)
	assert.False(t, negativeCost)
}

func TestTxProcessor_shouldIncreaseNonce(t *testing.T) {
	t.Parallel()

	t.Run("fix not enabled, should return true", func(t *testing.T) {
		args := createArgsForTxProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsRelayedNonceFixEnabledField: false,
		}
		txProc, _ := txproc.NewTxProcessor(args)

		assert.True(t, txProc.ShouldIncreaseNonce(nil))
	})
	t.Run("fix enabled, different errors should return true", func(t *testing.T) {
		args := createArgsForTxProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsRelayedNonceFixEnabledField: true,
		}
		txProc, _ := txproc.NewTxProcessor(args)

		assert.True(t, txProc.ShouldIncreaseNonce(nil))
		assert.True(t, txProc.ShouldIncreaseNonce(fmt.Errorf("random error")))
		assert.True(t, txProc.ShouldIncreaseNonce(process.ErrWrongTransaction))
	})
	t.Run("fix enabled, errors for an un-executable transaction should return false", func(t *testing.T) {
		args := createArgsForTxProcessor()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsRelayedNonceFixEnabledField: true,
		}
		txProc, _ := txproc.NewTxProcessor(args)

		assert.False(t, txProc.ShouldIncreaseNonce(process.ErrLowerNonceInTransaction))
		assert.False(t, txProc.ShouldIncreaseNonce(process.ErrHigherNonceInTransaction))
		assert.False(t, txProc.ShouldIncreaseNonce(process.ErrTransactionNotExecutable))
	})
}

func TestTxProcessor_AddNonExecutableLog(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	sender := []byte("sender")
	relayer := []byte("relayer")
	originalTx := &transaction.Transaction{
		SndAddr: relayer,
		RcvAddr: sender,
	}
	originalTxHash, errCalculateHash := core.CalculateHash(args.Marshalizer, args.Hasher, originalTx)
	assert.Nil(t, errCalculateHash)

	t.Run("not a non-executable error should not record log", func(t *testing.T) {
		t.Parallel()

		argsLocal := args
		argsLocal.TxLogsProcessor = &mock.TxLogsProcessorStub{
			SaveLogCalled: func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
				assert.Fail(t, "should have not called SaveLog")

				return nil
			},
		}
		txProc, _ := txproc.NewTxProcessor(argsLocal)
		err := txProc.AddNonExecutableLog(errors.New("random error"), originalTxHash, originalTx)
		assert.Nil(t, err)
	})
	t.Run("is non executable tx error should record log", func(t *testing.T) {
		t.Parallel()

		argsLocal := args
		numLogsSaved := 0
		argsLocal.TxLogsProcessor = &mock.TxLogsProcessorStub{
			SaveLogCalled: func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
				assert.Equal(t, originalTxHash, txHash)
				assert.Equal(t, originalTx, tx)
				assert.Equal(t, 1, len(vmLogs))
				firstLog := vmLogs[0]
				assert.Equal(t, core.SignalErrorOperation, string(firstLog.Identifier))
				assert.Equal(t, sender, firstLog.Address)
				assert.Empty(t, firstLog.Data)
				assert.Empty(t, firstLog.Topics)
				numLogsSaved++

				return nil
			},
		}

		txProc, _ := txproc.NewTxProcessor(argsLocal)
		err := txProc.AddNonExecutableLog(process.ErrLowerNonceInTransaction, originalTxHash, originalTx)
		assert.Nil(t, err)

		err = txProc.AddNonExecutableLog(process.ErrHigherNonceInTransaction, originalTxHash, originalTx)
		assert.Nil(t, err)

		err = txProc.AddNonExecutableLog(process.ErrTransactionNotExecutable, originalTxHash, originalTx)
		assert.Nil(t, err)

		assert.Equal(t, 3, numLogsSaved)
	})
}
