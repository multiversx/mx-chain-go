package state

func (adb *AccountsDB) LoadCode(accountWrapper AccountWrapper) error {
	return adb.loadCode(accountWrapper)
}

func (adb *AccountsDB) GetAccount(addressContainer AddressContainer) (AccountWrapper, error) {
	return adb.getAccount(addressContainer)
}
