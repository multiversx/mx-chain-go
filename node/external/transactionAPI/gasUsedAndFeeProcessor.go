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
	gasScheduleNotifier core.GasScheduleNotifier
	pubKeyConverter     core.PubkeyConverter
	argsParser          process.ArgumentsParser
	marshaller          marshal.Marshalizer
}

func newGasUsedAndFeeProcessor(
	txFeeCalculator feeComputer,
	gasScheduleNotifier core.GasScheduleNotifier,
	pubKeyConverter core.PubkeyConverter,
	argsParser process.ArgumentsParser,
	marshaller marshal.Marshalizer,
) *gasUsedAndFeeProcessor {
	return &gasUsedAndFeeProcessor{
		feeComputer:         txFeeCalculator,
		gasScheduleNotifier: gasScheduleNotifier,
		pubKeyConverter:     pubKeyConverter,
		argsParser:          argsParser,
		marshaller:          marshaller,
	}
}

func (gfp *gasUsedAndFeeProcessor) computeAndAttachGasUsedAndFee(tx *transaction.ApiTransactionResult) {
	gasUsed := gfp.feeComputer.ComputeGasLimit(tx)
	fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	tx.GasUsed = gasUsed
	tx.Fee = fee.String()

	if gfp.isESDTOperationWithSCCall(tx) {
		tx.GasUsed = tx.GasLimit
		tx.Fee = tx.InitiallyPaidFee
	}

	// if there is a guardian operation, SetGuardian/GuardAccount/UnGuardAccount
	// the pre-configured cost of the operation must be added separately
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

	if tx.IsRelayed {
		totalFee, isRelayed := gfp.getFeeOfRelayed(tx)
		if isRelayed {
			tx.Fee = totalFee.String()
			tx.InitiallyPaidFee = totalFee.String()
			tx.GasUsed = big.NewInt(0).Div(totalFee, big.NewInt(0).SetUint64(tx.GasPrice)).Uint64()
		}
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

func (gfp *gasUsedAndFeeProcessor) getFeeOfRelayed(tx *transaction.ApiTransactionResult) (*big.Int, bool) {
	if !tx.IsRelayed {
		return nil, false
	}

	if len(tx.InnerTransactions) > 0 {
		return gfp.feeComputer.ComputeTransactionFee(tx), true
	}

	if len(tx.Data) == 0 {
		return nil, false
	}

	funcName, args, err := gfp.argsParser.ParseCallData(string(tx.Data))
	if err != nil {
		return nil, false
	}

	if funcName == core.RelayedTransaction {
		return gfp.handleRelayedV1(args, tx)
	}

	if funcName == core.RelayedTransactionV2 {
		return gfp.handleRelayedV2(args, tx)
	}

	return nil, false
}

func (gfp *gasUsedAndFeeProcessor) handleRelayedV1(args [][]byte, tx *transaction.ApiTransactionResult) (*big.Int, bool) {
	if len(args) != 1 {
		return nil, false
	}

	innerTx := &transaction.Transaction{}
	err := gfp.marshaller.Unmarshal(innerTx, args[0])
	if err != nil {
		return nil, false
	}

	gasUsed := gfp.feeComputer.ComputeGasLimit(tx)
	fee := gfp.feeComputer.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	innerFee := gfp.feeComputer.ComputeTransactionFee(&transaction.ApiTransactionResult{
		Tx: innerTx,
	})

	return big.NewInt(0).Add(fee, innerFee), true
}

func (gfp *gasUsedAndFeeProcessor) handleRelayedV2(args [][]byte, tx *transaction.ApiTransactionResult) (*big.Int, bool) {
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

	innerFee := gfp.feeComputer.ComputeTransactionFee(&transaction.ApiTransactionResult{
		Tx: innerTx,
	})

	return big.NewInt(0).Add(fee, innerFee), true
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
