package hooks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"path"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
	BuiltInFunctions   vmcommon.BuiltInFunctionContainer
	NFTStorageHandler  vmcommon.SimpleESDTNFTStorageHandler
	CompiledSCPool     storage.Cacher
	ConfigSCStorage    config.StorageConfig
	EnableEpochs       config.EnableEpochs
	EpochNotifier      vmcommon.EpochNotifier
	WorkingDir         string
	NilCompiledSCStore bool
}

// BlockChainHookImpl is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type BlockChainHookImpl struct {
	accounts          state.AccountsAdapter
	pubkeyConv        core.PubkeyConverter
	storageService    dataRetriever.StorageService
	blockChain        data.ChainHandler
	shardCoordinator  sharding.Coordinator
	marshalizer       marshal.Marshalizer
	uint64Converter   typeConverters.Uint64ByteSliceConverter
	builtInFunctions  vmcommon.BuiltInFunctionContainer
	nftStorageHandler vmcommon.SimpleESDTNFTStorageHandler

	mutCurrentHdr sync.RWMutex
	currentHdr    data.HeaderHandler

	compiledScPool     storage.Cacher
	compiledScStorage  storage.Storer
	configSCStorage    config.StorageConfig
	workingDir         string
	nilCompiledSCStore bool

	isPayableBySCEnableEpoch       uint32
	flagIsPayableBySC              atomic.Flag
	optimizeNFTStoreEnableEpoch    uint32
	flagOptimizeNFTStore           atomic.Flag
	doNotReturnOldBlockEnableEpoch uint32
	flagDoNotReturnOldBlock        atomic.Flag
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
		accounts:                       args.Accounts,
		pubkeyConv:                     args.PubkeyConv,
		storageService:                 args.StorageService,
		blockChain:                     args.BlockChain,
		shardCoordinator:               args.ShardCoordinator,
		marshalizer:                    args.Marshalizer,
		uint64Converter:                args.Uint64Converter,
		builtInFunctions:               args.BuiltInFunctions,
		compiledScPool:                 args.CompiledSCPool,
		configSCStorage:                args.ConfigSCStorage,
		workingDir:                     args.WorkingDir,
		nilCompiledSCStore:             args.NilCompiledSCStore,
		nftStorageHandler:              args.NFTStorageHandler,
		isPayableBySCEnableEpoch:       args.EnableEpochs.IsPayableBySCEnableEpoch,
		optimizeNFTStoreEnableEpoch:    args.EnableEpochs.OptimizeNFTStoreEnableEpoch,
		doNotReturnOldBlockEnableEpoch: args.EnableEpochs.DoNotReturnOldBlockInBlockchainHookEnableEpoch,
	}

	log.Debug("blockchainHook: payable by SC", "epoch", blockChainHookImpl.isPayableBySCEnableEpoch)
	log.Debug("blockchainHook: optimize nft metadata store", "epoch", blockChainHookImpl.optimizeNFTStoreEnableEpoch)
	log.Debug("blockchainHook: do not return old block", "epoch", blockChainHookImpl.doNotReturnOldBlockEnableEpoch)

	err = blockChainHookImpl.makeCompiledSCStorage()
	if err != nil {
		return nil, err
	}

	blockChainHookImpl.ClearCompiledCodes()
	blockChainHookImpl.currentHdr = &block.Header{}

	args.EpochNotifier.RegisterNotifyHandler(blockChainHookImpl)

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
	if check.IfNil(args.NFTStorageHandler) {
		return process.ErrNilNFTStorageHandler
	}
	if check.IfNil(args.EpochNotifier) {
		return process.ErrNilEpochNotifier
	}

	return nil
}

// GetCode returns the code for the given account
func (bh *BlockChainHookImpl) GetCode(account vmcommon.UserAccountHandler) []byte {
	if check.IfNil(account) {
		return nil
	}

	return bh.accounts.GetCode(account.GetCodeHash())
}

func (bh *BlockChainHookImpl) isNotSystemAccountAndCrossShard(address []byte) bool {
	dstShardId := bh.shardCoordinator.ComputeId(address)
	return !core.IsSystemAccountAddress(address) && dstShardId != bh.shardCoordinator.SelfId()
}

