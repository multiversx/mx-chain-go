package transactionAPI

import (
	"encoding/hex"
	"fmt"
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
}

// NewAPITransactionProcessor will create a new instance of apiTransactionProcessor
func NewAPITransactionProcessor(args *ArgAPITransactionProcessor) (*apiTransactionProcessor, error) {
	err := checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	txUnmarshalerAndPreparer := newTransactionUnmarshaller(args.Marshalizer, args.AddressPubKeyConverter, args.Hasher)

	txResultsProc := newAPITransactionResultProcessor(
		args.AddressPubKeyConverter,
		args.HistoryRepository,
		args.StorageService,
		args.Marshalizer,
		txUnmarshalerAndPreparer,
		args.LogsFacade,
		args.ShardCoordinator.SelfId(),
	)

	refundDetector := newRefundDetector()

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
		refundDetector:              refundDetector,
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
	atp.populateComputedFieldsProcessingType(tx)
	atp.populateComputedFieldInitiallyPaidFee(tx)
	atp.populateComputedFieldIsRefund(tx)

	return tx, nil
}

func (atp *apiTransactionProcessor) doGetTransaction(hash []byte, withResults bool) (*transaction.ApiTransactionResult, error) {
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

// GetTransactionsPool will return a structure containing the transactions pool that is to be returned on API calls
func (atp *apiTransactionProcessor) GetTransactionsPool() (*common.TransactionsPoolAPIResponse, error) {
	txsPoolResponse := &common.TransactionsPoolAPIResponse{
		RegularTransactions:  txsHashesBytesToString(atp.dataPool.Transactions().Keys()),
		SmartContractResults: txsHashesBytesToString(atp.dataPool.UnsignedTransactions().Keys()),
		Rewards:              txsHashesBytesToString(atp.dataPool.RewardTransactions().Keys()),
	}

	return txsPoolResponse, nil
}

func txsHashesBytesToString(input [][]byte) []string {
	result := make([]string, 0, len(input))
	for _, txHashBytes := range input {
		result = append(result, hex.EncodeToString(txHashBytes))
	}

	return result
}

func (atp *apiTransactionProcessor) optionallyGetTransactionFromPool(hash []byte) (*transaction.ApiTransactionResult, error) {
	txObj, txType, found := atp.getTxObjFromDataPool(hash)
	if !found {
		return nil, nil
	}

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
	txsPool := atp.dataPool.Transactions()
	txObj, found := txsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeNormal, true
	}

	rewardTxsPool := atp.dataPool.RewardTransactions()
	txObj, found = rewardTxsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeReward, true
	}

	unsignedTxsPool := atp.dataPool.UnsignedTransactions()
	txObj, found = unsignedTxsPool.SearchFirstData(hash)
	if found && txObj != nil {
		return txObj, transaction.TxTypeUnsigned, true
	}

	return nil, transaction.TxTypeInvalid, false
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
func (atp *apiTransactionProcessor) UnmarshalTransaction(epoch uint32, txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	tx, err := atp.txUnmarshaller.unmarshalTransaction(txBytes, txType)
	if err != nil {
		return nil, err
	}

	tx.Epoch = epoch
	atp.populateComputedFieldsProcessingType(tx)
	atp.populateComputedFieldInitiallyPaidFee(tx)
	atp.populateComputedFieldIsRefund(tx)

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
