package process

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// EmptyChannel empties the given channel
func EmptyChannel(ch chan bool) int {
	readsCnt := 0
	for {
		select {
		case <-ch:
			readsCnt++
		default:
			return readsCnt
		}
	}
}

// GetShardHeader gets the header, which is associated with the given hash, from pool or storage
func GetShardHeader(
	hash []byte,
	cacher storage.Cacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.Header, error) {

	err := checkGetHeaderParamsForNil(cacher, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr, err := GetShardHeaderFromPool(hash, cacher)
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
	cacher storage.Cacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
) (*block.MetaBlock, error) {

	err := checkGetHeaderParamsForNil(cacher, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr, err := GetMetaHeaderFromPool(hash, cacher)
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
	cacher storage.Cacher,
) (*block.Header, error) {

	obj, err := getHeaderFromPool(hash, cacher)
	if err != nil {
		return nil, err
	}

	hdr, ok := obj.(*block.Header)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	return hdr, nil
}

// GetMetaHeaderFromPool gets the header, which is associated with the given hash, from pool
func GetMetaHeaderFromPool(
	hash []byte,
	cacher storage.Cacher,
) (*block.MetaBlock, error) {

	obj, err := getHeaderFromPool(hash, cacher)
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
) (*block.Header, error) {

	buffHdr, err := GetMarshalizedHeaderFromStorage(dataRetriever.BlockHeaderUnit, hash, marshalizer, storageService)
	if err != nil {
		return nil, err
	}

	hdr := &block.Header{}
	err = marshalizer.Unmarshal(hdr, buffHdr)
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

	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if storageService == nil || storageService.IsInterfaceNil() {
		return nil, ErrNilStorage
	}

	hdrStore := storageService.GetStorer(blockUnit)
	if hdrStore == nil || hdrStore.IsInterfaceNil() {
		return nil, ErrNilHeadersStorage
	}

	buffHdr, err := hdrStore.Get(hash)
	if err != nil {
		return nil, ErrMissingHeader
	}

	return buffHdr, nil
}

// GetShardHeaderWithNonce method returns a shard block header with a given nonce and shardId
func GetShardHeaderWithNonce(
	nonce uint64,
	shardId uint32,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) (*block.Header, []byte, error) {

	err := checkGetHeaderWithNonceParamsForNil(cacher, uint64SyncMapCacher, marshalizer, storageService, uint64Converter)
	if err != nil {
		return nil, nil, err
	}

	hdr, hash, err := GetShardHeaderFromPoolWithNonce(nonce, shardId, cacher, uint64SyncMapCacher)
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
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) (*block.MetaBlock, []byte, error) {

	err := checkGetHeaderWithNonceParamsForNil(cacher, uint64SyncMapCacher, marshalizer, storageService, uint64Converter)
	if err != nil {
		return nil, nil, err
	}

	hdr, hash, err := GetMetaHeaderFromPoolWithNonce(nonce, cacher, uint64SyncMapCacher)
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
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) (*block.Header, []byte, error) {

	obj, hash, err := getHeaderFromPoolWithNonce(nonce, shardId, cacher, uint64SyncMapCacher)
	if err != nil {
		return nil, nil, err
	}

	hdr, ok := obj.(*block.Header)
	if !ok {
		return nil, nil, ErrWrongTypeAssertion
	}

	return hdr, hash, nil
}

// GetMetaHeaderFromPoolWithNonce method returns a meta block header from pool with a given nonce
func GetMetaHeaderFromPoolWithNonce(
	nonce uint64,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) (*block.MetaBlock, []byte, error) {

	obj, hash, err := getHeaderFromPoolWithNonce(nonce, sharding.MetachainShardId, cacher, uint64SyncMapCacher)
	if err != nil {
		return nil, nil, err
	}

	hdr, ok := obj.(*block.MetaBlock)
	if !ok {
		return nil, nil, ErrWrongTypeAssertion
	}

	return hdr, hash, nil
}

// GetShardHeaderFromStorageWithNonce method returns a shard block header from storage with a given nonce and shardId
func GetShardHeaderFromStorageWithNonce(
	nonce uint64,
	shardId uint32,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
) (*block.Header, []byte, error) {

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
) (data.TransactionHandler, error) {

	err := checkGetTransactionParamsForNil(shardedDataCacherNotifier, storageService, marshalizer)
	if err != nil {
		return nil, err
	}

	tx, err := GetTransactionHandlerFromPool(senderShardID, destShardID, txHash, shardedDataCacherNotifier)
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
) (data.TransactionHandler, error) {

	if shardedDataCacherNotifier == nil {
		return nil, ErrNilShardedDataCacherNotifier
	}

	strCache := ShardCacherIdentifier(senderShardID, destShardID)
	txStore := shardedDataCacherNotifier.ShardDataStore(strCache)
	if txStore == nil {
		return nil, ErrNilStorage
	}

	val, ok := txStore.Peek(txHash)
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
	cacher storage.Cacher,
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
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	marshalizer marshal.Marshalizer,
	storageService dataRetriever.StorageService,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) error {

	err := checkGetHeaderParamsForNil(cacher, marshalizer, storageService)
	if err != nil {
		return err
	}
	if uint64SyncMapCacher == nil || uint64SyncMapCacher.IsInterfaceNil() {
		return ErrNilUint64SyncMapCacher
	}
	if uint64Converter == nil || uint64Converter.IsInterfaceNil() {
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
	cacher storage.Cacher,
) (interface{}, error) {

	if cacher == nil || cacher.IsInterfaceNil() {
		return nil, ErrNilCacher
	}

	obj, ok := cacher.Peek(hash)
	if !ok {
		return nil, ErrMissingHeader
	}

	return obj, nil
}

func getHeaderFromPoolWithNonce(
	nonce uint64,
	shardId uint32,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) (interface{}, []byte, error) {

	if cacher == nil || cacher.IsInterfaceNil() {
		return nil, nil, ErrNilCacher
	}
	if uint64SyncMapCacher == nil || uint64SyncMapCacher.IsInterfaceNil() {
		return nil, nil, ErrNilUint64SyncMapCacher
	}

	syncMap, ok := uint64SyncMapCacher.Get(nonce)
	if !ok {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	hash, ok := syncMap.Load(shardId)
	if hash == nil || !ok {
		return nil, nil, ErrMissingHashForHeaderNonce
	}

	obj, ok := cacher.Peek(hash)
	if !ok {
		return nil, nil, ErrMissingHeader
	}

	return obj, hash, nil
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

func ComputeGasConsumedByMiniBlockInShard(
	shardId uint32,
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
	economicsFee FeeHandler,
) (uint64, error) {

	gasConsumedByMiniBlockInSenderShard, gasConsumedByMiniBlockInReceiverShard, err := ComputeGasConsumedByMiniBlock(
		miniBlock,
		mapHashTx,
		economicsFee)

	if err != nil {
		return 0, err
	}

	gasConsumedByMiniBlock, err := ComputeGasConsumedInShard(
		shardId,
		miniBlock.SenderShardID,
		miniBlock.ReceiverShardID,
		gasConsumedByMiniBlockInSenderShard,
		gasConsumedByMiniBlockInReceiverShard)

	return gasConsumedByMiniBlock, err
}

func ComputeGasConsumedByMiniBlock(
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
	economicsFee FeeHandler,
) (uint64, uint64, error) {

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := mapHashTx[string(txHash)]
		if !ok {
			return 0, 0, ErrMissingTransaction
		}

		gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := ComputeGasConsumedByTx(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHandler,
			economicsFee)
		if err != nil {
			return 0, 0, err
		}

		gasConsumedByMiniBlockInSenderShard += gasConsumedByTxInSenderShard
		gasConsumedByMiniBlockInReceiverShard += gasConsumedByTxInReceiverShard
	}

	return gasConsumedByMiniBlockInSenderShard, gasConsumedByMiniBlockInReceiverShard, nil
}

func ComputeGasConsumedByTxInShard(
	shardId uint32,
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
	economicsFee FeeHandler,
) (uint64, error) {

	gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := ComputeGasConsumedByTx(
		txSenderShardId,
		txReceiverShardId,
		txHandler,
		economicsFee)
	if err != nil {
		return 0, err
	}

	gasConsumedByTx, err := ComputeGasConsumedInShard(
		shardId,
		txSenderShardId,
		txReceiverShardId,
		gasConsumedByTxInSenderShard,
		gasConsumedByTxInReceiverShard)

	return gasConsumedByTx, nil
}

func ComputeGasConsumedByTx(
	txSenderShardId uint32,
	txReceiverShardId uint32,
	txHandler data.TransactionHandler,
	economicsFee FeeHandler,
) (uint64, uint64, error) {

	tx, ok := txHandler.(*transaction.Transaction)
	if !ok {
		return 0, 0, ErrWrongTypeAssertion
	}

	if core.IsSmartContractAddress(tx.RcvAddr) {
		gasConsumedByTxInSenderShard := economicsFee.ComputeGasLimit(tx)
		gasConsumedByTxInReceiverShard := tx.GasLimit

		if txSenderShardId != txReceiverShardId {
			return gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, nil
		}

		gasConsumedByTx := gasConsumedByTxInSenderShard + gasConsumedByTxInReceiverShard
		return gasConsumedByTx, gasConsumedByTx, nil
	}

	gasConsumedByTx := economicsFee.ComputeGasLimit(tx)
	return gasConsumedByTx, gasConsumedByTx, nil
}

func ComputeGasConsumedInShard(
	shardId uint32,
	senderShardId uint32,
	receiverShardId uint32,
	gasConsumedInSenderShard uint64,
	gasConsumedInReceiverShard uint64,
) (uint64, error) {

	if shardId == senderShardId {
		return gasConsumedInSenderShard, nil
	}
	if shardId == receiverShardId {
		return gasConsumedInReceiverShard, nil
	}

	return 0, ErrInvalidShardId
}

func IsMaxGasLimitReached(
	gasConsumedByTxInSenderShard uint64,
	gasConsumedByTxInReceiverShard uint64,
	gasConsumedByTxInSelfShard uint64,
	currentGasConsumedByMiniBlockInSenderShard uint64,
	currentGasConsumedByMiniBlockInReceiverShard uint64,
	currentGasConsumedByBlockInSelfShard uint64,
	maxGasLimitPerBlock uint64,
) bool {

	gasConsumedByMiniBlockInSenderShard := currentGasConsumedByMiniBlockInSenderShard + gasConsumedByTxInSenderShard
	gasConsumedByMiniBlockInReceiverShard := currentGasConsumedByMiniBlockInReceiverShard + gasConsumedByTxInReceiverShard
	gasConsumedByBlockInSelfShard := currentGasConsumedByBlockInSelfShard + gasConsumedByTxInSelfShard

	isGasLimitPerMiniBlockInSenderShardReached := gasConsumedByMiniBlockInSenderShard > maxGasLimitPerBlock
	isGasLimitPerMiniBlockInReceiverShardReached := gasConsumedByMiniBlockInReceiverShard > maxGasLimitPerBlock
	isGasLimitPerBlockInSelfShardReached := gasConsumedByBlockInSelfShard > maxGasLimitPerBlock

	isMaxGasLimitReached := isGasLimitPerMiniBlockInSenderShardReached ||
		isGasLimitPerMiniBlockInReceiverShardReached ||
		isGasLimitPerBlockInSelfShardReached

	return isMaxGasLimitReached
}
