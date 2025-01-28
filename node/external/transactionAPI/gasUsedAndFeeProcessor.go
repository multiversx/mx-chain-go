package transactionAPI

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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

	userTx, initialTotalFee, isRelayedV1V2 := gfp.getFeeOfRelayedV1V2(tx)
	isRelayedAfterFix := isRelayedV1V2 && isFeeFixActive
	if isRelayedAfterFix {
		tx.InitiallyPaidFee = initialTotalFee.String()
		tx.Fee = initialTotalFee.String()
		tx.GasUsed = big.NewInt(0).Div(initialTotalFee, big.NewInt(0).SetUint64(tx.GasPrice)).Uint64()
	}

	isRelayedV3 := common.IsValidRelayedTxV3(tx.Tx)
	hasRefundForSender := false
	totalRefunds := big.NewInt(0)
	for _, scr := range tx.SmartContractResults {
		if !scr.IsRefund {
			continue
		}
		if !isRelayedV3 && scr.RcvAddr != tx.Sender {
			continue
		}
		if isRelayedV3 && scr.RcvAddr != tx.RelayerAddress {
			continue
		}

		hasRefundForSender = true
		totalRefunds.Add(totalRefunds, scr.Value)
	}

	if totalRefunds.Cmp(big.NewInt(0)) > 0 {
		gfp.setGasUsedAndFeeBaseOnRefundValue(tx, userTx, totalRefunds)
	}

	gfp.prepareTxWithResultsBasedOnLogs(tx, userTx, hasRefundForSender)
}

func (gfp *gasUsedAndFeeProcessor) getFeeOfRelayedV1V2(tx *transaction.ApiTransactionResult) (*transaction.ApiTransactionResult, *big.Int, bool) {
	if !tx.IsRelayed {
		return nil, nil, false
	}

	isRelayedV3 := common.IsValidRelayedTxV3(tx.Tx)
	if isRelayedV3 {
		return nil, nil, false
	}

	if len(tx.Data) == 0 {
		return nil, nil, false
	}

	funcName, args, err := gfp.argsParser.ParseCallData(string(tx.Data))
	if err != nil {
		return nil, nil, false
	}

	if funcName == core.RelayedTransaction {
		return gfp.handleRelayedV1(args, tx)
	}

	if funcName == core.RelayedTransactionV2 {
		return gfp.handleRelayedV2(args, tx)
	}

	return nil, nil, false
}

func (gfp *gasUsedAndFeeProcessor) handleRelayedV1(args [][]byte, tx *transaction.ApiTransactionResult) (*transaction.ApiTransactionResult, *big.Int, bool) {
	if len(args) != 1 {
		return nil, nil, false
	}

	innerTx := &transaction.Transaction{}
	err := gfp.marshaller.Unmarshal(innerTx, args[0])
	if err != nil {
		return nil, nil, false
	}

	gasUsed := gfp.feeComputer.ComputeGasLimit(tx)
	fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	innerTxApiResult := &transaction.ApiTransactionResult{
		Tx:    innerTx,
		Epoch: tx.Epoch,
	}
	innerFee := gfp.feeComputer.ComputeTransactionFee(innerTxApiResult)

	return innerTxApiResult, big.NewInt(0).Add(fee, innerFee), true
}

func (gfp *gasUsedAndFeeProcessor) handleRelayedV2(args [][]byte, tx *transaction.ApiTransactionResult) (*transaction.ApiTransactionResult, *big.Int, bool) {
	innerTx := &transaction.Transaction{}
	innerTx.RcvAddr = args[0]
	innerTx.Nonce = big.NewInt(0).SetBytes(args[1]).Uint64()
	innerTx.Data = args[2]
	innerTx.Signature = args[3]
	innerTx.Value = big.NewInt(0)
	innerTx.GasPrice = tx.GasPrice
	innerTx.GasLimit = tx.GasLimit - gfp.feeComputer.ComputeGasLimit(tx)
	innerTx.SndAddr = tx.Tx.GetRcvAddr()

	gasUsed := gfp.feeComputer.ComputeGasLimit(tx)
	fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	innerTxApiResult := &transaction.ApiTransactionResult{
		Tx:    innerTx,
		Epoch: tx.Epoch,
	}
	innerFee := gfp.feeComputer.ComputeTransactionFee(innerTxApiResult)

	return innerTxApiResult, big.NewInt(0).Add(fee, innerFee), true
}

func (gfp *gasUsedAndFeeProcessor) prepareTxWithResultsBasedOnLogs(
	tx *transaction.ApiTransactionResult,
	userTx *transaction.ApiTransactionResult,
	hasRefund bool,
) {
	if tx.Logs == nil || (tx.Function == "" && tx.Operation == datafield.OperationTransfer) {
		return
	}

	for _, event := range tx.Logs.Events {
		gfp.setGasUsedAndFeeBaseOnLogEvent(tx, userTx, hasRefund, event)
	}
}

func (gfp *gasUsedAndFeeProcessor) setGasUsedAndFeeBaseOnLogEvent(tx *transaction.ApiTransactionResult, userTx *transaction.ApiTransactionResult, hasRefund bool, event *transaction.Events) {
	if core.WriteLogIdentifier == event.Identifier && !hasRefund {
		gfp.setGasUsedAndFeeBaseOnRefundValue(tx, userTx, big.NewInt(0))
	}
	if core.SignalErrorOperation == event.Identifier {
		fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
		tx.GasUsed = tx.GasLimit
		tx.Fee = fee.String()
	}
}

func (gfp *gasUsedAndFeeProcessor) setGasUsedAndFeeBaseOnRefundValue(
	tx *transaction.ApiTransactionResult,
	userTx *transaction.ApiTransactionResult,
	refund *big.Int,
) {
	isRelayedV3 := len(tx.RelayerAddress) == len(tx.Sender) &&
		len(tx.RelayerSignature) == len(tx.Signature)
	isValidUserTxAfterBaseCostActivation := !check.IfNilReflect(userTx) && gfp.enableEpochsHandler.IsFlagEnabledInEpoch(common.FixRelayedBaseCostFlag, tx.Epoch)
	if isValidUserTxAfterBaseCostActivation && !isRelayedV3 {
		gasUsed, fee := gfp.feeComputer.ComputeGasUsedAndFeeBasedOnRefundValue(userTx, refund)
		gasUsedRelayedTx := gfp.feeComputer.ComputeGasLimit(tx)
		feeRelayedTx := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsedRelayedTx)

		tx.GasUsed = gasUsed + gasUsedRelayedTx

		fee.Add(fee, feeRelayedTx)
		tx.Fee = fee.String()

		return
	}

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
