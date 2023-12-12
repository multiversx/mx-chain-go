package transactionsfee

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
)

const loggerName = "outport/process/transactionsfee"

// ArgTransactionsFeeProcessor holds the arguments needed for creating a new instance of transactionsFeeProcessor
type ArgTransactionsFeeProcessor struct {
	Marshaller         marshal.Marshalizer
	TransactionsStorer storage.Storer
	ShardCoordinator   sharding.Coordinator
	TxFeeCalculator    FeesProcessorHandler
	PubKeyConverter    core.PubkeyConverter
}

type transactionsFeeProcessor struct {
	txGetter         transactionGetter
	txFeeCalculator  FeesProcessorHandler
	shardCoordinator sharding.Coordinator
	dataFieldParser  dataFieldParser
	log              logger.Logger
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
		txFeeCalculator:  arg.TxFeeCalculator,
		shardCoordinator: arg.ShardCoordinator,
		txGetter:         newTxGetter(arg.TransactionsStorer, arg.Marshaller),
		log:              logger.GetOrCreate(loggerName),
		dataFieldParser:  parser,
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

	return nil
}

// PutFeeAndGasUsed will compute and set in transactions pool fee and gas used
func (tep *transactionsFeeProcessor) PutFeeAndGasUsed(pool *outportcore.TransactionPool) error {
	tep.prepareInvalidTxs(pool)

	txsWithResultsMap := prepareTransactionsAndScrs(pool)
	tep.prepareNormalTxs(txsWithResultsMap)

	return tep.prepareScrsNoTx(txsWithResultsMap)
}

func (tep *transactionsFeeProcessor) prepareInvalidTxs(pool *outportcore.TransactionPool) {
	for _, invalidTx := range pool.InvalidTxs {
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(invalidTx.Transaction, invalidTx.Transaction.GasLimit)
		invalidTx.FeeInfo.SetGasUsed(invalidTx.Transaction.GetGasLimit())
		invalidTx.FeeInfo.SetFee(fee)
		invalidTx.FeeInfo.SetInitialPaidFee(fee)
	}
}

func (tep *transactionsFeeProcessor) prepareNormalTxs(transactionsAndScrs *transactionsAndScrsHolder) {
	for txHashHex, txWithResult := range transactionsAndScrs.txsWithResults {
		txHandler := txWithResult.GetTxHandler()

		gasUsed := tep.txFeeCalculator.ComputeGasLimit(txHandler)
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txHandler, gasUsed)
		initialPaidFee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txHandler, txHandler.GetGasLimit())

		feeInfo := txWithResult.GetFeeInfo()
		feeInfo.SetGasUsed(gasUsed)
		feeInfo.SetFee(fee)
		feeInfo.SetInitialPaidFee(initialPaidFee)

		if isRelayedTx(txWithResult) || tep.isESDTOperationWithSCCall(txHandler) {
			feeInfo.SetGasUsed(txWithResult.GetTxHandler().GetGasLimit())
			feeInfo.SetFee(initialPaidFee)
		}

		tep.prepareTxWithResults(txHashHex, txWithResult)
	}
}

func (tep *transactionsFeeProcessor) prepareTxWithResults(txHashHex string, txWithResults *transactionWithResults) {
	hasRefund := false
	for _, scrHandler := range txWithResults.scrs {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		if isSCRForSenderWithRefund(scr, txHashHex, txWithResults.GetTxHandler()) || isRefundForRelayed(scr, txWithResults.GetTxHandler()) {
			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults.GetTxHandler(), scr.Value)

			txWithResults.GetFeeInfo().SetGasUsed(gasUsed)
			txWithResults.GetFeeInfo().SetFee(fee)
			hasRefund = true
			break
		}
	}

	tep.prepareTxWithResultsBasedOnLogs(txWithResults, hasRefund)

}

func (tep *transactionsFeeProcessor) prepareTxWithResultsBasedOnLogs(
	txWithResults *transactionWithResults,
	hasRefund bool,
) {
	tx := txWithResults.GetTxHandler()
	res := tep.dataFieldParser.Parse(tx.GetData(), tx.GetSndAddr(), tx.GetRcvAddr(), tep.shardCoordinator.NumberOfShards())

	if check.IfNilReflect(txWithResults.log) || (res.Function == "" && res.Operation == "transfer") {
		return
	}

	for _, event := range txWithResults.log.GetLogEvents() {
		if core.WriteLogIdentifier == string(event.GetIdentifier()) && !hasRefund {
			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults.GetTxHandler(), big.NewInt(0))
			txWithResults.GetFeeInfo().SetGasUsed(gasUsed)
			txWithResults.GetFeeInfo().SetFee(fee)

			continue
		}
		if core.SignalErrorOperation == string(event.GetIdentifier()) {
			fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResults.GetTxHandler(), txWithResults.GetTxHandler().GetGasLimit())
			txWithResults.GetFeeInfo().SetGasUsed(txWithResults.GetTxHandler().GetGasLimit())
			txWithResults.GetFeeInfo().SetFee(fee)
		}
	}

}

func (tep *transactionsFeeProcessor) prepareScrsNoTx(transactionsAndScrs *transactionsAndScrsHolder) error {
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

		isForInitialTxSender := bytes.Equal(scr.RcvAddr, txFromStorage.SndAddr)
		if !isForInitialTxSender {
			continue
		}

		gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txFromStorage, scr.Value)

		scrHandler.GetFeeInfo().SetGasUsed(gasUsed)
		scrHandler.GetFeeInfo().SetFee(fee)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tep *transactionsFeeProcessor) IsInterfaceNil() bool {
	return tep == nil
}
