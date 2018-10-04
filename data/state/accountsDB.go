package state

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"strconv"
)

type AccountsDB struct {
	//should use a concurrent trie
	mainTrie trie.Trier
	hasher   hashing.Hasher
	marsh    marshal.Marshalizer

	prevRoot []byte
}

func NewAccountsDB(tr trie.Trier, hasher hashing.Hasher, marsh marshal.Marshalizer) *AccountsDB {
	adb := AccountsDB{mainTrie: tr, hasher: hasher, marsh: marsh}

	return &adb
}

func (adb *AccountsDB) RetrieveCode(state *AccountState) error {
	if state.CodeHash == nil {
		state.Code = nil
		return nil
	}

	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(state.CodeHash) != encoding.HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(encoding.HashLength) + "bytes")
	}

	val, err := adb.mainTrie.Get(state.CodeHash)

	if err != nil {
		return err
	}

	state.Code = val
	return nil
}

func (adb *AccountsDB) RetrieveData(state *AccountState) error {
	if state.Root == nil {
		state.Data = nil
		return nil
	}

	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(state.Root) != encoding.HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(encoding.HashLength) + "bytes")
	}

	dataTrie, err := adb.mainTrie.Recreate(state.Root, adb.mainTrie.DBW())
	if err != nil {
		//node does not exists, create new one
		dataTrie, err = adb.mainTrie.Recreate(make([]byte, 0), adb.mainTrie.DBW())
		if err != nil {
			return err
		}

		state.Root = dataTrie.Root()
	}

	state.Data = dataTrie
	return nil
}

func (adb *AccountsDB) PutCode(state *AccountState, code []byte) error {
	if (code == nil) || (len(code) == 0) {
		state.resetDataCode()
		return nil
	}

	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	state.CodeHash = state.hasher.Compute(string(code))
	state.Code = code

	err := adb.mainTrie.Update(state.CodeHash, state.Code)
	if err != nil {
		state.resetDataCode()
		return err
	}

	dataTrie, err := adb.mainTrie.Recreate(make([]byte, 0), adb.mainTrie.DBW())
	if err != nil {
		state.resetDataCode()
		return err
	}

	//should drop the variables from the code inside dataTrie, recomputing the root

	state.Data = dataTrie
	return nil
}

func (adb *AccountsDB) HasAccount(address Address) (bool, error) {
	if adb.mainTrie == nil {
		return false, errors.New("attempt to search on a nil trie")
	}

	adrHash := adb.hasher.Compute(string(address.Bytes()))

	val, err := adb.mainTrie.Get(adrHash)

	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (adb *AccountsDB) SaveAccountState(state *AccountState) error {
	if state == nil {
		return errors.New("can not save nil account")
	}

	if adb.mainTrie == nil {
		return errors.New("attempt to search on a nil trie")
	}

	buff, err := adb.marsh.Marshal(state.Account)

	if err != nil {
		return err
	}

	err = adb.mainTrie.Update(state.AddrHash, buff)
	if err != nil {
		return err
	}

	return nil
}

//func (adb *AccountsDB) GetOrCreateAccount(address Address) (*AccountState, error){
//	if adb.mainTrie == nil {
//		return nil, errors.New("attempt to search on a nil trie")
//	}
//
//	found, err := adb.HasAccount(address)
//	if err != nil{
//		return nil, err
//	}
//
//	if !found{
//		acntState := NewAccountState()
//
//
//	}
//}
