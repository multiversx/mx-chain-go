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

//------- NewExecTransaction

func TestNewExecTransaction_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(
		nil,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)
	assert.Equal(t, execution.ErrNilAccountsAdapter, err)
}

func TestNewExecTransaction_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		nil,
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)
	assert.Equal(t, execution.ErrNilHasher, err)
}

func TestNewExecTransaction_NilAddressConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		nil,
		&mock.MarshalizerMock{},
	)
	assert.Equal(t, execution.ErrNilAddressConverter, err)
}

func TestNewExecTransaction_NilMarshalizerMockShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		nil,
	)
	assert.Equal(t, execution.ErrNilMarshalizer, err)
}

func TestNewExecTransaction_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)
}

//------- SCHandler

func TestExecTransaction_GetSetSChandlerShouldWork(t *testing.T) {
	t.Parallel()

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	f := func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		return nil
	}

	execTx.SetSCHandler(f)
	assert.NotNil(t, execTx.SCHandler())
}

//------- getAddresses

func TestExecTransaction_GetAddressErrAddressConvShouldErr(t *testing.T) {
	t.Parallel()

	addressConv := &mock.AddressConverterMock{}

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		addressConv,
		&mock.MarshalizerMock{},
	)

	addressConv.Fail = true

	tx := transaction.Transaction{}

	_, _, err := execTx.GetAddresses(&tx)
	assert.NotNil(t, err)
}

func TestExecTransaction_GetAddressOkValsShouldWork(t *testing.T) {
	t.Parallel()

	addressConv := &mock.AddressConverterMock{}

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		addressConv,
		&mock.MarshalizerMock{},
	)

	tx := transaction.Transaction{}
	tx.RcvAddr = []byte{65, 66, 67}
	tx.SndAddr = []byte{32, 33, 34}

	adrSnd, adrRcv, err := execTx.GetAddresses(&tx)
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, adrRcv.Bytes())
	assert.Equal(t, []byte{32, 33, 34}, adrSnd.Bytes())
}

//------- getAccounts

func TestExecTransaction_GetAccountsMalfunctionAccountsShouldErr(t *testing.T) {
	accounts := mock.AccountsStub{}
	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		return nil, errors.New("failure")
	}

	execTx, _ := exTransaction.NewExecTransaction(
		&accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	_, _, err := execTx.GetAccounts(adr1, adr2)
	assert.NotNil(t, err)
}

func TestExecTransaction_GetAccountsOkValsShouldWork(t *testing.T) {
	accounts := mock.AccountsStub{}

	adr1 := mock.NewAddressMock([]byte{65})
	adr2 := mock.NewAddressMock([]byte{67})

	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)
	acnt2 := mock.NewJournalizedAccountWrapMock(adr1)

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if addressContainer == adr1 {
			return acnt1, nil
		}

		if addressContainer == adr2 {
			return acnt2, nil
		}

		return nil, errors.New("failure")
	}

	execTx, _ := exTransaction.NewExecTransaction(
		&accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.Equal(t, acnt2, a2)
}

//------- callSCHandler

func TestExecTransaction_NoCallSChandlerShouldErr(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	err := execTx.CallSChandler(nil)
	assert.Equal(t, execution.ErrNoVM, err)
}

func TestExecTransaction_WithCallSChandlerShouldWork(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	wasCalled := false
	errOutput := errors.New("not really error, just checking output")
	execTx.SetSChandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		wasCalled = true
		return errOutput
	})

	err := execTx.CallSChandler(nil)
	assert.Equal(t, errOutput, err)
	assert.True(t, wasCalled)
}

//------- checkTxValues

func TestExecTransaction_CheckTxValuesHigherNonceShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acnt1.BaseAccount().Nonce = 6

	err := execTx.CheckTxValues(acnt1, big.NewInt(0), 7)
	assert.Equal(t, execution.ErrHigherNonceInTransaction, err)
}

func TestExecTransaction_CheckTxValuesLowerNonceShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acnt1.BaseAccount().Nonce = 6

	err := execTx.CheckTxValues(acnt1, big.NewInt(0), 5)
	assert.Equal(t, execution.ErrLowerNonceInTransaction, err)
}

func TestExecTransaction_CheckTxValuesInsufficientFundsShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acnt1.BaseAccount().Balance = *big.NewInt(67)

	err := execTx.CheckTxValues(acnt1, big.NewInt(68), 0)
	assert.Equal(t, execution.ErrInsufficientFunds, err)
}

func TestExecTransaction_CheckTxValuesOkValsShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acnt1.BaseAccount().Balance = *big.NewInt(67)

	err := execTx.CheckTxValues(acnt1, big.NewInt(67), 0)
	assert.Nil(t, err)
}

