package txsimulator

import (
	"encoding/hex"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ArgsTxSimulator holds the arguments required for creating a new transaction simulator
type ArgsTxSimulator struct {
	TransactionProcessor      TransactionProcessor
	IntermediateProcContainer process.IntermediateProcessorContainer
	AddressPubKeyConverter    core.PubkeyConverter
	ShardCoordinator          sharding.Coordinator
	VMOutputCacher            storage.Cacher
	Hasher                    hashing.Hasher
	Marshalizer               marshal.Marshalizer
}

type transactionSimulator struct {
	mutOperation           sync.Mutex
	txProcessor            TransactionProcessor
	intermProcContainer    process.IntermediateProcessorContainer
	addressPubKeyConverter core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
	vmOutputCacher         storage.Cacher
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
}

// NewTransactionSimulator returns a new instance of a transactionSimulator
func NewTransactionSimulator(args ArgsTxSimulator) (*transactionSimulator, error) {
	if check.IfNil(args.TransactionProcessor) {
		return nil, ErrNilTxSimulatorProcessor
	}
	if check.IfNil(args.IntermediateProcContainer) {
		return nil, ErrNilIntermediateProcessorContainer
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, ErrNilPubkeyConverter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(args.VMOutputCacher) {
		return nil, ErrNilCacher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}

	return &transactionSimulator{
		txProcessor:            args.TransactionProcessor,
		intermProcContainer:    args.IntermediateProcContainer,
		addressPubKeyConverter: args.AddressPubKeyConverter,
		shardCoordinator:       args.ShardCoordinator,
		vmOutputCacher:         args.VMOutputCacher,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
	}, nil
}

// ProcessTx will process the transaction in a special environment, where state-writing is not allowed
func (ts *transactionSimulator) ProcessTx(tx *transaction.Transaction) (*txSimData.SimulationResults, error) {
	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	txStatus := transaction.TxStatusPending
	failReason := ""

	retCode, err := ts.txProcessor.ProcessTransaction(tx)
	if err != nil {
		failReason = err.Error()
		txStatus = transaction.TxStatusFail
	} else {
		if retCode == vmcommon.Ok {
			txStatus = transaction.TxStatusSuccess
		}
	}

	results := &txSimData.SimulationResults{
		Status:     txStatus,
		FailReason: failReason,
	}

	err = ts.addIntermediateTxsToResult(results)
	if err != nil {
		return nil, err
	}

	vmOutput, ok := ts.getVMOutputOfTx(tx)
	if ok {
		results.VMOutput = vmOutput

		var logs []txSimData.SimulationResultsLogEntry
		for _, vmOutputLog := range vmOutput.Logs {
			logs = append(
				logs,
				txSimData.SimulationResultsLogEntry{
					Identifier: vmOutputLog.Identifier,
					Address:    vmOutputLog.Address,
					Topics:     vmOutputLog.Topics,
					Data:       vmOutputLog.Data,
				},
			)
		}

		results.Logs = logs
	}

	return results, nil
}

func (ts *transactionSimulator) getVMOutputOfTx(tx *transaction.Transaction) (*vmcommon.VMOutput, bool) {
	txHash, err := core.CalculateHash(ts.marshalizer, ts.hasher, tx)
	if err != nil {
		return nil, false
	}

	defer ts.vmOutputCacher.Remove(txHash)

	vmOutputI, ok := ts.vmOutputCacher.Get(txHash)
	if !ok {
		return nil, false
	}

	vmOutput, ok := vmOutputI.(*vmcommon.VMOutput)
	if !ok {
		return nil, false
	}

	return vmOutput, true
}

func (ts *transactionSimulator) addIntermediateTxsToResult(result *txSimData.SimulationResults) error {
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

	receipts := make(map[string]*transaction.ApiReceipt)
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
	resScr := &transaction.ApiSmartContractResult{
		Nonce:          scr.Nonce,
		Value:          scr.Value,
		RcvAddr:        ts.addressPubKeyConverter.Encode(scr.RcvAddr),
		SndAddr:        ts.addressPubKeyConverter.Encode(scr.SndAddr),
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
	}

	if scr.OriginalSender != nil {
		resScr.OriginalSender = ts.addressPubKeyConverter.Encode(scr.OriginalSender)
	}
	if scr.RelayerAddr != nil {
		resScr.RelayerAddr = ts.addressPubKeyConverter.Encode(scr.RelayerAddr)
	}

	return resScr
}

func (ts *transactionSimulator) adaptReceipt(rcpt *receipt.Receipt) *transaction.ApiReceipt {
	return &transaction.ApiReceipt{
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
