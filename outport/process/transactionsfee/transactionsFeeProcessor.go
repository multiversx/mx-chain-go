package transactionsfee

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

const loggerName = "outport/process/transactionsfee"

// ArgTransactionsFeeProcessor holds the arguments needed for creating a new instance of transactionsFeeProcessor
type ArgTransactionsFeeProcessor struct {
	Marshaller          marshal.Marshalizer
	TransactionsStorer  storage.Storer
	ShardCoordinator    sharding.Coordinator
	TxFeeCalculator     FeesProcessorHandler
	PubKeyConverter     core.PubkeyConverter
	ArgsParser          process.ArgumentsParser
	EnableEpochsHandler common.EnableEpochsHandler
}

type transactionsFeeProcessor struct {
	txGetter            transactionGetter
	txFeeCalculator     FeesProcessorHandler
	shardCoordinator    sharding.Coordinator
	dataFieldParser     dataFieldParser
	log                 logger.Logger
	marshaller          marshal.Marshalizer
	argsParser          process.ArgumentsParser
	enableEpochsHandler common.EnableEpochsHandler
}

// NewTransactionsFeeProcessor will create a new instance of transactionsFeeProcessor
func NewTransactionsFeeProcessor(arg ArgTransactionsFeeProcessor) (*transactionsFeeProcessor, error) {
	err := checkArg(arg)
	if err != nil {
		return nil, err
	}

	parser, err := datafield.NewOperationDataFieldParser(&datafield.ArgsOperationDataFieldParser{
		AddressLength: arg.PubKeyConverter.Len(),
		Marshalizer:   arg.Marshaller,
	})
	if err != nil {
		return nil, err
	}

	return &transactionsFeeProcessor{
		txFeeCalculator:     arg.TxFeeCalculator,
		shardCoordinator:    arg.ShardCoordinator,
		txGetter:            newTxGetter(arg.TransactionsStorer, arg.Marshaller),
		log:                 logger.GetOrCreate(loggerName),
		dataFieldParser:     parser,
		marshaller:          arg.Marshaller,
		argsParser:          arg.ArgsParser,
		enableEpochsHandler: arg.EnableEpochsHandler,
	}, nil
}

func checkArg(arg ArgTransactionsFeeProcessor) error {
	if check.IfNil(arg.TransactionsStorer) {
		return ErrNilStorage
	}
	if check.IfNil(arg.ShardCoordinator) {
		return ErrNilShardCoordinator
	}
	if check.IfNil(arg.TxFeeCalculator) {
		return ErrNilTransactionFeeCalculator
	}
	if check.IfNil(arg.Marshaller) {
		return ErrNilMarshaller
	}
	if check.IfNil(arg.PubKeyConverter) {
		return core.ErrNilPubkeyConverter
	}
	if check.IfNil(arg.ArgsParser) {
		return process.ErrNilArgumentParser
	}
	if check.IfNil(arg.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}

	return nil
}

// PutFeeAndGasUsed will compute and set in transactions pool fee and gas used
func (tep *transactionsFeeProcessor) PutFeeAndGasUsed(pool *outportcore.TransactionPool, epoch uint32) error {
	tep.prepareInvalidTxs(pool)

	txsWithResultsMap := prepareTransactionsAndScrs(pool)
	tep.prepareNormalTxs(txsWithResultsMap, epoch)

	return tep.prepareScrsNoTx(txsWithResultsMap, epoch)
}

func (tep *transactionsFeeProcessor) prepareInvalidTxs(pool *outportcore.TransactionPool) {
	for _, invalidTx := range pool.InvalidTxs {
		fee := tep.txFeeCalculator.ComputeTxFee(invalidTx.Transaction)
		invalidTx.FeeInfo.SetGasUsed(invalidTx.Transaction.GetGasLimit())
		invalidTx.FeeInfo.SetFee(fee)
		invalidTx.FeeInfo.SetInitialPaidFee(fee)
	}
}

