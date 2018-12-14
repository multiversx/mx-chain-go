package exTransaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/exTransaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- NewExecTransaction

func TestNewExecTransactionNilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(nil, mock.HasherMock{}, &mock.AddressConverterMock{})
	assert.Equal(t, execution.ErrNilAccountsAdapter, err)
}

func TestNewExecTransactionNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(&mock.AccountsStub{}, nil, &mock.AddressConverterMock{})
	assert.Equal(t, execution.ErrNilHasher, err)
}

func TestNewExecTransactionNilAddressConverterMockShouldErr(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, nil)
	assert.Equal(t, execution.ErrNilAddressConverter, err)
}

func TestNewExecTransactionOkValsShouldWork(t *testing.T) {
	t.Parallel()

	_, err := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})
	assert.Nil(t, err)
}

//------- SChandler

func TestExecTransactionGetSetSChandlerShouldWork(t *testing.T) {
	t.Parallel()

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	f := func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
		return nil
	}

	execTx.SetSChandler(f)
	assert.NotNil(t, execTx.SChandler())
}

//------- RegisterHandler

func TestExecTransactionGetSetRegisterHandlerShouldWork(t *testing.T) {
	t.Parallel()

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	f := func(data []byte) error {
		return nil
	}

	execTx.SetRegisterHandler(f)
	assert.NotNil(t, execTx.RegisterHandler())
}

//------- getAddresses

func TestExecTransactionGetAddressErrAddressConvShouldErr(t *testing.T) {
	t.Parallel()

	addressConv := &mock.AddressConverterMock{}

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, addressConv)

	addressConv.Fail = true

	tx := transaction.Transaction{}

	_, _, err := execTx.GetAddresses(&tx)
	assert.NotNil(t, err)
}

func TestExecTransactionGetAddressOkValsShouldWork(t *testing.T) {
	t.Parallel()

	addressConv := &mock.AddressConverterMock{}

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, addressConv)

	tx := transaction.Transaction{}
	tx.RcvAddr = []byte{65, 66, 67}
	tx.SndAddr = []byte{32, 33, 34}

	adrSnd, adrRcv, err := execTx.GetAddresses(&tx)
	assert.Nil(t, err)
	assert.Equal(t, []byte{65, 66, 67}, adrRcv.Bytes())
	assert.Equal(t, []byte{32, 33, 34}, adrSnd.Bytes())
}

//------- getAccounts

func TestExecTransactionGetAccountsMalfunctionAccountsShouldErr(t *testing.T) {
	accounts := mock.AccountsStub{}
	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		return nil, errors.New("failure")
	}

	execTx, _ := exTransaction.NewExecTransaction(&accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	adr1 := mock.NewAddressMock([]byte{65}, []byte{66})
	adr2 := mock.NewAddressMock([]byte{67}, []byte{68})

	_, _, err := execTx.GetAccounts(adr1, adr2)
	assert.NotNil(t, err)
}

func TestExecTransactionGetAccountsOkValsShouldWork(t *testing.T) {
	accounts := mock.AccountsStub{}

	adr1 := mock.NewAddressMock([]byte{65}, []byte{66})
	adr2 := mock.NewAddressMock([]byte{67}, []byte{68})

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

	execTx, _ := exTransaction.NewExecTransaction(&accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	a1, a2, err := execTx.GetAccounts(adr1, adr2)
	assert.Nil(t, err)
	assert.Equal(t, acnt1, a1)
	assert.Equal(t, acnt2, a2)
}

//------- callSChandler

func TestExecTransactionNoCallSChandlerShouldErr(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	err := execTx.CallSChandler(nil)
	assert.Equal(t, execution.ErrNoVM, err)
}

func TestExecTransactionWithCallSChandlerShouldWork(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

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

//------- callRegisterHandler

func TestExecTransactionNoCallRegisterHandlerShouldErr(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	err := execTx.CallRegisterHandler(nil)
	assert.Equal(t, execution.ErrRegisterFunctionUndefined, err)
}

func TestExecTransactionWithCallRegisterHandlerShouldWork(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	wasCalled := false
	errOutput := errors.New("not really error, just checking output")
	execTx.SetRegisterHandler(func(data []byte) error {
		wasCalled = true
		return errOutput
	})

	err := execTx.CallRegisterHandler(nil)
	assert.Equal(t, errOutput, err)
	assert.True(t, wasCalled)
}

//------- checkTxValues

func TestExecTransactionCheckTxValuesHigherNonceShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65}, []byte{66})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acnt1.BaseAccount().Nonce = 6

	err := execTx.CheckTxValues(acnt1, big.NewInt(0), 7)
	assert.Equal(t, execution.ErrHigherNonceInTransaction, err)
}

func TestExecTransactionCheckTxValuesLowerNonceShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65}, []byte{66})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acnt1.BaseAccount().Nonce = 6

	err := execTx.CheckTxValues(acnt1, big.NewInt(0), 5)
	assert.Equal(t, execution.ErrLowerNonceInTransaction, err)
}

func TestExecTransactionCheckTxValuesInsufficientFundsShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65}, []byte{66})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acnt1.BaseAccount().Balance = *big.NewInt(67)

	err := execTx.CheckTxValues(acnt1, big.NewInt(68), 0)
	assert.Equal(t, execution.ErrInsufficientFunds, err)
}

func TestExecTransactionCheckTxValuesOkValsShouldErr(t *testing.T) {
	adr1 := mock.NewAddressMock([]byte{65}, []byte{66})
	acnt1 := mock.NewJournalizedAccountWrapMock(adr1)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acnt1.BaseAccount().Balance = *big.NewInt(67)

	err := execTx.CheckTxValues(acnt1, big.NewInt(67), 0)
	assert.Nil(t, err)
}