// GetUserAccount returns the balance of a shard account
func (bh *BlockChainHookImpl) GetUserAccount(address []byte) (vmcommon.UserAccountHandler, error) {
	defer stopMeasure(startMeasure("GetUserAccount"))

	if bh.isNotSystemAccountAndCrossShard(address) {
		return nil, state.ErrAccNotFound
	}

	acc, err := bh.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	dstAccount, ok := acc.(vmcommon.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return dstAccount, nil
}

// GetStorageData returns the storage value of a variable held in account's data trie
func (bh *BlockChainHookImpl) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	defer stopMeasure(startMeasure("GetStorageData"))

	userAcc, err := bh.GetUserAccount(accountAddress)
	if err == state.ErrAccNotFound {
		return make([]byte, 0), nil
	}
	if err != nil {
		return nil, err
	}

	value, err := userAcc.AccountDataHandler().RetrieveValue(index)
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
	if bh.flagDoNotReturnOldBlock.IsSet() {
		return nil, process.ErrInvalidNonceRequest
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

// LastNonce returns the nonce from the last committed block
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
	return make([]byte, 0)
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
	rootHash := bh.blockChain.GetCurrentBlockRootHash()
	if len(rootHash) > 0 {
		return rootHash
	}

	return make([]byte, 0)
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

	if !check.IfNil(dstAccount) && !bytes.Equal(input.CallerAddr, input.RecipientAddr) {
		err = bh.accounts.SaveAccount(dstAccount)
		if err != nil {
			return nil, err
		}
	}

	return vmOutput, nil
}

// SaveNFTMetaDataToSystemAccount will save NFT meta data to system account for the given transaction
func (bh *BlockChainHookImpl) SaveNFTMetaDataToSystemAccount(tx data.TransactionHandler) error {
	return bh.nftStorageHandler.SaveNFTMetaDataToSystemAccount(tx)
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
func (bh *BlockChainHookImpl) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	if core.IsSystemAccountAddress(recvAddress) {
		return false, nil
	}

	if !bh.IsSmartContract(recvAddress) {
		return true, nil
	}

	if bh.isNotSystemAccountAndCrossShard(recvAddress) {
		return true, nil
	}

	userAcc, err := bh.GetUserAccount(recvAddress)
	if err == state.ErrAccNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	metadata := vmcommon.CodeMetadataFromBytes(userAcc.GetCodeMetadata())
	if bh.flagIsPayableBySC.IsSet() && bh.IsSmartContract(sndAddress) {
		return metadata.Payable || metadata.PayableBySC, nil
	}

	return metadata.Payable, nil
}

func (bh *BlockChainHookImpl) getUserAccounts(
	input *vmcommon.ContractCallInput,
) (vmcommon.UserAccountHandler, vmcommon.UserAccountHandler, error) {
	var sndAccount vmcommon.UserAccountHandler
	sndShardId := bh.shardCoordinator.ComputeId(input.CallerAddr)
	if sndShardId == bh.shardCoordinator.SelfId() {
		acc, err := bh.accounts.GetExistingAccount(input.CallerAddr)
		if err != nil {
			return nil, nil, err
		}

		var ok bool
		sndAccount, ok = acc.(vmcommon.UserAccountHandler)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		if bytes.Equal(input.CallerAddr, input.RecipientAddr) {
			return sndAccount, sndAccount, nil
		}
	}

	var dstAccount vmcommon.UserAccountHandler
	dstShardId := bh.shardCoordinator.ComputeId(input.RecipientAddr)
	if dstShardId == bh.shardCoordinator.SelfId() {
		acc, err := bh.accounts.LoadAccount(input.RecipientAddr)
		if err != nil {
			return nil, nil, err
		}

		var ok bool
		dstAccount, ok = acc.(vmcommon.UserAccountHandler)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
	}

	return sndAccount, dstAccount, nil
}

// GetBuiltinFunctionNames returns the built in function names
func (bh *BlockChainHookImpl) GetBuiltinFunctionNames() vmcommon.FunctionNames {
	return bh.builtInFunctions.Keys()
}

// GetBuiltinFunctionsContainer returns the built in functions container
func (bh *BlockChainHookImpl) GetBuiltinFunctionsContainer() vmcommon.BuiltInFunctionContainer {
	return bh.builtInFunctions
}

// GetAllState returns the underlying state of a given account
// TODO remove this func completely
func (bh *BlockChainHookImpl) GetAllState(_ []byte) (map[string][]byte, error) {
	return nil, nil
}

// GetESDTToken returns the unmarshalled esdt data for the given key
func (bh *BlockChainHookImpl) GetESDTToken(address []byte, tokenID []byte, nonce uint64) (*esdt.ESDigitalToken, error) {
	userAcc, err := bh.GetUserAccount(address)
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0)}
	if err == state.ErrAccNotFound {
		return esdtData, nil
	}
	if err != nil {
		return nil, err
	}

	esdtTokenKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + string(tokenID))
	if !bh.flagOptimizeNFTStore.IsSet() {
		return bh.returnESDTTokenByLegacyMethod(userAcc, esdtData, esdtTokenKey, nonce)
	}

	esdtData, _, err = bh.nftStorageHandler.GetESDTNFTTokenOnDestination(userAcc, esdtTokenKey, nonce)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}

