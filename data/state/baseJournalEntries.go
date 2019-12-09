package state

//------- BaseJournalEntryCreation

// BaseJournalEntryCreation creates a new account entry in the state trie
// through updater it can revert the created changes.
type BaseJournalEntryCreation struct {
	key     []byte
	updater Updater
}

// NewBaseJournalEntryCreation outputs a new BaseJournalEntry implementation used to revert an account creation
func NewBaseJournalEntryCreation(key []byte, updater Updater) (*BaseJournalEntryCreation, error) {
	if updater == nil || updater.IsInterfaceNil() {
		return nil, ErrNilUpdater
	}
	if len(key) == 0 {
		return nil, ErrNilOrEmptyKey
	}

	return &BaseJournalEntryCreation{
		key:     key,
		updater: updater,
	}, nil
}

// Revert applies undo operation
func (bjec *BaseJournalEntryCreation) Revert() (AccountHandler, error) {
	return nil, bjec.updater.Update(bjec.key, nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjec *BaseJournalEntryCreation) IsInterfaceNil() bool {
	if bjec == nil {
		return true
	}
	return false
}

//------- BaseJournalEntryCodeHash

// BaseJournalEntryCodeHash creates a code hash change in account
type BaseJournalEntryCodeHash struct {
	account     AccountHandler
	oldCodeHash []byte
}

// NewBaseJournalEntryCodeHash outputs a new BaseJournalEntry implementation used to save and revert a code hash change
func NewBaseJournalEntryCodeHash(account AccountHandler, oldCodeHash []byte) (*BaseJournalEntryCodeHash, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryCodeHash{
		account:     account,
		oldCodeHash: oldCodeHash,
	}, nil
}

// Revert applies undo operation
func (bjech *BaseJournalEntryCodeHash) Revert() (AccountHandler, error) {
	bjech.account.SetCodeHash(bjech.oldCodeHash)

	return bjech.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjech *BaseJournalEntryCodeHash) IsInterfaceNil() bool {
	if bjech == nil {
		return true
	}
	return false
}

//------- BaseJournalEntryNonce

// BaseJournalEntryNonce is used to revert a nonce change
type BaseJournalEntryNonce struct {
	account  AccountHandler
	oldNonce uint64
}

// NewBaseJournalEntryNonce outputs a new JournalEntry implementation used to revert a nonce change
func NewBaseJournalEntryNonce(account AccountHandler, oldNonce uint64) (*BaseJournalEntryNonce, error) {
	if account == nil || account.IsInterfaceNil() {
		return nil, ErrNilAccountHandler
	}

	return &BaseJournalEntryNonce{
		account:  account,
		oldNonce: oldNonce,
	}, nil
}

// Revert applies undo operation
func (bjen *BaseJournalEntryNonce) Revert() (AccountHandler, error) {
	bjen.account.SetNonce(bjen.oldNonce)

	return bjen.account, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bjen *BaseJournalEntryNonce) IsInterfaceNil() bool {
	if bjen == nil {
		return true
	}
	return false
}
