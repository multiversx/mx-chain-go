package exTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"math/big"

	"bytes"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
)

// execTransaction implements TransactionExecutor interface and can modify account states according to a transaction
type execTransaction struct {
	accounts          state.AccountsAdapter
	adrConv           state.AddressConverter
	hasher            hashing.Hasher
	scHandler         func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
	registerHandler   func(data []byte) error
	unregisterHandler func(data []byte) error
}

// NewExecTransaction creates a new execTransaction engine
func NewExecTransaction(accounts state.AccountsAdapter, hasher hashing.Hasher,
	addressConv state.AddressConverter) (*execTransaction, error) {

	if accounts == nil {
		return nil, execution.ErrNilAccountsAdapter
	}

	if hasher == nil {
		return nil, execution.ErrNilHasher
	}

	if addressConv == nil {
		return nil, execution.ErrNilAddressConverter
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

// RegisterHandler returns the registration execution function
func (et *execTransaction) RegisterHandler() func(data []byte) error {
	return et.registerHandler
}

// SetRegisterHandler sets the registration execution function
func (et *execTransaction) SetRegisterHandler(f func(data []byte) error) {
	et.registerHandler = f
}

// UnregisterHandler returns the unregistration execution function
func (et *execTransaction) UnregisterHandler() func(data []byte) error {
	return et.unregisterHandler
}

// SetUnregisterHandler sets the unregistration execution function
func (et *execTransaction) SetUnregisterHandler(f func(data []byte) error) {
	et.unregisterHandler = f
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (et *execTransaction) ProcessTransaction(tx *transaction.Transaction) error {
	if tx == nil {
		return execution.ErrNilTransaction
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
		return execution.ErrNilValue
	}

	if acntDest.Code() != nil {
		return et.callSChandler(tx)
	}

	if bytes.Equal(adrDest.Bytes(), []byte(RegisterAddress)) {
		err = et.callRegisterHandler(tx.Data)
		if err != nil {
			return err
		}
	}

	if bytes.Equal(adrDest.Bytes(), []byte(UnregisterAddress)) {
		err = et.callUnregisterHandler(tx.Data)
		if err != nil {
			return err
		}
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
		err = execution.ErrNilValue
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
		return execution.ErrNoVM
	}

	return et.scHandler(et.accounts, tx)
}

func (et *execTransaction) callRegisterHandler(data []byte) error {
	if et.registerHandler == nil {
		return execution.ErrRegisterFunctionUndefined
	}

	return et.registerHandler(data)
}

func (et *execTransaction) callUnregisterHandler(data []byte) error {
	if et.unregisterHandler == nil {
		return execution.ErrUnregisterFunctionUndefined
	}

	return et.unregisterHandler(data)
}

func (et *execTransaction) checkTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	if acntSrc.BaseAccount().Nonce < nonce {
		return execution.ErrHigherNonceInTransaction
	}

	if acntSrc.BaseAccount().Nonce > nonce {
		return execution.ErrLowerNonceInTransaction
	}

	//negative balance test is done in transaction interceptor as the transaction is invalid and thus shall not disseminate

	if acntSrc.BaseAccount().Balance.Cmp(value) < 0 {
		return execution.ErrInsufficientFunds
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
