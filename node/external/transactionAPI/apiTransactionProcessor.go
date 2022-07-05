package transactionAPI

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	rewardTxData "github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

var log = logger.GetOrCreate("node/transactionAPI")

type apiTransactionProcessor struct {
	roundDuration               uint64
	genesisTime                 time.Time
	marshalizer                 marshal.Marshalizer
	addressPubKeyConverter      core.PubkeyConverter
	shardCoordinator            sharding.Coordinator
	historyRepository           dblookupext.HistoryRepository
	storageService              dataRetriever.StorageService
	dataPool                    dataRetriever.PoolsHolder
	uint64ByteSliceConverter    typeConverters.Uint64ByteSliceConverter
	txUnmarshaller              *txUnmarshaller
	transactionResultsProcessor *apiTransactionResultsProcessor
}

// NewAPITransactionProcessor will create a new instance of apiTransactionProcessor
func NewAPITransactionProcessor(args *ArgAPITransactionProcessor) (*apiTransactionProcessor, error) {
	err := checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	txUnmarshalerAndPreparer := newTransactionUnmarshaller(args.Marshalizer, args.AddressPubKeyConverter, args.DataFieldParser)
	txResultsProc := newAPITransactionResultProcessor(
		args.AddressPubKeyConverter,
		args.HistoryRepository,
		args.StorageService,
		args.Marshalizer,
		txUnmarshalerAndPreparer,
		args.ShardCoordinator.SelfId(),
		args.DataFieldParser,
	)

	return &apiTransactionProcessor{
		roundDuration:               args.RoundDuration,
		genesisTime:                 args.GenesisTime,
		marshalizer:                 args.Marshalizer,
		addressPubKeyConverter:      args.AddressPubKeyConverter,
		shardCoordinator:            args.ShardCoordinator,
		historyRepository:           args.HistoryRepository,
		storageService:              args.StorageService,
		dataPool:                    args.DataPool,
		uint64ByteSliceConverter:    args.Uint64ByteSliceConverter,
		txUnmarshaller:              txUnmarshalerAndPreparer,
		transactionResultsProcessor: txResultsProc,
	}, nil
}

// GetTransaction gets the transaction based on the given hash. It will search in the cache and the storage and
// will return the transaction in a format which can be respected by all types of transactions (normal, reward or unsigned)
func (atp *apiTransactionProcessor) GetTransaction(txHash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	tx, err := atp.optionallyGetTransactionFromPool(hash)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		return tx, nil
	}

	if atp.historyRepository.IsEnabled() {
		return atp.lookupHistoricalTransaction(hash, withResults)
	}

	return atp.getTransactionFromStorage(hash)
}

// GetTransactionsPool will return a structure containing the transactions pool fields that is to be returned on API calls
func (atp *apiTransactionProcessor) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	transactions := &common.TransactionsPoolAPIResponse{}
	requestedFieldsHandler := newFieldsHandler(fields)

	var err error
	transactions.RegularTransactions, err = atp.getRegularTransactionsFromPool(requestedFieldsHandler)
	if err != nil {
		return nil, err
	}

	transactions.Rewards, err = atp.getRewardTransactionsFromPool(requestedFieldsHandler)
	if err != nil {
		return nil, err
	}

	transactions.SmartContractResults, err = atp.getUnsignedTransactionsFromPool(requestedFieldsHandler)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsPoolForSender will return a structure containing the transactions for sender that is to be returned on API calls
func (atp *apiTransactionProcessor) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	senderAddr, err := atp.addressPubKeyConverter.Decode(sender)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", ErrInvalidAddress.Error(), err)
	}

	senderShard := atp.shardCoordinator.ComputeId(senderAddr)

	fields = strings.Trim(fields, " ")
	fields = strings.ToLower(fields)
	wrappedTxs := atp.fetchTxsForSender(string(senderAddr), senderShard)
	if len(wrappedTxs) == 0 {
		return nil, fmt.Errorf("%w, no transaction in pool for sender", ErrCannotRetrieveTransactions)
	}

	requestedFieldsHandler := newFieldsHandler(fields)
	transactions := &common.TransactionsPoolForSenderApiResponse{}
	for _, wrappedTx := range wrappedTxs {
		tx := atp.extractRequestedTxInfo(wrappedTx, requestedFieldsHandler)
		transactions.Transactions = append(transactions.Transactions, tx)
	}

	return transactions, nil
}

