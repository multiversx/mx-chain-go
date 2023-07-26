package smartContract

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

var _ process.SCQueryService = (*SCQueryService)(nil)

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer                  process.VirtualMachinesContainer
	economicsFee                 process.FeeHandler
	mutRunSc                     sync.Mutex
	blockChainHook               process.BlockChainHookHandler
	blockChain                   data.ChainHandler
	numQueries                   int
	gasForQuery                  uint64
	wasmVMChangeLocker           common.Locker
	bootstrapper                 process.Bootstrapper
	allowExternalQueriesChan     chan struct{}
	historyRepository            dblookupext.HistoryRepository
	shardCoordinator             sharding.Coordinator
	storageService               dataRetriever.StorageService
	marshaller                   marshal.Marshalizer
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
}

// ArgsNewSCQueryService defines the arguments needed for the sc query service
type ArgsNewSCQueryService struct {
	VmContainer                  process.VirtualMachinesContainer
	EconomicsFee                 process.FeeHandler
	BlockChainHook               process.BlockChainHookHandler
	BlockChain                   data.ChainHandler
	WasmVMChangeLocker           common.Locker
	Bootstrapper                 process.Bootstrapper
	AllowExternalQueriesChan     chan struct{}
	MaxGasLimitPerQuery          uint64
	HistoryRepository            dblookupext.HistoryRepository
	ShardCoordinator             sharding.Coordinator
	StorageService               dataRetriever.StorageService
	Marshaller                   marshal.Marshalizer
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	Uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
}

// NewSCQueryService returns a new instance of SCQueryService
func NewSCQueryService(
	args ArgsNewSCQueryService,
) (*SCQueryService, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gasForQuery := uint64(math.MaxUint64)
	if args.MaxGasLimitPerQuery > 0 {
		gasForQuery = args.MaxGasLimitPerQuery
	}
	return &SCQueryService{
		vmContainer:                  args.VmContainer,
		economicsFee:                 args.EconomicsFee,
		blockChain:                   args.BlockChain,
		blockChainHook:               args.BlockChainHook,
		wasmVMChangeLocker:           args.WasmVMChangeLocker,
		bootstrapper:                 args.Bootstrapper,
		gasForQuery:                  gasForQuery,
		allowExternalQueriesChan:     args.AllowExternalQueriesChan,
		historyRepository:            args.HistoryRepository,
		shardCoordinator:             args.ShardCoordinator,
		storageService:               args.StorageService,
		marshaller:                   args.Marshaller,
		scheduledTxsExecutionHandler: args.ScheduledTxsExecutionHandler,
		uint64ByteSliceConverter:     args.Uint64ByteSliceConverter,
	}, nil
}

func checkArgs(args ArgsNewSCQueryService) error {
	if check.IfNil(args.VmContainer) {
		return process.ErrNoVM
	}
	if check.IfNil(args.EconomicsFee) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.BlockChainHook) {
		return process.ErrNilBlockChainHook
	}
	if check.IfNil(args.BlockChain) {
		return process.ErrNilBlockChain
	}
	if check.IfNilReflect(args.WasmVMChangeLocker) {
		return process.ErrNilLocker
	}
	if check.IfNil(args.Bootstrapper) {
		return process.ErrNilBootstrapper
	}
	if args.AllowExternalQueriesChan == nil {
		return process.ErrNilAllowExternalQueriesChan
	}
	if check.IfNil(args.HistoryRepository) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.StorageService) {
		return process.ErrNilStorageService
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ScheduledTxsExecutionHandler) {
		return process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return process.ErrNilUint64Converter
	}

	return nil
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

	accountsAdapter := service.blockChainHook.GetAccountsAdapter()
	blockRootHash, err := service.extractBlockRootHash(query)
	if err != nil {
		return nil, err
	}
	if len(blockRootHash) > 0 {
		err = accountsAdapter.RecreateTrie(blockRootHash)
		if err != nil {
			return nil, err
		}

		// Temporary setting the root hash to the desired one until the sc call is ready
		_ = service.blockChain.SetCurrentBlockHeaderAndRootHash(nil, blockRootHash)
	}

	query = prepareScQuery(query)
	vmInput := service.createVMCallInput(query, gasPrice)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
	service.wasmVMChangeLocker.RUnlock()
	if err != nil {
		// Cleaning the current root hash so the real one would be returned further
		_ = service.blockChain.SetCurrentBlockHeaderAndRootHash(nil, nil)
		return nil, err
	}

	if service.hasRetriableExecutionError(vmOutput) {
		log.Error("Retriable execution error detected. Will retry (once) executeScCall()", "returnCode", vmOutput.ReturnCode, "returnMessage", vmOutput.ReturnMessage)

		vmOutput, err = vm.RunSmartContractCall(vmInput)
		if err != nil {
			// Cleaning the current root hash so the real one would be returned further
			_ = service.blockChain.SetCurrentBlockHeaderAndRootHash(nil, nil)
			return nil, err
		}
	}

	// Cleaning the current root hash so the real one would be returned further
	_ = service.blockChain.SetCurrentBlockHeaderAndRootHash(nil, nil)

	if query.SameScState {
		err = service.checkForRootHashChanges(rootHashBeforeExecution)
		if err != nil {
			return nil, err
		}
	}

	return vmOutput, nil
}

