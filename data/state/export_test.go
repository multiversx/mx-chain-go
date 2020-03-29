package state

func (adb *AccountsDB) LoadCode(accountHandler UserAccountHandler) error {
	return adb.loadCode(accountHandler)
}

func (adb *AccountsDB) LoadDataTrie(accountHandler UserAccountHandler) error {
	return adb.loadDataTrie(accountHandler)
}

func (adb *AccountsDB) GetAccount(addressContainer AddressContainer) (AccountHandler, error) {
	return adb.getAccount(addressContainer)
}

func (tdaw *TrackableDataTrie) OriginalData() map[string][]byte {
	return tdaw.originalData
}
