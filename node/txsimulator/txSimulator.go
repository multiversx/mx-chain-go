package txsimulator

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgsTxSimulator holds the arguments required for creating a new transaction simulator
type ArgsTxSimulator struct {
	TransactionProcessor       TransactionProcessor
	IntermmediateProcContainer process.IntermediateProcessorContainer
	AddressPubKeyConverter     core.PubkeyConverter
	ShardCoordinator           sharding.Coordinator
}

type transactionSimulator struct {
	txProcessor            TransactionProcessor
	intermProcContainer    process.IntermediateProcessorContainer
	addressPubKeyConverter core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
}

// NewTransactionSimulator returns a new instance of a transactionSimulator
func NewTransactionSimulator(args ArgsTxSimulator) (*transactionSimulator, error) {
	if check.IfNil(args.TransactionProcessor) {
		return nil, node.ErrNilTxSimulatorProcessor
	}
	if check.IfNil(args.IntermmediateProcContainer) {
		return nil, node.ErrNilIntermediateProcessorContainer
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, node.ErrNilPubkeyConverter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, node.ErrNilShardCoordinator
	}

	return &transactionSimulator{
		txProcessor:            args.TransactionProcessor,
		intermProcContainer:    args.IntermmediateProcContainer,
		addressPubKeyConverter: args.AddressPubKeyConverter,
		shardCoordinator:       args.ShardCoordinator,
	}, nil
}

// ProcessTx will process the transaction in a special environment, where state-writing is not allowed
func (ts *transactionSimulator) ProcessTx(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
	txStatus := transaction.TxStatusPending
	failReason := ""
	// TODO check return code also here, 0, error -> TxStatusSuccess and failReason not nil!
	retCode, err := ts.txProcessor.ProcessTransaction(tx)
	if err != nil {
		failReason = err.Error()
		txStatus = transaction.TxStatusFail
	}

	if retCode == vmcommon.Ok {
		txStatus = transaction.TxStatusSuccess
	}

	results := &transaction.SimulationResults{
		Status:     txStatus,
		FailReason: failReason,
	}

	err = ts.addIntermediateTxsToResult(results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (ts *transactionSimulator) addIntermediateTxsToResult(result *transaction.SimulationResults) error {
	defer func() {
		processorsKeys := ts.intermProcContainer.Keys()
		for _, procKey := range processorsKeys {
			processor, errGetProc := ts.intermProcContainer.Get(procKey)
			if errGetProc != nil || processor == nil {
				continue
			}

			processor.CreateBlockStarted()
		}
	}()

	scrForwarder, err := ts.intermProcContainer.Get(block.SmartContractResultBlock)
	if err != nil {
		return err
	}

	scResults := make(map[string]*transaction.ApiSmartContractResult)
	for hash, value := range scrForwarder.GetAllCurrentFinishedTxs() {
		scr, ok := value.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		scResults[hex.EncodeToString([]byte(hash))] = ts.adaptSmartContractResult(scr)
	}
	result.ScResults = scResults

	if ts.shardCoordinator.SelfId() == core.MetachainShardId {
		return nil
	}

	receiptsForwarder, err := ts.intermProcContainer.Get(block.ReceiptBlock)
	if err != nil {
		return err
	}

	receipts := make(map[string]*transaction.ReceiptApi)
	for hash, value := range receiptsForwarder.GetAllCurrentFinishedTxs() {
		rcpt, ok := value.(*receipt.Receipt)
		if !ok {
			continue
		}
		receipts[hex.EncodeToString([]byte(hash))] = ts.adaptReceipt(rcpt)
	}
	result.Receipts = receipts

	return nil
}

func (ts *transactionSimulator) adaptSmartContractResult(scr *smartContractResult.SmartContractResult) *transaction.ApiSmartContractResult {
	return &transaction.ApiSmartContractResult{
		Nonce:          scr.Nonce,
		Value:          scr.Value,
		RcvAddr:        ts.addressPubKeyConverter.Encode(scr.RcvAddr),
		SndAddr:        ts.addressPubKeyConverter.Encode(scr.SndAddr),
		RelayerAddr:    ts.addressPubKeyConverter.Encode(scr.RelayerAddr),
		RelayedValue:   scr.RelayedValue,
		Code:           string(scr.Code),
		Data:           string(scr.Data),
		PrevTxHash:     hex.EncodeToString(scr.PrevTxHash),
		OriginalTxHash: hex.EncodeToString(scr.OriginalTxHash),
		GasLimit:       scr.GasLimit,
		GasPrice:       scr.GasPrice,
		CallType:       scr.CallType,
		CodeMetadata:   string(scr.CodeMetadata),
		ReturnMessage:  string(scr.ReturnMessage),
		OriginalSender: ts.addressPubKeyConverter.Encode(scr.OriginalSender),
	}
}

func (ts *transactionSimulator) adaptReceipt(rcpt *receipt.Receipt) *transaction.ReceiptApi {
	return &transaction.ReceiptApi{
		Value:   rcpt.Value,
		SndAddr: ts.addressPubKeyConverter.Encode(rcpt.SndAddr),
		Data:    string(rcpt.Data),
		TxHash:  hex.EncodeToString(rcpt.TxHash),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *transactionSimulator) IsInterfaceNil() bool {
	return ts == nil
}
