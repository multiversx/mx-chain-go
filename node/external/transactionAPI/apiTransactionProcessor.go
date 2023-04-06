package transactionAPI

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	rewardTxData "github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/txstatus"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	logger "github.com/multiversx/mx-chain-logger-go"
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
	feeComputer                 feeComputer
	txTypeHandler               process.TxTypeHandler
	txUnmarshaller              *txUnmarshaller
	transactionResultsProcessor *apiTransactionResultsProcessor
	refundDetector              *refundDetector
	gasUsedAndFeeProcessor      *gasUsedAndFeeProcessor
}

// NewAPITransactionProcessor will create a new instance of apiTransactionProcessor
func NewAPITransactionProcessor(args *ArgAPITransactionProcessor) (*apiTransactionProcessor, error) {
	err := checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	txUnmarshalerAndPreparer := newTransactionUnmarshaller(args.Marshalizer, args.AddressPubKeyConverter, args.DataFieldParser, args.ShardCoordinator)
	txResultsProc := newAPITransactionResultProcessor(
		args.AddressPubKeyConverter,
		args.HistoryRepository,
		args.StorageService,
		args.Marshalizer,
		txUnmarshalerAndPreparer,
		args.LogsFacade,
		args.ShardCoordinator,
		args.DataFieldParser,
	)

	refundDetectorInstance := newRefundDetector()
	gasUsedAndFeeProc := newGasUsedAndFeeProcessor(args.FeeComputer, args.AddressPubKeyConverter)

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
		feeComputer:                 args.FeeComputer,
		txTypeHandler:               args.TxTypeHandler,
		txUnmarshaller:              txUnmarshalerAndPreparer,
		transactionResultsProcessor: txResultsProc,
		refundDetector:              refundDetectorInstance,
		gasUsedAndFeeProcessor:      gasUsedAndFeeProc,
	}, nil
}

// GetTransaction gets the transaction based on the given hash. It will search in the cache and the storage and
// will return the transaction in a format which can be respected by all types of transactions (normal, reward or unsigned)
func (atp *apiTransactionProcessor) GetTransaction(txHash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	tx, err := atp.doGetTransaction(hash, withResults)
	if err != nil {
		return nil, err
	}

	tx.Hash = txHash
	atp.PopulateComputedFields(tx)

	if withResults {
		atp.gasUsedAndFeeProcessor.computeAndAttachGasUsedAndFee(tx)
	}

	return tx, nil
}

func (atp *apiTransactionProcessor) doGetTransaction(hash []byte, withResults bool) (*transaction.ApiTransactionResult, error) {
	tx := atp.optionallyGetTransactionFromPool(hash)
	if tx != nil {
		return tx, nil
	}

	if atp.historyRepository.IsEnabled() {
		return atp.lookupHistoricalTransaction(hash, withResults)
	}

	return atp.getTransactionFromStorage(hash)
}

// PopulateComputedFields populates (computes) transaction fields such as processing type(s), initially paid fee etc.
func (atp *apiTransactionProcessor) PopulateComputedFields(tx *transaction.ApiTransactionResult) {
	atp.populateComputedFieldsProcessingType(tx)
	atp.populateComputedFieldInitiallyPaidFee(tx)
	atp.populateComputedFieldIsRefund(tx)
}

func (atp *apiTransactionProcessor) populateComputedFieldsProcessingType(tx *transaction.ApiTransactionResult) {
	typeOnSource, typeOnDestination := atp.txTypeHandler.ComputeTransactionType(tx.Tx)
	tx.ProcessingTypeOnSource = typeOnSource.String()
	tx.ProcessingTypeOnDestination = typeOnDestination.String()
}

func (atp *apiTransactionProcessor) populateComputedFieldInitiallyPaidFee(tx *transaction.ApiTransactionResult) {
	// Only user-initiated transactions will present an initially paid fee.
	if tx.Type != string(transaction.TxTypeNormal) && tx.Type != string(transaction.TxTypeInvalid) {
		return
	}

	fee := atp.feeComputer.ComputeTransactionFee(tx)
	// For user-initiated transactions, we can assume the fee is always strictly positive (note: BigInt(0) is stringified as "").
	tx.InitiallyPaidFee = fee.String()
}