func (bh *BlockChainHookImpl) returnESDTTokenByLegacyMethod(
	userAcc vmcommon.UserAccountHandler,
	esdtData *esdt.ESDigitalToken,
	esdtTokenKey []byte,
	nonce uint64,
) (*esdt.ESDigitalToken, error) {
	if nonce > 0 {
		esdtTokenKey = append(esdtTokenKey, big.NewInt(0).SetUint64(nonce).Bytes()...)
	}

	value, err := userAcc.AccountDataHandler().RetrieveValue(esdtTokenKey)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return esdtData, nil
	}

	err = bh.marshalizer.Unmarshal(esdtData, value)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}

// NumberOfShards returns the number of shards
func (bh *BlockChainHookImpl) NumberOfShards() uint32 {
	return bh.shardCoordinator.NumberOfShards()
}

func hashFromAddressAndNonce(creatorAddress []byte, creatorNonce uint64) []byte {
	buffNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffNonce, creatorNonce)
	adrAndNonce := append(creatorAddress, buffNonce...)
	scAddress := keccak.NewKeccak().Compute(string(adrAndNonce))

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
	log.LogIfError(err, "func", "BlockChainHookImpl.SaveCompiledCode: compiledScStorage.Put", "codeHash", codeHash)
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
	log.LogIfError(err, "func", "BlockChainHookImpl.DeleteCompiledCode: compiledScStorage.Remove", "codeHash", codeHash)
}

// Close closes/cleans up the blockchain hook
func (bh *BlockChainHookImpl) Close() error {
	bh.compiledScPool.Clear()
	return bh.compiledScStorage.DestroyUnit()
}

// ClearCompiledCodes deletes the compiled codes from storage and cache
func (bh *BlockChainHookImpl) ClearCompiledCodes() {
	bh.compiledScPool.Clear()
	err := bh.compiledScStorage.DestroyUnit()
	log.LogIfError(err, "func", "BlockChainHookImpl.ClearCompiledCodes().compiledScStorage.DestroyUnit")

	err = bh.makeCompiledSCStorage()
	log.LogIfError(err, "func", "BlockChainHookImpl.makeCompiledSCStorage")
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
	)
	if err != nil {
		return err
	}

	bh.compiledScStorage = store
	return nil
}

// GetSnapshot gets the number of entries in the journal as a snapshot id
func (bh *BlockChainHookImpl) GetSnapshot() int {
	return bh.accounts.JournalLen()
}

// RevertToSnapshot reverts snapshots up to the specified one
func (bh *BlockChainHookImpl) RevertToSnapshot(snapshot int) error {
	return bh.accounts.RevertToSnapshot(snapshot)
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (bh *BlockChainHookImpl) EpochConfirmed(epoch uint32, _ uint64) {
	bh.flagIsPayableBySC.SetValue(epoch >= bh.isPayableBySCEnableEpoch)
	log.Debug("blockchainHookImpl is payable by SC", "enabled", bh.flagIsPayableBySC.IsSet())

	bh.flagOptimizeNFTStore.SetValue(epoch >= bh.optimizeNFTStoreEnableEpoch)
	log.Debug("blockchainHookImpl optimize nft metadata store", "enabled", bh.flagOptimizeNFTStore.IsSet())

	bh.flagDoNotReturnOldBlock.SetValue(epoch >= bh.doNotReturnOldBlockEnableEpoch)
	log.Debug("blockchainHookImpl: do not return old block", "enabled", bh.flagDoNotReturnOldBlock.IsSet())
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
