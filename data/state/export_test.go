package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func (mdaw *ModifyingDataAccountWrap) OriginalData() map[string][]byte {
	return mdaw.originalData
}

func (adb *AccountsDB) SetMainTrie(trie trie.PatriciaMerkelTree) {
	adb.mainTrie = trie
}

func (adb *AccountsDB) MainTrie() trie.PatriciaMerkelTree {
	return adb.mainTrie
}

func (adb *AccountsDB) Marshalizer() marshal.Marshalizer {
	return adb.marshalizer
}

func (adb *AccountsDB) GetDbAccount(addressContainer AddressContainer) (DbAccountContainer, error) {
	return adb.getDbAccount(addressContainer)
}

func (adb *AccountsDB) LoadCode(journalizedAccountWrapper JournalizedAccountWrapper) error {
	return adb.loadCode(journalizedAccountWrapper)
}

func (adb *AccountsDB) LoadJournalizedAccountWrapper(dBaccount DbAccountContainer,
	addressContainer AddressContainer) (JournalizedAccountWrapper, error) {

	return adb.loadJournalizedAccountWrapper(dBaccount, addressContainer)
}
