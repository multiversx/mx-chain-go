package transaction

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type TxProcessor *txProcessor

// GetAccounts calls the un-exported method getAccounts
func (txProc *txProcessor) GetAccounts(adrSrc, adrDst []byte,
) (acntSrc, acntDst state.UserAccountHandler, err error) {
	return txProc.getAccounts(adrSrc, adrDst)
}

// CheckTxValues calls the un-exported method checkTxValues
func (txProc *txProcessor) CheckTxValues(tx *transaction.Transaction, acntSnd, acntDst state.UserAccountHandler, isUserTxOfRelayed bool) error {
	return txProc.checkTxValues(tx, acntSnd, acntDst, isUserTxOfRelayed)
}

// IncreaseNonce calls IncreaseNonce on the provided account
func (txProc *txProcessor) IncreaseNonce(acntSrc state.UserAccountHandler) {
	acntSrc.IncreaseNonce(1)
}

// ProcessTxFee calls the un-exported method processTxFee
func (txProc *txProcessor) ProcessTxFee(
	tx *transaction.Transaction,
	acntSnd, acntDst state.UserAccountHandler,
	txType process.TransactionType,
	isUserTxOfRelayed bool,
) (*big.Int, *big.Int, error) {
	return txProc.processTxFee(tx, acntSnd, acntDst, txType, isUserTxOfRelayed)
}

// SetWhitelistHandler sets the un-exported field whiteListerVerifiedTxs
func (inTx *InterceptedTransaction) SetWhitelistHandler(handler process.WhiteListHandler) {
	inTx.whiteListerVerifiedTxs = handler
}

// IsCrossTxFromMe calls the un-exported method isCrossTxFromMe
func (txProc *baseTxProcessor) IsCrossTxFromMe(adrSrc, adrDst []byte) bool {
	return txProc.isCrossTxFromMe(adrSrc, adrDst)
}

// ProcessUserTx calls the un-exported method processUserTx
func (txProc *txProcessor) ProcessUserTx(
	originalTx *transaction.Transaction,
	userTx *transaction.Transaction,
	relayedTxValue *big.Int,
	relayedNonce uint64,
) (vmcommon.ReturnCode, error) {
	return txProc.processUserTx(originalTx, userTx, relayedTxValue, relayedNonce)
}

// ProcessMoveBalanceCostRelayedUserTx calls the un-exported method processMoveBalanceCostRelayedUserTx
func (txProc *txProcessor) ProcessMoveBalanceCostRelayedUserTx(
	userTx *transaction.Transaction,
	userScr *smartContractResult.SmartContractResult,
	userAcc state.UserAccountHandler,
	originalTxHash []byte,
) error {
	return txProc.processMoveBalanceCostRelayedUserTx(userTx, userScr, userAcc, originalTxHash)
}

// ExecuteFailedRelayedTransaction calls the un-exported method executeFailedRelayedUserTx
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

// CheckMaxGasPrice calls the un-exported method checkMaxGasPrice
func (inTx *InterceptedTransaction) CheckMaxGasPrice() error {
	return inTx.checkMaxGasPrice()
}

// VerifyGuardian calls the un-exported method verifyGuardian
func (txProc *txProcessor) VerifyGuardian(tx *transaction.Transaction, account state.UserAccountHandler) error {
	return txProc.verifyGuardian(tx, account)
}

// ShouldIncreaseNonce calls the un-exported method shouldIncreaseNonce
func (txProc *txProcessor) ShouldIncreaseNonce(executionErr error) bool {
	return txProc.shouldIncreaseNonce(executionErr)
}

// AddNonExecutableLog calls the un-exported method addNonExecutableLog
func (txProc *txProcessor) AddNonExecutableLog(executionErr error, originalTxHash []byte, originalTx data.TransactionHandler) error {
	return txProc.addNonExecutableLog(executionErr, originalTxHash, originalTx)
}
