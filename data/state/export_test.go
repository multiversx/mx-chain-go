package state

import "github.com/ElrondNetwork/elrond-go-sandbox/data/trie"

func (adb *AccountsDB) LoadCode(accountWrapper AccountWrapper) error {
	return adb.loadCode(accountWrapper)
}

func (adb *AccountsDB) GetAccount(addressContainer AddressContainer) (AccountWrapper, error) {
	return adb.getAccount(addressContainer)
}

func (tdaw *TrackableDataTrie) OriginalData() map[string][]byte {
	return tdaw.originalData
}

func (tdaw *TrackableDataTrie) DataTrie() trie.PatriciaMerkelTree {
	return tdaw.tr
}
