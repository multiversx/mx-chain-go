package preprocess

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
)

type accountNonceProvider struct {
	accountsAdapter state.AccountsAdapter
	mutex           sync.RWMutex
}

func newAccountNonceProvider(accountsAdapter state.AccountsAdapter) (*accountNonceProvider, error) {
	if check.IfNil(accountsAdapter) {
		return nil, errors.ErrNilAccountsAdapter
	}

	return &accountNonceProvider{
		accountsAdapter: accountsAdapter,
	}, nil
}

// GetAccountNonce returns the nonce for an account.
// Will be called by mempool during transaction selection.
func (provider *accountNonceProvider) GetAccountNonce(address []byte) (uint64, error) {
	account, err := provider.accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return 0, err
	}

	return account.GetNonce(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *accountNonceProvider) IsInterfaceNil() bool {
	return provider == nil
}