// GetLastPoolNonceForSender will return the last nonce from pool for sender that is to be returned on API calls
func (atp *apiTransactionProcessor) GetLastPoolNonceForSender(sender string) (uint64, error) {
	senderAddr, err := atp.addressPubKeyConverter.Decode(sender)
	if err != nil {
		return 0, fmt.Errorf("%s, %w", ErrInvalidAddress.Error(), err)
	}

	senderShard := atp.shardCoordinator.ComputeId(senderAddr)
	lastNonce, err := atp.fetchLastNonceForSender(string(senderAddr), senderShard)
	if err != nil {
		return 0, err
	}

	return lastNonce, nil
}

// GetTransactionsPoolNonceGapsForSender will return the nonce gaps from pool for sender, if exists, that is to be returned on API calls
func (atp *apiTransactionProcessor) GetTransactionsPoolNonceGapsForSender(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	senderAddr, err := atp.addressPubKeyConverter.Decode(sender)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", ErrInvalidAddress.Error(), err)
	}

	senderShard := atp.shardCoordinator.ComputeId(senderAddr)
	nonceGaps, err := atp.extractNonceGaps(string(senderAddr), senderShard)
	if err != nil {
		return nil, err
	}

	return &common.TransactionsPoolNonceGapsForSenderApiResponse{
		Sender: sender,
		Gaps:   nonceGaps,
	}, nil
}

func (atp *apiTransactionProcessor) extractRequestedTxInfoFromObj(txObj interface{}, txType transaction.TxType, txHash []byte, requestedFieldsHandler fieldsHandler) (common.Transaction, error) {
	txResult, err := atp.getApiResultFromObj(txObj, txType)
	if err != nil {
		return common.Transaction{}, err
	}

	wrappedTx := &txcache.WrappedTransaction{
		Tx:     txResult.Tx,
		TxHash: txHash,
	}

	return atp.extractRequestedTxInfo(wrappedTx, requestedFieldsHandler), nil
}

func (atp *apiTransactionProcessor) getRegularTransactionsFromPool(requestedFieldsHandler fieldsHandler) ([]common.Transaction, error) {
	regularTxKeys := atp.dataPool.Transactions().Keys()
	regularTxs := make([]common.Transaction, len(regularTxKeys))
	for idx, key := range regularTxKeys {
		txObj, found := atp.getRegularTxObjFromDataPool(key)
		if !found {
			continue
		}

		tx, err := atp.extractRequestedTxInfoFromObj(txObj, transaction.TxTypeNormal, key, requestedFieldsHandler)
		if err != nil {
			return nil, err
		}

		regularTxs[idx] = tx
	}

	return regularTxs, nil
}

func (atp *apiTransactionProcessor) getRewardTransactionsFromPool(requestedFieldsHandler fieldsHandler) ([]common.Transaction, error) {
	rewardTxKeys := atp.dataPool.RewardTransactions().Keys()
	rewardTxs := make([]common.Transaction, len(rewardTxKeys))
	for idx, key := range rewardTxKeys {
		txObj, found := atp.getRewardTxObjFromDataPool(key)
		if !found {
			continue
		}

		tx, err := atp.extractRequestedTxInfoFromObj(txObj, transaction.TxTypeReward, key, requestedFieldsHandler)
		if err != nil {
			return nil, err
		}

		rewardTxs[idx] = tx
	}

	return rewardTxs, nil
}

func (atp *apiTransactionProcessor) getUnsignedTransactionsFromPool(requestedFieldsHandler fieldsHandler) ([]common.Transaction, error) {
	unsignedTxKeys := atp.dataPool.UnsignedTransactions().Keys()
	unsignedTxs := make([]common.Transaction, len(unsignedTxKeys))
	for idx, key := range unsignedTxKeys {
		txObj, found := atp.getUnsignedTxObjFromDataPool(key)
		if !found {
			continue
		}

		tx, err := atp.extractRequestedTxInfoFromObj(txObj, transaction.TxTypeUnsigned, key, requestedFieldsHandler)
		if err != nil {
			return nil, err
		}

		unsignedTxs[idx] = tx
	}

	return unsignedTxs, nil
}

