package state

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"sync"
)

type AccountsDB_A struct {
	tr trie.Trier

	prevRoot []byte

	mutStates      sync.RWMutex
	statesAccessed []AccountState
}

func NewAccounts(tr trie.Trier) (*AccountsDB_A, error) {
	if tr == nil {
		return nil, ErrorNilTrie
	}

	ac := &AccountsDB_A{tr: tr}
	ac.prevRoot = ac.tr.Root()
	ac.mutStates = sync.RWMutex{}
	ac.statesAccessed = make([]AccountState, 0)

	return ac, nil
}

//func (a *Accounts) Put(key []byte, value []byte) error {
//	return a.tr.Update(key, value)
//}
//
//func (a *Accounts) Get(key []byte) ([]byte, error) {
//	return a.tr.Get(key)
//}
//
//func (a *Accounts) Delete(key []byte) error {
//	return a.tr.Delete(key)
//}

func (a *AccountsDB_A) Commit() error {
	a.mutStates.Lock()
	defer a.mutStates.Unlock()

	hash, err := a.tr.Commit(nil)

	if err != nil {
		return err
	}

	a.prevRoot = hash
	return nil
}

func (a *AccountsDB_A) Undo() error {
	a.mutStates.Lock()
	defer a.mutStates.Unlock()

	ac, err := a.tr.Recreate(a.prevRoot, a.tr.DBW())

	if err != nil {
		return err
	}

	a.tr = ac

	//undo on all data tries belonging to last accounts accessed

	return nil
}

func (a *AccountsDB_A) Root() []byte {
	return a.tr.Root()
}

//func (a *Accounts) GetOrCreateAcntState(address Address) (*AccountState, error){
//	adrHash := DefHasher.Compute(address.String())
//
//	data, err := a.Get(adrHash)
//	if err != nil{
//		return nil, err
//	}
//
//	if data == nil{
//		acState := NewAccountState(address, Account{})
//
//
//	}
//
//
//
//
//
//	account := Account{}
//	err = DefMarsh.Unmarshal(&account, data)
//	if err != nil{
//		return nil, err
//	}
//
//	//fetch code from code hash
//	if account.CodeHash != nil{
//		c, err := a.Get(account.CodeHash)
//		if err != nil{
//			return nil, err
//		}
//
//		account.SetCode(c)
//	}
//
//	//fetch data from data hash
//	if account.DataHash != nil{
//		d, err := a.Get(account.DataHash)
//		if err != nil{
//			return nil, err
//		}
//
//		account.SetData(d)
//	}
//
//	return &account, nil
//}
//
//func (a *Accounts) CreateAcntState(address Address) (*AccountState, error){
//	adrHash := DefHasher.Compute(address.String())
//
//	data, err := a.Get(adrHash)
//	if err != nil{
//		return nil, err
//	}
//
//	if data != nil{
//		return nil, errors.New("account already exists")
//	}
//
//
//}
//
//func (a *Accounts) createAcntState(address Address) (*AccountState, error) {
//	acState := NewAccountState(address, Account{})
//
//
//
//}
//
//func (a *Accounts) PutAcntState(state AccountState) error{
//
//
//
//}
//
//
//func (a *Accounts) PutAccount(address Address, account Account) error{
//	//compute code hash and save it to accounter
//	if !(account.Code() == nil || len(account.Code()) == 0){
//		codeHash := DefHasher.Compute(string(account.Code()))
//		account.CodeHash = codeHash
//
//		a.Put(account.CodeHash, account.Code())
//	}
//
//	//compute data hash and save it to accounter
//	if !(account.Data() == nil || len(account.Data()) == 0) {
//		dataHash := DefHasher.Compute(string(account.Data()))
//		account.DataHash = dataHash
//
//		a.Put(account.DataHash, account.Data())
//	}
//
//	//save entire account
//	bAccount, err := DefMarsh.Marshal(account)
//	if err != nil{
//		return err
//	}
//
//	adrHash := DefHasher.Compute(address.String())
//
//	err = a.Put(adrHash, bAccount)
//	return err
//}