func (tep *transactionsFeeProcessor) prepareNormalTxs(transactionsAndScrs *transactionsAndScrsHolder, epoch uint32) {
	for txHashHex, txWithResult := range transactionsAndScrs.txsWithResults {
		txHandler := txWithResult.GetTxHandler()

		gasUsed := tep.txFeeCalculator.ComputeGasLimit(txHandler)
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txHandler, gasUsed)
		initialPaidFee := tep.txFeeCalculator.ComputeTxFee(txHandler)

		feeInfo := txWithResult.GetFeeInfo()
		feeInfo.SetGasUsed(gasUsed)
		feeInfo.SetFee(fee)
		feeInfo.SetInitialPaidFee(initialPaidFee)

		isRelayed := isRelayedTx(txWithResult)
		isFeeFixActive := tep.enableEpochsHandler.IsFlagEnabledInEpoch(common.FixRelayedBaseCostFlag, epoch)
		isRelayedBeforeFix := isRelayed && !isFeeFixActive
		if isRelayedBeforeFix || tep.isESDTOperationWithSCCall(txHandler) {
			feeInfo.SetGasUsed(txWithResult.GetTxHandler().GetGasLimit())
			feeInfo.SetFee(initialPaidFee)
		}

		userTx, totalFee, isRelayedV1V2 := tep.getFeeOfRelayedV1V2(txWithResult)
		isRelayedAfterFix := isRelayedV1V2 && isFeeFixActive
		if isRelayedAfterFix {
			feeInfo.SetFee(totalFee)
			feeInfo.SetInitialPaidFee(totalFee)
			feeInfo.SetGasUsed(big.NewInt(0).Div(totalFee, big.NewInt(0).SetUint64(txHandler.GetGasPrice())).Uint64())
		}

		tep.prepareTxWithResults(txHashHex, txWithResult, userTx, epoch)
	}
}

func (tep *transactionsFeeProcessor) prepareTxWithResults(
	txHashHex string,
	txWithResults *transactionWithResults,
	userTx data.TransactionHandler,
	epoch uint32,
) {
	hasRefund := false
	totalRefunds := big.NewInt(0)
	for _, scrHandler := range txWithResults.scrs {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		if isSCRForSenderWithRefund(scr, txHashHex, txWithResults.GetTxHandler()) || isRefundForRelayed(scr, txWithResults.GetTxHandler()) {
			hasRefund = true
			totalRefunds.Add(totalRefunds, scr.Value)
		}
	}

	if totalRefunds.Cmp(big.NewInt(0)) > 0 {
		tep.setGasUsedAndFeeBasedOnRefundValue(txWithResults, userTx, totalRefunds, epoch)

	}

	tep.prepareTxWithResultsBasedOnLogs(txHashHex, txWithResults, userTx, hasRefund, epoch)
}

func (tep *transactionsFeeProcessor) getFeeOfRelayedV1V2(tx *transactionWithResults) (data.TransactionHandler, *big.Int, bool) {
	if common.IsValidRelayedTxV3(tx.GetTxHandler()) {
		return nil, nil, false
	}

	if len(tx.GetTxHandler().GetData()) == 0 {
		return nil, nil, false
	}

	funcName, args, err := tep.argsParser.ParseCallData(string(tx.GetTxHandler().GetData()))
	if err != nil {
		return nil, nil, false
	}

	if funcName == core.RelayedTransaction {
		return tep.handleRelayedV1(args, tx)
	}

	if funcName == core.RelayedTransactionV2 {
		return tep.handleRelayedV2(args, tx)
	}

	return nil, nil, false
}

func (tep *transactionsFeeProcessor) handleRelayedV1(args [][]byte, tx *transactionWithResults) (data.TransactionHandler, *big.Int, bool) {
	if len(args) != 1 {
		return nil, nil, false
	}

	innerTx := &transaction.Transaction{}
	err := tep.marshaller.Unmarshal(innerTx, args[0])
	if err != nil {
		return nil, nil, false
	}

	txHandler := tx.GetTxHandler()
	gasUsed := tep.txFeeCalculator.ComputeGasLimit(txHandler)
	fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txHandler, gasUsed)

	innerFee := tep.txFeeCalculator.ComputeTxFee(innerTx)

	return innerTx, big.NewInt(0).Add(fee, innerFee), true
}

