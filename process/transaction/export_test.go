package transaction

import (
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

type TxProcessor *txProcessor

var mutex sync.Mutex

func (txProc *txProcessor) GetAddresses(tx *transaction.Transaction) (adrSrc, adrDst state.AddressContainer, err error) {
	return txProc.getAddresses(tx)
}

func (txProc *txProcessor) GetAccounts(adrSrc, adrDst state.AddressContainer,
) (acntSrc, acntDst *state.Account, err error) {
	return txProc.getAccounts(adrSrc, adrDst)
}

func (txProc *txProcessor) CheckTxValues(tx *transaction.Transaction, acntSnd state.AccountHandler) error {
	return txProc.checkTxValues(tx, acntSnd)
}

func (txProc *txProcessor) MoveBalances(acntSrc, acntDst *state.Account, value *big.Int) error {
	return txProc.moveBalances(acntSrc, acntDst, value)
}

func (txProc *txProcessor) IncreaseNonce(acntSrc *state.Account) error {
	return txProc.increaseNonce(acntSrc)
}

func (txProc *txProcessor) SetMinTxFee(minTxFee uint64) {
	mutex.Lock()
	preprocess.MinTxFee = minTxFee
	mutex.Unlock()
}

func (txProc *txProcessor) SetMinGasPrice(minGasPrice uint64) {
	mutex.Lock()
	preprocess.MinGasPrice = minGasPrice
	mutex.Unlock()
}
