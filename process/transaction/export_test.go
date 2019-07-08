package transaction

import (
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

func (txProc *txProcessor) CheckTxValues(acntSrc *state.Account, value *big.Int, nonce uint64) error {
	return txProc.checkTxValues(acntSrc, value, nonce)
}

func (txProc *txProcessor) MoveBalances(acntSrc, acntDst *state.Account, value *big.Int) error {
	return txProc.moveBalances(acntSrc, acntDst, value)
}

func (txProc *txProcessor) IncreaseNonce(acntSrc *state.Account) error {
	return txProc.increaseNonce(acntSrc)
}

func (txProc *txProcessor) SetMinTxFee(minTxFee int64) {
	mutex.Lock()
	MinTxFee = minTxFee
	mutex.Unlock()
}

func (txProc *txProcessor) SetMinGasPrice(minGasPrice int64) {
	mutex.Lock()
	MinGasPrice = minGasPrice
	mutex.Unlock()
}
