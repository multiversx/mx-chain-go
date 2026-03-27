package state

import (
	"errors"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// AccountsEphemeralProvider is an ephemeral, never-invalidating cache over AccountsAdapter.
// Thread-safe: mutCache allows concurrent reads during parallel txpool selection.
// Usage contract: create privately per selection/processing phase, never reuse across phases
// (cache never invalidates — stale nonces will cause incorrect tx selection).
type AccountsEphemeralProvider struct {
	accounts AccountsAdapter
	mutCache sync.RWMutex
	cache    map[string]UserAccountHandler
}

// NewAccountsEphemeralProvider creates a new "ephemeral" provider for accounts.
func NewAccountsEphemeralProvider(accounts AccountsAdapter) (*AccountsEphemeralProvider, error) {
	if check.IfNil(accounts) {
		return nil, ErrNilAccountsAdapter
	}

	return &AccountsEphemeralProvider{
		accounts: accounts,
		cache:    make(map[string]UserAccountHandler),
	}, nil
}

// GetRootHash returns the current root hash
func (provider *AccountsEphemeralProvider) GetRootHash() ([]byte, error) {
	return provider.accounts.RootHash()
}

// GetAccountNonce returns the nonce of the account, and whether it's currently existing on-chain.
func (provider *AccountsEphemeralProvider) GetAccountNonce(address []byte) (uint64, bool, error) {
	account, err := provider.GetUserAccount(address)
	if err != nil {
		// Unexpected failure.
		return 0, false, err
	}
	if check.IfNil(account) {
		// New (unknown) account.
		return 0, false, nil
	}

	return account.GetNonce(), true, nil
}

// GetAccountNonceAndBalance returns the nonce of the account, the balance of the account, and whether it's currently existing on-chain.
func (provider *AccountsEphemeralProvider) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
	account, err := provider.GetUserAccount(address)
	if err != nil {
		// Unexpected failure.
		return 0, nil, false, err
	}
	if check.IfNil(account) {
		// New (unknown) account.
		return 0, big.NewInt(0), false, nil
	}

	return account.GetNonce(), account.GetBalance(), true, nil
}

// GetUserAccount returns the user account. Thread-safe via RWMutex on cache.
func (provider *AccountsEphemeralProvider) GetUserAccount(address []byte) (UserAccountHandler, error) {
	addrKey := string(address)

	provider.mutCache.RLock()
	account, ok := provider.cache[addrKey]
	provider.mutCache.RUnlock()
	if ok {
		return account, nil
	}

	provider.mutCache.Lock()
	// Re-check under write lock to avoid duplicate trie reads
	account, ok = provider.cache[addrKey]
	if ok {
		provider.mutCache.Unlock()
		return account, nil
	}

	fetched, err := provider.getExistingAccountTypedAsUserAccount(address)

	var errAccountNotFoundAtBlock *ErrAccountNotFoundAtBlock
	isAccountNotFoundError := errors.Is(err, ErrAccNotFound) || errors.As(err, &errAccountNotFoundAtBlock)

	if err != nil && !isAccountNotFoundError {
		provider.mutCache.Unlock()
		return nil, err
	}

	provider.cache[addrKey] = fetched
	provider.mutCache.Unlock()

	return fetched, nil
}

func (provider *AccountsEphemeralProvider) getExistingAccountTypedAsUserAccount(address []byte) (UserAccountHandler, error) {
	account, err := provider.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(UserAccountHandler)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return userAccount, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *AccountsEphemeralProvider) IsInterfaceNil() bool {
	return provider == nil
}