func (atp *apiTransactionProcessor) extractRequestedTxInfo(wrappedTx *txcache.WrappedTransaction, requestedFieldsHandler fieldsHandler) common.Transaction {
	tx := common.Transaction{
		TxFields: make(map[string]interface{}),
	}

	tx.TxFields[hashField] = hex.EncodeToString(wrappedTx.TxHash)

	if requestedFieldsHandler.HasNonce {
		tx.TxFields[nonceField] = wrappedTx.Tx.GetNonce()
	}
	if requestedFieldsHandler.HasSender {
		tx.TxFields[senderField] = atp.addressPubKeyConverter.Encode(wrappedTx.Tx.GetSndAddr())
	}
	if requestedFieldsHandler.HasReceiver {
		tx.TxFields[receiverField] = atp.addressPubKeyConverter.Encode(wrappedTx.Tx.GetRcvAddr())
	}
	if requestedFieldsHandler.HasGasLimit {
		tx.TxFields[gasLimitField] = wrappedTx.Tx.GetGasLimit()
	}
	if requestedFieldsHandler.HasGasPrice {
		tx.TxFields[gasPriceField] = wrappedTx.Tx.GetGasPrice()
	}
	if requestedFieldsHandler.HasRcvUsername {
		tx.TxFields[rcvUsernameField] = wrappedTx.Tx.GetRcvUserName()
	}
	if requestedFieldsHandler.HasData {
		tx.TxFields[dataField] = wrappedTx.Tx.GetData()
	}
	if requestedFieldsHandler.HasValue {
		tx.TxFields[valueField] = wrappedTx.Tx.GetValue()
	}

	return tx
}

func (atp *apiTransactionProcessor) getDataStoresForSender(senderShard uint32) []storage.Cacher {
	cachers := make([]storage.Cacher, 0)
	numOfShards := atp.shardCoordinator.NumberOfShards()
	for shard := uint32(0); shard < numOfShards; shard++ {
		cacheId := process.ShardCacherIdentifier(senderShard, shard)
		shardCache := atp.dataPool.Transactions().ShardDataStore(cacheId)
		cachers = append(cachers, shardCache)
	}

	cacheId := process.ShardCacherIdentifier(senderShard, common.MetachainShardId)
	shardCache := atp.dataPool.Transactions().ShardDataStore(cacheId)
	cachers = append(cachers, shardCache)

	return cachers
}

func (atp *apiTransactionProcessor) fetchTxsForSender(sender string, senderShard uint32) []*txcache.WrappedTransaction {
	txsForSender := make([]*txcache.WrappedTransaction, 0)
	cachers := atp.getDataStoresForSender(senderShard)
	for _, cache := range cachers {
		txCache, ok := cache.(*txcache.TxCache)
		if !ok {
			continue
		}

		txs := txCache.GetTransactionsPoolForSender(sender)
		txsForSender = append(txsForSender, txs...)
	}

	sort.Slice(txsForSender, func(i, j int) bool {
		return txsForSender[i].Tx.GetNonce() < txsForSender[j].Tx.GetNonce()
	})

	return txsForSender
}

func (atp *apiTransactionProcessor) fetchLastNonceForSender(sender string, senderShard uint32) (uint64, error) {
	wrappedTxs := atp.fetchTxsForSender(sender, senderShard)
	if len(wrappedTxs) == 0 {
		return 0, fmt.Errorf("%w, no transaction in pool for sender", ErrCannotRetrieveTransactions)
	}

	lastTx := wrappedTxs[len(wrappedTxs)-1]
	return lastTx.Tx.GetNonce(), nil
}

func (atp *apiTransactionProcessor) extractNonceGaps(sender string, senderShard uint32) ([]common.NonceGapApiResponse, error) {
	wrappedTxs := atp.fetchTxsForSender(sender, senderShard)
	if len(wrappedTxs) == 0 {
		return nil, fmt.Errorf("%w, no transaction in pool for sender", ErrCannotRetrieveTransactions)
	}

	nonceGaps := make([]common.NonceGapApiResponse, 0)
	for i := 0; i < len(wrappedTxs)-1; i++ {
		nextNonce := wrappedTxs[i+1].Tx.GetNonce()
		currentNonce := wrappedTxs[i].Tx.GetNonce()
		nonceDiff := nextNonce - currentNonce
		if nonceDiff > 1 {
			nonceGap := common.NonceGapApiResponse{
				From: currentNonce + 1,
				To:   nextNonce - 1,
			}
			nonceGaps = append(nonceGaps, nonceGap)
		}
	}

	return nonceGaps, nil
}

func (atp *apiTransactionProcessor) optionallyGetTransactionFromPool(hash []byte) (*transaction.ApiTransactionResult, error) {
	txObj, txType, found := atp.getTxObjFromDataPool(hash)
	if !found {
		return nil, nil
	}

	return atp.getApiResultFromObj(txObj, txType)
}

