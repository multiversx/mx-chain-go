package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	txproc "github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func createAccountStub(sndAddr, rcvAddr []byte,
	acntSrc, acntDst *mock.JournalizedAccountWrapMock,
) *mock.AccountsStub {
	accounts := mock.AccountsStub{}

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), sndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), rcvAddr) {
			return acntDst, nil
		}

		return nil, errors.New("failure")
	}

	return &accounts
}

func createTxProcessor() txproc.TxProcessor {
	txProc, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	return txProc
}

//------- NewTxProcessor

func TestNewTxProcessor_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		nil,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		nil,
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilAddressConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		nil,
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilAddressConverter, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilMarshalizerMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_NilShardCoordinatorMockShouldErr(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, txProc)
}

func TestNewTxProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	txProc, err := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, txProc)
}

//------- SChandler

func TestTxProcessor_GetSetSChandlerShouldWork(t *testing.T) {
	t.Parallel()

	execTx := *createTxProcessor()

	f := func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		return nil
	}

	execTx.SetSCHandler(f)
	assert.NotNil(t, execTx.SCHandler())
}

//------- getAddresses

func TestTxProcessor_GetAddressErrAddressConvShouldErr(t *testing.T) {
	t.Parallel()

	addressConv := &mock.AddressConverterMock{}

	execTx, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		addressConv,
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	addressConv.Fail = true

	tx := transaction.Transaction{}

	_, _, err := execTx.GetAddresses(&tx)
	assert.NotNil(t, err)
}

func TestTxProcessor_GetAddressOkValsShouldWork(t *testing.T) {
	t.Parallel()

	execTx := *createTxProcessor()

	tx := transaction.Transaction{}
	tx.RcvAddr = []byte{65, 66, 67}
	tx.SndAddr = []byte{32, 33, 34}

	adrSnd, adrRcv, err := execTx.GetAddresses(&tx)
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, adrRcv.Bytes())
	assert.Equal(t, []byte{32, 33, 34}, adrSnd.Bytes())
}

//------- getAccounts

func TestTxProcessor_GetAccountsShouldErrNilAddressContainer(t *testing.T) {
	accounts := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	_, _, err := execTx.GetAccounts(nil, adr2)
	assert.Equal(t, process.ErrNilAddressContainer, err)

	_, _, err = execTx.GetAccounts(adr1, nil)
	assert.Equal(t, process.ErrNilAddressContainer, err)
}

func TestTxProcessor_GetAccountsMalfunctionAccountsShouldErr(t *testing.T) {
	accounts := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	_, _, err := execTx.GetAccounts(adr1, adr2)
	assert.NotNil(t, err)
}

func TestTxProcessor_GetAccountsOkValsSrcShouldWork(t *testing.T) {
	accounts := mock.AccountsStub{}

	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)
	acnt2 := mock.NewJournalizedAccountWrapMock(adr2)

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if addressContainer == adr1 {
			return acnt1, nil
		}

		if addressContainer == adr2 {
			return nil, errors.New("failure on destination")
		}

		return nil, errors.New("failure")
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	execTx, _ := txproc.NewTxProcessor(
		&accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
	)

	shardCoordinator.ComputeShardForAddressCalled = func(container state.AddressContainer, converter state.AddressConverter) uint32 {
		if bytes.Equal(container.Bytes(), adr2.Bytes()) {
			return 1
		}

		return 0
	}

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.NotEqual(t, acnt2, a2)
	assert.Equal(t, nil, a2)
}

func TestTxProcessor_GetAccountsOkValsDsthouldWork(t *testing.T) {
	accounts := mock.AccountsStub{}

	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)
	acnt2 := mock.NewJournalizedAccountWrapMock(adr2)

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if addressContainer == adr1 {
			return nil, errors.New("failure on source")
		}

		if addressContainer == adr2 {
			return acnt2, nil
		}

		return nil, errors.New("failure")
	}

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	execTx, _ := txproc.NewTxProcessor(
		&accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
	)

	shardCoordinator.ComputeShardForAddressCalled = func(container state.AddressContainer, converter state.AddressConverter) uint32 {
		if bytes.Equal(container.Bytes(), adr1.Bytes()) {
			return 1
		}

		return 0
	}

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.NotEqual(t, acnt1, a1)
	assert.Equal(t, nil, a1)
	assert.Equal(t, acnt2, a2)
}

