package mock

import (
	"github.com/pkg/errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
)

type JournalizedAccountWrapMock struct {
	*state.Account
	address               state.AddressContainer
	code                  []byte
	dataTrie              trie.PatriciaMerkelTree
	ClearDataCachesCalled func()
	originalData          map[string][]byte
	dirtyData             map[string][]byte

	Fail bool
}

func NewJournalizedAccountWrapMock(address state.AddressContainer) *JournalizedAccountWrapMock {
	return &JournalizedAccountWrapMock{
		Account:      state.NewAccount(),
		address:      address,
		originalData: make(map[string][]byte),
		dirtyData:    make(map[string][]byte),
	}
}

func (jawm *JournalizedAccountWrapMock) AppendRegistrationData(data *state.RegistrationData) error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) CleanRegistrationData() error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) TrimLastRegistrationData() error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) AppendDataRegistrationWithJournal(*state.RegistrationData) error {
	panic("implement me")
}

func (jawm *JournalizedAccountWrapMock) BaseAccount() *state.Account {
	return jawm.Account
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
	if jawm.Fail {
		return nil, errors.New("failure")
	}

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

func (jawm *JournalizedAccountWrapMock) SetNonceWithJournal(nonce uint64) error {
	if jawm.Fail {
		return errors.New("failure")
	}

	jawm.Nonce = nonce
	return nil
}

func (jawm *JournalizedAccountWrapMock) SetBalanceWithJournal(balance big.Int) error {
	if jawm.Fail {
		return errors.New("failure")
	}

	jawm.Balance = balance
	return nil
}

func (jawm *JournalizedAccountWrapMock) SetCodeHashWithJournal(codeHash []byte) error {
	if jawm.Fail {
		return errors.New("failure")
	}

	jawm.Account.CodeHash = codeHash
	return nil
}

func (jawm *JournalizedAccountWrapMock) SetRootHashWithJournal([]byte) error {
	if jawm.Fail {
		return errors.New("failure")
	}

	panic("implement me")
}
