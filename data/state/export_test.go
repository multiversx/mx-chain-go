package state

func NewEmptyBaseAccount(container AddressContainer, tracker DataTrieTracker) *baseAccount {
	return &baseAccount{addressContainer: container, dataTrieTracker: tracker}
}

func (adb *AccountsDB) LoadCode(accountHandler baseAccountHandler) error {
	return adb.loadCode(accountHandler)
}

func (adb *AccountsDB) LoadDataTrie(accountHandler baseAccountHandler) error {
	return adb.loadDataTrie(accountHandler)
}

func (adb *AccountsDB) GetAccount(addressContainer AddressContainer) (AccountHandler, error) {
	return adb.getAccount(addressContainer)
}

func (tdaw *TrackableDataTrie) OriginalData() map[string][]byte {
	return tdaw.originalData
}
