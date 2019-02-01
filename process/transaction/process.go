package transaction

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

// txProcessor implements TransactionProcessor interface and can modify account states according to a transaction
type txProcessor struct {
	accounts    state.AccountsAdapter
	adrConv     state.AddressConverter
	hasher      hashing.Hasher
	scHandler   func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error
	marshalizer marshal.Marshalizer
}

// NewTxProcessor creates a new txProcessor engine
func NewTxProcessor(
	accounts state.AccountsAdapter,
	hasher hashing.Hasher,
	addressConv state.AddressConverter,
	marshalizer marshal.Marshalizer,
) (*txProcessor, error) {

	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if addressConv == nil {
		return nil, process.ErrNilAddressConverter
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	return &txProcessor{
		accounts:    accounts,
		hasher:      hasher,
		adrConv:     addressConv,
		marshalizer: marshalizer,
	}, nil
}

// SCHandler returns the smart contract execution function
func (txProc *txProcessor) SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	return txProc.scHandler
}

// SetSCHandler sets the smart contract execution function
func (txProc *txProcessor) SetSCHandler(f func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	txProc.scHandler = f
}

// ProcessTransaction modifies the account states in respect with the transaction data
func (txProc *txProcessor) ProcessTransaction(tx *transaction.Transaction, roundIndex int32) error {
	if tx == nil {
		return process.ErrNilTransaction
	}

	adrSrc, adrDest, err := txProc.getAddresses(tx)
	if err != nil {
		return err
	}

	acntSrc, acntDest, err := txProc.getAccounts(adrSrc, adrDest)
	if err != nil {
		return err
	}

	if acntSrc == nil || acntDest == nil {
		return process.ErrNilValue
	}

	if acntDest.Code() != nil {
		return txProc.callSCHandler(tx)
	}

	if bytes.Equal(adrDest.Bytes(), state.RegistrationAddress.Bytes()) {
		regAccount, err := txProc.accounts.GetJournalizedAccount(state.RegistrationAddress)
		if err != nil {
			return err
		}

		regData := &state.RegistrationData{}
		err = txProc.marshalizer.Unmarshal(regData, tx.Data)
		if err != nil {
			return err
		}
		regData.OriginatorPubKey = adrSrc.Bytes()
		regData.RoundIndex = roundIndex

		err = regAccount.AppendDataRegistrationWithJournal(regData)
		if err != nil {
			return err
		}
	}

	value := tx.Value

	err = txProc.checkTxValues(acntSrc, value, tx.Nonce)
	if err != nil {
		return err
	}

	err = txProc.moveBalances(acntSrc, acntDest, value)
	if err != nil {
		return err
	}

	err = txProc.increaseNonceAcntSrc(acntSrc)
	if err != nil {
		return err
	}

	return nil
}

// SetBalancesToTrie adds balances to trie
func (txProc *txProcessor) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	if txProc.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if accBalance == nil {
		return nil, process.ErrNilValue
	}

	for i, v := range accBalance {
		err := txProc.setBalanceToTrie([]byte(i), v)

		if err != nil {
			return nil, err
		}
	}

	rootHash, err = txProc.accounts.Commit()

	if err != nil {
		err2 := txProc.accounts.RevertToSnapshot(0)

		if err2 != nil {
			log.Error(err2.Error())
		}
		return nil, err
	}

	return rootHash, err
}

func (txProc *txProcessor) setBalanceToTrie(addr []byte, balance *big.Int) error {
	if addr == nil {
		return process.ErrNilValue
	}

	addrContainer, err := txProc.adrConv.CreateAddressFromPublicKeyBytes(addr)

	if err != nil {
		return err
	}

	if addrContainer == nil {
		return process.ErrNilAddressContainer
	}

	account, err := txProc.accounts.GetJournalizedAccount(addrContainer)

	if err != nil {
		return err
	}

	return account.SetBalanceWithJournal(balance)
}

func (txProc *txProcessor) getAddresses(tx *transaction.Transaction) (adrSrc, adrDest state.AddressContainer, err error) {
	//for now we assume that the address = public key
	adrSrc, err = txProc.adrConv.CreateAddressFromPublicKeyBytes(tx.SndAddr)
	if err != nil {
		return
	}
	adrDest, err = txProc.adrConv.CreateAddressFromPublicKeyBytes(tx.RcvAddr)
	return
}

func (txProc *txProcessor) getAccounts(adrSrc, adrDest state.AddressContainer) (
	acntSrc state.JournalizedAccountWrapper,
	acntDest state.JournalizedAccountWrapper,
	err error) {

	if adrSrc == nil || adrDest == nil {
		return nil, nil, process.ErrNilValue
	}

	if bytes.Equal(adrSrc.Bytes(), adrDest.Bytes()) {
		acnt, err := txProc.accounts.GetJournalizedAccount(adrSrc)
		if err != nil {
			return nil, nil, err
		}

		return acnt, acnt, nil
	}

	acntSrc, err = txProc.accounts.GetJournalizedAccount(adrSrc)
	if err != nil {
		return nil, nil, err
	}
	acntDest, err = txProc.accounts.GetJournalizedAccount(adrDest)

	return acntSrc, acntDest, err
}

func (txProc *txProcessor) callSCHandler(tx *transaction.Transaction) error {
	if txProc.scHandler == nil {
		return process.ErrNoVM
	}

	return txProc.scHandler(txProc.accounts, tx)
}

func (txProc *txProcessor) checkTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	//TODO: undo this for nonce checking and un-skip tests
	//if acntSrc.BaseAccount().Nonce < nonce {
	//	return process.ErrHigherNonceInTransaction
	//}
	//
	//if acntSrc.BaseAccount().Nonce > nonce {
	//	return process.ErrLowerNonceInTransaction
	//}

	//negative balance test is done in transaction interceptor as the transaction is invalid and thus shall not disseminate

	if acntSrc.BaseAccount().Balance.Cmp(value) < 0 {
		return process.ErrInsufficientFunds
	}

	return nil
}

func (txProc *txProcessor) moveBalances(acntSrc, acntDest state.JournalizedAccountWrapper, value *big.Int) error {
	operation1 := big.NewInt(0)
	operation2 := big.NewInt(0)

	err := acntSrc.SetBalanceWithJournal(operation1.Sub(acntSrc.BaseAccount().Balance, value))
	if err != nil {
		return err
	}
	err = acntDest.SetBalanceWithJournal(operation2.Add(acntDest.BaseAccount().Balance, value))
	if err != nil {
		return err
	}

	return nil
}

func (txProc *txProcessor) increaseNonceAcntSrc(acntSrc state.JournalizedAccountWrapper) error {
	return acntSrc.SetNonceWithJournal(acntSrc.BaseAccount().Nonce + 1)
}