//------- moveBalances

func TestExecTransactionMoveBalancesFailureAcnt1ShouldErr(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65}, []byte{66})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDest := mock.NewAddressMock([]byte{67}, []byte{68})
	acntDest := mock.NewJournalizedAccountWrapMock(adrDest)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acntSrc.Fail = true

	err := execTx.MoveBalances(acntSrc, acntDest, big.NewInt(0))
	assert.NotNil(t, err)
}

func TestExecTransactionMoveBalancesFailureAcnt2ShouldErr(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65}, []byte{66})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDest := mock.NewAddressMock([]byte{67}, []byte{68})
	acntDest := mock.NewJournalizedAccountWrapMock(adrDest)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acntDest.Fail = true

	err := execTx.MoveBalances(acntSrc, acntDest, big.NewInt(0))
	assert.NotNil(t, err)
}

func TestExecTransactionMoveBalancesOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65}, []byte{66})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	adrDest := mock.NewAddressMock([]byte{67}, []byte{68})
	acntDest := mock.NewJournalizedAccountWrapMock(adrDest)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acntSrc.Balance = *big.NewInt(64)
	acntDest.Balance = *big.NewInt(31)

	err := execTx.MoveBalances(acntSrc, acntDest, big.NewInt(14))
	assert.Nil(t, err)
	assert.Equal(t, *big.NewInt(50), acntSrc.Balance)
	assert.Equal(t, *big.NewInt(45), acntDest.Balance)

}

//------- increaseNonceAcntSrc

func TestExecTransactionIncreaseNonceOkValsShouldWork(t *testing.T) {
	adrSrc := mock.NewAddressMock([]byte{65}, []byte{66})
	acntSrc := mock.NewJournalizedAccountWrapMock(adrSrc)

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	acntSrc.Nonce = 45

	err := execTx.IncreaseNonceAcntSrc(acntSrc)
	assert.Nil(t, err)
	assert.Equal(t, uint64(46), acntSrc.Nonce)
}

//------- ProcessTransaction

func TestExecTransactionProcessTransactionNilTxShouldErr(t *testing.T) {
	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, &mock.AddressConverterMock{})

	err := execTx.ProcessTransaction(nil)
	assert.Equal(t, execution.ErrNilTransaction, err)
}

func TestExecTransactionProcessTransactionErrAddressConvShouldErr(t *testing.T) {
	addressConv := &mock.AddressConverterMock{}

	execTx, _ := exTransaction.NewExecTransaction(&mock.AccountsStub{}, mock.HasherMock{}, addressConv)

	addressConv.Fail = true

	err := execTx.ProcessTransaction(&transaction.Transaction{})
	assert.NotNil(t, err)
}

func TestExecTransactionProcessTransactionMalfunctionAccountsShouldErr(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 45

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx)
	assert.NotNil(t, err)
}

func TestExecTransactionProcessTransactionScTxShouldWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

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

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr, nil))
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr, nil))
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

	err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestExecTransactionProcessTransactionRegisterTxShouldWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	wasCalled := false
	execTx.SetRegisterHandler(func(data []byte) error {
		wasCalled = true
		return nil
	})

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte(exTransaction.RegisterAddress)
	tx.Value = 0
	tx.Data = []byte("REGISTER ME")

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr, nil))
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr, nil))

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestExecTransactionProcessTransactionRegisterTxShouldNotWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	wasCalled := false
	execTx.SetRegisterHandler(func(data []byte) error {
		wasCalled = true
		return errors.New("register failed")
	})

	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte(exTransaction.RegisterAddress)
	tx.Value = 0
	tx.Data = []byte("REGISTER ME")

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr, nil))
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr, nil))

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx)
	assert.NotNil(t, err)
	assert.True(t, wasCalled)
}

func TestExecTransactionProcessCheckNotPassShouldErr(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 45

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr, nil))
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr, nil))

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx)
	assert.Equal(t, execution.ErrHigherNonceInTransaction, err)
}

func TestExecTransactionProcessMoveBalancesFailShouldErr(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 0

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr, nil))
	acntSrc.Fail = true
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr, nil))

	accounts.GetJournalizedAccountCalled = func(addressContainer state.AddressContainer) (state.JournalizedAccountWrapper, error) {
		if bytes.Equal(addressContainer.Bytes(), tx.SndAddr) {
			return acntSrc, nil
		}

		if bytes.Equal(addressContainer.Bytes(), tx.RcvAddr) {
			return acntDest, nil
		}

		return nil, errors.New("failure")
	}

	err := execTx.ProcessTransaction(&tx)
	assert.NotNil(t, err)
}

func TestExecTransactionProcessOkValsShouldWork(t *testing.T) {
	accounts := &mock.AccountsStub{}

	execTx, _ := exTransaction.NewExecTransaction(accounts, mock.HasherMock{}, &mock.AddressConverterMock{})

	//these values will trigger ErrHigherNonceInTransaction
	tx := transaction.Transaction{}
	tx.Nonce = 4
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DEST")
	tx.Value = 61

	acntSrc := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.SndAddr, nil))
	acntSrc.Nonce = 4
	acntSrc.Balance = *big.NewInt(90)
	acntDest := mock.NewJournalizedAccountWrapMock(mock.NewAddressMock(tx.RcvAddr, nil))
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

	err := execTx.ProcessTransaction(&tx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), acntSrc.Nonce)
	assert.Equal(t, *big.NewInt(29), acntSrc.Balance)
	assert.Equal(t, *big.NewInt(71), acntDest.Balance)
}