func TestTxProcessor_GetAccountsOkValsShouldWork(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)
	acnt2 := mock.NewJournalizedAccountWrapMock(adr2)

	accounts := createAccountStub(adr1.Bytes(), adr2.Bytes(), acnt1, acnt2)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.Equal(t, acnt2, a2)
}

func TestTxProcessor_GetSameAccountShouldWork(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{65})

	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)
	acnt2 := mock.NewJournalizedAccountWrapMock(adr2)

	accounts := createAccountStub(adr1.Bytes(), adr2.Bytes(), acnt1, acnt2)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	a1, a2, err := execTx.GetAccounts(adr1, adr1)
	assert.Nil(t, err)
	assert.True(t, a1 == a2)
}

//------- callSCHandler

func TestTxProcessor_NoCallSCHandlerShouldErr(t *testing.T) {
	execTx := *createTxProcessor()

	err := execTx.CallSCHandler(nil)
	assert.Equal(t, process.ErrNoVM, err)
}

func TestTxProcessor_WithCallSCHandlerShouldWork(t *testing.T) {
	execTx := *createTxProcessor()

	wasCalled := false
	errOutput := errors.New("not really error, just checking output")
	execTx.SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		wasCalled = true
		return errOutput
	})

	err := execTx.CallSCHandler(nil)
	assert.Equal(t, errOutput, err)
	assert.True(t, wasCalled)
}

//------- checkTxValues

func TestTxProcessor_CheckTxValuesHigherNonceShouldErr(t *testing.T) {
	t.Skip()
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx := *createTxProcessor()

	acnt1.BaseAccount().Nonce = 6

	err := execTx.CheckTxValues(acnt1, big.NewInt(0), 7)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestTxProcessor_CheckTxValuesLowerNonceShouldErr(t *testing.T) {
	t.Skip()
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx := *createTxProcessor()

	acnt1.BaseAccount().Nonce = 6

	err := execTx.CheckTxValues(acnt1, big.NewInt(0), 5)
	assert.Equal(t, process.ErrLowerNonceInTransaction, err)
}

func TestTxProcessor_CheckTxValuesInsufficientFundsShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx := *createTxProcessor()

	acnt1.BaseAccount().Balance = big.NewInt(67)

	err := execTx.CheckTxValues(acnt1, big.NewInt(68), 0)
	assert.Equal(t, process.ErrInsufficientFunds, err)
}

func TestTxProcessor_CheckTxValuesOkValsShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx := *createTxProcessor()

	acnt1.BaseAccount().Balance = big.NewInt(67)

	err := execTx.CheckTxValues(acnt1, big.NewInt(67), 0)
	assert.Nil(t, err)
}

//------- moveBalances

func TestTxProcessor_MoveBalancesFailureAcnt1ShouldErr(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDst := mock.NewAddressMock([]byte{67})
	acntDst := mock.NewJournalizedAccountWrapMock(adrDst)

	execTx := *createTxProcessor()

	acntSrc.Fail = true

	err := execTx.MoveBalances(acntSrc, acntDst, big.NewInt(0))
	assert.NotNil(t, err)
}

func TestTxProcessor_MoveBalancesFailureAcnt2ShouldErr(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDst := mock.NewAddressMock([]byte{67})
	acntDst := mock.NewJournalizedAccountWrapMock(adrDst)

	execTx := *createTxProcessor()

	acntDst.Fail = true

	err := execTx.MoveBalances(acntSrc, acntDst, big.NewInt(0))
	assert.NotNil(t, err)
}

func TestTxProcessor_MoveBalancesShouldNotFailWhenAcntSrcIsNotInNodeShard(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDst := mock.NewAddressMock([]byte{67})
	acntDst := mock.NewJournalizedAccountWrapMock(adrDst)

	execTx := *createTxProcessor()

	acntSrc.Fail = true

	err := execTx.MoveBalances(nil, acntDst, big.NewInt(0))
	assert.Nil(t, err)
}

