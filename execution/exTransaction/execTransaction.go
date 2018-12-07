package exTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// execTransaction implements TransactionExecutor interface and can modify account states according to a transaction
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
		return nil, ErrNilAccountsAdapter
	}

	if hasher == nil {
		return nil, ErrNilHasher
	}

	if addressConv == nil {
		return nil, ErrNilAddressConverter
	}

	return &execTransaction{
		accounts: accounts,
		hasher:   hasher,
		adrConv:  addressConv,
	}, nil
}

// SChandler returns the smart contract execution function
func (et *execTransaction) SChandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	return et.scHandler
}

// SetSChandler sets the smart contract execution function
func (et *execTransaction) SetSChandler(f func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	et.scHandler = f
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (et *execTransaction) ProcessTransaction(tx *transaction.Transaction) error {
	if tx == nil {
		return ErrNilTransaction
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
		return ErrNilValue
	}

	if acntDest.Code() != nil {
		return et.callSChandler(tx)
	}

	//TODO change to big int implementation
	value := big.NewInt(0)
	value.SetUint64(tx.Value)

	err = et.checkTxValues(acntSrc, value, tx.Nonce)
	if err != nil {
		return err
	}

	err = et.moveBalances(acntSrc, acntDest, value)
	if err != nil {
		return err
	}

	err = et.increaseNonceAcntSrc(acntSrc)
	if err != nil {
		return err
	}

	return nil
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
		err = ErrNilValue
		return
	}

	acntSrc, err = et.accounts.GetJournalizedAccount(adrSrc)
	if err != nil {
		return
	}
	acntDest, err = et.accounts.GetJournalizedAccount(adrDest)

	return
}

func (et *execTransaction) callSChandler(tx *transaction.Transaction) error {
	if et.scHandler == nil {
		return ErrNoVM
	}

	return et.scHandler(et.accounts, tx)
}

func (et *execTransaction) checkTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	if acntSrc.BaseAccount().Nonce < nonce {
		return ErrHigherNonceInTransaction
	}

	if acntSrc.BaseAccount().Nonce > nonce {
		return ErrLowerNonceInTransaction
	}

	//negative balance test is done in transaction interceptor as the transaction is invalid and thus shall not disseminate

	if acntSrc.BaseAccount().Balance.Cmp(value) < 0 {
		return ErrInsufficientFunds
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
