package processProxy

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("processProxy")

var _ process.SmartContractProcessorFacade = (*scProcessorProxy)(nil)

type configuredProcessor uint8

const (
	procV1 = iota
	procV2
)

type scProcessorProxy struct {
	configuredProcessor configuredProcessor
	args                scrCommon.ArgsNewSmartContractProcessor
	processor           process.SmartContractProcessorFacade
	processorsCache     map[configuredProcessor]process.SmartContractProcessorFacade
	mutRc               sync.RWMutex
}

// TODO -> remove the epochNotifier usage and instead extend EnableEpochsHandler
//   that will notify a new epoch *** after *** all its epoch flags are set

// NewSmartContractProcessorProxy creates a smart contract processor proxy
func NewSmartContractProcessorProxy(args scrCommon.ArgsNewSmartContractProcessor) (*scProcessorProxy, error) {
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	proxy := &scProcessorProxy{
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
			EpochNotifier:       args.EpochNotifier,
		},
	}

	proxy.processorsCache = make(map[configuredProcessor]process.SmartContractProcessorFacade)

	err := proxy.createProcessorV1()
	if err != nil {
		return nil, err
	}

	err = proxy.createProcessorV2()
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(proxy)

	return proxy, nil
}

func (proxy *scProcessorProxy) createProcessorV1() error {
	processor, err := smartContract.NewSmartContractProcessor(proxy.args)
	proxy.processorsCache[procV1] = processor
	return err
}

func (proxy *scProcessorProxy) createProcessorV2() error {
	processor, err := processorV2.NewSmartContractProcessorV2(proxy.args)
	proxy.processorsCache[procV2] = processor
	return err
}

func (proxy *scProcessorProxy) setActiveProcessorV1() {
	proxy.setActiveProcessor(procV1)
}

func (proxy *scProcessorProxy) setActiveProcessorV2() {
	proxy.setActiveProcessor(procV2)
}

func (proxy *scProcessorProxy) setActiveProcessor(version configuredProcessor) {
	log.Info("processorProxy", "configured", version)
	proxy.configuredProcessor = version
	proxy.processor = proxy.processorsCache[version]
}

func (proxy *scProcessorProxy) getProcessor() process.SmartContractProcessorFacade {
	proxy.mutRc.RLock()
	defer proxy.mutRc.RUnlock()
	return proxy.processor
}

// ExecuteSmartContractTransaction delegates to selected processor
func (proxy *scProcessorProxy) ExecuteSmartContractTransaction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
}

// ExecuteBuiltInFunction delegates to selected processor
func (proxy *scProcessorProxy) ExecuteBuiltInFunction(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ExecuteBuiltInFunction(tx, acntSrc, acntDst)
}

// DeploySmartContract delegates to selected processor
func (proxy *scProcessorProxy) DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().DeploySmartContract(tx, acntSrc)
}

// ProcessIfError delegates to selected processor
func (proxy *scProcessorProxy) ProcessIfError(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error {
	return proxy.getProcessor().ProcessIfError(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked)
}

// IsPayable delegates to selected processor
func (proxy *scProcessorProxy) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	return proxy.getProcessor().IsPayable(sndAddress, recvAddress)
}

// ProcessSmartContractResult delegates to selected processor
func (proxy *scProcessorProxy) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return proxy.getProcessor().ProcessSmartContractResult(scr)
}

// CheckBuiltinFunctionIsExecutable delegates to selected professor
func (proxy *scProcessorProxy) CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction string, tx data.TransactionHandler) error {
	return proxy.getProcessor().CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction, tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (proxy *scProcessorProxy) IsInterfaceNil() bool {
	return proxy == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (proxy *scProcessorProxy) EpochConfirmed(_ uint32, _ uint64) {
	proxy.mutRc.Lock()
	defer proxy.mutRc.Unlock()

	if proxy.args.EnableEpochsHandler.IsFlagEnabled(common.SCProcessorV2Flag) {
		proxy.setActiveProcessorV2()
		return
	}

	proxy.setActiveProcessorV1()
}
