package smartContract

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

var _ process.SCQueryService = (*SCQueryService)(nil)

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer              process.VirtualMachinesContainer
	economicsFee             process.FeeHandler
	mutRunSc                 sync.Mutex
	blockChainHook           process.BlockChainHookHandler
	blockChain               data.ChainHandler
	numQueries               int
	gasForQuery              uint64
	wasmVMChangeLocker       common.Locker
	bootstrapper             process.Bootstrapper
	allowExternalQueriesChan chan struct{}
}

// ArgsNewSCQueryService defines the arguments needed for the sc query service
type ArgsNewSCQueryService struct {
	VmContainer              process.VirtualMachinesContainer
	EconomicsFee             process.FeeHandler
	BlockChainHook           process.BlockChainHookHandler
	BlockChain               data.ChainHandler
	WasmVMChangeLocker       common.Locker
	Bootstrapper             process.Bootstrapper
	AllowExternalQueriesChan chan struct{}
	MaxGasLimitPerQuery      uint64
}

// NewSCQueryService returns a new instance of SCQueryService
func NewSCQueryService(
	args ArgsNewSCQueryService,
) (*SCQueryService, error) {
	if check.IfNil(args.VmContainer) {
		return nil, process.ErrNoVM
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.BlockChainHook) {
		return nil, process.ErrNilBlockChainHook
	}
	if check.IfNil(args.BlockChain) {
		return nil, process.ErrNilBlockChain
	}
	if check.IfNilReflect(args.WasmVMChangeLocker) {
		return nil, process.ErrNilLocker
	}
	if check.IfNil(args.Bootstrapper) {
		return nil, process.ErrNilBootstrapper
	}
	if args.AllowExternalQueriesChan == nil {
		return nil, process.ErrNilAllowExternalQueriesChan
	}

	gasForQuery := uint64(math.MaxUint64)
	if args.MaxGasLimitPerQuery > 0 {
		gasForQuery = args.MaxGasLimitPerQuery
	}
	return &SCQueryService{
		vmContainer:              args.VmContainer,
		economicsFee:             args.EconomicsFee,
		blockChain:               args.BlockChain,
		blockChainHook:           args.BlockChainHook,
		wasmVMChangeLocker:       args.WasmVMChangeLocker,
		bootstrapper:             args.Bootstrapper,
		gasForQuery:              gasForQuery,
		allowExternalQueriesChan: args.AllowExternalQueriesChan,
	}, nil
}

// ExecuteQuery returns the VMOutput resulted upon running the function on the smart contract
func (service *SCQueryService) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if !service.shouldAllowQueriesExecution() {
		return nil, process.ErrQueriesNotAllowedYet
	}

	if query.ScAddress == nil {
		return nil, process.ErrNilScAddress
	}
	if len(query.FuncName) == 0 {
		return nil, process.ErrEmptyFunctionName
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	return service.executeScCall(query, 0)
}

func (service *SCQueryService) shouldAllowQueriesExecution() bool {
	select {
	case <-service.allowExternalQueriesChan:
		return true
	default:
		return false
	}
}

func (service *SCQueryService) executeScCall(query *process.SCQuery, gasPrice uint64) (*vmcommon.VMOutput, error) {
	log.Trace("executeScCall", "function", query.FuncName, "numQueries", service.numQueries)
	service.numQueries++

	shouldEarlyExitBecauseOfSyncState := query.ShouldBeSynced && service.bootstrapper.GetNodeState() == common.NsNotSynchronized
	if shouldEarlyExitBecauseOfSyncState {
		return nil, process.ErrNodeIsNotSynced
	}

	shouldCheckRootHashChanges := query.SameScState
	rootHashBeforeExecution := make([]byte, 0)

	if shouldCheckRootHashChanges {
		rootHashBeforeExecution = service.blockChain.GetCurrentBlockRootHash()
	}

	service.blockChainHook.SetCurrentHeader(service.blockChain.GetCurrentBlockHeader())

	service.wasmVMChangeLocker.RLock()
	vm, _, err := scrCommon.FindVMByScAddress(service.vmContainer, query.ScAddress)
	if err != nil {
		service.wasmVMChangeLocker.RUnlock()
		return nil, err
	}

	query = prepareScQuery(query)
	vmInput := service.createVMCallInput(query, gasPrice)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
	service.wasmVMChangeLocker.RUnlock()
	if err != nil {
		return nil, err
	}

	if service.hasRetriableExecutionError(vmOutput) {
		log.Error("Retriable execution error detected. Will retry (once) executeScCall()", "returnCode", vmOutput.ReturnCode, "returnMessage", vmOutput.ReturnMessage)

		vmOutput, err = vm.RunSmartContractCall(vmInput)
		if err != nil {
			return nil, err
		}
	}

	if query.SameScState {
		err = service.checkForRootHashChanges(rootHashBeforeExecution)
		if err != nil {
			return nil, err
		}
	}

	return vmOutput, nil
}

func (service *SCQueryService) checkForRootHashChanges(rootHashBefore []byte) error {
	rootHashAfter := service.blockChain.GetCurrentBlockRootHash()

	if bytes.Equal(rootHashBefore, rootHashAfter) {
		return nil
	}

	return process.ErrStateChangedWhileExecutingVmQuery
}

func prepareScQuery(query *process.SCQuery) *process.SCQuery {
	if query.CallerAddr == nil {
		query.CallerAddr = query.ScAddress
	}
	if query.CallValue == nil {
		query.CallValue = big.NewInt(0)
	}

	return query
}

func (service *SCQueryService) createVMCallInput(query *process.SCQuery, gasPrice uint64) *vmcommon.ContractCallInput {
	vmInput := vmcommon.VMInput{
		CallerAddr:  query.CallerAddr,
		CallValue:   query.CallValue,
		GasPrice:    gasPrice,
		GasProvided: service.gasForQuery,
		Arguments:   query.Arguments,
		CallType:    vmData.DirectCall,
	}

	vmContractCallInput := &vmcommon.ContractCallInput{
		RecipientAddr: query.ScAddress,
		Function:      query.FuncName,
		VMInput:       vmInput,
	}

	return vmContractCallInput
}

func (service *SCQueryService) hasRetriableExecutionError(vmOutput *vmcommon.VMOutput) bool {
	return vmOutput.ReturnMessage == "allocation error"
}

// ComputeScCallGasLimit will estimate how many gas a transaction will consume
func (service *SCQueryService) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	argParser := parsers.NewCallArgsParser()

	function, arguments, err := argParser.ParseData(string(tx.Data))
	if err != nil {
		return 0, err
	}

	query := &process.SCQuery{
		ScAddress:  tx.RcvAddr,
		CallerAddr: tx.SndAddr,
		FuncName:   function,
		Arguments:  arguments,
		CallValue:  tx.Value,
	}

	service.mutRunSc.Lock()
	defer service.mutRunSc.Unlock()

	vmOutput, err := service.executeScCall(query, 1)
	if err != nil {
		return 0, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return 0, errors.New(vmOutput.ReturnMessage)
	}

	moveBalanceGasLimit := service.economicsFee.ComputeGasLimit(tx)
	gasConsumedExecution := service.gasForQuery - vmOutput.GasRemaining

	gasLimit := moveBalanceGasLimit + gasConsumedExecution

	return gasLimit, nil
}

// Close closes all underlying components
func (service *SCQueryService) Close() error {
	return service.vmContainer.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (service *SCQueryService) IsInterfaceNil() bool {
	return service == nil
}
