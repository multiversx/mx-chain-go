package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// execTransaction implements TransactionProcessor interface and can modify account states according to a transaction
type execTransaction struct {
	accounts  state.AccountsAdapter
	adrConv   state.AddressConverter
	hasher    hashing.Hasher
	scHandler func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
}

// NewExecTransaction creates a new execTransaction engine
func NewExecTransaction(accounts state.AccountsAdapter, hasher hashing.Hasher,
	addressConv state.AddressConverter) (*execTransaction, error) {

	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if addressConv == nil {
		return nil, process.ErrNilAddressConverter
	}

	return &execTransaction{
		accounts: accounts,
		hasher:   hasher,
		adrConv:  addressConv,
	}, nil
}

// SCHandler returns the smart contract execution function
func (et *execTransaction) SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	return et.scHandler
}

// SetSCHandler sets the smart contract execution function
func (et *execTransaction) SetSCHandler(f func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	et.scHandler = f
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (et *execTransaction) ProcessTransaction(tx *transaction.Transaction) error {
	if tx == nil {
		return process.ErrNilTransaction
	}

	adrSrc, adrDest, err := et.getAddresses(tx)
	if err != nil {
		return err
	}

	acntSrc, acntDest, err := et.getAccounts(adrSrc, adrDest)
	if err != nil {
		return err
	}

	if acntSrc == nil || acntDest == nil {
		return process.ErrNilValue
	}

	if acntDest.Code() != nil {
		return et.callSCHandler(tx)
	}

	value := tx.Value

	err = et.checkTxValues(acntSrc, &value, tx.Nonce)
	if err != nil {
		return err
	}

	err = et.moveBalances(acntSrc, acntDest, &value)
	if err != nil {
		return err
	}

	err = et.increaseNonceAcntSrc(acntSrc)
	if err != nil {
		return err
	}

	return nil
}

// SetBalancesToTrie adds balances to trie
func (et *execTransaction) SetBalancesToTrie(accBalance map[string]big.Int) (rootHash []byte, err error) {

	if et.accounts.JournalLen() != 0 {
		return nil, err
	}

	for i, v := range accBalance {
		err := et.setBalanceToTrie([]byte(i), v)

		if err != nil {
			return nil, process.ErrAccountStateDirty
		}
	}

	rootHash, err = et.accounts.Commit()

	if err != nil {
		et.accounts.RevertToSnapshot(0)
	}

	return rootHash, err
}

func (et *execTransaction) setBalanceToTrie(addr []byte, balance big.Int) error {
	if addr == nil {
		return process.ErrNilValue
	}

	addrContainer, err := et.adrConv.CreateAddressFromPublicKeyBytes(addr)

	if err != nil {
		return err
	}

	if addrContainer == nil {
		return process.ErrNilAddressContainer
	}

	account, err := et.accounts.GetJournalizedAccount(addrContainer)

	if err != nil {
		return err
	}

	return account.SetBalanceWithJournal(balance)
}

func (et *execTransaction) getAddresses(tx *transaction.Transaction) (adrSrc, adrDest state.AddressContainer, err error) {
	//for now we assume that the address = public key
	adrSrc, err = et.adrConv.CreateAddressFromPublicKeyBytes(tx.SndAddr)
	if err != nil {
		return
	}
	adrDest, err = et.adrConv.CreateAddressFromPublicKeyBytes(tx.RcvAddr)
	return
}

func (et *execTransaction) getAccounts(adrSrc, adrDest state.AddressContainer) (acntSrc, acntDest state.JournalizedAccountWrapper, err error) {
	if adrSrc == nil || adrDest == nil {
		err = process.ErrNilValue
		return
	}

	acntSrc, err = et.accounts.GetJournalizedAccount(adrSrc)
	if err != nil {
		return
	}
	acntDest, err = et.accounts.GetJournalizedAccount(adrDest)

	return
}

func (et *execTransaction) callSCHandler(tx *transaction.Transaction) error {
	if et.scHandler == nil {
		return process.ErrNoVM
	}

	return et.scHandler(et.accounts, tx)
}

func (et *execTransaction) checkTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	if acntSrc.BaseAccount().Nonce < nonce {
		return process.ErrHigherNonceInTransaction
	}

	if acntSrc.BaseAccount().Nonce > nonce {
		return process.ErrLowerNonceInTransaction
	}

	//negative balance test is done in transaction interceptor as the transaction is invalid and thus shall not disseminate

	if acntSrc.BaseAccount().Balance.Cmp(value) < 0 {
		return process.ErrInsufficientFunds
	}

	return nil
}

func (et *execTransaction) moveBalances(acntSrc, acntDest state.JournalizedAccountWrapper, value *big.Int) error {
	operation1 := big.NewInt(0)
	operation2 := big.NewInt(0)

	err := acntSrc.SetBalanceWithJournal(*operation1.Sub(&acntSrc.BaseAccount().Balance, value))
	if err != nil {
		return err
	}
	err = acntDest.SetBalanceWithJournal(*operation2.Add(&acntDest.BaseAccount().Balance, value))
	if err != nil {
		return err
	}

	return nil
}

func (et *execTransaction) increaseNonceAcntSrc(acntSrc state.JournalizedAccountWrapper) error {
	return acntSrc.SetNonceWithJournal(acntSrc.BaseAccount().Nonce + 1)
}
