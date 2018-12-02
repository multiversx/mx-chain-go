package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func (adb *AccountsDB) Marsh() marshal.Marshalizer {
	return adb.marsh
}

func (adb *AccountsDB) RetrieveCode(state *AccountState) error {
	return adb.retrieveCode(state)
}

func (as *AccountState) OriginalData() map[string][]byte {
	return as.originalData
}
