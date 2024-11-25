package preprocess

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

type selectionSession struct {
	accountsAdapter       state.AccountsAdapter
	transactionsProcessor process.TransactionProcessor
}

func newSelectionSession(accountsAdapter state.AccountsAdapter, transactionsProcessor process.TransactionProcessor) (*selectionSession, error) {
	if check.IfNil(accountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(transactionsProcessor) {
		return nil, process.ErrNilTxProcessor
	}

	return &selectionSession{
		accountsAdapter:       accountsAdapter,
		transactionsProcessor: transactionsProcessor,
	}, nil
}

// GetAccountState returns the state of an account.
// Will be called by mempool during transaction selection.
func (session *selectionSession) GetAccountState(address []byte) (*txcache.AccountState, error) {
	account, err := session.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return &txcache.AccountState{
		Nonce:   userAccount.GetNonce(),
		Balance: userAccount.GetBalance(),
	}, nil
}

// IsBadlyGuarded checks if a transaction is badly guarded (not executable).
// Will be called by mempool during transaction selection.
func (session *selectionSession) IsBadlyGuarded(tx data.TransactionHandler) bool {
	address := tx.GetSndAddr()
	account, err := session.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return false
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return false
	}

	txTyped, ok := tx.(*transaction.Transaction)
	if !ok {
		return false
	}

	err = session.transactionsProcessor.VerifyGuardian(txTyped, userAccount)
	return errors.Is(err, process.ErrTransactionNotExecutable)
}

// IsInterfaceNil returns true if there is no value under the interface
func (session *selectionSession) IsInterfaceNil() bool {
	return session == nil
}
