package processProxy

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ scrCommon.TestSmartContractProcessor = (*scProcessorTestProxy)(nil)

type scProcessorTestProxy struct {
	configuredProcessor configuredProcessor
	args                scrCommon.ArgsNewSmartContractProcessor
	testScProcessor     scrCommon.TestSmartContractProcessor
	testProcessorsCache map[configuredProcessor]scrCommon.TestSmartContractProcessor
	mutRc               sync.Mutex
}

// NewTestSmartContractProcessorProxy creates a smart contract processor proxy
func NewTestSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor, epochNotifier vmcommon.EpochNotifier) (*scProcessorTestProxy, error) {
	scProcessorTestProxy := &scProcessorTestProxy{
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
			EnableRoundsHandler: args.EnableRoundsHandler,
			EnableEpochsHandler: args.EnableEpochsHandler,
			EnableEpochs:        args.EnableEpochs,
			VMOutputCacher:      args.VMOutputCacher,
			WasmVMChangeLocker:  args.WasmVMChangeLocker,
			IsGenesisProcessing: args.IsGenesisProcessing,
		},
	}

	scProcessorTestProxy.testProcessorsCache = make(map[configuredProcessor]scrCommon.TestSmartContractProcessor)

	var err error
	err = scProcessorTestProxy.createProcessorV1()
	if err != nil {
		return nil, err
	}

	err = scProcessorTestProxy.createProcessorV2()
	if err != nil {
		return nil, err
	}

	epochNotifier.RegisterNotifyHandler(scProcessorTestProxy)

	return scProcessorTestProxy, nil
}

func (proxy *scProcessorTestProxy) createProcessorV1() error {
	processor, err := smartContract.NewSmartContractProcessor(proxy.args)
	proxy.testProcessorsCache[procV1] = smartContract.NewTestScProcessor(processor)
	return err
}

func (proxy *scProcessorTestProxy) createProcessorV2() error {
	processor, err := processorV2.NewSmartContractProcessorV2(proxy.args)
	proxy.testProcessorsCache[procV2] = processorV2.NewTestScProcessor(processor)
	return err
}

func (proxy *scProcessorTestProxy) setActiveProcessorV1() {
	proxy.setActiveProcessor(procV1)
}

func (proxy *scProcessorTestProxy) setActiveProcessorV2() {
	proxy.setActiveProcessor(procV2)
}

func (proxy *scProcessorTestProxy) setActiveProcessor(version configuredProcessor) {
	log.Info("processorTestProxy", "configured", version)
	proxy.configuredProcessor = version
	proxy.testScProcessor = proxy.testProcessorsCache[version]
}

func (proxy *scProcessorTestProxy) getProcessor() process.SmartContractProcessorFacade {
	proxy.mutRc.Lock()
	defer proxy.mutRc.Unlock()
	return proxy.testScProcessor
}

// ExecuteSmartContractTransaction delegates to selected processor
func (proxy *scProcessorTestProxy) ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

// ExecuteBuiltInFunction delegates to selected processor
func (proxy *scProcessorTestProxy) ExecuteBuiltInFunction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

// DeploySmartContract delegates to selected processor
func (proxy *scProcessorTestProxy) DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().DeploySmartContract(tx, acntSrc)
}

// ProcessIfError delegates to selected processor
func (proxy *scProcessorTestProxy) ProcessIfError(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
	return proxy.getProcessor().ProcessIfError(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked)
}

// IsPayable delegates to selected processor
func (proxy *scProcessorTestProxy) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	return proxy.getProcessor().IsPayable(sndAddress, recvAddress)
}

// ProcessSmartContractResult delegates to selected processor
func (proxy *scProcessorTestProxy) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ProcessSmartContractResult(scr)
}

// CheckBuiltinFunctionIsExecutable -
func (proxy *scProcessorTestProxy) CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction string, tx data.TransactionHandler) error {
	return proxy.getProcessor().CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction, tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (proxy *scProcessorTestProxy) IsInterfaceNil() bool {
	return proxy == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (proxy *scProcessorTestProxy) EpochConfirmed(_ uint32, _ uint64) {
	proxy.mutRc.Lock()
	defer proxy.mutRc.Unlock()

	if proxy.args.EnableEpochsHandler.IsSCProcessorV2FlagEnabled() {
		proxy.setActiveProcessorV2()
		return
	}

	proxy.setActiveProcessorV1()
}

// GetCompositeTestError delegates to the selected testScProcessor
func (proxy *scProcessorTestProxy) GetCompositeTestError() error {
	return proxy.testScProcessor.GetCompositeTestError()
}

// GetGasRemaining delegates to the selected testScProcessor
func (proxy *scProcessorTestProxy) GetGasRemaining() uint64 {
	return proxy.testScProcessor.GetGasRemaining()
}

// GetAllSCRs delegates to the selected testScProcessor
func (proxy *scProcessorTestProxy) GetAllSCRs() []data.TransactionHandler {
	return proxy.testScProcessor.GetAllSCRs()
}

// CleanGasRefunded delegates to the selected testScProcessor
func (proxy *scProcessorTestProxy) CleanGasRefunded() {
	proxy.testScProcessor.CleanGasRefunded()
}

// CheckSCRBeforeProcessing delegates to the selected testScProcessor
func (proxy *scProcessorTestProxy) CheckSCRBeforeProcessing(scr *smartContractResult.SmartContractResult) (process.ScrProcessingDataHandler, error) {
	return proxy.testScProcessor.CheckSCRBeforeProcessing(scr)
}