func (tep *transactionsFeeProcessor) handleRelayedV2(args [][]byte, tx *transactionWithResults) (data.TransactionHandler, *big.Int, bool) {
	txHandler := tx.GetTxHandler()

	innerTx := &transaction.Transaction{}
	innerTx.RcvAddr = args[0]
	innerTx.Nonce = big.NewInt(0).SetBytes(args[1]).Uint64()
	innerTx.Data = args[2]
	innerTx.Signature = args[3]
	innerTx.Value = big.NewInt(0)
	innerTx.GasPrice = txHandler.GetGasPrice()
	innerTx.GasLimit = txHandler.GetGasLimit() - tep.txFeeCalculator.ComputeGasLimit(txHandler)
	innerTx.SndAddr = txHandler.GetRcvAddr()

	gasUsed := tep.txFeeCalculator.ComputeGasLimit(txHandler)
	fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txHandler, gasUsed)

	innerFee := tep.txFeeCalculator.ComputeTxFee(innerTx)

	return innerTx, big.NewInt(0).Add(fee, innerFee), true
}

func (tep *transactionsFeeProcessor) prepareTxWithResultsBasedOnLogs(
	txHashHex string,
	txWithResults *transactionWithResults,
	userTx data.TransactionHandler,
	hasRefund bool,
	epoch uint32,
) {
	tx := txWithResults.GetTxHandler()
	if check.IfNil(tx) {
		tep.log.Warn("tep.prepareTxWithResultsBasedOnLogs nil transaction handler", "txHash", txHashHex)
		return
	}

	res := tep.dataFieldParser.Parse(tx.GetData(), tx.GetSndAddr(), tx.GetRcvAddr(), tep.shardCoordinator.NumberOfShards())
	if check.IfNilReflect(txWithResults.log) || (res.Function == "" && res.Operation == datafield.OperationTransfer) {
		return
	}

	for _, event := range txWithResults.log.GetLogEvents() {
		if core.WriteLogIdentifier == string(event.GetIdentifier()) && !hasRefund {
			tep.setGasUsedAndFeeBasedOnRefundValue(txWithResults, userTx, big.NewInt(0), epoch)
			continue
		}
		if core.SignalErrorOperation == string(event.GetIdentifier()) {
			fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResults.GetTxHandler(), txWithResults.GetTxHandler().GetGasLimit())
			txWithResults.GetFeeInfo().SetGasUsed(txWithResults.GetTxHandler().GetGasLimit())
			txWithResults.GetFeeInfo().SetFee(fee)
		}
	}
}

func (tep *transactionsFeeProcessor) setGasUsedAndFeeBasedOnRefundValue(
	txWithResults *transactionWithResults,
	userTx data.TransactionHandler,
	refund *big.Int,
	epoch uint32,
) {
	txWithResults.GetFeeInfo().SetHadRefund()

	isValidUserTxAfterBaseCostActivation := !check.IfNil(userTx) && tep.enableEpochsHandler.IsFlagEnabledInEpoch(common.FixRelayedBaseCostFlag, epoch)
	if isValidUserTxAfterBaseCostActivation && !common.IsValidRelayedTxV3(txWithResults.GetTxHandler()) {
		gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(userTx, refund)

		tx := txWithResults.GetTxHandler()
		gasUsedRelayedTx := tep.txFeeCalculator.ComputeGasLimit(tx)
		feeRelayedTx := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, gasUsedRelayedTx)

		txWithResults.GetFeeInfo().SetGasUsed(gasUsed + gasUsedRelayedTx)
		txWithResults.GetFeeInfo().SetFee(fee.Add(fee, feeRelayedTx))

		return
	}

	gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults.GetTxHandler(), refund)
	txWithResults.GetFeeInfo().SetGasUsed(gasUsed)
	txWithResults.GetFeeInfo().SetFee(fee)
}

