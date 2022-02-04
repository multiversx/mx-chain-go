package state

import (
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func NewEmptyBaseAccount(address []byte, tracker DataTrieTracker) *baseAccount {
	return &baseAccount{
		address:         address,
		dataTrieTracker: tracker,
	}
}

func (adb *AccountsDB) LoadCode(accountHandler baseAccountHandler) error {
	return adb.loadCode(accountHandler)
}

func (adb *AccountsDB) LoadDataTrie(accountHandler baseAccountHandler) error {
	return adb.loadDataTrie(accountHandler)
}

func (adb *AccountsDB) GetAccount(address []byte) (vmcommon.AccountHandler, error) {
	return adb.getAccount(address)
}

func (adb *AccountsDB) GetObsoleteHashes() map[string][][]byte {
	return adb.obsoleteDataTrieHashes
}

func GetCode(account baseAccountHandler) []byte {
	return account.GetCodeHash()
}

func GetCodeEntry(codeHash []byte, trie Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	return getCodeEntry(codeHash, trie, marshalizer)
}

// RecreateTrieIfNecessary -
func (accountsDB *accountsDBApi) RecreateTrieIfNecessary() error {
	return accountsDB.recreateTrieIfNecessary()
}

// DoRecreateTrie -
func (accountsDB *accountsDBApi) DoRecreateTrie(targetRootHash []byte) error {
	return accountsDB.doRecreateTrie(targetRootHash)
}

// SetLastRootHash -
func (accountsDB *accountsDBApi) SetLastRootHash(rootHash []byte) {
	accountsDB.mutLastRootHash.Lock()
	accountsDB.lastRootHash = rootHash
	accountsDB.mutLastRootHash.Unlock()
}
