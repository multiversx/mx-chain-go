package smartContract

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

var _ process.SCQueryService = (*SCQueryService)(nil)

var logQueryService = logger.GetOrCreate("process/smartcontract.queryService")

// MaxGasLimitPerQuery - each unit is the equivalent of 1 nanosecond processing time
const MaxGasLimitPerQuery = 300_000_000_000

// SCQueryService can execute Get functions over SC to fetch stored values
type SCQueryService struct {
	vmContainer              process.VirtualMachinesContainer
	economicsFee             process.FeeHandler
	mutRunSc                 sync.Mutex
	blockChainHook           process.BlockChainHookWithAccountsAdapter
	mainBlockChain           data.ChainHandler
	apiBlockChain            data.ChainHandler
	gasForQuery              uint64
	wasmVMChangeLocker       common.Locker
	bootstrapper             process.Bootstrapper
	allowExternalQueriesChan chan struct{}
	historyRepository        dblookupext.HistoryRepository
	shardCoordinator         sharding.Coordinator
	storageService           dataRetriever.StorageService
	marshaller               marshal.Marshalizer
	hasher                   hashing.Hasher
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	latestQueriedEpoch       core.OptionalUint32
}

// ArgsNewSCQueryService defines the arguments needed for the sc query service
type ArgsNewSCQueryService struct {
	VmContainer              process.VirtualMachinesContainer
	EconomicsFee             process.FeeHandler
	BlockChainHook           process.BlockChainHookWithAccountsAdapter
	MainBlockChain           data.ChainHandler
	APIBlockChain            data.ChainHandler
	WasmVMChangeLocker       common.Locker
	Bootstrapper             process.Bootstrapper
	AllowExternalQueriesChan chan struct{}
	MaxGasLimitPerQuery      uint64
	HistoryRepository        dblookupext.HistoryRepository
	ShardCoordinator         sharding.Coordinator
	StorageService           dataRetriever.StorageService
	Marshaller               marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewSCQueryService returns a new instance of SCQueryService
func NewSCQueryService(
	args ArgsNewSCQueryService,
) (*SCQueryService, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gasForQuery := uint64(MaxGasLimitPerQuery)
	if args.MaxGasLimitPerQuery > 0 {
		gasForQuery = args.MaxGasLimitPerQuery
	}
	return &SCQueryService{
		vmContainer:              args.VmContainer,
		economicsFee:             args.EconomicsFee,
		mainBlockChain:           args.MainBlockChain,
		apiBlockChain:            args.APIBlockChain,
		blockChainHook:           args.BlockChainHook,
		wasmVMChangeLocker:       args.WasmVMChangeLocker,
		bootstrapper:             args.Bootstrapper,
		gasForQuery:              gasForQuery,
		allowExternalQueriesChan: args.AllowExternalQueriesChan,
		historyRepository:        args.HistoryRepository,
		shardCoordinator:         args.ShardCoordinator,
		storageService:           args.StorageService,
		marshaller:               args.Marshaller,
		hasher:                   args.Hasher,
		uint64ByteSliceConverter: args.Uint64ByteSliceConverter,
		latestQueriedEpoch:       core.OptionalUint32{},
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
	if check.IfNil(args.MainBlockChain) {
		return fmt.Errorf("%w for main blockchain", process.ErrNilBlockChain)
	}
	if check.IfNil(args.APIBlockChain) {
		return fmt.Errorf("%w for api blockchain", process.ErrNilBlockChain)
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
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return process.ErrNilUint64Converter
	}

	return nil
}

// ExecuteQuery returns the VMOutput resulted upon running the function on the smart contract
func (service *SCQueryService) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
	if !service.shouldAllowQueriesExecution() {
		return nil, nil, process.ErrQueriesNotAllowedYet
	}

	if query.ScAddress == nil {
		return nil, nil, process.ErrNilScAddress
	}
	if len(query.FuncName) == 0 {
		return nil, nil, process.ErrEmptyFunctionName
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

func (service *SCQueryService) executeScCall(query *process.SCQuery, gasPrice uint64) (*vmcommon.VMOutput, common.BlockInfo, error) {
	logQueryService.Trace("executeScCall", "address", query.ScAddress, "function", query.FuncName, "blockNonce", query.BlockNonce.Value, "blockHash", query.BlockHash)

	shouldEarlyExitBecauseOfSyncState := query.ShouldBeSynced && service.bootstrapper.GetNodeState() == common.NsNotSynchronized
	if shouldEarlyExitBecauseOfSyncState {
		return nil, nil, process.ErrNodeIsNotSynced
	}

	blockHeader, blockRootHash, err := service.extractBlockHeaderAndRootHash(query)
	if err != nil {
		return nil, nil, err
	}

	if len(blockRootHash) > 0 {
		err = service.apiBlockChain.SetCurrentBlockHeaderAndRootHash(blockHeader, blockRootHash)
		if err != nil {
			return nil, nil, err
		}

		err = service.recreateTrie(blockRootHash, blockHeader)
		if err != nil {
			return nil, nil, err
		}
		service.blockChainHook.SetCurrentHeader(blockHeader)
	}

	shouldCheckRootHashChanges := query.SameScState
	rootHashBeforeExecution := make([]byte, 0)

	if shouldCheckRootHashChanges {
		rootHashBeforeExecution = service.apiBlockChain.GetCurrentBlockRootHash()
	}

	service.wasmVMChangeLocker.RLock()
	vm, _, err := scrCommon.FindVMByScAddress(service.vmContainer, query.ScAddress)
	if err != nil {
		service.wasmVMChangeLocker.RUnlock()
		return nil, nil, err
	}

	query = prepareScQuery(query)
	vmInput := service.createVMCallInput(query, gasPrice)
	vmOutput, err := vm.RunSmartContractCall(vmInput)
	service.wasmVMChangeLocker.RUnlock()
	if err != nil {
		return nil, nil, err
	}

	if query.SameScState {
		err = service.checkForRootHashChanges(rootHashBeforeExecution)
		if err != nil {
			return nil, nil, err
		}
	}

	var blockHash []byte
	var blockNonce uint64
	if !check.IfNil(blockHeader) {
		blockNonce = blockHeader.GetNonce()
		blockHash, err = core.CalculateHash(service.marshaller, service.hasher, blockHeader)
		if err != nil {
			return nil, nil, err
		}
	}
	blockInfo := holders.NewBlockInfo(blockHash, blockNonce, blockRootHash)
	return vmOutput, blockInfo, nil
}

func (service *SCQueryService) recreateTrie(blockRootHash []byte, blockHeader data.HeaderHandler) error {
	if check.IfNil(blockHeader) {
		return process.ErrNilBlockHeader
	}

	accountsAdapter := service.blockChainHook.GetAccountsAdapter()

	if service.isLatestQueriedEpoch(blockHeader.GetEpoch()) {
		logQueryService.Trace("calling RecreateTrie, for recent history", "block", blockHeader.GetNonce(), "rootHash", blockRootHash)

		err := accountsAdapter.RecreateTrie(blockRootHash)
		if err != nil {
			return err
		}

		service.rememberQueriedEpoch(blockHeader.GetEpoch())
	}

	logQueryService.Trace("calling RecreateTrieFromEpoch, for older history", "block", blockHeader.GetNonce(), "rootHash", blockRootHash)
	holder := holders.NewRootHashHolder(blockRootHash, core.OptionalUint32{Value: blockHeader.GetEpoch(), HasValue: true})

	err := accountsAdapter.RecreateTrieFromEpoch(holder)
	if err != nil {
		return err
	}

	service.rememberQueriedEpoch(blockHeader.GetEpoch())
	return err
}

func (service *SCQueryService) isLatestQueriedEpoch(epoch uint32) bool {
	return service.latestQueriedEpoch.HasValue && service.latestQueriedEpoch.Value == epoch
}

func (service *SCQueryService) rememberQueriedEpoch(epoch uint32) {
	service.latestQueriedEpoch = core.OptionalUint32{Value: epoch, HasValue: true}
}

func (service *SCQueryService) getCurrentEpoch() uint32 {
	header := service.mainBlockChain.GetCurrentBlockHeader()
	if check.IfNil(header) {
		return 0
	}

	return header.GetEpoch()
}

// TODO: extract duplicated code with nodeBlocks.go
func (service *SCQueryService) extractBlockHeaderAndRootHash(query *process.SCQuery) (data.HeaderHandler, []byte, error) {
	if len(query.BlockHash) > 0 {
		currentHeader, err := service.getBlockHeaderByHash(query.BlockHash)
		if err != nil {
			return nil, nil, err
		}

		return service.getRootHashForBlock(currentHeader)
	}

	if query.BlockNonce.HasValue {
		currentHeader, _, err := service.getBlockHeaderByNonce(query.BlockNonce.Value)
		if err != nil {
			return nil, nil, err
		}

		return service.getRootHashForBlock(currentHeader)
	}

	return service.mainBlockChain.GetCurrentBlockHeader(), service.mainBlockChain.GetCurrentBlockRootHash(), nil
}

func (service *SCQueryService) getRootHashForBlock(currentHeader data.HeaderHandler) (data.HeaderHandler, []byte, error) {
	blockHeader, _, err := service.getBlockHeaderByNonce(currentHeader.GetNonce() + 1)
	if err != nil {
		return nil, nil, err
	}

	additionalData := blockHeader.GetAdditionalData()
	if check.IfNil(additionalData) {
		return currentHeader, currentHeader.GetRootHash(), nil
	}

	return blockHeader, additionalData.GetScheduledRootHash(), nil
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
	rootHashAfter := service.apiBlockChain.GetCurrentBlockRootHash()

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

	vmOutput, _, err := service.executeScCall(query, 1)
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
