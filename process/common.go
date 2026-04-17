package process

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
)

var log = logger.GetOrCreate("process")

const maxSelfNotarizedLookback = 50
const VMStoragePrefix = "VM@"

// maxForwardPendingWalkIter bounds the backward meta walk during repair. In a
// healthy chain nonces decrease monotonically via PrevHash and the loop exits
// on the nonce check; the bound only matters if storage is corrupt or cyclic.
const maxForwardPendingWalkIter = 100000

// ShardedCacheSearchMethod defines the algorithm for searching through a sharded cache
type ShardedCacheSearchMethod byte

const (
	// SearchMethodJustPeek will make the algorithm invoke just Peek method
	SearchMethodJustPeek ShardedCacheSearchMethod = iota

	// SearchMethodSearchFirst will make the algorithm invoke just SearchFirst method
	SearchMethodSearchFirst

	// SearchMethodPeekWithFallbackSearchFirst will first try a Peek method. If the data is not found will fall back
	// to SearchFirst method
	SearchMethodPeekWithFallbackSearchFirst
)

// ToString converts the ShardedCacheSearchMethod to its string representation
func (method ShardedCacheSearchMethod) ToString() string {
	switch method {
	case SearchMethodJustPeek:
		return "just peek"
	case SearchMethodSearchFirst:
		return "search first"
	case SearchMethodPeekWithFallbackSearchFirst:
		return "peek with fallback to search first"
	default:
		return fmt.Sprintf("unknown method %d", method)
	}
}

