package state

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type accountsEphemeralProvider struct {
	accounts AccountsAdapter
	// Not concurrency-safe, but should never be accessed concurrently.
	cache map[string]UserAccountHandler
}

// NewAccountsEphemeralProvider creates a new "ephemeral" provider for accounts.
// Make sure such a provider isn't reused among multiple transactions selections or multiple processing (of blocks) phases,
// since it contains a never-invalidating cache (deliberate, by design).
// Create it, use it, forget it. Don't keep references to it.
func NewAccountsEphemeralProvider(accounts AccountsAdapter) (*accountsEphemeralProvider, error) {
	if check.IfNil(accounts) {
		return nil, ErrNilAccountsAdapter
	}

	return &accountsEphemeralProvider{
		accounts: accounts,
		cache:    make(map[string]UserAccountHandler),
	}, nil
}

// GetRootHash returns the current root hash
func (provider *accountsEphemeralProvider) GetRootHash() ([]byte, error) {
	return provider.accounts.RootHash()
}

// GetAccountNonceAndBalance returns the nonce of the account, the balance of the account, and whether it's currently existing on-chain.
func (provider *accountsEphemeralProvider) GetAccountNonceAndBalance(address []byte) (uint64, *big.Int, bool, error) {
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

// GetUserAccount returns the user account, as found on blockchain. If missing (account not found), nil is returned (with no error).
func (provider *accountsEphemeralProvider) GetUserAccount(address []byte) (UserAccountHandler, error) {
	account, ok := provider.cache[string(address)]
	if ok {
		// Existing or new (unknown) account, previously-cached.
		return account, nil
	}

	account, err := provider.getExistingAccountTypedAsUserAccount(address)
	if err != nil && err != ErrAccNotFound {
		// Unexpected failure (error different from "ErrAccNotFound").
		// Account won't be cached.
		return nil, err
	}

	// Existing account or new (unknown), we'll cache it (actual object or nil).
	provider.cache[string(address)] = account

	// Generally speaking, this isn't a good pattern: returning both nil (for unknown accounts), and a nil error.
	// However, this is a non-exported method, which should only be called with care, within this struct only.
	return account, nil
}

func (provider *accountsEphemeralProvider) getExistingAccountTypedAsUserAccount(address []byte) (UserAccountHandler, error) {
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
func (provider *accountsEphemeralProvider) IsInterfaceNil() bool {
	return provider == nil
}