//------- moveBalances

func TestExecTransaction_MoveBalancesFailureAcnt1ShouldErr(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDest := mock.NewAddressMock([]byte{67})
	acntDest := mock.NewJournalizedAccountWrapMock(adrDest)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acntSrc.Fail = true

	err := execTx.MoveBalances(acntSrc, acntDest, big.NewInt(0))
	assert.NotNil(t, err)
}

func TestExecTransaction_MoveBalancesFailureAcnt2ShouldErr(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDest := mock.NewAddressMock([]byte{67})
	acntDest := mock.NewJournalizedAccountWrapMock(adrDest)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acntDest.Fail = true

	err := execTx.MoveBalances(acntSrc, acntDest, big.NewInt(0))
	assert.NotNil(t, err)
}

func TestExecTransaction_MoveBalancesOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDest := mock.NewAddressMock([]byte{67})
	acntDest := mock.NewJournalizedAccountWrapMock(adrDest)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acntSrc.Balance = *big.NewInt(64)
	acntDest.Balance = *big.NewInt(31)

	err := execTx.MoveBalances(acntSrc, acntDest, big.NewInt(14))
	assert.Nil(t, err)
	assert.Equal(t, *big.NewInt(50), acntSrc.Balance)
	assert.Equal(t, *big.NewInt(45), acntDest.Balance)

}

//------- increaseNonceAcntSrc

func TestExecTransaction_IncreaseNonceOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	acntSrc.Nonce = 45

	err := execTx.IncreaseNonceAcntSrc(acntSrc)
	assert.Nil(t, err)
	assert.Equal(t, uint64(46), acntSrc.Nonce)
}

//------- ProcessTransaction

func TestExecTransaction_ProcessTransactionNilTxShouldErr(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(
		&mock.AccountsStub{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	err := execTx.ProcessTransaction(nil, 4)
	assert.Equal(t, execution.ErrNilTransaction, err)
}

func TestExecTransaction_ProcessTransactionErrAddressConvShouldErr(t *testing.T) {
	addressConv := &mock.AddressConverterMock{}

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, addressConv, &mock.MarshalizerMock{})

	addressConv.Fail = true

	err := execTx.ProcessTransaction(&transaction.Transaction{}, 4)
	assert.NotNil(t, err)
}

func TestExecTransaction_ProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 45

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx, 4)
	assert.NotNil(t, err)
}

func TestExecTransaction_ProcessTransactionScTxShouldWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	wasCalled := false
	execTx.SetSChandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		wasCalled = true
		return nil
	})

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 45

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))
	acntDest.SetCode([]byte{65})

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestExecTransaction_ProcessTransactionRegisterTxShouldWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	rd := state.RegistrationData{
		OriginatorPubKey: []byte("a"),
		NodePubKey:       []byte("b"),
		RoundIndex:       6,
		Action:           state.ArUnregister,
		Stake:            *big.NewInt(45),
	}

	marshalizer := mock.MarshalizerMock{}
	buff, err := marshalizer.Marshal(&rd)
	assert.Nil(t, err)

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = state.RegistrationAddress.Bytes()
	tx.Value = 0
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

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), state.RegistrationAddress.Bytes()) {
			return acntReg, nil
		}

		return nil, errors.New("failure")
	}

	err = execTx.ProcessTransaction(&tx, 1)
	assert.Nil(t, err)
	assert.True(t, wasCalledAppend)
	assert.Equal(t, *big.NewInt(45), data2.Stake)
	assert.Equal(t, []byte("SRC"), data2.OriginatorPubKey)
	assert.Equal(t, []byte("b"), data2.NodePubKey)
	assert.Equal(t, int32(1), data2.RoundIndex)
	assert.Equal(t, state.ArUnregister, data2.Action)

}

func TestExecTransaction_ProcessCheckNotPassShouldErr(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 45

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Equal(t, execution.ErrHigherNonceInTransaction, err)
}

func TestExecTransaction_ProcessMoveBalancesFailShouldErr(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 0

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Fail = true
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx, 4)
	assert.NotNil(t, err)
}

func TestExecTransaction_ProcessOkValsShouldWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(
		accounts,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		&mock.MarshalizerMock{},
	)

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 4
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 61

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr))
	acntSrc.Nonce = 4
	acntSrc.Balance = *big.NewInt(90)
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr))
	acntDest.Balance = *big.NewInt(10)

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx, 4)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.Nonce)
	assert.Equal(t, *big.NewInt(29), acntSrc.Balance)
	assert.Equal(t, *big.NewInt(71), acntDest.Balance)
}
