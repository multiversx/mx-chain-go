package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func (j *Journal) DirtyAddresses() map[*Address]int {
	return j.dirtyAddresses
}

func (adb *AccountsDB) Marsh() marshal.Marshalizer {
	return adb.marsh
}

func (as *AccountState) OriginalData() map[string][]byte {
	return as.originalData
}
