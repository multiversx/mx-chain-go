package accountsCache

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/common"
)

// TODO add mutex
type accountsCache struct {
	modifiedAccounts map[string]struct{}
	previousCache    map[string][]byte
	currentCache     map[string][]byte
}

func NewAccountsCache() *accountsCache {
	return &accountsCache{
		modifiedAccounts: make(map[string]struct{}),
		previousCache:    make(map[string][]byte),
		currentCache:     make(map[string][]byte),
	}
}

func (ac *accountsCache) SaveAccount(address []byte, accountBytes []byte) {
	// TODO log size of the cache
	ac.modifiedAccounts[string(address)] = struct{}{}
	ac.currentCache[string(address)] = accountBytes
}

func (ac *accountsCache) GetAccount(address []byte) []byte {
	account, found := ac.currentCache[string(address)]
	if found {
		return account
	}

	account, found = ac.previousCache[string(address)]
	if found {
		return account
	}

	return nil
}

func (ac *accountsCache) UpdateTrieWithLatestChanges(trie common.Trie) error {
	for address := range ac.modifiedAccounts {
		account, found := ac.currentCache[address]
		if !found {
			return fmt.Errorf("account not found in cache")
		}

		err := trie.Update([]byte(address), account)
		if err != nil {
			return err
		}
	}

	ac.previousCache = ac.currentCache
	ac.currentCache = make(map[string][]byte)
	ac.modifiedAccounts = make(map[string]struct{})

	return nil
}

func (ac *accountsCache) RevertLatestChanges() {
	ac.currentCache = ac.previousCache
	ac.previousCache = make(map[string][]byte)
	ac.modifiedAccounts = make(map[string]struct{})
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *accountsCache) IsInterfaceNil() bool {
	return ac == nil
}
