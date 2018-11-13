package mock

import (
	"errors"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

var errorMockAccountsDBMock = errors.New("AccountsDBMock general failure")

// AccountsDBMock is the struct used for accessing accounts in "dummy" mode (only in mem)
type AccountsDBMock struct {
	//should use a concurrent trie
	MainTrie trie.PatriciaMerkelTree
	hasher   hashing.Hasher
	marsh    marshal.Marshalizer

	mutUsedAccounts   sync.RWMutex
	usedStateAccounts map[string]*state.AccountState

	FailSaveAccountState        bool
	NoToFailureSaveAccountState int

	FailGetOrCreateAccount        bool
	NoToFailureGetOrCreateAccount int
}

// NewAccountsDBMock creates a new mock account manager
func NewAccountsDBMock() *AccountsDBMock {
	h := &HasherMock{}
	m := &MarshalizerMock{}

	adb := AccountsDBMock{MainTrie: NewMockTrie(), hasher: h, marsh: m}

	adb.mutUsedAccounts = sync.RWMutex{}
	adb.usedStateAccounts = make(map[string]*state.AccountState, 0)

	return &adb
}

// RetrieveCode retrieves and saves the SC code inside AccountState object. Errors if something went wrong
func (adb *AccountsDBMock) RetrieveCode(s *state.AccountState) error {
	if s.CodeHash == nil {
		s.Code = nil
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(s.CodeHash) != encoding.HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(encoding.HashLength) + "bytes")
	}

	val, err := adb.MainTrie.Get(s.CodeHash)

	if err != nil {
		return err
	}

	s.Code = val
	return nil
}

// RetrieveData retrieves and saves the SC data inside AccountState object. Errors if something went wrong
func (adb *AccountsDBMock) RetrieveData(state *state.AccountState) error {
	if state.Root == nil {
		state.Data = nil
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(state.Root) != encoding.HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(encoding.HashLength) + "bytes")
	}

	dataTrie, err := adb.MainTrie.Recreate(state.Root, adb.MainTrie.DBW())
	if err != nil {
		//node does not exists, create new one
		dataTrie, err = adb.MainTrie.Recreate(make([]byte, 0), adb.MainTrie.DBW())
		if err != nil {
			return err
		}

		state.Root = dataTrie.Root()
	}

	state.Data = dataTrie
	return nil
}

// PutCode sets the SC plain code in AccountState object and trie, code hash in AccountState. Errors if something went wrong
func (adb *AccountsDBMock) PutCode(s *state.AccountState, code []byte) error {
	if (code == nil) || (len(code) == 0) {
		s.Reset()
		return nil
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	s.CodeHash = adb.hasher.Compute(string(code))
	s.Code = code

	err := adb.MainTrie.Update(s.CodeHash, s.Code)
	if err != nil {
		s.Reset()
		return err
	}

	dataTrie, err := adb.MainTrie.Recreate(make([]byte, 0), adb.MainTrie.DBW())
	if err != nil {
		s.Reset()
		return err
	}

	s.Data = dataTrie
	return nil
}

// HasAccount searches for an account based on the address. Errors if something went wrong and outputs if the account exists or not
func (adb *AccountsDBMock) HasAccount(address state.Address) (bool, error) {
	if adb.MainTrie == nil {
		return false, errors.New("attempt to search on a nil trie")
	}

	val, err := adb.MainTrie.Get(address.Hash(adb.hasher))

	if err != nil {
		return false, err
	}

	return val != nil, nil
}

// SaveAccountState saves the account WITHOUT data trie inside main trie. Errors if something went wrong
func (adb *AccountsDBMock) SaveAccountState(s *state.AccountState) error {
	if adb.FailSaveAccountState {
		if adb.NoToFailureSaveAccountState > 0 {
			adb.NoToFailureSaveAccountState--
		} else {
			return errorMockAccountsDBMock
		}
	}

	if s == nil {
		return errors.New("can not save nil account state")
	}

	if adb.MainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	buff, err := adb.marsh.Marshal(s.Account)

	if err != nil {
		return err
	}

	adb.trackAccountState(s)

	err = adb.MainTrie.Update(s.Addr.Hash(adb.hasher), buff)
	if err != nil {
		return err
	}

	return nil
}

// GetOrCreateAccount fetches the account based on the address. Creates an empty account if the account is missing.
func (adb *AccountsDBMock) GetOrCreateAccount(address state.Address) (*state.AccountState, error) {
	if adb.FailGetOrCreateAccount {
		if adb.NoToFailureGetOrCreateAccount > 0 {
			adb.NoToFailureGetOrCreateAccount--
		} else {
			return nil, errorMockAccountsDBMock
		}
	}

	has, err := adb.HasAccount(address)
	if err != nil {
		return nil, err
	}

	if has {
		val, err := adb.MainTrie.Get(address.Hash(adb.hasher))
		if err != nil {
			return nil, err
		}

		acnt := state.Account{}

		err = adb.marsh.Unmarshal(&acnt, val)
		if err != nil {
			return nil, err
		}

		s := state.NewAccountState(address, acnt)

		adb.trackAccountState(s)

		return s, nil
	}

	s := state.NewAccountState(address, state.Account{})

	err = adb.SaveAccountState(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Commit will persist all data inside
func (adb *AccountsDBMock) Commit() ([]byte, error) {
	adb.mutUsedAccounts.Lock()
	defer adb.mutUsedAccounts.Unlock()

	//Step 1. iterate through tracked account states map and commits the new tries
	for _, v := range adb.usedStateAccounts {
		if v.Dirty() {
			hash, err := v.Data.Commit(nil)
			if err != nil {
				return nil, err
			}

			v.Root = hash
			v.PrevRoot = hash
		}

		buff, err := adb.marsh.Marshal(v.Account)
		if err != nil {
			return nil, err
		}
		adb.MainTrie.Update(v.Addr.Hash(adb.hasher), buff)
	}

	//step 2. clean used accounts map
	adb.usedStateAccounts = make(map[string]*state.AccountState, 0)

	//Step 3. commit main trie
	hash, err := adb.MainTrie.Commit(nil)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (adb *AccountsDBMock) trackAccountState(state *state.AccountState) {
	//add account to used accounts to track the modifications in data trie for committing/undoing
	adb.mutUsedAccounts.Lock()
	_, ok := adb.usedStateAccounts[string(state.Addr.Hash(adb.hasher))]
	if !ok {
		adb.usedStateAccounts[string(state.Addr.Hash(adb.hasher))] = state
	}
	adb.mutUsedAccounts.Unlock()
}
