package hooks

import (
	"encoding/binary"
	"fmt"
	"path"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var _ process.BlockChainHookHandler = (*BlockChainHookImpl)(nil)

var log = logger.GetOrCreate("process/smartcontract/blockchainhook")

const defaultCompiledSCPath = "compiledSCStorage"
const executeDurationAlarmThreshold = time.Duration(50) * time.Millisecond

// ArgBlockChainHook represents the arguments structure for the blockchain hook
type ArgBlockChainHook struct {
	Accounts           state.AccountsAdapter
	PubkeyConv         core.PubkeyConverter
	StorageService     dataRetriever.StorageService
	DataPool           dataRetriever.PoolsHolder
	BlockChain         data.ChainHandler
	ShardCoordinator   sharding.Coordinator
	Marshalizer        marshal.Marshalizer
	Uint64Converter    typeConverters.Uint64ByteSliceConverter
	BuiltInFunctions   process.BuiltInFunctionContainer
	CompiledSCPool     storage.Cacher
	ConfigSCStorage    config.StorageConfig
	WorkingDir         string
	NilCompiledSCStore bool
}

// BlockChainHookImpl is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type BlockChainHookImpl struct {
	accounts         state.AccountsAdapter
	pubkeyConv       core.PubkeyConverter
	storageService   dataRetriever.StorageService
	blockChain       data.ChainHandler
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	uint64Converter  typeConverters.Uint64ByteSliceConverter
	builtInFunctions process.BuiltInFunctionContainer

	mutCurrentHdr sync.RWMutex
	currentHdr    data.HeaderHandler

	compiledScPool     storage.Cacher
	compiledScStorage  storage.Storer
	configSCStorage    config.StorageConfig
	workingDir         string
	nilCompiledSCStore bool
}

// NewBlockChainHookImpl creates a new BlockChainHookImpl instance
func NewBlockChainHookImpl(
	args ArgBlockChainHook,
) (*BlockChainHookImpl, error) {
	err := checkForNil(args)
	if err != nil {
		return nil, err
	}

	blockChainHookImpl := &BlockChainHookImpl{
		accounts:           args.Accounts,
		pubkeyConv:         args.PubkeyConv,
		storageService:     args.StorageService,
		blockChain:         args.BlockChain,
		shardCoordinator:   args.ShardCoordinator,
		marshalizer:        args.Marshalizer,
		uint64Converter:    args.Uint64Converter,
		builtInFunctions:   args.BuiltInFunctions,
		compiledScPool:     args.CompiledSCPool,
		configSCStorage:    args.ConfigSCStorage,
		workingDir:         args.WorkingDir,
		nilCompiledSCStore: args.NilCompiledSCStore,
	}

	err = blockChainHookImpl.makeCompiledSCStorage()
	if err != nil {
		return nil, err
	}

	blockChainHookImpl.ClearCompiledCodes()
	blockChainHookImpl.currentHdr = &block.Header{}

	return blockChainHookImpl, nil
}

func checkForNil(args ArgBlockChainHook) error {
	if check.IfNil(args.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.PubkeyConv) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.StorageService) {
		return process.ErrNilStorage
	}
	if check.IfNil(args.BlockChain) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.Uint64Converter) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(args.BuiltInFunctions) {
		return process.ErrNilBuiltInFunction
	}
	if check.IfNil(args.CompiledSCPool) {
		return process.ErrNilCacher
	}

	return nil
}

// GetUserAccount returns the balance of a shard account
func (bh *BlockChainHookImpl) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	defer stopMeasure(startMeasure("GetUserAccount"))

	dstShardId := bh.shardCoordinator.ComputeId(address)
	if !core.IsSystemAccountAddress(address) && dstShardId != bh.shardCoordinator.SelfId() {
		return nil, nil
	}

	acc, err := bh.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	dstAccount, ok := acc.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return dstAccount, nil
}

// GetStorageData returns the storage value of a variable held in account's data trie
func (bh *BlockChainHookImpl) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	defer stopMeasure(startMeasure("GetStorageData"))

	account, err := bh.GetUserAccount(accountAddress)
	if err == state.ErrAccNotFound {
		return make([]byte, 0), nil
	}
	if err != nil {
		return nil, err
	}

	userAcc, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	value, err := userAcc.DataTrieTracker().RetrieveValue(index)
	messages := []interface{}{
		"address", accountAddress,
		"rootHash", userAcc.GetRootHash(),
		"key", index,
		"value", value,
	}
	if err != nil {
		messages = append(messages, "error")
		messages = append(messages, err)
	}
	log.Trace("GetStorageData ", messages...)
	return value, err
}

