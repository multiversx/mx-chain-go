package smartContract

import (
	"errors"
	"math"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmData "github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

var _ process.SCQueryService = (*SCQueryService)(nil)

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer       process.VirtualMachinesContainer
	economicsFee      process.FeeHandler
	mutRunSc          sync.Mutex
	blockChainHook    process.BlockChainHookHandler
	blockChain        data.ChainHandler
	numQueries        int
	gasForQuery       uint64
	arwenChangeLocker process.Locker
}

// ArgsNewSCQueryService defines the arguments needed for the sc query service
type ArgsNewSCQueryService struct {
	VmContainer       process.VirtualMachinesContainer
	EconomicsFee      process.FeeHandler
	BlockChainHook    process.BlockChainHookHandler
	BlockChain        data.ChainHandler
	ArwenChangeLocker process.Locker
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
	if check.IfNilReflect(args.ArwenChangeLocker) {
		return nil, process.ErrNilLocker
	}

	return &SCQueryService{
		vmContainer:       args.VmContainer,
		economicsFee:      args.EconomicsFee,
		blockChain:        args.BlockChain,
		blockChainHook:    args.BlockChainHook,
		arwenChangeLocker: args.ArwenChangeLocker,
		gasForQuery:       math.MaxUint64,
	}, nil
}

// ExecuteQuery returns the VMOutput resulted upon running the function on the smart contract
func (service *SCQueryService) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
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

func (service *SCQueryService) executeScCall(query *process.SCQuery, gasPrice uint64) (*vmcommon.VMOutput, error) {
	log.Trace("executeScCall", "function", query.FuncName, "numQueries", service.numQueries)
	service.numQueries++

	service.blockChainHook.SetCurrentHeader(service.blockChain.GetCurrentBlockHeader())

	service.arwenChangeLocker.RLock()
	vm, err := findVMByScAddress(service.vmContainer, query.ScAddress)
	if err != nil {
		service.arwenChangeLocker.RUnlock()
		return nil, err
	}

	query = prepareScQuery(query)
	vmInput := service.createVMCallInput(query, gasPrice)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
	service.arwenChangeLocker.RUnlock()
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

	return vmOutput, nil
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
