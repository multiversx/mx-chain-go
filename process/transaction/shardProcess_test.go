package transaction_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	txproc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func generateRandomByteSlice(size int) []byte {
	buff := make([]byte, size)
	_, _ = rand.Reader.Read(buff)

	return buff
}

func feeHandlerMock() *mock.FeeHandlerStub {
	return &mock.FeeHandlerStub{
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
		ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(0)
		},
	}
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

func createTxProcessor() txproc.TxProcessor {
	txProc, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	return txProc
}

//------- NewTxProcessor

func TestNewTxProcessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		nil,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		nil,
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilPubkeyConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilMarshalizerMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		nil,
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilShardCoordinatorMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		nil,
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilSCProcessorShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilSmartContractProcessor, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		nil,
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Equal(t, process.ErrNilUnsignedTxHandler, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

//------- getAccounts

func TestTxProcessor_GetAccountsShouldErrNilAddressContainer(t *testing.T) {
	t.Parallel()

	adb := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

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

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	adr1 := []byte{65}
	adr2 := []byte{67}

	_, _, err := execTx.GetAccounts(adr1, adr2)
	assert.NotNil(t, err)
}

func TestTxProcessor_GetAccountsOkValsSrcShouldWork(t *testing.T) {
	t.Parallel()

	adb := mock.AccountsStub{}

	adr1 := []byte{65}
	adr2 := []byte{67}

	acnt1, _ := state.NewUserAccount(adr1)
	acnt2, _ := state.NewUserAccount(adr2)

	adb.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
		if bytes.Equal(address, adr1) {
			return acnt1, nil
		}

		if bytes.Equal(address, adr2) {
			return nil, errors.New("failure on destination")
		}

		return nil, errors.New("failure")
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	execTx, _ := txproc.NewTxProcessor(
		&adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

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

	adb := mock.AccountsStub{}

	adr1 := []byte{65}
	adr2 := []byte{67}

	acnt1, _ := state.NewUserAccount(adr1)
	acnt2, _ := state.NewUserAccount(adr2)

	adb.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
		if bytes.Equal(address, adr1) {
			return nil, errors.New("failure on source")
		}

		if bytes.Equal(address, adr2) {
			return acnt2, nil
		}

		return nil, errors.New("failure")
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	execTx, _ := txproc.NewTxProcessor(
		&adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

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

	acnt1, _ := state.NewUserAccount(adr1)
	acnt2, _ := state.NewUserAccount(adr2)

	adb := createAccountStub(adr1, adr2, acnt1, acnt2)

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.Equal(t, acnt2, a2)
}

func TestTxProcessor_GetSameAccountShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	adr2 := []byte{65}

	acnt1, _ := state.NewUserAccount(adr1)
	acnt2, _ := state.NewUserAccount(adr2)

	adb := createAccountStub(adr1, adr2, acnt1, acnt2)

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	a1, a2, err := execTx.GetAccounts(adr1, adr1)
	assert.Nil(t, err)
	assert.True(t, a1 == a2)
}

//------- checkTxValues

func TestTxProcessor_CheckTxValuesHigherNonceShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1, err := state.NewUserAccount(adr1)
	assert.Nil(t, err)

	execTx := *createTxProcessor()

	acnt1.Nonce = 6

	err = execTx.CheckTxValues(&transaction.Transaction{Nonce: 7}, acnt1, nil)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestTxProcessor_CheckTxValuesLowerNonceShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1, err := state.NewUserAccount(adr1)
	assert.Nil(t, err)

	execTx := *createTxProcessor()

	acnt1.Nonce = 6

	err = execTx.CheckTxValues(&transaction.Transaction{Nonce: 5}, acnt1, nil)
	assert.Equal(t, process.ErrLowerNonceInTransaction, err)
}

func TestTxProcessor_CheckTxValuesInsufficientFundsShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1, err := state.NewUserAccount(adr1)
	assert.Nil(t, err)

	execTx := *createTxProcessor()

	acnt1.Balance = big.NewInt(67)

	err = execTx.CheckTxValues(&transaction.Transaction{Value: big.NewInt(68)}, acnt1, nil)
	assert.Equal(t, process.ErrInsufficientFunds, err)
}

func TestTxProcessor_CheckTxValuesOkValsShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1, err := state.NewUserAccount(adr1)
	assert.Nil(t, err)

	execTx := *createTxProcessor()

	acnt1.Balance = big.NewInt(67)

	err = execTx.CheckTxValues(&transaction.Transaction{Value: big.NewInt(67)}, acnt1, nil)
	assert.Nil(t, err)
}