// GetBlockhash returns the header hash for a requested nonce delta
func (bh *BlockChainHookImpl) GetBlockhash(nonce uint64) ([]byte, error) {
	defer stopMeasure(startMeasure("GetBlockhash"))

	hdr := bh.blockChain.GetCurrentBlockHeader()

	if check.IfNil(hdr) {
		return nil, process.ErrNilBlockHeader
	}
	if nonce > hdr.GetNonce() {
		return nil, process.ErrInvalidNonceRequest
	}
	if nonce == hdr.GetNonce() {
		return bh.blockChain.GetCurrentBlockHeaderHash(), nil
	}

	header, hash, err := process.GetHeaderFromStorageWithNonce(
		nonce,
		bh.shardCoordinator.SelfId(),
		bh.storageService,
		bh.uint64Converter,
		bh.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	if header.GetEpoch() != hdr.GetEpoch() {
		return nil, process.ErrInvalidBlockRequestOldEpoch
	}

	return hash, nil
}

// LastNonce returns the nonce from from the last committed block
func (bh *BlockChainHookImpl) LastNonce() uint64 {
	if !check.IfNil(bh.blockChain.GetCurrentBlockHeader()) {
		return bh.blockChain.GetCurrentBlockHeader().GetNonce()
	}
	return 0
}

// LastRound returns the round from the last committed block
func (bh *BlockChainHookImpl) LastRound() uint64 {
	if !check.IfNil(bh.blockChain.GetCurrentBlockHeader()) {
		return bh.blockChain.GetCurrentBlockHeader().GetRound()
	}
	return 0
}

// LastTimeStamp returns the timeStamp from the last committed block
func (bh *BlockChainHookImpl) LastTimeStamp() uint64 {
	if !check.IfNil(bh.blockChain.GetCurrentBlockHeader()) {
		return bh.blockChain.GetCurrentBlockHeader().GetTimeStamp()
	}
	return 0
}

// LastRandomSeed returns the random seed from the last committed block
func (bh *BlockChainHookImpl) LastRandomSeed() []byte {
	if !check.IfNil(bh.blockChain.GetCurrentBlockHeader()) {
		return bh.blockChain.GetCurrentBlockHeader().GetRandSeed()
	}
	return []byte{}
}

// LastEpoch returns the epoch from the last committed block
func (bh *BlockChainHookImpl) LastEpoch() uint32 {
	if !check.IfNil(bh.blockChain.GetCurrentBlockHeader()) {
		return bh.blockChain.GetCurrentBlockHeader().GetEpoch()
	}
	return 0
}

// GetStateRootHash returns the state root hash from the last committed block
func (bh *BlockChainHookImpl) GetStateRootHash() []byte {
	if !check.IfNil(bh.blockChain.GetCurrentBlockHeader()) {
		return bh.blockChain.GetCurrentBlockHeader().GetRootHash()
	}
	return []byte{}
}

// CurrentNonce returns the nonce from the current block
func (bh *BlockChainHookImpl) CurrentNonce() uint64 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()

	return bh.currentHdr.GetNonce()
}

// CurrentRound returns the round from the current block
func (bh *BlockChainHookImpl) CurrentRound() uint64 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetRound()
}

// CurrentTimeStamp return the timestamp from the current block
func (bh *BlockChainHookImpl) CurrentTimeStamp() uint64 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetTimeStamp()
}

// CurrentRandomSeed returns the random seed from the current header
func (bh *BlockChainHookImpl) CurrentRandomSeed() []byte {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetRandSeed()
}

// CurrentEpoch returns the current epoch
func (bh *BlockChainHookImpl) CurrentEpoch() uint32 {
	bh.mutCurrentHdr.RLock()
	defer bh.mutCurrentHdr.RUnlock()
	return bh.currentHdr.GetEpoch()
}

// NewAddress is a hook which creates a new smart contract address from the creators address and nonce
// The address is created by applied keccak256 on the appended value off creator address and nonce
// Prefix mask is applied for first 8 bytes 0, and for bytes 9-10 - VM type
// Suffix mask is applied - last 2 bytes are for the shard ID - mask is applied as suffix mask
func (bh *BlockChainHookImpl) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	addressLength := bh.pubkeyConv.Len()
	if len(creatorAddress) != addressLength {
		return nil, ErrAddressLengthNotCorrect
	}

	if len(vmType) != core.VMTypeLen {
		return nil, ErrVMTypeLengthIsNotCorrect
	}

	base := hashFromAddressAndNonce(creatorAddress, creatorNonce)
	prefixMask := createPrefixMask(vmType)
	suffixMask := createSuffixMask(creatorAddress)

	copy(base[:core.NumInitCharactersForScAddress], prefixMask)
	copy(base[len(base)-core.ShardIdentiferLen:], suffixMask)

	return base, nil
}

