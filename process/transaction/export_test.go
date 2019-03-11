package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type TxProcessor *txProcessor

func (txProc *txProcessor) GetAddresses(tx *transaction.Transaction) (adrSrc, adrDst state.AddressContainer, err error) {
	return txProc.getAddresses(tx)
}

func (txProc *txProcessor) GetAccounts(adrSrc, adrDst state.AddressContainer,
	srcInShard, dstInShard bool,
) (acntSrc, acntDst state.JournalizedAccountWrapper, err error) {
	return txProc.getAccounts(adrSrc, adrDst, srcInShard, dstInShard)
}

func (txProc *txProcessor) CallSCHandler(tx *transaction.Transaction) error {
	return txProc.callSCHandler(tx)
}

func (txProc *txProcessor) CheckTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	return txProc.checkTxValues(acntSrc, value, nonce)
}

func (txProc *txProcessor) MoveBalances(acntSrc, acntDst state.JournalizedAccountWrapper,
	srcInShard, dstInShard bool,
	value *big.Int,
) error {
	return txProc.moveBalances(acntSrc, acntDst, srcInShard, dstInShard, value)
}

func (txProc *txProcessor) IncreaseNonce(acntSrc state.JournalizedAccountWrapper) error {
	return txProc.increaseNonce(acntSrc)
}
