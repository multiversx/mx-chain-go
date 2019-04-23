package state

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/pkg/errors"
)

// JournalEntryCreation is used to revert an account creation
type JournalEntryCreation struct {
	key     []byte
	updater Updater
}

// JournalEntryNonce is used to revert a nonce change
type JournalEntryNonce struct {
	account  *Account
	oldNonce uint64
}

// JournalEntryBalance is used to revert a balance change
type JournalEntryBalance struct {
	account    *Account
	oldBalance *big.Int
}

// JournalEntryCodeHash is used to revert a code hash change
type JournalEntryCodeHash struct {
	account     AccountWrapper
	oldCodeHash []byte
}

// JournalEntryRootHash is used to revert an account's root hash change
type JournalEntryRootHash struct {
	account     AccountWrapper
	oldRootHash []byte
}

// JournalEntryData is used to mark an account's data change
type JournalEntryData struct {
	trie    trie.PatriciaMerkelTree
	account AccountWrapper
}

//TODO(jls) move
var ErrNilUpdater = errors.New("")
var ErrNilAccountWrapper = errors.New("")
var ErrNilOrEmptyKey = errors.New("")

//------- JournalEntryCreation

// NewJournalEntryCreation outputs a new JournalEntry implementation used to revert an account creation
func NewJournalEntryCreation(key []byte, updater Updater) (*JournalEntryCreation, error) {
	if updater == nil {
		return nil, ErrNilUpdater
	}
	if len(key) == 0 {
		return nil, ErrNilOrEmptyKey
	}

	return &JournalEntryCreation{
		key:     key,
		updater: updater,
	}, nil
}

// Revert applies undo operation
func (jec *JournalEntryCreation) Revert() (AccountWrapper, error) {
	return nil, jec.updater.Update(jec.key, nil)
}

//------- JournalEntryNonce

// NewJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewJournalEntryNonce(account *Account, oldNonce uint64) (*JournalEntryNonce, error) {
	if account == nil {
		return nil, ErrNilAccount
	}

	return &JournalEntryNonce{
		account:  account,
		oldNonce: oldNonce,
	}, nil
}

// Revert applies undo operation
func (jen *JournalEntryNonce) Revert() (AccountWrapper, error) {
	jen.account.Nonce = jen.oldNonce

	return jen.account, nil
}

//------- JournalEntryBalance

// NewJournalEntryBalance outputs a new JournalEntry implementation used to revert a balance change
func NewJournalEntryBalance(account *Account, oldBalance *big.Int) (*JournalEntryBalance, error) {
	if account == nil {
		return nil, ErrNilAccount
	}

	return &JournalEntryBalance{
		account:    account,
		oldBalance: oldBalance,
	}, nil
}

// Revert applies undo operation
func (jeb *JournalEntryBalance) Revert() (AccountWrapper, error) {
	jeb.account.Balance = jeb.oldBalance

	return jeb.account, nil
}

//------- JournalEntryCodeHash

// NewJournalEntryCodeHash outputs a new JournalEntry implementation used to revert a code hash change
func NewJournalEntryCodeHash(account AccountWrapper, oldCodeHash []byte) (*JournalEntryCodeHash, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &JournalEntryCodeHash{
		account:     account,
		oldCodeHash: oldCodeHash,
	}, nil
}

// Revert applies undo operation
func (jech *JournalEntryCodeHash) Revert() (AccountWrapper, error) {
	jech.account.SetCodeHash(jech.oldCodeHash)

	return jech.account, nil
}

//------- JournalEntryRoot

// NewJournalEntryRootHash outputs a new JournalEntry implementation used to revert an account's root hash change
func NewJournalEntryRootHash(account AccountWrapper, oldRootHash []byte) (*JournalEntryRootHash, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}
	if account.AddressContainer() == nil {
		return nil, ErrNilAddressContainer
	}

	return &JournalEntryRootHash{
		account:     account,
		oldRootHash: oldRootHash,
	}, nil
}

// Revert applies undo operation
func (jer *JournalEntryRootHash) Revert() (AccountWrapper, error) {
	jer.account.SetRootHash(jer.oldRootHash)
	//TODO(jls) fix this
	//err := jer.accountTracker.LoadDataTrie(jer.account)
	//if err != nil {
	//	return nil, err
	//}
	return jer.account, nil
}

//------- JournalEntryData

// NewJournalEntryData outputs a new JournalEntry implementation used to keep track of data change.
// The revert will practically empty the dirty data map
func NewJournalEntryData(account AccountWrapper, trie trie.PatriciaMerkelTree) (*JournalEntryData, error) {
	if account == nil {
		return nil, ErrNilAccountWrapper
	}

	return &JournalEntryData{
		account: account,
		trie:    trie,
	}, nil
}

// Revert will empty the dirtyData map from AccountState
func (jed *JournalEntryData) Revert() (AccountWrapper, error) {
	//TODO(jls) call clear cache here from TrackableDataTrie
	//jed.account.ClearDataCaches()

	return nil, nil
}

// Trie returns the referenced PatriciaMerkelTree for committing the changes
func (jed *JournalEntryData) Trie() trie.PatriciaMerkelTree {
	return jed.trie
}
