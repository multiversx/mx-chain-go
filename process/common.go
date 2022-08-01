package process

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("process")

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

	hdr, err := CreateShardHeader(marshalizer, buffHdr)
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

	hdrStore := storageService.GetStorer(blockUnit)
	if check.IfNil(hdrStore) {
		return nil, ErrNilHeadersStorage
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

	hash, err := getHeaderHashFromStorageWithNonce(
		nonce,
		storageService,
		uint64Converter,
		marshalizer,
		dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardId))
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

	hash, err := getHeaderHashFromStorageWithNonce(
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
	searchFirst bool,
) (data.TransactionHandler, error) {

	err := checkGetTransactionParamsForNil(shardedDataCacherNotifier, storageService, marshalizer)
	if err != nil {
		return nil, err
	}

	tx, err := GetTransactionHandlerFromPool(senderShardID, destShardID, txHash, shardedDataCacherNotifier, searchFirst)
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
	searchFirst bool,
) (data.TransactionHandler, error) {

	if shardedDataCacherNotifier == nil {
		return nil, ErrNilShardedDataCacherNotifier
	}

	var val interface{}
	ok := false
	if searchFirst {
		val, ok = shardedDataCacherNotifier.SearchFirstData(txHash)
		if !ok {
			return nil, ErrTxNotFound
		}
	} else {
		strCache := ShardCacherIdentifier(senderShardID, destShardID)
		txStore := shardedDataCacherNotifier.ShardDataStore(strCache)
		if txStore == nil {
			return nil, ErrNilStorage
		}

		val, ok = txStore.Peek(txHash)
	}

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

	//TODO what should we do when we get from pool more than one header with same nonce and shardId
	return headers[len(headers)-1], hashes[len(hashes)-1], nil
}

func getHeaderHashFromStorageWithNonce(
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

	headerStore := storageService.GetStorer(blockUnit)
	if headerStore == nil {
		return nil, ErrNilHeadersStorage
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

	receiver := ""
	if len(txHandler.GetRcvAddr()) == addressPubkeyConverter.Len() {
		receiver = addressPubkeyConverter.Encode(txHandler.GetRcvAddr())
	}

	sender := ""
	if len(txHandler.GetSndAddr()) == addressPubkeyConverter.Len() {
		sender = addressPubkeyConverter.Encode(txHandler.GetSndAddr())
	}

	log.Trace("executing transaction",
		"txHash", txHash,
		"nonce", txHandler.GetNonce(),
		"value", txHandler.GetValue(),
		"gas limit", txHandler.GetGasLimit(),
		"gas price", txHandler.GetGasPrice(),
		"data", hex.EncodeToString(txHandler.GetData()),
		"sender", sender,
		"receiver", receiver)
}

// IsAllowedToSaveUnderKey returns if saving key-value in data tries under given key is allowed
func IsAllowedToSaveUnderKey(key []byte) bool {
	prefixLen := len(core.ElrondProtectedKeyPrefix)
	if len(key) < prefixLen {
		return true
	}

	trimmedKey := key[:prefixLen]
	return !bytes.Equal(trimmedKey, []byte(core.ElrondProtectedKeyPrefix))
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

// CreateShardHeader creates a shard header from the given byte array
func CreateShardHeader(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
	hdr, err := CreateHeaderV2(marshalizer, hdrBuff)
	if err == nil {
		return hdr, nil
	}

	hdr, err = CreateHeaderV1(marshalizer, hdrBuff)
	return hdr, err
}

// CreateHeaderV2 creates a header with version 2 from the given byte array
func CreateHeaderV2(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
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

// CreateHeaderV1 creates a header with version 1 from the given byte array
func CreateHeaderV1(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.ShardHeaderHandler, error) {
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
