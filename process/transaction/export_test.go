package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type TxProcessor *txProcessor

func (txProc *txProcessor) GetAccounts(adrSrc, adrDst []byte,
) (acntSrc, acntDst state.UserAccountHandler, err error) {
	return txProc.getAccounts(adrSrc, adrDst)
}

func (txProc *txProcessor) CheckTxValues(tx *transaction.Transaction, acntSnd, acntDst state.UserAccountHandler, isUserTxOfRelayed bool) error {
	return txProc.checkTxValues(tx, acntSnd, acntDst, isUserTxOfRelayed)
}

func (txProc *txProcessor) IncreaseNonce(acntSrc state.UserAccountHandler) {
	acntSrc.IncreaseNonce(1)
}

func (txProc *txProcessor) ProcessTxFee(
	tx *transaction.Transaction,
	acntSnd, acntDst state.UserAccountHandler,
	txType process.TransactionType,
	isUserTxOfRelayed bool,
) (*big.Int, *big.Int, error) {
	return txProc.processTxFee(tx, acntSnd, acntDst, txType, isUserTxOfRelayed)
}

func (inTx *InterceptedTransaction) SetWhitelistHandler(handler process.WhiteListHandler) {
	inTx.whiteListerVerifiedTxs = handler
}

func (txProc *baseTxProcessor) IsCrossTxFromMe(adrSrc, adrDst []byte) bool {
	return txProc.isCrossTxFromMe(adrSrc, adrDst)
}

func (txProc *txProcessor) ProcessUserTx(
	originalTx *transaction.Transaction,
	userTx *transaction.Transaction,
	relayedTxValue *big.Int,
	relayedNonce uint64,
	txHash []byte,
) (vmcommon.ReturnCode, error) {
	return txProc.processUserTx(originalTx, userTx, relayedTxValue, relayedNonce, txHash)
}

func (txProc *txProcessor) ProcessMoveBalanceCostRelayedUserTx(
	userTx *transaction.Transaction,
	userScr *smartContractResult.SmartContractResult,
	userAcc state.UserAccountHandler,
	originalTxHash []byte,
) error {
	return txProc.processMoveBalanceCostRelayedUserTx(userTx, userScr, userAcc, originalTxHash)
}

func (txProc *txProcessor) ExecuteFailedRelayedTransaction(
	userTx *transaction.Transaction,
	relayerAdr []byte,
	relayedTxValue *big.Int,
	relayedNonce uint64,
	originalTx *transaction.Transaction,
	originalTxHash []byte,
	errorMsg string,
) error {
	return txProc.executeFailedRelayedUserTx(
		userTx,
		relayerAdr,
		relayedTxValue,
		relayedNonce,
		originalTx,
		originalTxHash,
		errorMsg)
}
