package processProxy

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/processorV2"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/scrCommon"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("processProxy")

var _ scrCommon.TestSmartContractProcessor = (*SCProcessorProxy)(nil)

type configuredProcessor string

const (
	procV1 configuredProcessor = "processorV1"
	procV2 configuredProcessor = "processorV2"
)

type SCProcessorProxy struct {
	configuredProcessor configuredProcessor
	args                scrCommon.ArgsNewSmartContractProcessor
	processor           process.SmartContractProcessorFacade
	processorsCache     map[configuredProcessor]process.SmartContractProcessorFacade
	testScProcessor     scrCommon.TestSmartContractProcessor
	testProcessorsCache map[configuredProcessor]scrCommon.TestSmartContractProcessor
	mutRc               sync.Mutex
}

// NewSmartContractProcessorProxy creates a smart contract processor proxy
func NewSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier) (*SCProcessorProxy, error) {
	scProcessorProxy := &SCProcessorProxy{
		args: scrCommon.ArgsNewSmartContractProcessor{
			VmContainer:         args.VmContainer,
			ArgsParser:          args.ArgsParser,
			Hasher:              args.Hasher,
			Marshalizer:         args.Marshalizer,
			AccountsDB:          args.AccountsDB,
			BlockChainHook:      args.BlockChainHook,
			BuiltInFunctions:    args.BuiltInFunctions,
			PubkeyConv:          args.PubkeyConv,
			ShardCoordinator:    args.ShardCoordinator,
			ScrForwarder:        args.ScrForwarder,
			TxFeeHandler:        args.TxFeeHandler,
			EconomicsFee:        args.EconomicsFee,
			TxTypeHandler:       args.TxTypeHandler,
			GasHandler:          args.GasHandler,
			GasSchedule:         args.GasSchedule,
			TxLogsProcessor:     args.TxLogsProcessor,
			BadTxForwarder:      args.BadTxForwarder,
			EnableEpochsHandler: args.EnableEpochsHandler,
			EnableEpochs:        args.EnableEpochs,
			VMOutputCacher:      args.VMOutputCacher,
			ArwenChangeLocker:   args.ArwenChangeLocker,
			IsGenesisProcessing: args.IsGenesisProcessing,
		},
	}

	scProcessorProxy.processorsCache = make(map[configuredProcessor]process.SmartContractProcessorFacade)
	scProcessorProxy.testProcessorsCache = make(map[configuredProcessor]scrCommon.TestSmartContractProcessor)

	var err error
	err = scProcessorProxy.createProcessorV1()
	if err != nil {
		return nil, err
	}

	err = scProcessorProxy.createProcessorV2()
	if err != nil {
		return nil, err
	}

	epochNotifier.RegisterNotifyHandler(scProcessorProxy)

	return scProcessorProxy, nil
}

func (scPProxy *SCProcessorProxy) createProcessorV1() error {
	processor, err := smartContract.NewSmartContractProcessor(scPProxy.args)
	scPProxy.processorsCache[procV1] = processor
	scPProxy.testProcessorsCache[procV1] = smartContract.NewTestScProcessor(processor)
	return err
}

func (scPProxy *SCProcessorProxy) createProcessorV2() error {
	processor, err := processorV2.NewSmartContractProcessorV2(scPProxy.args)
	scPProxy.processorsCache[procV2] = processor
	scPProxy.testProcessorsCache[procV2] = processorV2.NewTestScProcessor(processor)
	return err
}

func (scPProxy *SCProcessorProxy) setActiveProcessorV1() {
	scPProxy.setActiveProcessor(procV1)
}

func (scPProxy *SCProcessorProxy) setActiveProcessorV2() {
	scPProxy.setActiveProcessor(procV2)
}

func (scPProxy *SCProcessorProxy) setActiveProcessor(version configuredProcessor) {
	scPProxy.mutRc.Lock()
	defer scPProxy.mutRc.Unlock()
	log.Info("processorProxy", "configured", version)
	scPProxy.configuredProcessor = version
	scPProxy.processor = scPProxy.processorsCache[version]
	scPProxy.testScProcessor = scPProxy.testProcessorsCache[version]
}

// ExecuteSmartContractTransaction delegates to selected processor
func (scPProxy *SCProcessorProxy) ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

// ExecuteBuiltInFunction delegates to selected processor
func (scPProxy *SCProcessorProxy) ExecuteBuiltInFunction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

// DeploySmartContract delegates to selected processor
func (scPProxy *SCProcessorProxy) DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.DeploySmartContract(tx, acntSrc)
}

// ProcessIfError delegates to selected processor
func (scPProxy *SCProcessorProxy) ProcessIfError(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
	return scPProxy.processor.ProcessIfError(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked)
}

// IsPayable delegates to selected processor
func (scPProxy *SCProcessorProxy) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	return scPProxy.processor.IsPayable(sndAddress, recvAddress)
}

// ProcessSmartContractResult delegates to selected processor
func (scPProxy *SCProcessorProxy) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return scPProxy.processor.ProcessSmartContractResult(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scPProxy *SCProcessorProxy) IsInterfaceNil() bool {
	return scPProxy == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (scPProxy *SCProcessorProxy) EpochConfirmed(epoch uint32, _ uint64) {
	scPProxy.mutRc.Lock()
	defer scPProxy.mutRc.Unlock()

	if scPProxy.args.EnableEpochsHandler.IsSCProcessorV2FlagEnabled() {
		scPProxy.setActiveProcessorV2()
		return
	}

	scPProxy.setActiveProcessorV1()
}

// GetCompositeTestError delegates to the selected testScProcessor
func (scPProxy *SCProcessorProxy) GetCompositeTestError() error {
	return scPProxy.testScProcessor.GetCompositeTestError()
}

// GetGasRemaining delegates to the selected testScProcessor
func (scPProxy *SCProcessorProxy) GetGasRemaining() uint64 {
	return scPProxy.testScProcessor.GetGasRemaining()
}

// GetAllSCRs delegates to the selected testScProcessor
func (scPProxy *SCProcessorProxy) GetAllSCRs() []data.TransactionHandler {
	return scPProxy.testScProcessor.GetAllSCRs()
}

// CleanGasRefunded delegates to the selected testScProcessor
func (scPProxy *SCProcessorProxy) CleanGasRefunded() {
	scPProxy.testScProcessor.CleanGasRefunded()
}
