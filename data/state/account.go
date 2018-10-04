package state

import (
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"math/big"
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

func NewAccountState(address Address, account Account, hasher hashing.Hasher) *AccountState {
	acState := AccountState{Account: account, Addr: address}
	acState.hasher = hasher

	acState.AddrHash = acState.hasher.Compute(address.Hex(hasher))

	return &acState
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
