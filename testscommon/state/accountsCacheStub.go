package state

import "github.com/multiversx/mx-chain-go/common"

// AccountsCacheStub -
type AccountsCacheStub struct {
	SaveAccountCalled                 func(address []byte, accountBytes []byte)
	GetAccountCalled                  func(address []byte) []byte
	UpdateTrieWithLatestChangesCalled func(trie common.Trie) error
	RevertLatestChangesCalled         func()
}

// SaveAccount -
func (a *AccountsCacheStub) SaveAccount(address []byte, accountBytes []byte) {
	if a.SaveAccountCalled != nil {
		a.SaveAccountCalled(address, accountBytes)
	}
}

// GetAccount -
func (a *AccountsCacheStub) GetAccount(address []byte) []byte {
	if a.GetAccountCalled != nil {
		return a.GetAccountCalled(address)
	}
	return nil
}

// UpdateTrieWithLatestChanges -
func (a *AccountsCacheStub) UpdateTrieWithLatestChanges(trie common.Trie) error {
	if a.UpdateTrieWithLatestChangesCalled != nil {
		return a.UpdateTrieWithLatestChangesCalled(trie)
	}
	return nil
}

// RevertLatestChanges -
func (a *AccountsCacheStub) RevertLatestChanges() {
	if a.RevertLatestChangesCalled != nil {
		a.RevertLatestChangesCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *AccountsCacheStub) IsInterfaceNil() bool {
	return a == nil
}
