package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type TxProcessor *txProcessor

func (txProc *txProcessor) GetAddresses(tx *transaction.Transaction) (adrSrc, adrDest state.AddressContainer, err error) {
	return txProc.getAddresses(tx)
}

func (txProc *txProcessor) GetAccounts(
	adrSrc, adrDest state.AddressContainer,
	adrSrcInNodeShard, adrDestInNodeShard bool,
) (acntSrc, acntDest state.JournalizedAccountWrapper, err error) {
	return txProc.getAccounts(adrSrc, adrDest, adrSrcInNodeShard, adrDestInNodeShard)
}

func (txProc *txProcessor) CallSCHandler(tx *transaction.Transaction) error {
	return txProc.callSCHandler(tx)
}

func (txProc *txProcessor) CheckTxValues(acntSrc state.JournalizedAccountWrapper, value *big.Int, nonce uint64) error {
	return txProc.checkTxValues(acntSrc, value, nonce)
}

func (txProc *txProcessor) MoveBalances(
	acntSrc, acntDest state.JournalizedAccountWrapper,
	adrSrcInNodeShard, adrDestInNodeShard bool,
	value *big.Int,
) error {
	return txProc.moveBalances(acntSrc, acntDest, adrSrcInNodeShard, adrDestInNodeShard, value)
}

func (txProc *txProcessor) IncreaseNonceAcntSrc(acntSrc state.JournalizedAccountWrapper) error {
	return txProc.increaseNonceAcntSrc(acntSrc)
}
