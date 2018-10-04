package state

import (
	"bytes"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash []byte
	Root     []byte
}

type AccountState struct {
	Account
	Addr     Address
	AddrHash []byte
	Code     []byte
	Data     trie.Trier
	hasher   hashing.Hasher
}

func NewAccountState(address Address, account Account) *AccountState {
	acState := AccountState{Account: account, Addr: address}
	acState.hasher = DefHasher

	acState.AddrHash = DefHasher.Compute(address.String())

	return &acState
}

func (as *AccountState) RetrieveCode(srch trie.Trier) error {
	if as.CodeHash == nil {
		as.Code = nil
		return nil
	}

	if srch == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(as.CodeHash) != encoding.HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(encoding.HashLength) + "bytes")
	}

	val, err := srch.Get(as.CodeHash)

	if err != nil {
		return err
	}

	as.Code = val
	return nil
}

func (as *AccountState) RetrieveData(srch trie.Trier) error {
	if as.Root == nil {
		as.Data = nil
		return nil
	}

	if srch == nil {
		return errors.New("attempt to search on a nil trie")
	}

	if len(as.Root) != encoding.HashLength {
		return errors.New("attempt to search a hash not normalized to" +
			strconv.Itoa(encoding.HashLength) + "bytes")
	}

	if !bytes.Equal(as.Root, as.AddrHash) {
		return errors.New("can not retrieve data from other root hash")
	}

	dataTrie, err := srch.Recreate(as.Root, srch.DBW())
	if err != nil {
		return err
	}

	as.Data = dataTrie
	return nil
}

func (as *AccountState) PutCode(srch trie.Trier, code []byte) error {
	if (code == nil) || (len(code) == 0) {
		as.resetDataCode()
		return nil
	}

	as.CodeHash = as.hasher.Compute(string(code))
	as.Code = code

	err := srch.Update(as.CodeHash, as.Code)
	if err != nil {
		as.resetDataCode()
		return err
	}

	as.Root = as.AddrHash

	dataTrie, err := srch.Recreate(as.Root, srch.DBW())
	if err != nil {
		as.resetDataCode()
		return err
	}

	as.Data = dataTrie
	return nil
}

func (as *AccountState) resetDataCode() {
	//reset code and data
	as.CodeHash = nil
	as.Code = nil
	as.Root = nil
	as.Data = nil
}

//func (as *AccountState)Get(key []byte) ([]byte, error){
//	if as.Data == nil{
//		return nil, errors.New("nil data trie, can not retrieve")
//	}
//
//	if len(key) != encoding.HashLength{
//		return nil, errors.New("key is not normalized")
//	}
//
//	return as.Data.Get(key)
//}
//
//func (as *AccountState)Put(key []byte, value []byte) (error){
//	if as.Data == nil{
//		return errors.New("nil data trie, can not put")
//	}
//
//	//try to get, maybe it has same value, won't set dirty flag
//	data, err := as.Data.Get(key)
//	if err != nil{
//		return err
//	}
//
//	if bytes.Equal(data, value){
//		return nil
//	}
//
//	as.dirty = true
//
//	return as.Data.Update(key, value)
//}
//
//func (as *AccountState)Dirty() bool{
//	return as.dirty
//}
