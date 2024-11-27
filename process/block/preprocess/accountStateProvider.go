package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

type accountStateProvider struct {
	accountsAdapter state.AccountsAdapter
}

func newAccountStateProvider(accountsAdapter state.AccountsAdapter) (*accountStateProvider, error) {
	if check.IfNil(accountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}

	return &accountStateProvider{
		accountsAdapter: accountsAdapter,
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
		return nil, errors.ErrWrongTypeAssertion
	}

	return &txcache.AccountState{
		Nonce:   userAccount.GetNonce(),
		Balance: userAccount.GetBalance(),
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *accountStateProvider) IsInterfaceNil() bool {
	return provider == nil
}
