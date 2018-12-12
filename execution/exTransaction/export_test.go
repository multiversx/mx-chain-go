package exTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
)

func (et *execTransaction) GetAddresses(tx *transaction.Transaction) (adrSrc, adrDest state.AddressContainer, err error) {
	return et.getAddresses(tx)
}

func (et *execTransaction) GetAccounts(adrSrc, adrDest state.AddressContainer) (acntSrc, acntDest state.JournalizedAccountWrapper, err error) {
	return et.getAccounts(adrSrc, adrDest)
}

func (et *execTransaction) CallSChandler(tx *transaction.Transaction) error {
	return et.callSChandler(tx)
}

func (et *execTransaction) CallRegisterHandler(data []byte) error {
	return et.callRegisterHandler(data)
}

func (et *execTransaction) CallUnregisterHandler(data []byte) error {
	return et.callUnregisterHandler(data)
}

func (et *execTransaction) CheckTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	return et.checkTxValues(acntSrc, value, nonce)
}

func (et *execTransaction) MoveBalances(acntSrc, acntDest state.JournalizedAccountWrapper, value *big.Int) error {
	return et.moveBalances(acntSrc, acntDest, value)
}

func (et *execTransaction) IncreaseNonceAcntSrc(acntSrc state.JournalizedAccountWrapper) error {
	return et.increaseNonceAcntSrc(acntSrc)
}

func (rt *registerTransaction) Register(data []byte) error {
	return rt.register(data)
}

func (rt *registerTransaction) Unregister(data []byte) error {
	return rt.unregister(data)
}
