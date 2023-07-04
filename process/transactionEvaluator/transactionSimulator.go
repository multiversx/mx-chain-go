package transactionEvaluator

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
	"github.com/multiversx/mx-chain-go/node/external/transactionAPI"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("process/txSimulator")

// ArgsTxSimulator holds the arguments required for creating a new transaction simulator
type ArgsTxSimulator struct {
	TransactionProcessor      TransactionProcessor
	IntermediateProcContainer process.IntermediateProcessorContainer
	AddressPubKeyConverter    core.PubkeyConverter
	ShardCoordinator          sharding.Coordinator
	VMOutputCacher            storage.Cacher
	Hasher                    hashing.Hasher
	Marshalizer               marshal.Marshalizer
	DataFieldParser           DataFieldParser
}

type refundHandler interface {
	IsRefund(input transactionAPI.RefundDetectorInput) bool
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
	refundDetector         refundHandler
	dataFieldParser        DataFieldParser
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
	if check.IfNilReflect(args.DataFieldParser) {
		return nil, ErrNilDataFieldParser
	}

	return &transactionSimulator{
		txProcessor:            args.TransactionProcessor,
		intermProcContainer:    args.IntermediateProcContainer,
		addressPubKeyConverter: args.AddressPubKeyConverter,
		shardCoordinator:       args.ShardCoordinator,
		vmOutputCacher:         args.VMOutputCacher,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		refundDetector:         transactionAPI.NewRefundDetector(),
		dataFieldParser:        args.DataFieldParser,
	}, nil
}

// ProcessTx will process the transaction in a special environment, where state-writing is not allowed
func (ts *transactionSimulator) ProcessTx(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
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

	results := &txSimData.SimulationResultsWithVMOutput{
		SimulationResults: transaction.SimulationResults{
			Status:     txStatus,
			FailReason: failReason,
		},
	}

	err = ts.addIntermediateTxsToResult(results)
	if err != nil {
		return nil, err
	}

	vmOutput, ok := ts.getVMOutputOfTx(tx)
	if ok {
		results.VMOutput = vmOutput
	}

	ts.addLogsFromVmOutput(results, vmOutput)

	return results, nil
}

func (ts *transactionSimulator) addLogsFromVmOutput(results *txSimData.SimulationResultsWithVMOutput, vmOutput *vmcommon.VMOutput) {
	if vmOutput == nil || len(vmOutput.Logs) == 0 {
		return
	}

	results.Logs = &transaction.ApiLogs{
		Events: make([]*transaction.Events, 0, len(vmOutput.Logs)),
	}

	for _, entry := range vmOutput.Logs {
		results.Logs.Events = append(results.Logs.Events, &transaction.Events{
			Address:    ts.addressPubKeyConverter.SilentEncode(entry.Address, log),
			Identifier: string(entry.Identifier),
			Topics:     entry.Topics,
			Data:       entry.Data,
		})
	}
}

func (ts *transactionSimulator) getVMOutputOfTx(tx *transaction.Transaction) (*vmcommon.VMOutput, bool) {
	txHash, err := core.CalculateHash(ts.marshalizer, ts.hasher, tx)
	if err != nil {
		return nil, false
	}

	defer ts.vmOutputCacher.Remove(txHash)

	vmOutputInterface, ok := ts.vmOutputCacher.Get(txHash)
	if !ok || check.IfNilReflect(vmOutputInterface) {
		return nil, false
	}

	vmOutput, ok := vmOutputInterface.(*vmcommon.VMOutput)
	if !ok {
		return nil, false
	}

	return vmOutput, true
}

func (ts *transactionSimulator) addIntermediateTxsToResult(result *txSimData.SimulationResultsWithVMOutput) error {
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
	isRefund := ts.refundDetector.IsRefund(transactionAPI.RefundDetectorInput{
		Value:         scr.Value.String(),
		Data:          scr.Data,
		ReturnMessage: string(scr.ReturnMessage),
		GasLimit:      scr.GasLimit,
	})
	res := ts.dataFieldParser.Parse(scr.Data, scr.SndAddr, scr.RcvAddr, ts.shardCoordinator.NumberOfShards())

	receiversEncoded, err := ts.addressPubKeyConverter.EncodeSlice(res.Receivers)
	if err != nil {
		log.Warn("transactionSimulator.adaptSmartContractResult: cannot encode receivers slice", "error", err)
	}

	resScr := &transaction.ApiSmartContractResult{
		Nonce:             scr.Nonce,
		Value:             scr.Value,
		RcvAddr:           ts.addressPubKeyConverter.SilentEncode(scr.RcvAddr, log),
		SndAddr:           ts.addressPubKeyConverter.SilentEncode(scr.SndAddr, log),
		RelayedValue:      scr.RelayedValue,
		Code:              hex.EncodeToString(scr.Code),
		Data:              string(scr.Data),
		PrevTxHash:        hex.EncodeToString(scr.PrevTxHash),
		OriginalTxHash:    hex.EncodeToString(scr.OriginalTxHash),
		GasLimit:          scr.GasLimit,
		GasPrice:          scr.GasPrice,
		CallType:          scr.CallType,
		CodeMetadata:      hex.EncodeToString(scr.CodeMetadata),
		ReturnMessage:     string(scr.ReturnMessage),
		IsRefund:          isRefund,
		Operation:         res.Operation,
		Function:          res.Function,
		ESDTValues:        res.ESDTValues,
		Tokens:            res.Tokens,
		Receivers:         receiversEncoded,
		ReceiversShardIDs: res.ReceiversShardID,
		IsRelayed:         res.IsRelayed,
	}

	if scr.OriginalSender != nil {
		resScr.OriginalSender = ts.addressPubKeyConverter.SilentEncode(scr.OriginalSender, log)
	}

	if scr.RelayerAddr != nil {
		resScr.RelayerAddr = ts.addressPubKeyConverter.SilentEncode(scr.RelayerAddr, log)
	}

	return resScr
}

func (ts *transactionSimulator) adaptReceipt(rcpt *receipt.Receipt) *transaction.ApiReceipt {
	receiptSenderAddress := ts.addressPubKeyConverter.SilentEncode(rcpt.SndAddr, log)

	return &transaction.ApiReceipt{
		Value:   rcpt.Value,
		SndAddr: receiptSenderAddress,
		Data:    string(rcpt.Data),
		TxHash:  hex.EncodeToString(rcpt.TxHash),
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *transactionSimulator) IsInterfaceNil() bool {
	return ts == nil
}
