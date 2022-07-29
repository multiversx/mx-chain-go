package transactionsfee

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	writeLogOperation    = "writeLog"
	signalErrorOperation = "signalError"
)

// ArgTransactionsFeeProcessor holds the arguments needed for creating a new instance of transactionsFeeProcessor
type ArgTransactionsFeeProcessor struct {
	Marshaller         marshal.Marshalizer
	TransactionsStorer storage.Storer
	ShardCoordinator   sharding.Coordinator
	TxFeeCalculator    FeesProcessorHandler
}

type transactionsFeeProcessor struct {
	txGetter         *txGetter
	txFeeCalculator  FeesProcessorHandler
	shardCoordinator sharding.Coordinator
}

// NewTransactionFeeProcessor will create a new instance of transactionsFeeProcessor
func NewTransactionFeeProcessor(arg ArgTransactionsFeeProcessor) (*transactionsFeeProcessor, error) {
	err := checkArg(arg)
	if err != nil {
		return nil, err
	}

	return &transactionsFeeProcessor{
		txFeeCalculator:  arg.TxFeeCalculator,
		shardCoordinator: arg.ShardCoordinator,
		txGetter:         newTxGetter(arg.TransactionsStorer, arg.Marshaller),
	}, nil
}

func (tep *transactionsFeeProcessor) PutFeeAndGasUsed(pool *outportcore.Pool) error {
	tep.prepareInvalidTxs(pool)

	txsWithResultsMap := prepareTransactionsAndScrs(pool)
	tep.prepareNormalTxs(txsWithResultsMap)

	return tep.prepareScrsNoTx(txsWithResultsMap)
}

func (tep *transactionsFeeProcessor) prepareInvalidTxs(pool *outportcore.Pool) {
	for _, invalidTx := range pool.Invalid {
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(invalidTx, invalidTx.GetGasLimit())
		invalidTx.SetGasUsed(invalidTx.GetGasLimit())
		invalidTx.SetFee(fee)
	}
}

func (tep *transactionsFeeProcessor) prepareNormalTxs(transactionsAndScrs *transactionsAndScrsHolder) {
	for txHash, txWithResult := range transactionsAndScrs.txsWithResults {
		gasUsed := tep.txFeeCalculator.ComputeGasLimit(txWithResult)
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResult, gasUsed)

		txWithResult.SetGasUsed(gasUsed)
		txWithResult.SetFee(fee)

		tep.prepareTxWithResults([]byte(txHash), txWithResult)
	}
}

func (tep *transactionsFeeProcessor) prepareTxWithResults(txHash []byte, txWithResults *transactionWithResults) {
	txHashHasRefund := make(map[string]struct{})
	for _, scrHandler := range txWithResults.scrs {
		scr, ok := scrHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		if isSCRForSenderWithRefund(scr, txHash, txWithResults) || isRefundForRelayed(scr, txWithResults) {
			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults, scr.Value)

			txWithResults.SetGasUsed(gasUsed)
			txWithResults.SetFee(fee)
			txHashHasRefund[string(txHash)] = struct{}{}
		}
	}

	if check.IfNil(txWithResults.logs) {
		return
	}

	for _, event := range txWithResults.logs.GetLogEvents() {
		identifier := string(event.GetIdentifier())
		switch identifier {
		case signalErrorOperation:
			fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResults, txWithResults.GetGasLimit())
			txWithResults.SetGasUsed(txWithResults.GetGasLimit())
			txWithResults.SetFee(fee)
			return
		case writeLogOperation:
			_, found := txHashHasRefund[string(txHash)]
			if !found {
				return
			}

			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults, big.NewInt(0))
			txWithResults.SetGasUsed(gasUsed)
			txWithResults.SetFee(fee)
		default:
			continue
		}
	}
}

func (tep *transactionsFeeProcessor) prepareScrsNoTx(transactionsAndScrs *transactionsAndScrsHolder) error {
	for _, scrHandler := range transactionsAndScrs.scrsNoTx {
		scrTxHandler, ok := scrHandler.(data.TransactionHandler)
		if !ok {
			continue
		}
		scr, ok := scrTxHandler.(*smartContractResult.SmartContractResult)
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

		txFromStorage, err := tep.txGetter.getTxByHash(scr.OriginalTxHash)
		if err != nil {
			return err
		}

		gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txFromStorage, scr.Value)

		scrHandler.SetGasUsed(gasUsed)
		scrHandler.SetFee(fee)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tep *transactionsFeeProcessor) IsInterfaceNil() bool {
	return tep == nil
}
