package transaction

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type TxProcessor *txProcessor

func (txProc *txProcessor) GetAccounts(adrSrc, adrDst []byte,
) (acntSrc, acntDst common.UserAccountHandler, err error) {
	return txProc.getAccounts(adrSrc, adrDst)
}

func (txProc *txProcessor) CheckTxValues(tx *transaction.Transaction, acntSnd, acntDst common.UserAccountHandler, isUserTxOfRelayed bool) error {
	return txProc.checkTxValues(tx, acntSnd, acntDst, isUserTxOfRelayed)
}

func (txProc *txProcessor) IncreaseNonce(acntSrc common.UserAccountHandler) {
	acntSrc.IncreaseNonce(1)
}

func (txProc *txProcessor) ProcessTxFee(
	tx *transaction.Transaction,
	acntSnd, acntDst common.UserAccountHandler,
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
	userAcc common.UserAccountHandler,
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

func (inTx *InterceptedTransaction) CheckMaxGasPrice() error {
	return inTx.checkMaxGasPrice()
}

func (txProc *txProcessor) VerifyGuardian(tx *transaction.Transaction, account common.UserAccountHandler) error {
	return txProc.verifyGuardian(tx, account)
}

// ShouldIncreaseNonce -
func (txProc *txProcessor) ShouldIncreaseNonce(executionErr error) bool {
	return txProc.shouldIncreaseNonce(executionErr)
}
