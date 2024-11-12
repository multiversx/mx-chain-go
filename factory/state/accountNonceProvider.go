package state

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

// NewAccountNonceProvider creates a new instance of accountNonceProvider.
// When the accounts adapter is not yet available, client code is allowed to pass a nil reference in the constructor.
// In that case, the accounts adapter should be set at a later time, by means of SetAccountsAdapter.
func NewAccountNonceProvider(accountsAdapter state.AccountsAdapter) (*accountNonceProvider, error) {
	return &accountNonceProvider{
		accountsAdapter: accountsAdapter,
	}, nil
}

// SetAccountsAdapter sets the accounts adapter
func (provider *accountNonceProvider) SetAccountsAdapter(accountsAdapter state.AccountsAdapter) error {
	if check.IfNil(accountsAdapter) {
		return errors.ErrNilAccountsAdapter
	}

	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.accountsAdapter = accountsAdapter
	return nil
}

// GetAccountNonce returns the nonce for an account.
// Will be called by "shardedTxPool" on every transaction added to the pool.
func (provider *accountNonceProvider) GetAccountNonce(address []byte) (uint64, error) {
	provider.mutex.RLock()
	accountsAdapter := provider.accountsAdapter
	provider.mutex.RUnlock()

	// No need for double check locking here (we are just guarding against a programming mistake, not against a specific runtime condition).
	if check.IfNil(accountsAdapter) {
		return 0, errors.ErrNilAccountsAdapter
	}

	account, err := accountsAdapter.GetExistingAccount(address)
	if err != nil {
		return 0, err
	}

	return account.GetNonce(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *accountNonceProvider) IsInterfaceNil() bool {
	return provider == nil
}
