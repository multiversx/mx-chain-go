package transactionAPI

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

type gasUsedAndFeeProcessor struct {
	feeComputer         feeComputer
	gasScheduleNotifier core.GasScheduleNotifier
	pubKeyConverter     core.PubkeyConverter
}

func newGasUsedAndFeeProcessor(txFeeCalculator feeComputer, gasScheduleNotifier core.GasScheduleNotifier, pubKeyConverter core.PubkeyConverter) *gasUsedAndFeeProcessor {
	return &gasUsedAndFeeProcessor{
		feeComputer:         txFeeCalculator,
		gasScheduleNotifier: gasScheduleNotifier,
		pubKeyConverter:     pubKeyConverter,
	}
}

func (gfp *gasUsedAndFeeProcessor) computeAndAttachGasUsedAndFee(tx *transaction.ApiTransactionResult) {
	gasUsed := gfp.feeComputer.ComputeGasLimit(tx)
	fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	tx.GasUsed = gasUsed
	tx.Fee = fee.String()

	if tx.IsRelayed || gfp.isESDTOperationWithSCCall(tx) {
		tx.GasUsed = tx.GasLimit
		tx.Fee = tx.InitiallyPaidFee
	}

	if gfp.isGuardianOperation(tx) {
		gasUsed = gfp.feeComputer.ComputeGasLimit(tx)
		guardianOperationCost := gfp.getGuardianOperationCost(tx)
		gasUsed += guardianOperationCost
		tx.GasUsed = gasUsed

		fee = big.NewInt(0).SetUint64(gasUsed * tx.GasPrice)
		tx.Fee = fee.String()

		initiallyPaidFee := gfp.feeComputer.ComputeMoveBalanceFee(tx)
		tx.InitiallyPaidFee = initiallyPaidFee.String()

		return
	}

	hasRefundForSender := false
	for _, scr := range tx.SmartContractResults {
		if !scr.IsRefund || scr.RcvAddr != tx.Sender {
			continue
		}
		if scr.RcvAddr != tx.Sender {
			continue
		}

		gfp.setGasUsedAndFeeBaseOnRefundValue(tx, scr.Value)
		hasRefundForSender = true
		break
	}

	gfp.prepareTxWithResultsBasedOnLogs(tx, hasRefundForSender)
}

func (gfp *gasUsedAndFeeProcessor) getGuardianOperationCost(tx *transaction.ApiTransactionResult) uint64 {
	gasSchedule, err := gfp.gasScheduleNotifier.GasScheduleForEpoch(tx.Epoch)
	if err != nil {
		return 0
	}

	switch tx.Operation {
	case core.BuiltInFunctionSetGuardian:
		return gasSchedule[common.BuiltInCost][core.BuiltInFunctionSetGuardian]
	case core.BuiltInFunctionGuardAccount:
		return gasSchedule[common.BuiltInCost][core.BuiltInFunctionGuardAccount]
	case core.BuiltInFunctionUnGuardAccount:
		return gasSchedule[common.BuiltInCost][core.BuiltInFunctionUnGuardAccount]
	default:
		return 0
	}
}

func (gfp *gasUsedAndFeeProcessor) isGuardianOperation(tx *transaction.ApiTransactionResult) bool {
	return tx.Operation == core.BuiltInFunctionSetGuardian ||
		tx.Operation == core.BuiltInFunctionGuardAccount ||
		tx.Operation == core.BuiltInFunctionUnGuardAccount
}

func (gfp *gasUsedAndFeeProcessor) prepareTxWithResultsBasedOnLogs(
	tx *transaction.ApiTransactionResult,
	hasRefund bool,
) {
	if tx.Logs == nil || (tx.Function == "" && tx.Operation == datafield.OperationTransfer) {
		return
	}

	for _, event := range tx.Logs.Events {
		gfp.setGasUsedAndFeeBaseOnLogEvent(tx, hasRefund, event)
	}
}

func (gfp *gasUsedAndFeeProcessor) setGasUsedAndFeeBaseOnLogEvent(tx *transaction.ApiTransactionResult, hasRefund bool, event *transaction.Events) {
	if core.WriteLogIdentifier == event.Identifier && !hasRefund {
		gasUsed, fee := gfp.feeComputer.ComputeGasUsedAndFeeBasedOnRefundValue(tx, big.NewInt(0))
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
	}
	if core.SignalErrorOperation == event.Identifier {
		fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
		tx.GasUsed = tx.GasLimit
		tx.Fee = fee.String()
	}
}

func (gfp *gasUsedAndFeeProcessor) setGasUsedAndFeeBaseOnRefundValue(tx *transaction.ApiTransactionResult, refund *big.Int) {
	gasUsed, fee := gfp.feeComputer.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refund)
	tx.GasUsed = gasUsed
	tx.Fee = fee.String()
}

func (gfp *gasUsedAndFeeProcessor) isESDTOperationWithSCCall(tx *transaction.ApiTransactionResult) bool {
	isESDTTransferOperation := tx.Operation == core.BuiltInFunctionESDTTransfer ||
		tx.Operation == core.BuiltInFunctionESDTNFTTransfer || tx.Operation == core.BuiltInFunctionMultiESDTNFTTransfer

	isReceiverSC := core.IsSmartContractAddress(tx.Tx.GetRcvAddr())
	hasFunction := tx.Function != ""
	if !hasFunction {
		return false
	}

	if tx.Sender != tx.Receiver {
		return isESDTTransferOperation && isReceiverSC && hasFunction
	}

	if len(tx.Receivers) == 0 {
		return false
	}

	receiver := tx.Receivers[0]
	decodedReceiver, err := gfp.pubKeyConverter.Decode(receiver)
	if err != nil {
		log.Warn("gasUsedAndFeeProcessor.isESDTOperationWithSCCall cannot decode receiver address", "error", err.Error())
		return false
	}

	isReceiverSC = core.IsSmartContractAddress(decodedReceiver)

	return isESDTTransferOperation && isReceiverSC && hasFunction
}