func (atp *apiTransactionProcessor) populateComputedFieldIsRefund(tx *transaction.ApiTransactionResult) {
	if tx.Type != string(transaction.TxTypeUnsigned) {
		return
	}

	tx.IsRefund = atp.refundDetector.isRefund(refundDetectorInput{
		Value:         tx.Value,
		Data:          tx.Data,
		ReturnMessage: tx.ReturnMessage,
		GasLimit:      tx.GasLimit,
	})
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

	transactions.SmartContractResults = atp.getUnsignedTransactionsFromPool(requestedFieldsHandler)

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
		return &common.TransactionsPoolForSenderApiResponse{
			Transactions: []common.Transaction{},
		}, nil
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
func (atp *apiTransactionProcessor) GetTransactionsPoolNonceGapsForSender(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	senderAddr, err := atp.addressPubKeyConverter.Decode(sender)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", ErrInvalidAddress.Error(), err)
	}

	senderShard := atp.shardCoordinator.ComputeId(senderAddr)
	nonceGaps, err := atp.extractNonceGaps(string(senderAddr), senderShard, senderAccountNonce)
	if err != nil {
		return nil, err
	}

	return &common.TransactionsPoolNonceGapsForSenderApiResponse{
		Sender: sender,
		Gaps:   nonceGaps,
	}, nil
}

func (atp *apiTransactionProcessor) extractRequestedTxInfoFromObj(txObj interface{}, txType transaction.TxType, txHash []byte, requestedFieldsHandler fieldsHandler) common.Transaction {
	txResult := atp.getApiResultFromObj(txObj, txType)

	wrappedTx := &txcache.WrappedTransaction{
		Tx:     txResult.Tx,
		TxHash: txHash,
	}

	requestedTxInfo := atp.extractRequestedTxInfo(wrappedTx, requestedFieldsHandler)

	return requestedTxInfo
}

func (atp *apiTransactionProcessor) getRegularTransactionsFromPool(requestedFieldsHandler fieldsHandler) ([]common.Transaction, error) {
	regularTxKeys := atp.dataPool.Transactions().Keys()
	regularTxs := make([]common.Transaction, len(regularTxKeys))
	for idx, key := range regularTxKeys {
		txObj, found := atp.getRegularTxObjFromDataPool(key)
		if !found {
			continue
		}

		tx := atp.extractRequestedTxInfoFromObj(txObj, transaction.TxTypeNormal, key, requestedFieldsHandler)

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

		tx := atp.extractRequestedTxInfoFromObj(txObj, transaction.TxTypeReward, key, requestedFieldsHandler)

		rewardTxs[idx] = tx
	}

	return rewardTxs, nil
}

func (atp *apiTransactionProcessor) getUnsignedTransactionsFromPool(requestedFieldsHandler fieldsHandler) []common.Transaction {
	unsignedTxKeys := atp.dataPool.UnsignedTransactions().Keys()
	unsignedTxs := make([]common.Transaction, len(unsignedTxKeys))
	for idx, key := range unsignedTxKeys {
		txObj, found := atp.getUnsignedTxObjFromDataPool(key)
		if !found {
			continue
		}

		tx := atp.extractRequestedTxInfoFromObj(txObj, transaction.TxTypeUnsigned, key, requestedFieldsHandler)

		unsignedTxs[idx] = tx
	}

	return unsignedTxs
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
		tx.TxFields[senderField], _ = atp.addressPubKeyConverter.Encode(wrappedTx.Tx.GetSndAddr())
	}

	if requestedFieldsHandler.HasReceiver {
		tx.TxFields[receiverField], _ = atp.addressPubKeyConverter.Encode(wrappedTx.Tx.GetRcvAddr())
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
		tx.TxFields[valueField] = getTxValue(wrappedTx)
	}

	return tx
}