// GetShardHeader gets the header, which is associated with the given hash, from pool or storage
func GetShardHeader(
	hash []byte,
	headersCacher dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (data.ShardHeaderHandler, error) {

	err := checkGetHeaderParamsForNil(headersCacher, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr, err := GetShardHeaderFromPool(hash, headersCacher)
	if err != nil {
		hdr, err = GetShardHeaderFromStorage(hash, marshalizer, storageService)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// GetMetaHeader gets the header, which is associated with the given hash, from pool or storage
func GetMetaHeader(
	hash []byte,
	headersCacher dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {

	err := checkGetHeaderParamsForNil(headersCacher, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr, err := GetMetaHeaderFromPool(hash, headersCacher)
	if err != nil {
		hdr, err = GetMetaHeaderFromStorage(hash, marshalizer, storageService)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// GetShardHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetShardHeaderFromPool(
	hash []byte,
	headersCacher dataRetriever.HeadersPool,
) (data.ShardHeaderHandler, error) {

	obj, err := getHeaderFromPool(hash, headersCacher)
	if err != nil {
		return nil, err
	}

	hdr, ok := obj.(data.ShardHeaderHandler)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return hdr, nil
}

// GetMetaHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetMetaHeaderFromPool(
	hash []byte,
	headersCacher dataRetriever.HeadersPool,
) (*block.MetaBlock, error) {

	obj, err := getHeaderFromPool(hash, headersCacher)
	if err != nil {
		return nil, err
	}

	hdr, ok := obj.(*block.MetaBlock)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return hdr, nil
}

// GetHeaderFromStorage method returns a block header from storage
func GetHeaderFromStorage(
	shardId uint32,
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (data.HeaderHandler, error) {
	if shardId == core.MetachainShardId {
		return GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	}
	return GetShardHeaderFromStorage(hash, marshalizer, storageService)
}

// GetShardHeaderFromStorage gets the header, which is associated with the given hash, from storage
func GetShardHeaderFromStorage(
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (data.ShardHeaderHandler, error) {

	buffHdr, err := GetMarshalizedHeaderFromStorage(dataRetriever.BlockHeaderUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr, err := UnmarshalShardHeader(marshalizer, buffHdr)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return hdr, nil
}

// GetMetaHeaderFromStorage gets the header, which is associated with the given hash, from storage
func GetMetaHeaderFromStorage(
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {

	buffHdr, err := GetMarshalizedHeaderFromStorage(dataRetriever.MetaBlockUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr := &block.MetaBlock{}
	err = marshalizer.Unmarshal(hdr, buffHdr)
	if err != nil {
		return nil, ErrUnmarshalWithoutSuccess
	}

	return hdr, nil
}

// GetMarshalizedHeaderFromStorage gets the marshalized header, which is associated with the given hash, from storage
func GetMarshalizedHeaderFromStorage(
	blockUnit dataRetriever.UnitType,
	hash []byte,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) ([]byte, error) {

	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(storageService) {
		return nil, ErrNilStorage
	}

	hdrStore, err := storageService.GetStorer(blockUnit)
	if err != nil {
		return nil, err
	}

	buffHdr, err := hdrStore.Get(hash)
	if err != nil {
		return nil, fmt.Errorf("%w : GetMarshalizedHeaderFromStorage hash = %s",
			ErrMissingHeader, logger.DisplayByteSlice(hash))
	}

	return buffHdr, nil
}

// GetShardHeaderWithNonce method returns a shard block header with a given nonce and shardId
func GetShardHeaderWithNonce(
	nonce uint64,
	shardId uint32,
	headersCacher dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) (data.HeaderHandler, []byte, error) {

	err := checkGetHeaderWithNonceParamsForNil(headersCacher, marshalizer, storageService, uint64Converter)
	if err != nil {
		return nil, nil, err
	}

	hdr, hash, err := GetShardHeaderFromPoolWithNonce(nonce, shardId, headersCacher)
	if err != nil {
		hdr, hash, err = GetShardHeaderFromStorageWithNonce(nonce, shardId, storageService, uint64Converter, marshalizer)
		if err != nil {
			return nil, nil, err
		}
	}

	return hdr, hash, nil
}

// GetMetaHeaderWithNonce method returns a meta block header with a given nonce
func GetMetaHeaderWithNonce(
	nonce uint64,
	headersCacher dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) (*block.MetaBlock, []byte, error) {

	err := checkGetHeaderWithNonceParamsForNil(headersCacher, marshalizer, storageService, uint64Converter)
	if err != nil {
		return nil, nil, err
	}

	hdr, hash, err := GetMetaHeaderFromPoolWithNonce(nonce, headersCacher)
	if err != nil {
		hdr, hash, err = GetMetaHeaderFromStorageWithNonce(nonce, storageService, uint64Converter, marshalizer)
		if err != nil {
			return nil, nil, err
		}
	}

	return hdr, hash, nil
}

// GetShardHeaderFromPoolWithNonce method returns a shard block header from pool with a given nonce and shardId
func GetShardHeaderFromPoolWithNonce(
	nonce uint64,
	shardId uint32,
	headersCacher dataRetriever.HeadersPool,
) (data.HeaderHandler, []byte, error) {

	obj, hash, err := getHeaderFromPoolWithNonce(nonce, shardId, headersCacher)
	if err != nil {
		return nil, nil, err
	}

	hdr, ok := obj.(data.ShardHeaderHandler)
	if !ok {
		return nil, nil, ErrWrongTypeAssertion
	}

	return hdr, hash, nil
}

// GetMetaHeaderFromPoolWithNonce method returns a meta block header from pool with a given nonce
func GetMetaHeaderFromPoolWithNonce(
	nonce uint64,
	headersCacher dataRetriever.HeadersPool,
) (*block.MetaBlock, []byte, error) {

	obj, hash, err := getHeaderFromPoolWithNonce(nonce, core.MetachainShardId, headersCacher)
	if err != nil {
		return nil, nil, err
	}

	hdr, ok := obj.(*block.MetaBlock)
	if !ok {
		return nil, nil, ErrWrongTypeAssertion
	}

	return hdr, hash, nil
}

// GetHeaderFromStorageWithNonce method returns a block header from storage with a given nonce and shardId
func GetHeaderFromStorageWithNonce(
	nonce uint64,
	shardId uint32,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
) (data.HeaderHandler, []byte, error) {

	if shardId == core.MetachainShardId {
		return GetMetaHeaderFromStorageWithNonce(nonce, storageService, uint64Converter, marshalizer)
	}
	return GetShardHeaderFromStorageWithNonce(nonce, shardId, storageService, uint64Converter, marshalizer)
}

// GetShardHeaderFromStorageWithNonce method returns a shard block header from storage with a given nonce and shardId
func GetShardHeaderFromStorageWithNonce(
	nonce uint64,
	shardId uint32,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
) (data.HeaderHandler, []byte, error) {

	hash, err := GetHeaderHashFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer,
		dataRetriever.GetHdrNonceHashDataUnit(shardId))
	if err != nil {
		return nil, nil, err
	}

	hdr, err := GetShardHeaderFromStorage(hash, marshalizer, storageService)
	if err != nil {
		return nil, nil, err
	}

	return hdr, hash, nil
}

// GetMetaHeaderFromStorageWithNonce method returns a meta block header from storage with a given nonce
func GetMetaHeaderFromStorageWithNonce(
	nonce uint64,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
) (*block.MetaBlock, []byte, error) {

	hash, err := GetHeaderHashFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer,
		dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		return nil, nil, err
	}

	hdr, err := GetMetaHeaderFromStorage(hash, marshalizer, storageService)
	if err != nil {
		return nil, nil, err
	}

	return hdr, hash, nil
}

// GetTransactionHandler gets the transaction with a given sender/receiver shardId and txHash
func GetTransactionHandler(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
	shardedDataCacherNotifier dataRetriever.ShardedDataCacherNotifier,
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	method ShardedCacheSearchMethod,
) (data.TransactionHandler, error) {

	err := checkGetTransactionParamsForNil(shardedDataCacherNotifier, storageService, marshalizer)
	if err != nil {
		return nil, err
	}

	tx, err := GetTransactionHandlerFromPool(senderShardID, destShardID, txHash, shardedDataCacherNotifier, method)
	if err != nil {
		tx, err = GetTransactionHandlerFromStorage(txHash, storageService, marshalizer)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// GetTransactionHandlerFromPool gets the transaction from pool with a given sender/receiver shardId and txHash
func GetTransactionHandlerFromPool(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
	shardedDataCacherNotifier dataRetriever.ShardedDataCacherNotifier,
	method ShardedCacheSearchMethod,
) (data.TransactionHandler, error) {

	if check.IfNil(shardedDataCacherNotifier) {
		return nil, ErrNilShardedDataCacherNotifier
	}

	return getTransactionHandlerFromPool(senderShardID, destShardID, txHash, shardedDataCacherNotifier, method)
}

func getTransactionHandlerFromPool(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
	shardedDataCacherNotifier dataRetriever.ShardedDataCacherNotifier,
	method ShardedCacheSearchMethod,
) (data.TransactionHandler, error) {
	var val interface{}
	var ok bool

	if method == SearchMethodSearchFirst {
		val, ok = shardedDataCacherNotifier.SearchFirstData(txHash)

		return castDataFromCacheAsTransactionHandler(val, ok)
	}

	strCache := ShardCacherIdentifier(senderShardID, destShardID)
	txStore := shardedDataCacherNotifier.ShardDataStore(strCache)
	if txStore == nil {
		return nil, ErrNilStorage
	}

	switch method {
	case SearchMethodJustPeek:
		val, ok = txStore.Peek(txHash)
	case SearchMethodPeekWithFallbackSearchFirst:
		val, ok = txStore.Peek(txHash)
		if !ok {
			val, ok = shardedDataCacherNotifier.SearchFirstData(txHash)
		}
	default:
		return nil, fmt.Errorf("%w for provided method: %s in getTransactionHandlerFromPool",
			ErrInvalidValue, method.ToString())
	}

	return castDataFromCacheAsTransactionHandler(val, ok)
}

func castDataFromCacheAsTransactionHandler(val interface{}, ok bool) (data.TransactionHandler, error) {
	if !ok {
		return nil, ErrTxNotFound
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		return nil, ErrInvalidTxInPool
	}

	return tx, nil
}

// GetTransactionHandlerFromStorage gets the transaction from storage with a given sender/receiver shardId and txHash
func GetTransactionHandlerFromStorage(
	txHash []byte,
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
) (data.TransactionHandler, error) {

	if storageService == nil || storageService.IsInterfaceNil() {
		return nil, ErrNilStorage
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}

	txBuff, err := storageService.Get(dataRetriever.TransactionUnit, txHash)
	if err != nil {
		return nil, err
	}

	tx := transaction.Transaction{}
	err = marshalizer.Unmarshal(&tx, txBuff)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func checkGetHeaderParamsForNil(
	cacher dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) error {

	if cacher == nil || cacher.IsInterfaceNil() {
		return ErrNilCacher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return ErrNilMarshalizer
	}
	if storageService == nil || storageService.IsInterfaceNil() {
		return ErrNilStorage
	}

	return nil
}

func checkGetHeaderWithNonceParamsForNil(
	headersCacher dataRetriever.HeadersPool,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) error {

	err := checkGetHeaderParamsForNil(headersCacher, marshalizer, storageService)
	if err != nil {
		return err
	}
	if check.IfNil(uint64Converter) {
		return ErrNilUint64Converter
	}

	return nil
}

func checkGetTransactionParamsForNil(
	shardedDataCacherNotifier dataRetriever.ShardedDataCacherNotifier,
	storageService dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
) error {

	if shardedDataCacherNotifier == nil || shardedDataCacherNotifier.IsInterfaceNil() {
		return ErrNilShardedDataCacherNotifier
	}
	if storageService == nil || storageService.IsInterfaceNil() {
		return ErrNilStorage
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return ErrNilMarshalizer
	}

	return nil
}

func getHeaderFromPool(
	hash []byte,
	headersCacher dataRetriever.HeadersPool,
) (interface{}, error) {

	if check.IfNil(headersCacher) {
		return nil, ErrNilCacher
	}

	obj, err := headersCacher.GetHeaderByHash(hash)
	if err != nil {
		return nil, fmt.Errorf("%w : getHeaderFromPool hash = %s",
			ErrMissingHeader, logger.DisplayByteSlice(hash))
	}

	return obj, nil
}

func getHeaderFromPoolWithNonce(
	nonce uint64,
	shardId uint32,
	headersCacher dataRetriever.HeadersPool,
) (interface{}, []byte, error) {

	if check.IfNil(headersCacher) {
		return nil, nil, ErrNilCacher
	}

	headers, hashes, err := headersCacher.GetHeadersByNonceAndShardId(nonce, shardId)
	if err != nil {
		return nil, nil, fmt.Errorf("%w : getHeaderFromPoolWithNonce shard = %d nonce = %d",
			ErrMissingHeader, shardId, nonce)
	}

	// TODO what should we do when we get from pool more than one header with same nonce and shardId
	return headers[len(headers)-1], hashes[len(hashes)-1], nil
}

// GetHeaderHashFromStorageWithNonce gets a header hash, given a nonce
func GetHeaderHashFromStorageWithNonce(
	nonce uint64,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
	blockUnit dataRetriever.UnitType,
) ([]byte, error) {

	if storageService == nil || storageService.IsInterfaceNil() {
		return nil, ErrNilStorage
	}
	if uint64Converter == nil || uint64Converter.IsInterfaceNil() {
		return nil, ErrNilUint64Converter
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}

	headerStore, err := storageService.GetStorer(blockUnit)
	if err != nil {
		return nil, err
	}

	nonceToByteSlice := uint64Converter.ToByteSlice(nonce)
	hash, err := headerStore.Get(nonceToByteSlice)
	if err != nil {
		return nil, ErrMissingHashForHeaderNonce
	}

	return hash, nil
}

// SortHeadersByNonce will sort a given list of headers by nonce
func SortHeadersByNonce(headers []data.HeaderHandler) {
	if len(headers) > 1 {
		sort.Slice(headers, func(i, j int) bool {
			return headers[i].GetNonce() < headers[j].GetNonce()
		})
	}
}

// IsInProperRound checks if the given round index satisfies the round modulus trigger
func IsInProperRound(index int64) bool {
	return index%RoundModulusTrigger == 0
}

// AddHeaderToBlackList adds a hash to black list handler. Logs if the operation did not succeed
func AddHeaderToBlackList(blackListHandler TimeCacher, hash []byte) {
	blackListHandler.Sweep()
	err := blackListHandler.Add(string(hash))
	if err != nil {
		log.Trace("blackListHandler.Add", "error", err.Error())
	}

	log.Debug("header has been added to blacklist",
		"hash", hash)
}

// ForkInfo hold the data related to a detected fork
type ForkInfo struct {
	IsDetected bool
	Nonce      uint64
	Round      uint64
	Hash       []byte
}

// NewForkInfo creates a new ForkInfo object
func NewForkInfo() *ForkInfo {
	return &ForkInfo{IsDetected: false, Nonce: math.MaxUint64, Round: math.MaxUint64, Hash: nil}
}

// DisplayProcessTxDetails displays information related to the tx which should be executed
func DisplayProcessTxDetails(
	message string,
	accountHandler vmcommon.AccountHandler,
	txHandler data.TransactionHandler,
	txHash []byte,
	addressPubkeyConverter core.PubkeyConverter,
) {
	if log.GetLevel() > logger.LogTrace {
		return
	}

	if !check.IfNil(accountHandler) {
		account, ok := accountHandler.(state.UserAccountHandler)
		if ok {
			log.Trace(message,
				"nonce", account.GetNonce(),
				"balance", account.GetBalance(),
			)
		}
	}

	if check.IfNil(addressPubkeyConverter) {
		return
	}
	if check.IfNil(txHandler) {
		return
	}

	receiverAddress, _ := addressPubkeyConverter.Encode(txHandler.GetRcvAddr())
	senderAddress, _ := addressPubkeyConverter.Encode(txHandler.GetSndAddr())

	log.Trace("executing transaction",
		"txHash", txHash,
		"nonce", txHandler.GetNonce(),
		"value", txHandler.GetValue(),
		"gas limit", txHandler.GetGasLimit(),
		"gas price", txHandler.GetGasPrice(),
		"data", hex.EncodeToString(txHandler.GetData()),
		"sender", senderAddress,
		"receiver", receiverAddress)
}

// IsAllowedToSaveUnderKey returns if saving key-value in data tries under given key is allowed
func IsAllowedToSaveUnderKey(key []byte) bool {
	vmStoragePrefix := core.ProtectedKeyPrefix + VMStoragePrefix
	vmPrefixLen := len(vmStoragePrefix)
	if len(key) > vmPrefixLen {
		trimmedKey := key[:len(vmStoragePrefix)]
		if bytes.Equal(trimmedKey, []byte(vmStoragePrefix)) {
			return true
		}
	}

	prefixLen := len(core.ProtectedKeyPrefix)
	if len(key) < prefixLen {
		return true
	}

	trimmedKey := key[:prefixLen]
	return !bytes.Equal(trimmedKey, []byte(core.ProtectedKeyPrefix))
}

// SortVMOutputInsideData returns the output accounts as a sorted list
func SortVMOutputInsideData(vmOutput *vmcommon.VMOutput) []*vmcommon.OutputAccount {
	sort.Slice(vmOutput.DeletedAccounts, func(i, j int) bool {
		return bytes.Compare(vmOutput.DeletedAccounts[i], vmOutput.DeletedAccounts[j]) < 0
	})
	sort.Slice(vmOutput.TouchedAccounts, func(i, j int) bool {
		return bytes.Compare(vmOutput.TouchedAccounts[i], vmOutput.TouchedAccounts[j]) < 0
	})

	outPutAccounts := make([]*vmcommon.OutputAccount, len(vmOutput.OutputAccounts))
	i := 0
	for _, outAcc := range vmOutput.OutputAccounts {
		outPutAccounts[i] = outAcc
		i++
	}

	sort.Slice(outPutAccounts, func(i, j int) bool {
		return bytes.Compare(outPutAccounts[i].Address, outPutAccounts[j].Address) < 0
	})

	return outPutAccounts
}

// GetSortedStorageUpdates returns the storage updates as a sorted list
func GetSortedStorageUpdates(account *vmcommon.OutputAccount) []*vmcommon.StorageUpdate {
	storageUpdates := make([]*vmcommon.StorageUpdate, len(account.StorageUpdates))
	i := 0
	for _, update := range account.StorageUpdates {
		storageUpdates[i] = update
		i++
	}

	sort.Slice(storageUpdates, func(i, j int) bool {
		return bytes.Compare(storageUpdates[i].Offset, storageUpdates[j].Offset) < 0
	})

	return storageUpdates
}

// GetHeader tries to get the header from pool first and if not found, searches for it through storer
func GetHeader(
	headerHash []byte,
	headersPool dataRetriever.HeadersPool,
	headersStorer dataRetriever.StorageService,
	marshaller marshal.Marshalizer,
	shardID uint32,
) (data.HeaderHandler, error) {
	if shardID == core.MetachainShardId {
		return GetMetaHeader(headerHash, headersPool, marshaller, headersStorer)
	}

	return GetShardHeader(headerHash, headersPool, marshaller, headersStorer)
}

// UnmarshalHeader unmarshalls a block header
func UnmarshalHeader(shardId uint32, marshalizer marshal.Marshalizer, headerBuffer []byte) (data.HeaderHandler, error) {
	if shardId == core.MetachainShardId {
		return UnmarshalMetaHeader(marshalizer, headerBuffer)
	} else {
		return UnmarshalShardHeader(marshalizer, headerBuffer)
	}
}

// UnmarshalMetaHeader unmarshalls a meta header
func UnmarshalMetaHeader(marshalizer marshal.Marshalizer, headerBuffer []byte) (data.MetaHeaderHandler, error) {
	header := &block.MetaBlock{}
	err := marshalizer.Unmarshal(header, headerBuffer)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// UnmarshalShardHeader unmarshalls a shard header
func UnmarshalShardHeader(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
	hdr, err := UnmarshalShardHeaderV2(marshalizer, hdrBuff)
	if err == nil {
		return hdr, nil
	}

	hdr, err = UnmarshalShardHeaderV1(marshalizer, hdrBuff)
	return hdr, err
}

// UnmarshalShardHeaderV2 unmarshalls a header with version 2
func UnmarshalShardHeaderV2(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
	hdrV2 := &block.HeaderV2{}
	err := marshalizer.Unmarshal(hdrV2, hdrBuff)
	if err != nil {
		return nil, err
	}
	if check.IfNil(hdrV2.Header) {
		return nil, fmt.Errorf("%w while checking inner header", ErrNilHeaderHandler)
	}

	return hdrV2, nil
}

// UnmarshalShardHeaderV1 unmarshalls a header with version 1
func UnmarshalShardHeaderV1(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
	hdr := &block.Header{}
	err := marshalizer.Unmarshal(hdr, hdrBuff)
	if err != nil {
		return nil, err
	}

	return hdr, nil
}

// IsScheduledMode returns true if the first mini block from the given body is marked as a scheduled
func IsScheduledMode(
	header data.HeaderHandler,
	body *block.Body,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) (bool, error) {
	if body == nil || len(body.MiniBlocks) == 0 {
		return false, nil
	}

	miniBlockHash, err := core.CalculateHash(marshalizer, hasher, body.MiniBlocks[0])
	if err != nil {
		return false, err
	}

	for _, miniBlockHeader := range header.GetMiniBlockHeaderHandlers() {
		if bytes.Equal(miniBlockHash, miniBlockHeader.GetHash()) {
			return miniBlockHeader.GetProcessingType() == int32(block.Scheduled), nil
		}
	}

	return false, nil
}

const additionalTimeForCreatingScheduledMiniBlocks = 150 * time.Millisecond

// HaveAdditionalTime returns if the additional time allocated for scheduled mini blocks is elapsed
func HaveAdditionalTime() func() bool {
	startTime := time.Now()
	return func() bool {
		return additionalTimeForCreatingScheduledMiniBlocks > time.Since(startTime)
	}
}

// GetZeroGasAndFees returns a zero value structure for the gas and fees
func GetZeroGasAndFees() scheduled.GasAndFees {
	return scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
}

// ScheduledInfo holds all the info needed for scheduled SC execution
type ScheduledInfo struct {
	RootHash        []byte
	IntermediateTxs map[block.Type][]data.TransactionHandler
	GasAndFees      scheduled.GasAndFees
	MiniBlocks      block.MiniBlockSlice
}

// GetFinalCrossMiniBlockHashes returns all the finalized miniblocks hashes, from the given header and with the given destination
func GetFinalCrossMiniBlockHashes(header data.HeaderHandler, shardID uint32) map[string]uint32 {
	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(shardID)

	miniBlockHashes := make(map[string]uint32)
	for crossMiniBlockHash, senderShardID := range crossMiniBlockHashes {
		miniBlockHeader := GetMiniBlockHeaderWithHash(header, []byte(crossMiniBlockHash))
		if miniBlockHeader != nil && !miniBlockHeader.IsFinal() {
			log.Debug("GetFinalCrossMiniBlockHashes: skip mini block which is not final", "mb hash", miniBlockHeader.GetHash())
			continue
		}

		miniBlockHashes[crossMiniBlockHash] = senderShardID
	}

	return miniBlockHashes
}

// GetMiniBlockHeaderWithHash returns the miniblock header with the given hash
func GetMiniBlockHeaderWithHash(header data.HeaderHandler, miniBlockHash []byte) data.MiniBlockHeaderHandler {
	for _, miniBlockHeader := range header.GetMiniBlockHeaderHandlers() {
		if bytes.Equal(miniBlockHeader.GetHash(), miniBlockHash) {
			return miniBlockHeader
		}
	}
	return nil
}

// IsBuiltinFuncCallWithParam checks if the given transaction data represents a builtin function call with parameters
func IsBuiltinFuncCallWithParam(txData []byte, function string) bool {
	expectedTxDataPrefix := []byte(function + "@")
	return bytes.HasPrefix(txData, expectedTxDataPrefix)
}

// IsSetGuardianCall checks if the given transaction data represents the set guardian builtin function call
func IsSetGuardianCall(txData []byte) bool {
	return IsBuiltinFuncCallWithParam(txData, core.BuiltInFunctionSetGuardian)
}

// CheckIfIndexesAreOutOfBound checks if the given indexes are out of bound for the given mini block
func CheckIfIndexesAreOutOfBound(
	indexOfFirstTxToBeProcessed int32,
	indexOfLastTxToBeProcessed int32,
	miniBlock *block.MiniBlock,
) error {
	maxIndex := int32(len(miniBlock.TxHashes)) - 1

	isFirstIndexHigherThanLastIndex := indexOfFirstTxToBeProcessed > indexOfLastTxToBeProcessed
	isFirstIndexOutOfRange := indexOfFirstTxToBeProcessed < 0 || indexOfFirstTxToBeProcessed > maxIndex
	isLastIndexOutOfRange := indexOfLastTxToBeProcessed < 0 || indexOfLastTxToBeProcessed > maxIndex

	isIndexOutOfBound := isFirstIndexHigherThanLastIndex || isFirstIndexOutOfRange || isLastIndexOutOfRange
	if isIndexOutOfBound {
		return fmt.Errorf("%w: indexOfFirstTxToBeProcessed: %d, indexOfLastTxToBeProcessed = %d, maxIndex: %d",
			ErrIndexIsOutOfBound,
			indexOfFirstTxToBeProcessed,
			indexOfLastTxToBeProcessed,
			maxIndex,
		)
	}

	return nil
}

// CompleteMissingSelfNotarizedHeaders fills self-notarized headers for shards still at
// nonce 0. Fast path uses FirstPendingMetaBlock on epoch start metablocks; fallback walks
// back from startHash inspecting ShardInfo. Returns an error if any shard remains
// unresolved and the walk did not reach genesis (legitimate empty state).
func CompleteMissingSelfNotarizedHeaders(
	startHash []byte,
	numShards uint32,
	blockTracker BlockTracker,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) error {
	missingShards := make(map[uint32]bool)
	for shardID := uint32(0); shardID < numShards; shardID++ {
		lastSelfNotarized, _, err := blockTracker.GetLastSelfNotarizedHeader(shardID)
		if err != nil || check.IfNil(lastSelfNotarized) || lastSelfNotarized.GetNonce() == 0 {
			missingShards[shardID] = true
		}
	}

	if len(missingShards) == 0 {
		return nil
	}

	log.Debug("CompleteMissingSelfNotarizedHeaders",
		"numMissing", len(missingShards))

	startMetaBlock, err := GetMetaHeaderFromStorage(startHash, marshalizer, store)
	if err != nil {
		return fmt.Errorf("%w: could not load start metablock %s: %s",
			ErrMissingHeader, hex.EncodeToString(startHash), err.Error())
	}

	if startMetaBlock.IsStartOfEpochBlock() {
		deriveSelfNotarizedFromEpochStartData(startMetaBlock, missingShards, blockTracker, marshalizer, store)
		if len(missingShards) == 0 {
			return nil
		}
	}

	reachedGenesis := false
	currentHash := startHash
	for i := 0; i < maxSelfNotarizedLookback && len(missingShards) > 0 && len(currentHash) > 0; i++ {
		metaBlock, errGet := GetMetaHeaderFromStorage(currentHash, marshalizer, store)
		if errGet != nil {
			log.Debug("CompleteMissingSelfNotarizedHeaders: could not load meta block during walk back", "error", errGet.Error())
			break
		}

		for shardID := range missingShards {
			bestNonce, bestHeader, bestHash := findSelfNotarizedMetaHeaderInBlock(metaBlock, shardID, marshalizer, store)
			if bestHeader != nil {
				log.Debug("CompleteMissingSelfNotarizedHeaders: derived self-notarized header",
					"shardID", shardID,
					"metaNonce", bestNonce,
					"metaHash", bestHash)
				blockTracker.AddSelfNotarizedHeader(shardID, bestHeader, bestHash)
				delete(missingShards, shardID)
			}
		}

		if metaBlock.GetNonce() == 0 {
			reachedGenesis = true
			break
		}
		currentHash = metaBlock.GetPrevHash()
	}

	if len(missingShards) == 0 {
		return nil
	}

	if reachedGenesis {
		log.Debug("CompleteMissingSelfNotarizedHeaders: reached genesis, nothing to derive",
			"numStillMissing", len(missingShards))
		return nil
	}

	log.Warn("CompleteMissingSelfNotarizedHeaders: could not derive all self-notarized headers",
		"numStillMissing", len(missingShards))
	return fmt.Errorf("%w: could not derive self-notarized headers for %d shards",
		ErrMissingHeader, len(missingShards))
}

func deriveSelfNotarizedFromEpochStartData(
	epochStartMetaBlock data.MetaHeaderHandler,
	missingShards map[uint32]bool,
	blockTracker BlockTracker,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) {
	epochStartHandler := epochStartMetaBlock.GetEpochStartHandler()
	if epochStartHandler == nil {
		return
	}

	for _, shardData := range epochStartHandler.GetLastFinalizedHeaderHandlers() {
		shardID := shardData.GetShardID()
		if !missingShards[shardID] {
			continue
		}

		metaHash := shardData.GetFirstPendingMetaBlock()
		if len(metaHash) == 0 {
			continue
		}

		selfNotarizedMeta, err := GetMetaHeaderFromStorage(metaHash, marshalizer, store)
		if err != nil {
			log.Debug("deriveSelfNotarizedFromEpochStartData: could not load meta block",
				"shardID", shardID,
				"metaHash", metaHash,
				"error", err.Error())
			continue
		}

		log.Debug("deriveSelfNotarizedFromEpochStartData: derived self-notarized header",
			"shardID", shardID,
			"metaNonce", selfNotarizedMeta.GetNonce(),
			"metaHash", metaHash)
		blockTracker.AddSelfNotarizedHeader(shardID, selfNotarizedMeta, metaHash)
		delete(missingShards, shardID)
	}
}

func findSelfNotarizedMetaHeaderInBlock(
	metaBlock data.MetaHeaderHandler,
	shardID uint32,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) (uint64, data.HeaderHandler, []byte) {
	var bestNonce uint64
	var bestHeader data.HeaderHandler
	var bestHash []byte
	hadLoadErrors := false

	shardInfoHandlers := metaBlock.GetShardInfoHandlers()
	for i := range shardInfoHandlers {
		if shardInfoHandlers[i].GetShardID() != shardID {
			continue
		}

		headerHash := shardInfoHandlers[i].GetHeaderHash()
		shardHeader, err := GetShardHeaderFromStorage(headerHash, marshalizer, store)
		if err != nil {
			log.Debug("findSelfNotarizedMetaHeaderInBlock: could not load shard header",
				"shardID", shardID,
				"headerHash", headerHash,
				"error", err.Error())
			hadLoadErrors = true
			continue
		}

		for _, metaHash := range shardHeader.GetMetaBlockHashes() {
			metaHeader, errGet := GetMetaHeaderFromStorage(metaHash, marshalizer, store)
			if errGet != nil {
				continue
			}

			if metaHeader.GetNonce() > bestNonce {
				bestNonce = metaHeader.GetNonce()
				bestHeader = metaHeader
				bestHash = metaHash
			}
		}
	}

	if hadLoadErrors {
		return 0, nil, nil
	}

	return bestNonce, bestHeader, bestHash
}

// RepairPendingMiniBlocks recomputes, from chain state, the pending cross-shard
// miniblocks that should be held by the pending-miniblocks handler at the bootstrap
// point (last committed meta block). Per shard, it anchors the computation at the
// last meta that shard has referenced (block tracker's self-notarized meta header),
// combining:
//   - miniblocks inside the anchor meta destined for that shard that are not yet
//     marked final in the shard headers that reference it, and
//   - every cross-shard miniblock destined for that shard in meta blocks strictly
//     newer than the shard's anchor, up to lastCommittedMetaHash.
//
// If lastCommittedMetaHash itself is an epoch-start meta block, the anchor-based
// computation is skipped and the pending set is read directly from the epoch-start
// shard data (that is the post-reset handler state every running node has).
//
// Returns hashes grouped by receiver shard.
func RepairPendingMiniBlocks(
	lastCommittedMetaHash []byte,
	numShards uint32,
	blockTracker BlockTracker,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) (map[uint32][][]byte, error) {
	lastMeta, err := GetMetaHeaderFromStorage(lastCommittedMetaHash, marshalizer, store)
	if err != nil {
		return nil, fmt.Errorf("%w: could not load last meta block %s: %s",
			ErrMissingHeader, hex.EncodeToString(lastCommittedMetaHash), err.Error())
	}

	if lastMeta.IsStartOfEpochBlock() {
		return pendingFromEpochStart(lastMeta), nil
	}

	anchorByShard := make(map[uint32]data.HeaderHandler, numShards)
	anchorHashByShard := make(map[uint32][]byte, numShards)
	for shardID := uint32(0); shardID < numShards; shardID++ {
		anchorMeta, anchorHash, errAnchor := blockTracker.GetLastSelfNotarizedHeader(shardID)
		if errAnchor != nil || check.IfNil(anchorMeta) {
			log.Debug("RepairPendingMiniBlocks: no anchor for shard, skipping",
				"shardID", shardID,
				"error", errAnchor)
			continue
		}
		anchorByShard[shardID] = anchorMeta
		anchorHashByShard[shardID] = anchorHash
	}

	pendingByShard := make(map[uint32][][]byte, numShards)

	for shardID, anchorMeta := range anchorByShard {
		shardHdrs := collectShardHeadersReferencingAnchor(shardID, anchorHashByShard[shardID], blockTracker, marshalizer, store)
		pending := computePendingMBsAtAnchor(anchorMeta, shardID, shardHdrs)
		for hash := range pending {
			pendingByShard[shardID] = append(pendingByShard[shardID], []byte(hash))
		}
	}

	forward, err := collectForwardPendingMBsByShard(lastMeta, anchorByShard, marshalizer, store)
	if err != nil {
		return nil, err
	}
	for shardID, hashes := range forward {
		existing := make(map[string]struct{}, len(pendingByShard[shardID]))
		for _, h := range pendingByShard[shardID] {
			existing[string(h)] = struct{}{}
		}
		for _, h := range hashes {
			if _, ok := existing[string(h)]; ok {
				continue
			}
			pendingByShard[shardID] = append(pendingByShard[shardID], h)
		}
	}

	return pendingByShard, nil
}

func pendingFromEpochStart(epochStartMeta data.MetaHeaderHandler) map[uint32][][]byte {
	out := make(map[uint32][][]byte)
	epochStartHandler := epochStartMeta.GetEpochStartHandler()
	if epochStartHandler == nil {
		return out
	}
	for _, shardData := range epochStartHandler.GetLastFinalizedHeaderHandlers() {
		for _, mbh := range shardData.GetPendingMiniBlockHeaderHandlers() {
			if !shouldConsiderCrossShardMiniBlockForRepair(mbh.GetSenderShardID(), mbh.GetReceiverShardID()) {
				continue
			}
			out[mbh.GetReceiverShardID()] = append(out[mbh.GetReceiverShardID()], mbh.GetHash())
		}
	}
	return out
}

// shouldConsiderCrossShardMiniBlockForRepair mirrors the filter applied by the
// pending-miniblocks handler when building its map: same-shard, shard->meta and
// meta->all miniblocks are not tracked as cross-shard pending entries.
func shouldConsiderCrossShardMiniBlockForRepair(senderShardID uint32, receiverShardID uint32) bool {
	if senderShardID == receiverShardID {
		return false
	}
	if senderShardID != core.MetachainShardId && receiverShardID == core.MetachainShardId {
		return false
	}
	if senderShardID == core.MetachainShardId && receiverShardID == core.AllShardId {
		return false
	}
	return true
}

// collectShardHeadersReferencingAnchor walks the shard chain backward from the
// last cross-notarized shard header, returning the contiguous run of shard
// headers that still reference the anchor meta among their MetaBlockHashes.
// The shard cannot advance past a meta without processing every MB it owes to
// itself in that meta, so only those shard headers can have "drained" MBs from
// the anchor.
func collectShardHeadersReferencingAnchor(
	shardID uint32,
	anchorMetaHash []byte,
	blockTracker BlockTracker,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) []data.HeaderHandler {
	shardHdrs := make([]data.HeaderHandler, 0)
	if len(anchorMetaHash) == 0 {
		return shardHdrs
	}

	currShardHdr, _, err := blockTracker.GetLastCrossNotarizedHeader(shardID)
	if err != nil || check.IfNil(currShardHdr) {
		return shardHdrs
	}

	for i := 0; i < maxSelfNotarizedLookback; i++ {
		if !shardHeaderReferencesMeta(currShardHdr, anchorMetaHash) {
			break
		}
		shardHdrs = append(shardHdrs, currShardHdr)

		prevHash := currShardHdr.GetPrevHash()
		if len(prevHash) == 0 {
			break
		}
		prevShardHdr, errLoad := GetShardHeaderFromStorage(prevHash, marshalizer, store)
		if errLoad != nil {
			log.Debug("collectShardHeadersReferencingAnchor: could not load previous shard header",
				"shardID", shardID,
				"prevHash", prevHash,
				"error", errLoad.Error())
			break
		}
		currShardHdr = prevShardHdr
	}

	return shardHdrs
}

func shardHeaderReferencesMeta(shardHdr data.HeaderHandler, metaHash []byte) bool {
	shardHeader, ok := shardHdr.(data.ShardHeaderHandler)
	if !ok {
		return false
	}
	for _, h := range shardHeader.GetMetaBlockHashes() {
		if bytes.Equal(h, metaHash) {
			return true
		}
	}
	return false
}

// computePendingMBsAtAnchor returns cross-shard miniblocks inside anchorMeta that
// target shardID and have not been marked final by any of the given shard headers.
func computePendingMBsAtAnchor(
	anchorMeta data.HeaderHandler,
	shardID uint32,
	shardHdrs []data.HeaderHandler,
) map[string]struct{} {
	pending := getAllCrossShardMBsForShardInMeta(anchorMeta, shardID)

	for _, shardHdr := range shardHdrs {
		for _, smbh := range shardHdr.GetMiniBlockHeaderHandlers() {
			hash := string(smbh.GetHash())
			if _, ok := pending[hash]; !ok {
				continue
			}
			if smbh.IsFinal() {
				delete(pending, hash)
			}
		}
	}

	return pending
}

// getAllCrossShardMBsForShardInMeta returns cross-shard MB hashes destined for
// dstShardID in a meta block: meta-body MBs (e.g. rewards meta->shard) and
// outgoing MBs in other shards' notarized headers. ShardInfo[dstShardID] is
// skipped so the destination's own block's already-consumed incoming MBs are
// not misclassified as freshly produced.
func getAllCrossShardMBsForShardInMeta(metaBlock data.HeaderHandler, dstShardID uint32) map[string]struct{} {
	result := make(map[string]struct{})

	metaHdr, ok := metaBlock.(data.MetaHeaderHandler)
	if !ok {
		return result
	}

	for _, shardData := range metaHdr.GetShardInfoHandlers() {
		if shardData.GetShardID() == dstShardID {
			continue
		}
		for _, mbh := range shardData.GetShardMiniBlockHeaderHandlers() {
			if mbh.GetReceiverShardID() != dstShardID {
				continue
			}
			if !shouldConsiderCrossShardMiniBlockForRepair(mbh.GetSenderShardID(), mbh.GetReceiverShardID()) {
				continue
			}
			result[string(mbh.GetHash())] = struct{}{}
		}
	}

	for _, mbh := range metaHdr.GetMiniBlockHeaderHandlers() {
		if mbh.GetReceiverShardID() != dstShardID {
			continue
		}
		if !shouldConsiderCrossShardMiniBlockForRepair(mbh.GetSenderShardID(), mbh.GetReceiverShardID()) {
			continue
		}
		result[string(mbh.GetHash())] = struct{}{}
	}

	return result
}

// collectForwardPendingMBsByShard walks meta blocks backward from lastMeta via
// PrevHash, stopping when the nonce drops at or below the oldest anchor. For
// each visited meta, it adds cross-shard MBs to each receiver shard's bucket
// when the meta's nonce is strictly greater than that shard's anchor nonce.
// ShardInfo entries where the source shard equals the receiver are skipped so
// a shard's own block notarization (consumption side) is not mistaken for a
// fresh production.
func collectForwardPendingMBsByShard(
	lastMeta data.MetaHeaderHandler,
	anchorByShard map[uint32]data.HeaderHandler,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) (map[uint32][][]byte, error) {
	out := make(map[uint32][][]byte)
	if len(anchorByShard) == 0 {
		return out, nil
	}

	anchorNonceByShard := make(map[uint32]uint64, len(anchorByShard))
	oldestAnchorNonce := ^uint64(0)
	for shardID, anchor := range anchorByShard {
		nonce := anchor.GetNonce()
		anchorNonceByShard[shardID] = nonce
		if nonce < oldestAnchorNonce {
			oldestAnchorNonce = nonce
		}
	}

	setByShard := make(map[uint32]map[string]struct{}, len(anchorByShard))

	currMeta := lastMeta
	for i := 0; i < maxForwardPendingWalkIter; i++ {
		if currMeta.GetNonce() <= oldestAnchorNonce {
			break
		}
		appendForwardPendingFromMeta(currMeta, anchorNonceByShard, setByShard)

		prevHash := currMeta.GetPrevHash()
		if len(prevHash) == 0 {
			break
		}
		prevMeta, err := GetMetaHeaderFromStorage(prevHash, marshalizer, store)
		if err != nil {
			return nil, fmt.Errorf("%w: could not load meta %s during forward-pending walk: %s",
				ErrMissingHeader, hex.EncodeToString(prevHash), err.Error())
		}
		currMeta = prevMeta
	}

	for shardID, set := range setByShard {
		for hash := range set {
			out[shardID] = append(out[shardID], []byte(hash))
		}
	}
	return out, nil
}

func appendForwardPendingFromMeta(
	meta data.MetaHeaderHandler,
	anchorNonceByShard map[uint32]uint64,
	setByShard map[uint32]map[string]struct{},
) {
	metaNonce := meta.GetNonce()

	for _, shardData := range meta.GetShardInfoHandlers() {
		sourceShard := shardData.GetShardID()
		for _, mbh := range shardData.GetShardMiniBlockHeaderHandlers() {
			recv := mbh.GetReceiverShardID()
			if sourceShard == recv {
				continue
			}
			anchorNonce, ok := anchorNonceByShard[recv]
			if !ok {
				continue
			}
			if metaNonce <= anchorNonce {
				continue
			}
			if !shouldConsiderCrossShardMiniBlockForRepair(mbh.GetSenderShardID(), recv) {
				continue
			}
			addToSet(setByShard, recv, string(mbh.GetHash()))
		}
	}

	for _, mbh := range meta.GetMiniBlockHeaderHandlers() {
		recv := mbh.GetReceiverShardID()
		anchorNonce, ok := anchorNonceByShard[recv]
		if !ok {
			continue
		}
		if metaNonce <= anchorNonce {
			continue
		}
		if !shouldConsiderCrossShardMiniBlockForRepair(mbh.GetSenderShardID(), recv) {
			continue
		}
		addToSet(setByShard, recv, string(mbh.GetHash()))
	}
}

func addToSet(m map[uint32]map[string]struct{}, key uint32, value string) {
	if m[key] == nil {
		m[key] = make(map[string]struct{})
	}
	m[key][value] = struct{}{}
}
