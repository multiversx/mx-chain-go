package state

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type accountsDbWrapperSync struct {
	*AccountsDB
}

// NewAccountsDbWrapperSync creates a new account manager used when syncing
func NewAccountsDbWrapperSync(adapter AccountsAdapter) (*accountsDbWrapperSync, error) {
	if adapter == nil || adapter.IsInterfaceNil() {
		return nil, ErrNilAccountsAdapter
	}

	accountsDb, ok := adapter.(*AccountsDB)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return &accountsDbWrapperSync{accountsDb}, nil
}

// PruneTrie removes old values from the trie database
func (adb *accountsDbWrapperSync) PruneTrie(rootHash []byte) error {
	return adb.mainTrie.Prune(rootHash, data.NewRoot)
}

// CancelPrune clears the trie's evictionWaitingList
func (adb *accountsDbWrapperSync) CancelPrune(rootHash []byte) {
	adb.mainTrie.CancelPrune(rootHash, data.OldRoot)
}
