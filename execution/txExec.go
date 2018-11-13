package execution

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/pkg/errors"
)

const (
	// TxSuccess signals that the transaction has been executed correctly
	TxSuccess ExecCode = iota
	// TxNoSChandler signals that the transaction can not be executed because is a SC call and no SC handler has been set
	TxNoSChandler
	// TxInvalidParameters signals that the parameters are invalid
	TxInvalidParameters
	// TxLowerNonce signals that the transaction's nonce is lower, transaction can be discarded
	TxLowerNonce
	// TxHigherNonce signals that the transaction's nonce is higher, transaction can not be executed right now,
	// maybe in the future
	TxHigherNonce
	// TxInsufficentFunds signals that there are not enough funds to be transferred
	TxInsufficentFunds
	// TxAccountsError signals a general failure in accounts DB. Process should return immediately and restore
	// the DB trie as there might be inconsistencies
	TxAccountsError
)

// TxExec implements TransactionExecutor interface and can modify account states according to a transaction
type TxExec struct {
	scHandler func(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary
}

// NewTxExec creates a new TxExec engine
func NewTxExec() *TxExec {
	te := TxExec{}
	return &te
}

// SChandler returns the smart contract execution function
func (te *TxExec) SChandler() func(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary {
	return te.scHandler
}

// SetSChandler sets the smart contract execution function
func (te *TxExec) SetSChandler(f func(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary) {
	te.scHandler = f
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (te *TxExec) ProcessTransaction(accounts state.AccountsHandler, transaction *transaction.Transaction) *ExecSummary {
	if accounts == nil {
		return NewExecSummary(TxInvalidParameters, errors.New("nil accounts"))
	}

	if transaction == nil {
		return NewExecSummary(TxInvalidParameters, errors.New("nil transaction"))
	}

	if len(transaction.Data) > 0 {
		if te.scHandler == nil {
			return NewExecSummary(TxNoSChandler, errors.New("no VM wired"))
		} else {
			return te.scHandler(accounts, transaction)
		}
	}

	adrRcv := te.getAddress(transaction.RcvAddr)
	adrSnd := te.getAddress(transaction.SndAddr)

	acntRcv, err := accounts.GetOrCreateAccount(*adrRcv)
	if err != nil {
		return NewExecSummary(TxAccountsError, err)
	}

	acntSnd, err := accounts.GetOrCreateAccount(*adrSnd)
	if err != nil {
		return NewExecSummary(TxAccountsError, err)
	}

	//TODO change to big int implementation
	txValue := big.NewInt(0)
	txValue.SetUint64(transaction.Value)

	es := te.checkMoveBalanceTx(acntSnd, txValue, transaction)

	if es.Code() != TxSuccess {
		return es
	}

	acntSnd.Balance = acntSnd.Balance.Sub(acntSnd.Balance, txValue)
	acntRcv.Balance = acntRcv.Balance.Add(acntRcv.Balance, txValue)

	//save data
	err = accounts.SaveAccountState(acntSnd)
	if err != nil {
		return NewExecSummary(TxAccountsError, err)
	}
	err = accounts.SaveAccountState(acntRcv)
	if err != nil {
		return NewExecSummary(TxAccountsError, err)
	}

	return NewExecSummary(TxSuccess, nil)
}

func (te *TxExec) getAddress(buff []byte) *state.Address {
	adr := state.Address{}
	adr.SetBytes(buff)
	return &adr
}

func (te *TxExec) checkMoveBalanceTx(acntSnd *state.AccountState, txValue *big.Int, transaction *transaction.Transaction) *ExecSummary {

	if acntSnd.Nonce < transaction.Nonce {
		return NewExecSummary(TxHigherNonce, errors.New("higher Nonce in transaction: wait"))
	}

	if acntSnd.Nonce > transaction.Nonce {
		return NewExecSummary(TxLowerNonce, errors.New("lower Nonce in transaction: dropping tx"))
	}

	if acntSnd.Balance.Cmp(txValue) < 0 {
		return NewExecSummary(TxInsufficentFunds, errors.New("insufficient funds"))
	}

	return NewExecSummary(TxSuccess, nil)
}