func TestTxProcessor_MoveBalancesShouldNotFailWhenAcntDstIsNotInNodeShard(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDst := mock.NewAddressMock([]byte{67})
	acntDst := mock.NewJournalizedAccountWrapMock(adrDst)

	execTx := *createTxProcessor()

	acntDst.Fail = true

	err := execTx.MoveBalances(acntSrc, nil, big.NewInt(0))
	assert.Nil(t, err)
}

func TestTxProcessor_MoveBalancesOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDst := mock.NewAddressMock([]byte{67})
	acntDst := mock.NewJournalizedAccountWrapMock(adrDst)

	execTx := *createTxProcessor()

	acntSrc.Balance = big.NewInt(64)
	acntDst.Balance = big.NewInt(31)

	err := execTx.MoveBalances(acntSrc, acntDst, big.NewInt(14))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(50), acntSrc.Balance)
	assert.Equal(t, big.NewInt(45), acntDst.Balance)

}

func TestTxProcessor_MoveBalancesToSelfOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	acntDst := acntSrc

	execTx := *createTxProcessor()

	acntSrc.Balance = big.NewInt(64)

	err := execTx.MoveBalances(acntSrc, acntDst, big.NewInt(1))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(64), acntSrc.Balance)
	assert.Equal(t, big.NewInt(64), acntDst.Balance)

}

//------- increaseNonce

func TestTxProcessor_IncreaseNonceOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	execTx := *createTxProcessor()

	acntSrc.Nonce = 45

	err := execTx.IncreaseNonce(acntSrc)
	assert.Nil(t, err)
	assert.Equal(t, uint64(46), acntSrc.Nonce)
}

//------- ProcessTransaction

func TestTxProcessor_ProcessTransactionNilTxShouldErr(t *testing.T) {
	execTx := *createTxProcessor()

	err := execTx.ProcessTransaction(nil, 4)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestTxProcessor_ProcessTransactionErrAddressConvShouldErr(t *testing.T) {
	addressConv := &mock.AddressConverterMock{}

	execTx, _ := txproc.NewTxProcessor(
		&mock.AccountsStub{},
		mock.HasherMock{},
		addressConv,
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	addressConv.Fail = true

	err := execTx.ProcessTransaction(&transaction.Transaction{}, 4)
	assert.NotNil(t, err)
}

func TestTxProcessor_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	accounts := createAccountStub(nil, nil, nil, nil)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.NotNil(t, err)
}

func TestTxProcessor_ProcessTransactionScTxShouldWork(t *testing.T) {
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Balance = big.NewInt(45)
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))
	acntDst.SetCode([]byte{65})

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	wasCalled := false
	execTx.SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		wasCalled = true
		return nil
	})

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestTxProcessor_ProcessTransactionScTxShouldReturnErrWhenExecutionFails(t *testing.T) {
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Balance = big.NewInt(45)
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))
	acntDst.SetCode([]byte{65})

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := errors.New("sc execution error")
	execTx.SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		return err
	})

	err2 := execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, err, err2)
}

