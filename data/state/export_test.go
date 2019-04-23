package state

//
//import (
//	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
//	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
//	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
//)
//
//func (tdaw *TrackableDataAccountWrap) OriginalData() map[string][]byte {
//	return tdaw.originalData
//}
//
//func Generate(trie trie.PatriciaMerkelTree, hasher hashing.Hasher, marshalizer marshal.Marshalizer,
//	journal *Journal) AccountsDB {
//	return AccountsDB{trie, hasher, marshalizer, journal}
//}

func (adb *AccountsDB) LoadCode(accountWrapper AccountWrapper) error {
	return adb.loadCode(accountWrapper)
}

//func (adb *AccountsDB) LoadJournalizedAccountWrapper(account *Account,
//	addressContainer AddressContainer) (JournalizedAccountWrapper, error) {
//
//	return adb.loadJournalizedAccountWrapper(account, addressContainer)
//}

func (adb *AccountsDB) GetAccount(addressContainer AddressContainer) (AccountWrapper, error) {
	return adb.getAccount(addressContainer)
}
