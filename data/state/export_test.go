package state

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

func (tdaw *TrackableDataTrie) OriginalData() map[string][]byte {
	return tdaw.originalData
}