func TestTxProcessor_ProcessTransactionScTxShouldNotBeCalledWhenAdrDstIsNotInNodeShard(t *testing.T) {
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	shardCoordinator.ComputeShardForAddressCalled = func(container state.AddressContainer, converter state.AddressConverter) uint32 {
		if bytes.Equal(container.Bytes(), tx.RcvAddr) {
			return 1
		}

		return 0
	}

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Balance = big.NewInt(45)
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))
	acntDst.SetCode([]byte{65})

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
	)

	wasCalled := false
	execTx.SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		wasCalled = true
		return nil
	})

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestTxProcessor_ProcessTransactionRegisterTxShouldWork(t *testing.T) {
	rd := state.RegistrationData{
		OriginatorPubKey: []byte("a"),
		NodePubKey:       []byte("b"),
		RoundIndex:       6,
		Action:           state.ArUnregister,
		Stake:            big.NewInt(45),
	}

	marshalizer := mock.MarshalizerMock{}
	buff, err := marshalizer.Marshal(&rd)
	assert.Nil(t, err)

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = state.RegistrationAddress.Bytes()
	tx.Value = big.NewInt(0)
	tx.Data = buff

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntReg := mock.NewJournalizedAccountWrapMock(state.RegistrationAddress)

	wasCalledAppend := false

	data2 := &state.RegistrationData{}

	acntReg.AppendDataRegistrationWithJournalCalled = func(data *state.RegistrationData) error {
		wasCalledAppend = true
		data2 = data
		return nil
	}

	accounts := createAccountStub(tx.SndAddr, state.RegistrationAddress.Bytes(), acntSrc, acntReg)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	err = execTx.ProcessTransaction(&tx, 1)
	assert.Nil(t, err)
	assert.True(t, wasCalledAppend)
	assert.Equal(t, big.NewInt(45), data2.Stake)
	assert.Equal(t, []byte("SRC"), data2.OriginatorPubKey)
	assert.Equal(t, []byte("b"), data2.NodePubKey)
	assert.Equal(t, int32(1), data2.RoundIndex)
	assert.Equal(t, state.ArUnregister, data2.Action)
}

func TestTxProcessor_ProcessCheckNotPassShouldErr(t *testing.T) {
	t.Skip()

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestTxProcessor_ProcessCheckShouldPassWhenAdrSrcIsNotInNodeShard(t *testing.T) {
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	shardCoordinator.ComputeShardForAddressCalled = func(container state.AddressContainer, converter state.AddressConverter) uint32 {
		if bytes.Equal(container.Bytes(), tx.SndAddr) {
			return 1
		}

		return 0
	}

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
}

func TestTxProcessor_ProcessMoveBalancesFailShouldErr(t *testing.T) {
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Fail = true
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.NotNil(t, err)
}

func TestTxProcessor_ProcessMoveBalancesShouldPassWhenAdrSrcIsNotInNodeShard(t *testing.T) {
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	shardCoordinator.ComputeShardForAddressCalled = func(container state.AddressContainer, converter state.AddressConverter) uint32 {
		if bytes.Equal(container.Bytes(), tx.SndAddr) {
			return 1
		}

		return 0
	}

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Fail = true
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
}

func TestTxProcessor_ProcessIncreaseNonceFailShouldErr(t *testing.T) {
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.FailSetNonceWithJurnal = true
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, "failure setting nonce", err.Error())
}

func TestTxProcessor_ProcessIncreaseNonceShouldPassWhenAdrSrcIsNotInNodeShard(t *testing.T) {
	shardCoordinator := mock.NewOneShardCoordinatorMock()

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(0)

	shardCoordinator.ComputeShardForAddressCalled = func(container state.AddressContainer, converter state.AddressConverter) uint32 {
		if bytes.Equal(container.Bytes(), tx.SndAddr) {
			return 1
		}

		return 0
	}

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.FailSetNonceWithJurnal = true
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		shardCoordinator,
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
}

func TestTxProcessor_ProcessOkValsShouldWork(t *testing.T) {
	tx := transaction.Transaction{}
	tx.Nonce = 4
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(61)

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Nonce = 4
	acntSrc.Balance = big.NewInt(90)
	acntDst := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))
	acntDst.Balance = big.NewInt(10)

	accounts := createAccountStub(tx.SndAddr, tx.RcvAddr, acntSrc, acntDst)

	execTx, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.Nonce)
	assert.Equal(t, big.NewInt(29), acntSrc.Balance)
	assert.Equal(t, big.NewInt(71), acntDst.Balance)
}

//------- SetBalancesToTrie

