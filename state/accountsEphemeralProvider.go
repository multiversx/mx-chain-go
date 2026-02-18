package state

import (
	"errors"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
)

// AccountsEphemeralProvider acts as an "ephemeral" provider for accounts.
// Make sure such a provider isn't reused among multiple transactions selections or multiple processing (of blocks) phases,
// since it contains a **never-invalidating cache** (deliberate, by design).
// Create it privately (don't receive it in constructors), use it privately, make sure it's forgotten afterwards. Don't keep lasting references to it.
// Note: this structure is exported on purpose (less ceremonious code where it's being used, no extra interfaces needed).
type AccountsEphemeralProvider struct {
	accounts AccountsAdapter
	// Not concurrency-safe, but should never be accessed concurrently.
	cache map[string]UserAccountHandler
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
	account, err := provider.getAccountForNonceAndBalance(address)
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
	account, err := provider.getAccountForNonceAndBalance(address)
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

// getAccountForNonceAndBalance returns whatever account is cached (including lightweight accounts).
// On cache hit, returns the cached entry as-is (light or full) — no upgrade is triggered.
// On cache miss, fetches a full account from the underlying adapter and caches it.
// Unlike GetUserAccount, this method does NOT upgrade lightAccountInfo to full accounts,
// since nonce, balance, and guarded status are correctly represented by lightweight accounts.
func (provider *AccountsEphemeralProvider) getAccountForNonceAndBalance(address []byte) (UserAccountHandler, error) {
	account, ok := provider.cache[string(address)]
	if ok {
		return account, nil
	}

	// Not in cache — fetch and cache it (same logic as GetUserAccount).
	return provider.fetchAndCacheAccount(address)
}

// GetUserAccount returns the user account, as found on blockchain. If missing (account not found), nil is returned (with no error).
// IMPORTANT: If the cache contains a lightAccountInfo (from PrefetchAccounts), this method will
// re-fetch the full account from the underlying adapter. This ensures callers that need full
// account state (e.g., guardian checks via VerifyGuardian) always get complete account objects.
// For nonce/balance-only access, use GetAccountNonceAndBalance instead (it accepts light accounts).
func (provider *AccountsEphemeralProvider) GetUserAccount(address []byte) (UserAccountHandler, error) {
	account, ok := provider.cache[string(address)]
	if ok {
		// If the cached entry is a lightweight account, upgrade it to a full account.
		// lightAccountInfo only carries nonce, balance, codeMetadata, and rootHash;
		// callers of GetUserAccount may need full account state (e.g., guardian validation).
		if light, isLight := account.(*lightAccountInfo); isLight {
			return provider.upgradeFromLight(address, light)
		}

		// Full account or nil (unknown account), previously-cached.
		return account, nil
	}

	return provider.fetchAndCacheAccount(address)
}

// upgradeFromLight creates a full account from cached light data without re-reading the main trie.
// Falls back to fetchAndCacheAccount if the underlying adapter doesn't support AccountFromLightUpgrader.
func (provider *AccountsEphemeralProvider) upgradeFromLight(address []byte, light *lightAccountInfo) (UserAccountHandler, error) {
	upgrader, ok := provider.accounts.(AccountFromLightUpgrader)
	if !ok {
		return provider.fetchAndCacheAccount(address) // fallback: old path
	}

	account, err := upgrader.CreateFullAccountFromLight(address, light.nonce, light.GetBalance(), light.codeMetadata, light.rootHash)
	if err != nil {
		log.Debug("AccountsEphemeralProvider.upgradeFromLight: CreateFullAccountFromLight failed, falling back to trie read",
			"err", err)
		return provider.fetchAndCacheAccount(address) // fallback: old path
	}

	userAccount, ok := account.(UserAccountHandler)
	if !ok {
		log.Debug("AccountsEphemeralProvider.upgradeFromLight: type assertion to UserAccountHandler failed, falling back to trie read")
		return provider.fetchAndCacheAccount(address)
	}

	provider.cache[string(address)] = userAccount
	return userAccount, nil
}

// fetchAndCacheAccount fetches the account from the underlying adapter and caches it.
func (provider *AccountsEphemeralProvider) fetchAndCacheAccount(address []byte) (UserAccountHandler, error) {
	account, err := provider.getExistingAccountTypedAsUserAccount(address)

	var errAccountNotFoundAtBlock *ErrAccountNotFoundAtBlock
	isAccountNotFoundError := errors.Is(err, ErrAccNotFound) || errors.As(err, &errAccountNotFoundAtBlock)

	if err != nil && !isAccountNotFoundError {
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

// PrefetchAccounts batch-fetches multiple accounts and warms the ephemeral cache.
// This leverages batch trie reads (single lock acquisition) to minimize mutex contention
// when many unique accounts need to be read (e.g., 10k unique senders in a block).
// Addresses already in the cache are skipped. Non-existent accounts are cached as nil.
//
// Prefetch strategy (ordered by preference):
//  1. LightAccountsBatchReader — returns lightweight read-only accounts (nonce + balance only).
//     This is the fastest path: single trie lock, no TrackableDataTrie/DataTrieLeafParser allocation.
//     Used when the caller only needs nonce and balance (proposal phase).
//  2. AccountsBatchReader — returns full UserAccount objects via single trie lock.
//     Slower due to CreateAccount overhead, but still avoids N individual trie lock acquisitions.
//  3. Sequential fallback — individual GetUserAccount calls. Slowest path (N trie locks).
func (provider *AccountsEphemeralProvider) PrefetchAccounts(addresses [][]byte) {
	// Filter out already-cached addresses
	uncachedAddresses := make([][]byte, 0, len(addresses))
	for _, addr := range addresses {
		if _, ok := provider.cache[string(addr)]; !ok {
			uncachedAddresses = append(uncachedAddresses, addr)
		}
	}

	if len(uncachedAddresses) == 0 {
		return
	}

	startBatch := time.Now()

	// Strategy 1: Try lightweight batch read (nonce + balance only, no heavy object creation).
	// This is the optimal path for the proposal phase where only nonce and balance are needed.
	if provider.prefetchLightAccounts(uncachedAddresses, len(addresses), startBatch) {
		provider.upgradeGuardedAccounts()
		return
	}

	// Strategy 2: Try full batch read (single lock, but creates full UserAccount objects).
	if provider.prefetchFullAccounts(uncachedAddresses, len(addresses), startBatch) {
		return
	}

	// Strategy 3: Fallback to sequential reads (N individual trie lock acquisitions).
	// This path is taken if the underlying accounts adapter supports neither batch interface.
	for _, addr := range uncachedAddresses {
		_, _ = provider.GetUserAccount(addr)
	}
}

// prefetchLightAccounts attempts to batch-fetch accounts using the lightweight path
// (LightAccountsBatchReader), which returns read-only accounts with only nonce and balance.
// Returns true if the lightweight path was used (even if it fell back to sequential on error).
func (provider *AccountsEphemeralProvider) prefetchLightAccounts(
	uncachedAddresses [][]byte,
	totalAddresses int,
	startBatch time.Time,
) bool {
	lightReader, ok := provider.accounts.(LightAccountsBatchReader)
	if !ok {
		return false
	}

	lightAccounts, err := lightReader.GetLightAccountsBatch(uncachedAddresses)
	if err != nil {
		log.Warn("AccountsEphemeralProvider.PrefetchAccounts light batch read failed, falling back",
			"err", err,
			"num addresses", len(uncachedAddresses))
		// Don't fall back to sequential here — let the caller try the next strategy.
		return false
	}

	log.Trace("AccountsEphemeralProvider.PrefetchAccounts light batch read completed",
		"total addresses", totalAddresses,
		"uncached addresses", len(uncachedAddresses),
		"duration", time.Since(startBatch))

	// Cache results: lightAccountInfo for existing accounts, nil for non-existent.
	// The lightAccountInfo implements UserAccountHandler, so cache consumers (GetAccountNonce,
	// GetAccountNonceAndBalance) can call GetNonce()/GetBalance() without knowing the difference.
	for i, addr := range uncachedAddresses {
		provider.cache[string(addr)] = lightAccounts[i]
	}

	return true
}

// prefetchFullAccounts attempts to batch-fetch accounts using the full path
// (AccountsBatchReader), which returns complete UserAccount objects.
// Returns true if the full batch path was used.
func (provider *AccountsEphemeralProvider) prefetchFullAccounts(
	uncachedAddresses [][]byte,
	totalAddresses int,
	startBatch time.Time,
) bool {
	batchReader, ok := provider.accounts.(AccountsBatchReader)
	if !ok {
		return false
	}

	accounts, err := batchReader.GetExistingAccountsBatch(uncachedAddresses)
	if err != nil {
		log.Warn("AccountsEphemeralProvider.PrefetchAccounts full batch read failed, falling back to sequential",
			"err", err,
			"num addresses", len(uncachedAddresses))
		// Fall back to sequential
		for _, addr := range uncachedAddresses {
			_, _ = provider.GetUserAccount(addr)
		}
		return true
	}

	log.Trace("AccountsEphemeralProvider.PrefetchAccounts full batch read completed",
		"total addresses", totalAddresses,
		"uncached addresses", len(uncachedAddresses),
		"duration", time.Since(startBatch))

	// Cache all results (both existing accounts and nil for non-existent)
	for i, addr := range uncachedAddresses {
		var userAccount UserAccountHandler
		if accounts[i] != nil {
			typed, castOk := accounts[i].(UserAccountHandler)
			if castOk {
				userAccount = typed
			}
		}
		provider.cache[string(addr)] = userAccount
	}

	return true
}

// upgradeGuardedAccounts proactively upgrades all guarded light accounts in the cache
// to full accounts. This is called after the light batch prefetch so that the selection
// loop sees only cache hits for guardian checks (no individual trie reads).
func (provider *AccountsEphemeralProvider) upgradeGuardedAccounts() {
	startUpgrade := time.Now()
	upgraded := 0
	fallbacks := 0

	for addrStr, account := range provider.cache {
		light, isLight := account.(*lightAccountInfo)
		if !isLight || !light.IsGuarded() {
			continue
		}

		// Safe: Go allows modifying existing map entries during range iteration.
		// upgradeFromLight replaces this cache entry with a full account.
		_, err := provider.upgradeFromLight([]byte(addrStr), light)
		if err != nil {
			fallbacks++
		} else {
			upgraded++
		}
	}

	if upgraded > 0 || fallbacks > 0 {
		log.Trace("AccountsEphemeralProvider.upgradeGuardedAccounts completed",
			"upgraded", upgraded,
			"fallbacks", fallbacks,
			"duration", time.Since(startUpgrade))
	}
}

// IsAccountGuarded checks whether the account is guarded without upgrading
// lightAccountInfo to full accounts. Safe because lightAccountInfo correctly
// implements IsGuarded() via codeMetadata.
func (provider *AccountsEphemeralProvider) IsAccountGuarded(address []byte) (bool, error) {
	account, err := provider.getAccountForNonceAndBalance(address)
	if err != nil {
		return false, err
	}
	if check.IfNil(account) {
		return false, nil
	}

	return account.IsGuarded(), nil
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

// compile-time interface check
var _ common.AccountBatchPrefetcher = (*AccountsEphemeralProvider)(nil)

// IsInterfaceNil returns true if there is no value under the interface
func (provider *AccountsEphemeralProvider) IsInterfaceNil() bool {
	return provider == nil
}
