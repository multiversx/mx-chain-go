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

type accountStateProvider struct {
	accountsAdapter       state.AccountsAdapter
	transactionsProcessor process.TransactionProcessor
}

func newAccountStateProvider(accountsAdapter state.AccountsAdapter, transactionsProcessor process.TransactionProcessor) (*accountStateProvider, error) {
	if check.IfNil(accountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(transactionsProcessor) {
		return nil, process.ErrNilTxProcessor
	}

	return &accountStateProvider{
		accountsAdapter:       accountsAdapter,
		transactionsProcessor: transactionsProcessor,
	}, nil
}

// GetAccountState returns the state of an account.
// Will be called by mempool during transaction selection.
func (provider *accountStateProvider) GetAccountState(address []byte) (*txcache.AccountState, error) {
	account, err := provider.accountsAdapter.GetExistingAccount(address)
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

func (provider *accountStateProvider) IsBadlyGuarded(tx data.TransactionHandler) bool {
	address := tx.GetSndAddr()
	account, err := provider.accountsAdapter.GetExistingAccount(address)
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

	err = provider.transactionsProcessor.VerifyGuardian(txTyped, userAccount)
	return errors.Is(err, process.ErrTransactionNotExecutable)
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *accountStateProvider) IsInterfaceNil() bool {
	return provider == nil
}