//------- moveBalances
func TestTxProcessor_MoveBalancesShouldNotFailWhenAcntSrcIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	adrDst := []byte{67}
	acntDst, _ := state.NewUserAccount(adrDst)

	execTx := *createTxProcessor()
	err := execTx.MoveBalances(nil, acntDst, big.NewInt(0))
	assert.Nil(t, err)
}

func TestTxProcessor_MoveBalancesShouldNotFailWhenAcntDstIsNotInNodeShard(t *testing.T) {
	t.Parallel()

	adrSrc := []byte{65}
	acntSrc, _ := state.NewUserAccount(adrSrc)

	execTx := *createTxProcessor()
	err := execTx.MoveBalances(acntSrc, nil, big.NewInt(0))

	assert.Nil(t, err)
}

func TestTxProcessor_MoveBalancesOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adrSrc := []byte{65}
	acntSrc, err := state.NewUserAccount(adrSrc)
	assert.Nil(t, err)

	adrDst := []byte{67}
	acntDst, err := state.NewUserAccount(adrDst)
	assert.Nil(t, err)

	execTx := *createTxProcessor()

	acntSrc.Balance = big.NewInt(64)
	acntDst.Balance = big.NewInt(31)
	err = execTx.MoveBalances(acntSrc, acntDst, big.NewInt(14))

	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(50), acntSrc.Balance)
	assert.Equal(t, big.NewInt(45), acntDst.Balance)
}

func TestTxProcessor_MoveBalancesToSelfOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adrSrc := []byte{65}
	acntSrc, err := state.NewUserAccount(adrSrc)
	assert.Nil(t, err)

	acntDst := acntSrc

	execTx := *createTxProcessor()

	acntSrc.Balance = big.NewInt(64)

	err = execTx.MoveBalances(acntSrc, acntDst, big.NewInt(1))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(64), acntSrc.Balance)
	assert.Equal(t, big.NewInt(64), acntDst.Balance)
}

//------- increaseNonce

func TestTxProcessor_IncreaseNonceOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adrSrc := []byte{65}
	acntSrc, err := state.NewUserAccount(adrSrc)
	assert.Nil(t, err)

	execTx := *createTxProcessor()

	acntSrc.Nonce = 45

	execTx.IncreaseNonce(acntSrc)
	assert.Equal(t, uint64(46), acntSrc.Nonce)
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

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
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

func TestTxProcessor_ProcessMoveBalanceToSmartNonPayableContract(t *testing.T) {
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, process.ErrAccountNotPayable, err)
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	acntDst.CodeMetadata = []byte{0, vmcommon.METADATA_PAYABLE}

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	acntSrc.Nonce = 4
	acntSrc.Balance = big.NewInt(90)
	acntDst.Balance = big.NewInt(10)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.Nonce)
	assert.Equal(t, big.NewInt(29), acntSrc.Balance)
	assert.Equal(t, big.NewInt(71), acntDst.Balance)
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

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntDst, err := state.NewUserAccount(tx.RcvAddr)
	assert.Nil(t, err)

	acntSrc.Nonce = 4
	acntSrc.Balance = big.NewInt(90)
	acntDst.Balance = big.NewInt(10)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)
	adb.SaveAccountCalled = func(account state.AccountHandler) error {
		saveAccountCalled++
		return nil
	}

	txCost := big.NewInt(16)
	feeHandler := &mock.FeeHandlerStub{
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
		ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return txCost
		},
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		feeHandler,
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.Nonce)
	assert.Equal(t, big.NewInt(13), acntSrc.Balance)
	assert.Equal(t, big.NewInt(71), acntDst.Balance)
	assert.Equal(t, 2, saveAccountCalled)
}

func TestTxProcessor_ProcessTransactionScTxShouldWork(t *testing.T) {
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

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
				return process.SCInvoking
			},
		},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
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
		return vmcommon.UserError, process.ErrNoVM
	}

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
		scProcessorMock,
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
			return process.SCInvoking
		}},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
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
		return vmcommon.UserError, process.ErrNoVM
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  mock.NewPubkeyConverterMock(32),
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: make(map[string]struct{}),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	computeType, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
		scProcessorMock,
		&mock.FeeAccumulatorStub{},
		computeType,
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.False(t, wasCalled)
	assert.Equal(t, 1, saveAccountCalled)
}

