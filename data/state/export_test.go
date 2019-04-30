package state

import "github.com/ElrondNetwork/elrond-go-sandbox/data/trie"

func (adb *AccountsDB) LoadCode(accountHandler AccountHandler) error {
	return adb.loadCode(accountHandler)
}

func (adb *AccountsDB) GetAccount(addressContainer AddressContainer) (AccountHandler, error) {
	return adb.getAccount(addressContainer)
}

func (tdaw *TrackableDataTrie) OriginalData() map[string][]byte {
	return tdaw.originalData
}

func (tdaw *TrackableDataTrie) DataTrie() trie.PatriciaMerkelTree {
	return tdaw.tr
}
