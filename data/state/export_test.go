package state

import "github.com/ElrondNetwork/elrond-go/marshal"

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

func (adb *AccountsDB) GetAccount(address []byte) (AccountHandler, error) {
	return adb.getAccount(address)
}

func (adb *AccountsDB) GetObsoleteHashes() map[string][][]byte {
	return adb.obsoleteDataTrieHashes
}

func (adb *AccountsDB) GetPruningBufferLen() int {
	return adb.pruningBuffer.len()
}

func (adb *AccountsDB) RemoveAllPruningBufferData() [][]byte {
	return adb.pruningBuffer.removeAll()
}

func GetCode(account baseAccountHandler) []byte {
	return account.GetCodeHash()
}

func GetCodeEntry(codeHash []byte, trie Updater, marshalizer marshal.Marshalizer) (*CodeEntry, error) {
	return getCodeEntry(codeHash, trie, marshalizer)
}

func RemoveDuplicatedKeys(map1 map[string]struct{}, map2 map[string]struct{}) {
	removeDuplicatedKeys(map1, map2)
}