func (service *SCQueryService) extractBlockRootHash(query *process.SCQuery) ([]byte, error) {
	if len(query.BlockRootHash) > 0 {
		return query.BlockRootHash, nil
	}

	if len(query.BlockHash) > 0 {
		blockHeader, err := service.getBlockHeaderByHash(query.BlockHash)
		if err != nil {
			return make([]byte, 0), err
		}

		return service.getBlockRootHash(query.BlockHash, blockHeader), nil
	}

	if query.BlockNonce.HasValue {
		blockHeader, blockHash, err := service.getBlockHeaderByNonce(query.BlockNonce.Value)
		if err != nil {
			return make([]byte, 0), err
		}

		return service.getBlockRootHash(blockHash, blockHeader), nil
	}

	return make([]byte, 0), nil
}

func (service *SCQueryService) getBlockHeaderByHash(headerHash []byte) (data.HeaderHandler, error) {
	epoch, err := service.getOptionalEpochByHash(headerHash)
	if err != nil {
		return nil, err
	}

	header, err := service.getBlockHeaderInEpochByHash(headerHash, epoch)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (service *SCQueryService) getOptionalEpochByHash(hash []byte) (core.OptionalUint32, error) {
	if !service.historyRepository.IsEnabled() {
		return core.OptionalUint32{}, nil
	}

	epoch, err := service.historyRepository.GetEpochByHash(hash)
	if err != nil {
		return core.OptionalUint32{}, err
	}

	return core.OptionalUint32{Value: epoch, HasValue: true}, nil
}

func (service *SCQueryService) getBlockHeaderInEpochByHash(headerHash []byte, epoch core.OptionalUint32) (data.HeaderHandler, error) {
	shardId := service.shardCoordinator.SelfId()
	unitType := dataRetriever.GetHeadersDataUnit(shardId)
	storer, err := service.storageService.GetStorer(unitType)
	if err != nil {
		return nil, err
	}

	var headerBuffer []byte

	if epoch.HasValue {
		headerBuffer, err = storer.GetFromEpoch(headerHash, epoch.Value)
	} else {
		headerBuffer, err = storer.Get(headerHash)
	}
	if err != nil {
		return nil, err
	}

	header, err := process.UnmarshalHeader(shardId, service.marshaller, headerBuffer)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (service *SCQueryService) getBlockRootHash(headerHash []byte, header data.HeaderHandler) []byte {
	blockRootHash, err := service.scheduledTxsExecutionHandler.GetScheduledRootHashForHeaderWithEpoch(
		headerHash,
		header.GetEpoch())
	if err != nil {
		blockRootHash = header.GetRootHash()
	}

	return blockRootHash
}

func (service *SCQueryService) getBlockHeaderByNonce(nonce uint64) (data.HeaderHandler, []byte, error) {
	headerHash, err := service.getBlockHashByNonce(nonce)
	if err != nil {
		return nil, nil, err
	}

	header, err := service.getBlockHeaderByHash(headerHash)
	if err != nil {
		return nil, nil, err
	}

	return header, headerHash, nil
}

func (service *SCQueryService) getBlockHashByNonce(nonce uint64) ([]byte, error) {
	shardId := service.shardCoordinator.SelfId()
	hashByNonceUnit := dataRetriever.GetHdrNonceHashDataUnit(shardId)

	return process.GetHeaderHashFromStorageWithNonce(
		nonce,
		service.storageService,
		service.uint64ByteSliceConverter,
		service.marshaller,
		hashByNonceUnit,
	)
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
