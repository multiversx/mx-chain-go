package transactionAPI

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

type gasUsedAndFeeProcessor struct {
	feeComputer         feeComputer
	pubKeyConverter     core.PubkeyConverter
	argsParser          process.ArgumentsParser
	marshaller          marshal.Marshalizer
	enableEpochsHandler common.EnableEpochsHandler
}

func newGasUsedAndFeeProcessor(
	txFeeCalculator feeComputer,
	pubKeyConverter core.PubkeyConverter,
	argsParser process.ArgumentsParser,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
) *gasUsedAndFeeProcessor {
	return &gasUsedAndFeeProcessor{
		feeComputer:         txFeeCalculator,
		pubKeyConverter:     pubKeyConverter,
		argsParser:          argsParser,
		marshaller:          marshaller,
		enableEpochsHandler: enableEpochsHandler,
	}
}

func (gfp *gasUsedAndFeeProcessor) computeAndAttachGasUsedAndFee(tx *transaction.ApiTransactionResult) {
	gasUsed := gfp.feeComputer.ComputeGasLimit(tx)
	fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	tx.GasUsed = gasUsed
	tx.Fee = fee.String()

	isFeeFixActive := gfp.enableEpochsHandler.IsFlagEnabledInEpoch(common.FixRelayedBaseCostFlag, tx.Epoch)
	isRelayedBeforeFix := tx.IsRelayed && !isFeeFixActive
	if isRelayedBeforeFix || gfp.isESDTOperationWithSCCall(tx) {
		tx.GasUsed = tx.GasLimit
		tx.Fee = tx.InitiallyPaidFee
	}

	initialTotalFee, isRelayedV3 := gfp.getFeeOfRelayedV3(tx)
	if isRelayedV3 && isFeeFixActive {
		tx.InitiallyPaidFee = initialTotalFee.String()
		tx.Fee = initialTotalFee.String()
		tx.GasUsed = big.NewInt(0).Div(initialTotalFee, big.NewInt(0).SetUint64(tx.GasPrice)).Uint64()
	}

	hasRefundForSender := false
	totalRefunds := big.NewInt(0)
	for _, scr := range tx.SmartContractResults {
		if !scr.IsRefund || scr.RcvAddr != tx.Sender {
			continue
		}

		hasRefundForSender = true
		totalRefunds.Add(totalRefunds, scr.Value)
	}

	if totalRefunds.Cmp(big.NewInt(0)) > 0 {
		gasUsed, fee = gfp.feeComputer.ComputeGasUsedAndFeeBasedOnRefundValue(tx, totalRefunds)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
	}

	gfp.prepareTxWithResultsBasedOnLogs(tx, hasRefundForSender)
}

func (gfp *gasUsedAndFeeProcessor) getFeeOfRelayedV3(tx *transaction.ApiTransactionResult) (*big.Int, bool) {
	if !tx.IsRelayed {
		return nil, false
	}

	if len(tx.InnerTransactions) == 0 {
		return nil, false
	}

	return gfp.feeComputer.ComputeTransactionFee(tx), true
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