// ProcessBuiltInFunction is the hook through which a smart contract can execute a built in function
func (bh *BlockChainHookImpl) ProcessBuiltInFunction(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	defer stopMeasure(startMeasure("ProcessBuiltInFunction"))

	if input == nil {
		return nil, process.ErrNilVmInput
	}

	function, err := bh.builtInFunctions.Get(input.Function)
	if err != nil {
		return nil, err
	}

	sndAccount, dstAccount, err := bh.getUserAccounts(input)
	if err != nil {
		return nil, err
	}

	vmOutput, err := function.ProcessBuiltinFunction(sndAccount, dstAccount, input)
	if err != nil {
		return nil, err
	}

	if !check.IfNil(sndAccount) {
		err = bh.accounts.SaveAccount(sndAccount)
		if err != nil {
			return nil, err
		}
	}

	if !check.IfNil(dstAccount) {
		err = bh.accounts.SaveAccount(dstAccount)
		if err != nil {
			return nil, err
		}
	}

	return vmOutput, nil
}

// GetShardOfAddress is the hook that returns the shard of a given address
func (bh *BlockChainHookImpl) GetShardOfAddress(address []byte) uint32 {
	return bh.shardCoordinator.ComputeId(address)
}

// IsSmartContract returns whether the address points to a smart contract
func (bh *BlockChainHookImpl) IsSmartContract(address []byte) bool {
	return core.IsSmartContractAddress(address)
}

// IsPayable checks whether the provided address can receive ERD or not
func (bh *BlockChainHookImpl) IsPayable(address []byte) (bool, error) {
	if core.IsSystemAccountAddress(address) {
		return false, nil
	}

	if !bh.IsSmartContract(address) {
		return true, nil
	}

	userAcc, err := bh.GetUserAccount(address)
	if err == state.ErrAccNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if check.IfNil(userAcc) {
		return true, nil
	}

	metadata := vmcommon.CodeMetadataFromBytes(userAcc.GetCodeMetadata())
	return metadata.Payable, nil
}

func (bh *BlockChainHookImpl) getUserAccounts(
	input *vmcommon.ContractCallInput,
) (state.UserAccountHandler, state.UserAccountHandler, error) {
	var sndAccount state.UserAccountHandler
	sndShardId := bh.shardCoordinator.ComputeId(input.CallerAddr)
	if sndShardId == bh.shardCoordinator.SelfId() {
		acc, err := bh.accounts.GetExistingAccount(input.CallerAddr)
		if err != nil {
			return nil, nil, err
		}

		var ok bool
		sndAccount, ok = acc.(state.UserAccountHandler)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
	}

	var dstAccount state.UserAccountHandler
	dstShardId := bh.shardCoordinator.ComputeId(input.RecipientAddr)
	if dstShardId == bh.shardCoordinator.SelfId() {
		acc, err := bh.accounts.LoadAccount(input.RecipientAddr)
		if err != nil {
			return nil, nil, err
		}

		var ok bool
		dstAccount, ok = acc.(state.UserAccountHandler)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
	}

	return sndAccount, dstAccount, nil
}

// GetBuiltInFunctions returns the built in functions container
func (bh *BlockChainHookImpl) GetBuiltInFunctions() process.BuiltInFunctionContainer {
	return bh.builtInFunctions
}

// GetBuiltinFunctionNames returns the built in function names
func (bh *BlockChainHookImpl) GetBuiltinFunctionNames() vmcommon.FunctionNames {
	return bh.builtInFunctions.Keys()
}

// GetAllState returns the underlying state of a given account
func (bh *BlockChainHookImpl) GetAllState(address []byte) (map[string][]byte, error) {
	defer stopMeasure(startMeasure("GetAllState"))

	dstShardId := bh.shardCoordinator.ComputeId(address)
	if dstShardId != bh.shardCoordinator.SelfId() {
		return nil, process.ErrDestinationNotInSelfShard
	}

	acc, err := bh.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	dstAccount, ok := acc.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return dstAccount.DataTrie().GetAllLeaves()
}

// NumberOfShards returns the number of shards
func (bh *BlockChainHookImpl) NumberOfShards() uint32 {
	return bh.shardCoordinator.NumberOfShards()
}

