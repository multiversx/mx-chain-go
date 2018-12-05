package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type JournalizedAccountWrapMock struct {
	address               state.AddressContainer
	nonce                 uint64
	balance               big.Int
	codeHash              []byte
	rootHash              []byte
	code                  []byte
	dataTrie              trie.PatriciaMerkelTree
	ClearDataCachesCalled func()
	SaveToDbAccountCalled func(target state.DbAccountContainer) error
	originalData          map[string][]byte
	dirtyData             map[string][]byte
}

func NewJournalizedAccountWrapMock(address state.AddressContainer) *JournalizedAccountWrapMock {
	return &JournalizedAccountWrapMock{
		address:      address,
		originalData: make(map[string][]byte),
		dirtyData:    make(map[string][]byte),
	}
}

func (jawm *JournalizedAccountWrapMock) Nonce() uint64 {
	return jawm.nonce
}

func (jawm *JournalizedAccountWrapMock) SetNonce(nonce uint64) {
	jawm.nonce = nonce
}

func (jawm *JournalizedAccountWrapMock) CodeHash() []byte {
	return jawm.codeHash
}

func (jawm *JournalizedAccountWrapMock) SetCodeHash(codeHash []byte) {
	jawm.codeHash = codeHash
}

func (jawm *JournalizedAccountWrapMock) RootHash() []byte {
	return jawm.rootHash
}

func (jawm *JournalizedAccountWrapMock) SetRootHash(rootHash []byte) {
	jawm.rootHash = rootHash
}

func (jawm *JournalizedAccountWrapMock) Balance() big.Int {
	return jawm.balance
}

func (jawm *JournalizedAccountWrapMock) SetBalance(balance big.Int) {
	jawm.balance = balance
}

func (jawm *JournalizedAccountWrapMock) LoadFromDbAccount(source state.DbAccountContainer) error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) SaveToDbAccount(target state.DbAccountContainer) error {
	return jawm.SaveToDbAccountCalled(target)
}

func (jawm *JournalizedAccountWrapMock) AddressContainer() state.AddressContainer {
	return jawm.address
}

func (jawm *JournalizedAccountWrapMock) Code() []byte {
	return jawm.code
}

func (jawm *JournalizedAccountWrapMock) SetCode(code []byte) {
	jawm.code = code
}

func (jawm *JournalizedAccountWrapMock) DataTrie() trie.PatriciaMerkelTree {
	return jawm.dataTrie
}

func (jawm *JournalizedAccountWrapMock) SetDataTrie(trie trie.PatriciaMerkelTree) {
	jawm.dataTrie = trie
}

func (jawm *JournalizedAccountWrapMock) ClearDataCaches() {
	jawm.ClearDataCachesCalled()
}

func (jawm *JournalizedAccountWrapMock) DirtyData() map[string][]byte {
	return jawm.dirtyData
}

func (jawm *JournalizedAccountWrapMock) OriginalValue(key []byte) []byte {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) RetrieveValue(key []byte) ([]byte, error) {
	if jawm.DataTrie() == nil {
		return nil, state.ErrNilDataTrie
	}

	strKey := string(key)

	//search in dirty data cache
	data, found := jawm.dirtyData[strKey]
	if found {
		return data, nil
	}

	//search in original data cache
	data, found = jawm.originalData[strKey]
	if found {
		return data, nil
	}

	//ok, not in cache, retrieve from trie
	data, err := jawm.DataTrie().Get(key)
	if err != nil {
		return nil, err
	}

	//got the value, put it originalData cache as the next fetch will run faster
	jawm.originalData[string(key)] = data
	return data, nil
}

func (jawm *JournalizedAccountWrapMock) SaveKeyValue(key []byte, value []byte) {
	jawm.dirtyData[string(key)] = value
}

func (jawm *JournalizedAccountWrapMock) SetNonceWithJournal(uint64) error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) SetBalanceWithJournal(big.Int) error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) SetCodeHashWithJournal(codeHash []byte) error {
	jawm.codeHash = codeHash
	return nil
}

func (jawm *JournalizedAccountWrapMock) SetRootHashWithJournal([]byte) error {
	panic("implement me")
}
