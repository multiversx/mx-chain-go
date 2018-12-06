package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func (mdaw *ModifyingDataAccountWrap) OriginalData() map[string][]byte {
	return mdaw.originalData
}

func Generate(trie trie.PatriciaMerkelTree, hasher hashing.Hasher, marshalizer marshal.Marshalizer,
	journal *Journal) AccountsDB {
	return AccountsDB{trie, hasher, marshalizer, journal}
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