func hashFromAddressAndNonce(creatorAddress []byte, creatorNonce uint64) []byte {
	buffNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffNonce, creatorNonce)
	adrAndNonce := append(creatorAddress, buffNonce...)
	scAddress := keccak.Keccak{}.Compute(string(adrAndNonce))

	return scAddress
}

func createPrefixMask(vmType []byte) []byte {
	prefixMask := make([]byte, core.NumInitCharactersForScAddress-core.VMTypeLen)
	prefixMask = append(prefixMask, vmType...)

	return prefixMask
}

func createSuffixMask(creatorAddress []byte) []byte {
	return creatorAddress[len(creatorAddress)-2:]
}

// SetCurrentHeader sets current header to be used by smart contracts
func (bh *BlockChainHookImpl) SetCurrentHeader(hdr data.HeaderHandler) {
	if check.IfNil(hdr) {
		return
	}

	bh.mutCurrentHdr.Lock()
	bh.currentHdr = hdr
	bh.mutCurrentHdr.Unlock()
}

// SaveCompiledCode saves the compiled code to cache and storage
func (bh *BlockChainHookImpl) SaveCompiledCode(codeHash []byte, code []byte) {
	bh.compiledScPool.Put(codeHash, code, len(code))
	err := bh.compiledScStorage.Put(codeHash, code)
	if err != nil {
		log.Debug("SaveCompiledCode", "error", err, "codeHash", codeHash)
	}
}

// GetCompiledCode returns the compiled code if it is found in the cache or storage
func (bh *BlockChainHookImpl) GetCompiledCode(codeHash []byte) (bool, []byte) {
	val, found := bh.compiledScPool.Get(codeHash)
	if found {
		compiledCode, ok := val.([]byte)
		if ok {
			return true, compiledCode
		}
	}

	compiledCode, err := bh.compiledScStorage.Get(codeHash)
	if err != nil || len(compiledCode) == 0 {
		return false, nil
	}

	bh.compiledScPool.Put(codeHash, compiledCode, len(compiledCode))

	return true, compiledCode
}

// DeleteCompiledCode deletes the compiled code from storage and cache
func (bh *BlockChainHookImpl) DeleteCompiledCode(codeHash []byte) {
	bh.compiledScPool.Remove(codeHash)
	err := bh.compiledScStorage.Remove(codeHash)
	if err != nil {
		log.Debug("DeleteCompiledCode", "error", err, "codeHash", codeHash)
	}
}

// ClearCompiledCodes deletes the compiled codes from storage and cache
func (bh *BlockChainHookImpl) ClearCompiledCodes() {
	bh.compiledScPool.Clear()
	err := bh.compiledScStorage.DestroyUnit()
	if err != nil {
		log.Error("blockchainHook ClearCompiledCodes DestroyUnit", "error", err)
	}

	err = bh.compiledScStorage.Close()
	if err != nil {
		log.Error("blockchainHook ClearCompiledCodes Close", "error", err)
	}

	err = bh.makeCompiledSCStorage()
	if err != nil {
		log.Error("blockchainHook ClearCompiledCodes makeCompiledSCStorage", "error", err)
	}
}

func (bh *BlockChainHookImpl) makeCompiledSCStorage() error {
	if bh.nilCompiledSCStore {
		bh.compiledScStorage = storageUnit.NewNilStorer()
		return nil
	}

	dbConfig := factory.GetDBFromConfig(bh.configSCStorage.DB)
	dbConfig.FilePath = path.Join(bh.workingDir, defaultCompiledSCPath, bh.configSCStorage.DB.FilePath)
	store, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(bh.configSCStorage.Cache),
		dbConfig,
		factory.GetBloomFromConfig(bh.configSCStorage.Bloom),
	)
	if err != nil {
		return err
	}

	bh.compiledScStorage = store
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bh *BlockChainHookImpl) IsInterfaceNil() bool {
	return bh == nil
}

func startMeasure(hook string) (string, *core.StopWatch) {
	sw := core.NewStopWatch()
	sw.Start(hook)
	return hook, sw
}

func stopMeasure(hook string, sw *core.StopWatch) {
	sw.Stop(hook)

	duration := sw.GetMeasurement(hook)
	if duration > executeDurationAlarmThreshold {
		log.Debug(fmt.Sprintf("%s took > %s", hook, executeDurationAlarmThreshold), "duration", duration)
	} else {
		log.Trace(hook, "duration", duration)
	}
}