func (atp *apiTransactionProcessor) fetchTxsForSender(sender string, senderShard uint32) []*txcache.WrappedTransaction {
	cacheId := process.ShardCacherIdentifier(senderShard, senderShard)
	cache := atp.dataPool.Transactions().ShardDataStore(cacheId)
	txCache, ok := cache.(*txcache.TxCache)
	if !ok {
		log.Warn("fetchTxsForSender could not cast to TxCache")
		return nil
	}

	txsForSender := txCache.GetTransactionsPoolForSender(sender)

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

func (atp *apiTransactionProcessor) extractNonceGaps(sender string, senderShard uint32, senderAccountNonce uint64) ([]common.NonceGapApiResponse, error) {
	wrappedTxs := atp.fetchTxsForSender(sender, senderShard)
	if len(wrappedTxs) == 0 {
		return []common.NonceGapApiResponse{}, nil
	}

	nonceGaps := make([]common.NonceGapApiResponse, 0)
	firstNonceInPool := wrappedTxs[0].Tx.GetNonce()
	atp.appendGapFromAccountNonceIfNeeded(senderAccountNonce, firstNonceInPool, senderShard, &nonceGaps)

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

func (atp *apiTransactionProcessor) appendGapFromAccountNonceIfNeeded(
	senderAccountNonce uint64,
	firstNonceInPool uint64,
	senderShard uint32,
	nonceGaps *[]common.NonceGapApiResponse,
) {
	if atp.shardCoordinator.SelfId() != senderShard {
		return
	}

	nonceDif := firstNonceInPool - senderAccountNonce
	if nonceDif >= 1 {
		nonceGap := common.NonceGapApiResponse{
			From: senderAccountNonce,
			To:   firstNonceInPool - 1,
		}
		*nonceGaps = append(*nonceGaps, nonceGap)
	}
}

func (atp *apiTransactionProcessor) optionallyGetTransactionFromPool(hash []byte) *transaction.ApiTransactionResult {
	txObj, txType, found := atp.getTxObjFromDataPool(hash)
	if !found {
		return nil
	}

	return atp.getApiResultFromObj(txObj, txType)
}

func (atp *apiTransactionProcessor) getApiResultFromObj(txObj interface{}, txType transaction.TxType) *transaction.ApiTransactionResult {
	tx := atp.castObjToTransaction(txObj, txType)

	tx.SourceShard = atp.shardCoordinator.ComputeId(tx.Tx.GetSndAddr())
	tx.DestinationShard = atp.shardCoordinator.ComputeId(tx.Tx.GetRcvAddr())
	tx.Status = transaction.TxStatusPending

	return tx
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
		return nil, ErrCannotRetrieveTransaction
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
		err = atp.transactionResultsProcessor.putResultsInTransaction(hash, tx, miniblockMetadata.Epoch)
		if err != nil {
			return nil, err
		}
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
	txsStorer, err := store.GetStorer(dataRetriever.TransactionUnit)
	if err != nil {
		return nil, transaction.TxTypeInvalid, false
	}

	txBytes, err := txsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeNormal, true
	}

	rewardTxsStorer, err := store.GetStorer(dataRetriever.RewardTransactionUnit)
	if err != nil {
		return nil, transaction.TxTypeInvalid, false
	}

	txBytes, err = rewardTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeReward, true
	}

	unsignedTxsStorer, err := store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return nil, transaction.TxTypeInvalid, false
	}

	txBytes, err = unsignedTxsStorer.SearchFirst(hash)
	if err == nil {
		return txBytes, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (atp *apiTransactionProcessor) getTxBytesFromStorageByEpoch(hash []byte, epoch uint32) ([]byte, transaction.TxType, bool) {
	store := atp.storageService
	txsStorer, err := store.GetStorer(dataRetriever.TransactionUnit)
	if err != nil {
		return nil, transaction.TxTypeInvalid, false
	}

	txBytes, err := txsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeNormal, true
	}

	rewardTxsStorer, err := store.GetStorer(dataRetriever.RewardTransactionUnit)
	if err != nil {
		return nil, transaction.TxTypeInvalid, false
	}

	txBytes, err = rewardTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeReward, true
	}

	unsignedTxsStorer, err := store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return nil, transaction.TxTypeInvalid, false
	}

	txBytes, err = unsignedTxsStorer.GetFromEpoch(hash, epoch)
	if err == nil {
		return txBytes, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
}

func (atp *apiTransactionProcessor) castObjToTransaction(txObj interface{}, txType transaction.TxType) *transaction.ApiTransactionResult {
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
	return &transaction.ApiTransactionResult{Type: string(transaction.TxTypeInvalid)}
}

func getTxValue(wrappedTx *txcache.WrappedTransaction) string {
	txValue := wrappedTx.Tx.GetValue()
	if txValue != nil {
		return txValue.String()
	}
	return "0"
}

// UnmarshalTransaction will try to unmarshal the transaction bytes based on the transaction type
func (atp *apiTransactionProcessor) UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	tx, err := atp.txUnmarshaller.unmarshalTransaction(txBytes, txType)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// UnmarshalReceipt will try to unmarshal the provided receipts bytes
func (atp *apiTransactionProcessor) UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error) {
	return atp.txUnmarshaller.unmarshalReceipt(receiptBytes)
}

// IsInterfaceNil returns true if underlying object is nil
func (atp *apiTransactionProcessor) IsInterfaceNil() bool {
	return atp == nil
}