func (atp *apiTransactionProcessor) getApiResultFromObj(txObj interface{}, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	tx, err := atp.castObjToTransaction(txObj, txType)
	if err != nil {
		return nil, err
	}

	tx.SourceShard = atp.shardCoordinator.ComputeId(tx.Tx.GetSndAddr())
	tx.DestinationShard = atp.shardCoordinator.ComputeId(tx.Tx.GetRcvAddr())
	tx.Status = transaction.TxStatusPending

	return tx, nil
}

// computeTimestampForRound will return the timestamp for the given round
func (atp *apiTransactionProcessor) computeTimestampForRound(round uint64) int64 {
	if round == 0 {
		return 0
	}

	secondsSinceGenesis := round * atp.roundDuration
	timestamp := atp.genesisTime.Add(time.Duration(secondsSinceGenesis) * time.Millisecond)

	return timestamp.Unix()
}

func (atp *apiTransactionProcessor) lookupHistoricalTransaction(hash []byte, withResults bool) (*transaction.ApiTransactionResult, error) {
	miniblockMetadata, err := atp.historyRepository.GetMiniblockMetadataByTxHash(hash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrTransactionNotFound.Error(), err)
	}

	txBytes, txType, found := atp.getTxBytesFromStorageByEpoch(hash, miniblockMetadata.Epoch)
	if !found {
		log.Warn("lookupHistoricalTransaction(): unexpected condition, cannot find transaction in storage")
		return nil, fmt.Errorf("%s: %w", ErrCannotRetrieveTransaction.Error(), err)
	}

	// After looking up a transaction from storage, it's impossible to say whether it was successful or invalid
	// (since both successful and invalid transactions are kept in the same storage unit),
	// so we have to use our extra information from the "miniblockMetadata" to correct the txType if appropriate
	if block.Type(miniblockMetadata.Type) == block.InvalidBlock {
		txType = transaction.TxTypeInvalid
	}

	tx, err := atp.txUnmarshaller.unmarshalTransaction(txBytes, txType)
	if err != nil {
		log.Warn("lookupHistoricalTransaction(): unexpected condition, cannot unmarshal transaction")
		return nil, fmt.Errorf("%s: %w", ErrCannotRetrieveTransaction.Error(), err)
	}

	putMiniblockFieldsInTransaction(tx, miniblockMetadata)
	tx.Timestamp = atp.computeTimestampForRound(tx.Round)
	statusComputer, err := txstatus.NewStatusComputer(atp.shardCoordinator.SelfId(), atp.uint64ByteSliceConverter, atp.storageService)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrNilStatusComputer.Error(), err)
	}

	if ok, _ := statusComputer.SetStatusIfIsRewardReverted(
		tx,
		block.Type(miniblockMetadata.Type),
		miniblockMetadata.HeaderNonce,
		miniblockMetadata.HeaderHash); ok {
		return tx, nil
	}

	tx.Status, _ = statusComputer.ComputeStatusWhenInStorageKnowingMiniblock(
		block.Type(miniblockMetadata.Type), tx)

	if withResults {
		atp.transactionResultsProcessor.putResultsInTransaction(hash, tx, miniblockMetadata.Epoch)
	}

	return tx, nil
}

func putMiniblockFieldsInTransaction(tx *transaction.ApiTransactionResult, miniblockMetadata *dblookupext.MiniblockMetadata) *transaction.ApiTransactionResult {
	tx.Epoch = miniblockMetadata.Epoch
	tx.Round = miniblockMetadata.Round

	tx.MiniBlockType = block.Type(miniblockMetadata.Type).String()
	tx.MiniBlockHash = hex.EncodeToString(miniblockMetadata.MiniblockHash)
	tx.DestinationShard = miniblockMetadata.DestinationShardID
	tx.SourceShard = miniblockMetadata.SourceShardID

	tx.BlockNonce = miniblockMetadata.HeaderNonce
	tx.BlockHash = hex.EncodeToString(miniblockMetadata.HeaderHash)
	tx.NotarizedAtSourceInMetaNonce = miniblockMetadata.NotarizedAtSourceInMetaNonce
	tx.NotarizedAtSourceInMetaHash = hex.EncodeToString(miniblockMetadata.NotarizedAtSourceInMetaHash)
	tx.NotarizedAtDestinationInMetaNonce = miniblockMetadata.NotarizedAtDestinationInMetaNonce
	tx.NotarizedAtDestinationInMetaHash = hex.EncodeToString(miniblockMetadata.NotarizedAtDestinationInMetaHash)

	return tx
}