func TestTxProcessor_SetBalancesToTrieDirtyAccountsShouldErr(t *testing.T) {
	t.Parallel()

	accounts := &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 1
		},
	}

	txProc, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	hash, err := txProc.SetBalancesToTrie(make(map[string]*big.Int))

	assert.Nil(t, hash)
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestTxProcessor_SetBalancesToTrieNilMapShouldErr(t *testing.T) {
	t.Parallel()

	accounts := &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 0
		},
	}

	txProc, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	hash, err := txProc.SetBalancesToTrie(nil)

	assert.Nil(t, hash)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestTxProcessor_SetBalancesToTrieCommitFailsShouldRevert(t *testing.T) {
	t.Parallel()

	adr1 := []byte("accnt1")
	adr2 := []byte("accnt2")

	accnt1 := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(adr1))
	accnt2 := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(adr2))

	val1 := big.NewInt(10)
	val2 := big.NewInt(20)

	revertCalled := false
	errCommit := errors.New("should err")

	accounts := createAccountStub(adr1, adr2, accnt1, accnt2)

	accounts.JournalLenCalled = func() int {
		return 0
	}

	accounts.CommitCalled = func() (i []byte, e error) {
		return nil, errCommit
	}

	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		revertCalled = true
		return nil
	}

	txProc, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	m := make(map[string]*big.Int)
	m[string(adr1)] = val1
	m[string(adr2)] = val2

	hash, err := txProc.SetBalancesToTrie(m)

	assert.Nil(t, hash)
	assert.Equal(t, errCommit, err)
	assert.True(t, revertCalled)
}

func TestTxProcessor_SetBalancesToTrieNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte("accnt1")
	adr2 := []byte("accnt2")

	accnt1 := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(adr1))
	accnt2 := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(adr2))

	val1 := big.NewInt(10)
	val2 := big.NewInt(20)

	rootHash := []byte("resulted root hash")

	accounts := createAccountStub(adr1, adr2, accnt1, accnt2)

	accounts.JournalLenCalled = func() int {
		return 0
	}

	accounts.CommitCalled = func() (i []byte, e error) {
		return rootHash, nil
	}

	txProc, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	m := make(map[string]*big.Int)
	m[string(adr1)] = val1
	m[string(adr2)] = val2

	hash, err := txProc.SetBalancesToTrie(m)

	assert.Equal(t, rootHash, hash)
	assert.Nil(t, err)
	assert.Equal(t, val1, accnt1.Balance)
	assert.Equal(t, val2, accnt2.Balance)
}

func TestTxProcessor_SetBalancesToTrieAccountsFailShouldErr(t *testing.T) {
	t.Parallel()

	adr1 := []byte("accnt1")
	adr2 := []byte("accnt2")

	val1 := big.NewInt(10)
	val2 := big.NewInt(20)

	rootHash := []byte("resulted root hash")

	errAccounts := errors.New("accounts error")

	accounts := &mock.AccountsStub{
		JournalLenCalled: func() int {
			return 0
		},
		GetJournalizedAccountCalled: func(addressContainer state.AddressContainer) (wrapper state.JournalizedAccountWrapper, e error) {

			return nil, errAccounts
		},
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
	}

	txProc, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	m := make(map[string]*big.Int)
	m[string(adr1)] = val1
	m[string(adr2)] = val2

	hash, err := txProc.SetBalancesToTrie(m)

	assert.Nil(t, hash)
	assert.Equal(t, errAccounts, err)
}

func TestTxProcessor_SetBalancesToTrieOkValsShouldWork(t *testing.T) {
	t.Parallel()

	adr1 := []byte("accnt1")
	adr2 := []byte("accnt2")

	accnt1 := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(adr1))
	accnt2 := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(adr2))

	val1 := big.NewInt(10)
	val2 := big.NewInt(20)

	rootHash := []byte("resulted root hash")

	accounts := createAccountStub(adr1, adr2, accnt1, accnt2)

	accounts.JournalLenCalled = func() int {
		return 0
	}

	accounts.CommitCalled = func() (i []byte, e error) {
		return rootHash, nil
	}

	txProc, _ := txproc.NewTxProcessor(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	m := make(map[string]*big.Int)
	m[string(adr1)] = val1
	m[string(adr2)] = val2

	hash, err := txProc.SetBalancesToTrie(m)

	assert.Equal(t, rootHash, hash)
	assert.Nil(t, err)
	assert.Equal(t, val1, accnt1.Balance)
	assert.Equal(t, val2, accnt2.Balance)
}
