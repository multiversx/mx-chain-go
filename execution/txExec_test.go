package execution

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestTxExec_ProcessTransaction_InvalidParams_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	//invalid accounts
	ret := te.ProcessTransaction(nil, &transaction.Transaction{})
	assert.Equal(t, TxInvalidParameters, ret.Code())

	//invalid transaction
	ret = te.ProcessTransaction(mock.NewAccountsDBMock(), nil)
	assert.Equal(t, TxInvalidParameters, ret.Code())
}

func TestTxExec_ProcessTransaction_SCDataNoHandler_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	txSC := transaction.Transaction{Data: make([]byte, 1)}

	ret := te.ProcessTransaction(mock.NewAccountsDBMock(), &txSC)
	assert.Equal(t, TxNoSChandler, ret.Code())
}

func TestTxExec_ProcessTransaction_SCDataWithHandler_ShouldWork(t *testing.T) {
	te := NewTxExec()
	te.SetSChandler(func(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary {
		return NewExecSummary(-1, errors.New("OK actually"))
	})

	assert.NotNil(t, te.SChandler())

	txSC := transaction.Transaction{Data: make([]byte, 1)}

	ret := te.ProcessTransaction(mock.NewAccountsDBMock(), &txSC)
	assert.Equal(t, ExecCode(-1), ret.Code())
	fmt.Println(ret.Error())
}

func generateAddress() *state.Address {
	buff := make([]byte, state.AdrLen)

	rand.Read(buff)

	adr := state.Address{}
	adr.SetBytes(buff)

	return &adr
}

func TestTxExec_ProcessTransaction_OKVals_ShouldWork(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 0
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxSuccess, ret.Code())

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(900), acntSnd.Balance)
	assert.Equal(t, big.NewInt(100), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_HigherNonce_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 1
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxHigherNonce, ret.Code())

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(1000), acntSnd.Balance)
	assert.Equal(t, big.NewInt(0), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_LowerNonce_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	acntSnd.Nonce = 10
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 1
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxLowerNonce, ret.Code())

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(1000), acntSnd.Balance)
	assert.Equal(t, big.NewInt(0), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_InsufficientFunds_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	acntSnd.Nonce = 10
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 1001
	tx.Nonce = 10
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxInsufficentFunds, ret.Code())

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(1000), acntSnd.Balance)
	assert.Equal(t, big.NewInt(0), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_OKValsAccountsFailGet1_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	acntSnd.Nonce = 10
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 10
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	accounts.FailGetOrCreateAccount = true
	accounts.NoToFailureGetOrCreateAccount = 0

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxAccountsError, ret.Code())

	accounts.FailGetOrCreateAccount = false

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(1000), acntSnd.Balance)
	assert.Equal(t, big.NewInt(0), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_OKValsAccountsFailGet2_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	acntSnd.Nonce = 10
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 10
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	accounts.FailGetOrCreateAccount = true
	accounts.NoToFailureGetOrCreateAccount = 1

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxAccountsError, ret.Code())

	accounts.FailGetOrCreateAccount = false

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(1000), acntSnd.Balance)
	assert.Equal(t, big.NewInt(0), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_OKValsAccountsFailSave1_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	acntSnd.Nonce = 10
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 10
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	accounts.FailSaveAccountState = true
	accounts.NoToFailureSaveAccountState = 1

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxAccountsError, ret.Code())

	accounts.FailSaveAccountState = false

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

	assert.Equal(t, big.NewInt(1000), acntSnd.Balance)
	assert.Equal(t, big.NewInt(0), acntRcv.Balance)
}

func TestTxExec_ProcessTransaction_OKValsAccountsFailSave2_ShouldRetErr(t *testing.T) {
	te := NewTxExec()

	adrSnd := generateAddress()
	fmt.Printf("Generated sender address: %v\n", adrSnd.Hex(mock.HasherMock{}))

	adrRcv := generateAddress()
	fmt.Printf("Generated receiver address: %v\n", adrRcv.Hex(mock.HasherMock{}))

	accounts := mock.NewAccountsDBMock()

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated sender account: %v\n", acntSnd)

	acntRcv, err := accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Generated receiver account: %v\n", acntRcv)

	//minting sender account
	acntSnd.Balance = big.NewInt(1000)
	acntSnd.Nonce = 10
	accounts.SaveAccountState(acntSnd)

	//generating transaction
	tx := transaction.Transaction{}

	tx.Value = 100
	tx.Nonce = 10
	tx.SndAddr = adrSnd.Bytes()
	tx.RcvAddr = adrRcv.Bytes()

	accounts.FailSaveAccountState = true
	accounts.NoToFailureSaveAccountState = 2

	ret := te.ProcessTransaction(accounts, &tx)
	assert.Equal(t, TxAccountsError, ret.Code())

	accounts.FailSaveAccountState = false

	acntSnd, err = accounts.GetOrCreateAccount(*adrSnd)
	assert.Nil(t, err)
	fmt.Printf("Sender account: %v\n", acntSnd)

	acntRcv, err = accounts.GetOrCreateAccount(*adrRcv)
	assert.Nil(t, err)
	fmt.Printf("Receiver account: %v\n", acntRcv)

}