func (atp *apiTransactionProcessor) getTransactionFromStorage(hash []byte) (*transaction.ApiTransactionResult, error) {
	txBytes, txType, found := atp.getTxBytesFromStorage(hash)
	if !found {
		return nil, ErrTransactionNotFound
	}

	tx, err := atp.txUnmarshaller.unmarshalTransaction(txBytes, txType)
	if err != nil {
		return nil, err
	}

	tx.Timestamp = atp.computeTimestampForRound(tx.Round)

	// TODO: take care of this when integrating the adaptivity
	statusComputer, err := txstatus.NewStatusComputer(atp.shardCoordinator.SelfId(), atp.uint64ByteSliceConverter, atp.storageService)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrNilStatusComputer.Error(), err)
	}
	tx.Status, _ = statusComputer.ComputeStatusWhenInStorageNotKnowingMiniblock(
		atp.shardCoordinator.ComputeId(tx.Tx.GetRcvAddr()), tx)

	return tx, nil
}

func (atp *apiTransactionProcessor) getTxObjFromDataPool(hash []byte) (interface{}, transaction.TxType, bool) {
	txObj, found := atp.getRegularTxObjFromDataPool(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeNormal, true
	}

	txObj, found = atp.getRewardTxObjFromDataPool(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeReward, true
	}

	txObj, found = atp.getUnsignedTxObjFromDataPool(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (atp *apiTransactionProcessor) getRegularTxObjFromDataPool(hash []byte) (interface{}, bool) {
	txsPool := atp.dataPool.Transactions()
	txObj, found := txsPool.SearchFirstData(hash)
	return txObj, found
}

func (atp *apiTransactionProcessor) getRewardTxObjFromDataPool(hash []byte) (interface{}, bool) {
	rewardTxsPool := atp.dataPool.RewardTransactions()
	txObj, found := rewardTxsPool.SearchFirstData(hash)
	return txObj, found
}

func (atp *apiTransactionProcessor) getUnsignedTxObjFromDataPool(hash []byte) (interface{}, bool) {
	unsignedTxsPool := atp.dataPool.UnsignedTransactions()
	txObj, found := unsignedTxsPool.SearchFirstData(hash)
	return txObj, found
}

func (atp *apiTransactionProcessor) getTxBytesFromStorage(hash []byte) ([]byte, transaction.TxType, bool) {
	store := atp.storageService
	txsStorer := store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeNormal, true
	}

	rewardTxsStorer := store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeReward, true
	}

	unsignedTxsStorer := store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (atp *apiTransactionProcessor) getTxBytesFromStorageByEpoch(hash []byte, epoch uint32) ([]byte, transaction.TxType, bool) {
	store := atp.storageService
	txsStorer := store.GetStorer(dataRetriever.TransactionUnit)
	txBytes, err := txsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeNormal, true
	}

	rewardTxsStorer := store.GetStorer(dataRetriever.RewardTransactionUnit)
	txBytes, err = rewardTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeReward, true
	}

	unsignedTxsStorer := store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	txBytes, err = unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (atp *apiTransactionProcessor) castObjToTransaction(txObj interface{}, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case transaction.TxTypeNormal:
		if tx, ok := txObj.(*transaction.Transaction); ok {
			return atp.txUnmarshaller.prepareNormalTx(tx)
		}
	case transaction.TxTypeInvalid:
		if tx, ok := txObj.(*transaction.Transaction); ok {
			return atp.txUnmarshaller.prepareInvalidTx(tx)
		}
	case transaction.TxTypeReward:
		if tx, ok := txObj.(*rewardTxData.RewardTx); ok {
			return atp.txUnmarshaller.prepareRewardTx(tx)
		}
	case transaction.TxTypeUnsigned:
		if tx, ok := txObj.(*smartContractResult.SmartContractResult); ok {
			return atp.txUnmarshaller.prepareUnsignedTx(tx)
		}
	}

	log.Warn("castObjToTransaction() unexpected: unknown txType", "txType", txType)
	return &transaction.ApiTransactionResult{Type: string(transaction.TxTypeInvalid)}, nil
}

// UnmarshalTransaction will try to unmarshal the transaction bytes based on the transaction type
func (atp *apiTransactionProcessor) UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	return atp.txUnmarshaller.unmarshalTransaction(txBytes, txType)
}

// UnmarshalReceipt will try to unmarshal the provided receipts bytes
func (atp *apiTransactionProcessor) UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error) {
	return atp.txUnmarshaller.unmarshalReceipt(receiptBytes)
}

// IsInterfaceNil returns true if underlying object is nil
func (atp *apiTransactionProcessor) IsInterfaceNil() bool {
	return atp == nil
}