func (tep *transactionsFeeProcessor) prepareScrsNoTx(transactionsAndScrs *transactionsAndScrsHolder, epoch uint32) error {
	for _, scrHandler := range transactionsAndScrs.scrsNoTx {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		scrReceiverShardID := tep.shardCoordinator.ComputeId(scr.RcvAddr)
		if scrReceiverShardID != tep.shardCoordinator.SelfId() {
			continue
		}

		if !isSCRWithRefundNoTx(scr) {
			continue
		}

		txFromStorage, err := tep.txGetter.GetTxByHash(scr.OriginalTxHash)
		if err != nil {
			tep.log.Trace("transactionsFeeProcessor.prepareScrsNoTx: cannot find transaction in storage", "hash", scr.OriginalTxHash, "error", err.Error())
			continue
		}

		isRelayedV3 := common.IsValidRelayedTxV3(txFromStorage)
		isForInitialTxSender := bytes.Equal(scr.RcvAddr, txFromStorage.SndAddr)
		isForRelayerV3 := bytes.Equal(scr.RcvAddr, txFromStorage.RelayerAddr)
		shouldSkipRelayedV3 := isRelayedV3 && !isForRelayerV3
		shouldSkipTx := !isRelayedV3 && !isForInitialTxSender || shouldSkipRelayedV3
		if shouldSkipTx {
			continue
		}

		userTx := tep.getUserTxOfRelayed(txFromStorage)
		if check.IfNil(userTx) {
			// relayed v3 and other txs
			if isRelayedV3 {
				gasUnits := tep.txFeeCalculator.ComputeGasUnitsFromRefundValue(txFromStorage, scr.Value, epoch)
				scrHandler.GetFeeInfo().SetGasRefunded(gasUnits)
				scrHandler.GetFeeInfo().SetFee(scr.Value)
				continue
			}

			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txFromStorage, scr.Value)

			scrHandler.GetFeeInfo().SetGasUsed(gasUsed)
			scrHandler.GetFeeInfo().SetFee(fee)
		} else {
			// relayed v1 and v2
			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(userTx, scr.Value)

			gasUsedRelayedTx := tep.txFeeCalculator.ComputeGasLimit(txFromStorage)
			feeRelayedTx := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txFromStorage, gasUsedRelayedTx)

			scrHandler.GetFeeInfo().SetGasUsed(gasUsed + gasUsedRelayedTx)
			scrHandler.GetFeeInfo().SetFee(fee.Add(fee, feeRelayedTx))
		}
	}

	return nil
}

func (tep *transactionsFeeProcessor) getUserTxOfRelayed(tx data.TransactionHandler) data.TransactionHandler {
	if len(tx.GetData()) == 0 {
		return nil
	}

	funcName, args, err := tep.argsParser.ParseCallData(string(tx.GetData()))
	if err != nil {
		return nil
	}

	if funcName == core.RelayedTransaction {
		if len(args) != 1 {
			return nil
		}

		userTx := &transaction.Transaction{}
		err := tep.marshaller.Unmarshal(userTx, args[0])
		if err != nil {
			return nil
		}

		return userTx
	}

	if funcName == core.RelayedTransactionV2 {
		userTx := &transaction.Transaction{}
		userTx.RcvAddr = args[0]
		userTx.Nonce = big.NewInt(0).SetBytes(args[1]).Uint64()
		userTx.Data = args[2]
		userTx.Signature = args[3]
		userTx.Value = big.NewInt(0)
		userTx.GasPrice = tx.GetGasPrice()
		userTx.GasLimit = tx.GetGasLimit() - tep.txFeeCalculator.ComputeGasLimit(tx)
		userTx.SndAddr = tx.GetRcvAddr()

		return userTx
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tep *transactionsFeeProcessor) IsInterfaceNil() bool {
	return tep == nil
}
