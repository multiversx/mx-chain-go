package transactionsfee

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgTransactionsFeeProcessor holds the arguments needed for creating a new instance of transactionsFeeProcessor
type ArgTransactionsFeeProcessor struct {
	Marshaller         marshal.Marshalizer
	TransactionsStorer storage.Storer
	ShardCoordinator   sharding.Coordinator
	TxFeeCalculator    FeesProcessorHandler
}

type transactionsFeeProcessor struct {
	txGetter         transactionGetter
	txFeeCalculator  FeesProcessorHandler
	shardCoordinator sharding.Coordinator
}

// NewTransactionsFeeProcessor will create a new instance of transactionsFeeProcessor
func NewTransactionsFeeProcessor(arg ArgTransactionsFeeProcessor) (*transactionsFeeProcessor, error) {
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

	return nil
}

// PutFeeAndGasUsed will compute and set in transactions pool fee and gas used
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
		invalidTx.SetInitialPaidFee(fee)
	}
}

func (tep *transactionsFeeProcessor) prepareNormalTxs(transactionsAndScrs *transactionsAndScrsHolder) {
	for txHash, txWithResult := range transactionsAndScrs.txsWithResults {
		gasUsed := tep.txFeeCalculator.ComputeGasLimit(txWithResult)
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResult, gasUsed)
		initialPaidFee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResult, txWithResult.GetGasLimit())

		txWithResult.SetGasUsed(gasUsed)
		txWithResult.SetFee(fee)
		txWithResult.SetInitialPaidFee(initialPaidFee)

		tep.prepareTxWithResults([]byte(txHash), txWithResult)
	}
}

func (tep *transactionsFeeProcessor) prepareTxWithResults(txHash []byte, txWithResults *transactionWithResults) {
	hasRefund := false
	for _, scrHandler := range txWithResults.scrs {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		if isSCRForSenderWithRefund(scr, txHash, txWithResults) || isRefundForRelayed(scr, txWithResults) {
			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults, scr.Value)

			txWithResults.SetGasUsed(gasUsed)
			txWithResults.SetFee(fee)
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
	if check.IfNilReflect(txWithResults.log) {
		return
	}

	for _, event := range txWithResults.log.GetLogEvents() {
		switch {
		case core.WriteLogIdentifier == string(event.GetIdentifier()) && !hasRefund:
			{
				gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults, big.NewInt(0))
				txWithResults.SetGasUsed(gasUsed)
				txWithResults.SetFee(fee)
			}
		case core.SignalErrorOperation == string(event.GetIdentifier()):
			{
				fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResults, txWithResults.GetGasLimit())
				txWithResults.SetGasUsed(txWithResults.GetGasLimit())
				txWithResults.SetFee(fee)
			}
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