func TestTxProcessor_ProcessTxFeeIntraShard(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	negMoveBalanceFee := big.NewInt(0).Neg(moveBalanceFee)
	execTx, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(2),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
				return moveBalanceFee
			},
		},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)
	tx := &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	acntSnd := &mock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negMoveBalanceFee) == 0)
		return nil
	}}
	acntDst := &mock.UserAccountStub{}

	cost, err := execTx.ProcessTxFee(tx, acntSnd, acntDst)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
}

func TestTxProcessor_ProcessTxFeeCrossShardMoveBalance(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	negMoveBalanceFee := big.NewInt(0).Neg(moveBalanceFee)
	execTx, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(2),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
				return moveBalanceFee
			},
		},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)
	tx := &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	acntSnd := &mock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negMoveBalanceFee) == 0)
		return nil
	}}

	cost, err := execTx.ProcessTxFee(tx, acntSnd, nil)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)

	tx = &transaction.Transaction{
		RcvAddr:  []byte("aaa"),
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
		Data:     []byte("data"),
	}

	cost, err = execTx.ProcessTxFee(tx, acntSnd, nil)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)

	scAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	tx = &transaction.Transaction{
		RcvAddr:  scAddress,
		SndAddr:  []byte("bbb"),
		GasPrice: moveBalanceFee.Uint64(),
		GasLimit: moveBalanceFee.Uint64(),
	}

	cost, err = execTx.ProcessTxFee(tx, acntSnd, nil)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
}

func TestTxProcessor_ProcessTxFeeCrossShardSCCall(t *testing.T) {
	t.Parallel()

	moveBalanceFee := big.NewInt(50)
	execTx, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.NewMultiShardsCoordinatorMock(2),
		&mock.SCProcessorMock{},
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
				return moveBalanceFee
			},
		},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

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
	acntSnd := &mock.UserAccountStub{AddToBalanceCalled: func(value *big.Int) error {
		assert.True(t, value.Cmp(negTotalCost) == 0)
		return nil
	}}

	cost, err := execTx.ProcessTxFee(tx, acntSnd, nil)
	assert.Nil(t, err)
	assert.True(t, cost.Cmp(moveBalanceFee) == 0)
}

func TestTxProcessor_ProcessTransactionShouldReturnErrForInvalidMetaTx(t *testing.T) {
	t.Parallel()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = factory.StakingSCAddress
	tx.Value = big.NewInt(45)
	tx.GasPrice = 1
	tx.GasLimit = 1

	acntSrc, err := state.NewUserAccount(tx.SndAddr)
	assert.Nil(t, err)
	acntSrc.Balance = big.NewInt(45)

	adb := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, nil)
	scProcessorMock := &mock.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(5, 3)
	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardC,
		scProcessorMock,
		&mock.FeeAccumulatorStub{},
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
				return process.MoveBalance
			},
		},
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.ArgumentParserMock{},
		&mock.IntermediateTransactionHandlerMock{},
	)

	_, err = execTx.ProcessTransaction(&tx)
	assert.Equal(t, err, process.ErrFailedTransaction)
}

func TestTxProcessor_ProcessRelayedTransaction(t *testing.T) {
	t.Parallel()

	pubKeyConverter := mock.NewPubkeyConverterMock(4)

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

	acntSrc, _ := state.NewUserAccount(tx.SndAddr)
	acntSrc.Balance = big.NewInt(100)
	acntDst, _ := state.NewUserAccount(tx.RcvAddr)
	acntDst.Balance = big.NewInt(10)
	acntFinal, _ := state.NewUserAccount(userTx.RcvAddr)
	acntFinal.Balance = big.NewInt(10)

	adb := &mock.AccountsStub{}
	adb.LoadAccountCalled = func(address []byte) (state.AccountHandler, error) {
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
	scProcessorMock := &mock.SCProcessorMock{}
	shardC, _ := sharding.NewMultiShardCoordinator(1, 0)

	argTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pubKeyConverter,
		ShardCoordinator: shardC,
		BuiltInFuncNames: make(map[string]struct{}),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argTxTypeHandler)

	execTx, _ := txproc.NewTxProcessor(
		adb,
		mock.HasherMock{},
		pubKeyConverter,
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		shardC,
		scProcessorMock,
		&mock.FeeAccumulatorStub{},
		txTypeHandler,
		feeHandlerMock(),
		&mock.IntermediateTransactionHandlerMock{},
		&mock.IntermediateTransactionHandlerMock{},
		smartContract.NewArgumentParser(),
		&mock.IntermediateTransactionHandlerMock{},
	)

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
