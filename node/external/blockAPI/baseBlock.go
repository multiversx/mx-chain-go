package blockAPI

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/transaction/status"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/api/shared/logging"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	outportProcess "github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// BlockStatus is the status of a block
type BlockStatus string

const (
	// BlockStatusOnChain represents the identifier for an on-chain block
	BlockStatusOnChain = "on-chain"
	// BlockStatusReverted represent the identifier for a reverted block
	BlockStatusReverted = "reverted"
)

type baseAPIBlockProcessor struct {
	hasDbLookupExtensions        bool
	selfShardID                  uint32
	emptyReceiptsHash            []byte
	store                        dataRetriever.StorageService
	marshalizer                  marshal.Marshalizer
	uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
	historyRepo                  dblookupext.HistoryRepository
	hasher                       hashing.Hasher
	addressPubKeyConverter       core.PubkeyConverter
	txStatusComputer             status.StatusComputerHandler
	apiTransactionHandler        APITransactionHandler
	logsFacade                   logsFacade
	receiptsRepository           receiptsRepository
	alteredAccountsProvider      outportProcess.AlteredAccountsProviderHandler
	accountsRepository           state.AccountsRepository
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	enableEpochsHandler          common.EnableEpochsHandler
}

var log = logger.GetOrCreate("node/blockAPI")

func (bap *baseAPIBlockProcessor) getIntrashardMiniblocksFromReceiptsStorage(header data.HeaderHandler, headerHash []byte, options api.BlockQueryOptions) ([]*api.MiniBlock, error) {
	receiptsHolder, err := bap.receiptsRepository.LoadReceipts(header, headerHash)
	if err != nil {
		return nil, err
	}

	apiMiniblocks := make([]*api.MiniBlock, 0, len(receiptsHolder.GetMiniblocks()))
	for _, miniblock := range receiptsHolder.GetMiniblocks() {
		apiMiniblock, err := bap.convertMiniblockFromReceiptsStorageToApiMiniblock(miniblock, header, options)
		if err != nil {
			return nil, err
		}

		apiMiniblocks = append(apiMiniblocks, apiMiniblock)
	}

	return apiMiniblocks, nil
}

func (bap *baseAPIBlockProcessor) convertMiniblockFromReceiptsStorageToApiMiniblock(miniblock *block.MiniBlock, header data.HeaderHandler, options api.BlockQueryOptions) (*api.MiniBlock, error) {
	mbHash, err := core.CalculateHash(bap.marshalizer, bap.hasher, miniblock)
	if err != nil {
		return nil, err
	}

	miniblockAPI := &api.MiniBlock{
		Hash:                  hex.EncodeToString(mbHash),
		Type:                  miniblock.Type.String(),
		SourceShard:           miniblock.SenderShardID,
		DestinationShard:      miniblock.ReceiverShardID,
		IsFromReceiptsStorage: true,
		ProcessingType:        block.ProcessingType(miniblock.GetProcessingType()).String(),
		// It's a bit more complex (and not necessary at this point) to also set the construction state here.
	}

	if options.WithTransactions {
		firstProcessed := int32(0)
		lastProcessed := int32(len(miniblock.TxHashes) - 1)

		err = bap.getAndAttachTxsToMbByEpoch(mbHash, miniblock, header, miniblockAPI, firstProcessed, lastProcessed, options)
		if err != nil {
			return nil, err
		}
	}

	return miniblockAPI, nil
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMb(
	mbHeader data.MiniBlockHeaderHandler,
	header data.HeaderHandler,
	apiMiniblock *api.MiniBlock,
	options api.BlockQueryOptions,
) error {
	miniblockHash := mbHeader.GetHash()
	miniBlock, err := bap.getMiniblockByHashAndEpoch(miniblockHash, header.GetEpoch())
	if err != nil {
		return err
	}

	firstProcessed := mbHeader.GetIndexOfFirstTxProcessed()
	lastProcessed := mbHeader.GetIndexOfLastTxProcessed()
	return bap.getAndAttachTxsToMbByEpoch(miniblockHash, miniBlock, header, apiMiniblock, firstProcessed, lastProcessed, options)
}

func (bap *baseAPIBlockProcessor) getMiniblockByHashAndEpoch(miniblockHash []byte, epoch uint32) (*block.MiniBlock, error) {
	buff, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, miniblockHash, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v, hash = %s", errCannotLoadMiniblocks, err, hex.EncodeToString(miniblockHash))
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, buff)
	if err != nil {
		return nil, fmt.Errorf("%w: %v, hash = %s", errCannotUnmarshalMiniblocks, err, hex.EncodeToString(miniblockHash))
	}

	return miniBlock, nil
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMbByEpoch(
	miniblockHash []byte,
	miniBlock *block.MiniBlock,
	header data.HeaderHandler,
	apiMiniblock *api.MiniBlock,
	firstProcessedTxIndex int32,
	lastProcessedTxIndex int32,
	options api.BlockQueryOptions,
) error {
	var err error

	switch miniBlock.Type {
	case block.TxBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, header, transaction.TxTypeNormal, dataRetriever.TransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.RewardsBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, header, transaction.TxTypeReward, dataRetriever.RewardTransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.SmartContractResultBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, header, transaction.TxTypeUnsigned, dataRetriever.UnsignedTransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.InvalidBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, header, transaction.TxTypeInvalid, dataRetriever.TransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.ReceiptBlock:
		apiMiniblock.Receipts, err = bap.getReceiptsFromMiniblock(miniBlock, header.GetEpoch())
	}

	if err != nil {
		return err
	}

	if options.WithLogs {
		err = bap.logsFacade.IncludeLogsInTransactions(apiMiniblock.Transactions, miniBlock.TxHashes, header.GetEpoch())
		if err != nil {
			return err
		}
	}

	return nil
}

func (bap *baseAPIBlockProcessor) getReceiptsFromMiniblock(miniblock *block.MiniBlock, epoch uint32) ([]*transaction.ApiReceipt, error) {
	storer, err := bap.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	marshalledReceipts, err := storer.GetBulkFromEpoch(miniblock.TxHashes, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotLoadReceipts, err)
	}
	logging.LogAPIActionDurationIfNeeded(start, "GetBulkFromEpoch")

	apiReceipts := make([]*transaction.ApiReceipt, 0)
	for _, pair := range marshalledReceipts {
		receipt, errUnmarshal := bap.apiTransactionHandler.UnmarshalReceipt(pair.Value)
		if errUnmarshal != nil {
			return nil, fmt.Errorf("%w: %v, hash = %s", errCannotUnmarshalReceipts, errUnmarshal, hex.EncodeToString(pair.Key))
		}

		apiReceipts = append(apiReceipts, receipt)
	}

	return apiReceipts, nil
}

func (bap *baseAPIBlockProcessor) getTxsFromMiniblock(
	miniblock *block.MiniBlock,
	miniblockHash []byte,
	header data.HeaderHandler,
	txType transaction.TxType,
	unit dataRetriever.UnitType,
	firstProcessedTxIndex int32,
	lastProcessedTxIndex int32,
) ([]*transaction.ApiTransactionResult, error) {
	storer, err := bap.store.GetStorer(unit)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	executedTxHashes := extractExecutedTxHashes(miniblock.TxHashes, firstProcessedTxIndex, lastProcessedTxIndex)
	marshalledTxs, err := storer.GetBulkFromEpoch(executedTxHashes, header.GetEpoch())
	if err != nil {
		return nil, fmt.Errorf("%w: %v, miniblock = %s", errCannotLoadTransactions, err, hex.EncodeToString(miniblockHash))
	}
	logging.LogAPIActionDurationIfNeeded(start, "GetBulkFromEpoch")

	start = time.Now()
	txs := make([]*transaction.ApiTransactionResult, 0)
	for _, pair := range marshalledTxs {
		tx, errUnmarshalTx := bap.apiTransactionHandler.UnmarshalTransaction(pair.Value, txType)
		if errUnmarshalTx != nil {
			return nil, fmt.Errorf("%w: %v, miniblock = %s", errCannotUnmarshalTransactions, err, hex.EncodeToString(miniblockHash))
		}
		tx.Hash = hex.EncodeToString(pair.Key)
		tx.HashBytes = pair.Key
		tx.MiniBlockType = miniblock.Type.String()
		tx.MiniBlockHash = hex.EncodeToString(miniblockHash)
		tx.SourceShard = miniblock.SenderShardID
		tx.DestinationShard = miniblock.ReceiverShardID
		tx.Epoch = header.GetEpoch()
		tx.Round = header.GetRound()
		bap.apiTransactionHandler.PopulateComputedFields(tx)

		// TODO : should check if tx is reward reverted
		tx.Status, _ = bap.txStatusComputer.ComputeStatusWhenInStorageKnowingMiniblock(miniblock.Type, tx)

		txs = append(txs, tx)
	}
	logging.LogAPIActionDurationIfNeeded(start, "UnmarshalTransactions")

	return txs, nil
}

func (bap *baseAPIBlockProcessor) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
	if !bap.hasDbLookupExtensions {
		return bap.store.Get(unit, key)
	}

	epoch, err := bap.historyRepo.GetEpochByHash(key)
	if err != nil {
		return nil, err
	}

	storer, err := bap.store.GetStorer(unit)
	if err != nil {
		return nil, err
	}

	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBlockProcessor) getFromStorerWithEpoch(unit dataRetriever.UnitType, key []byte, epoch uint32) ([]byte, error) {
	storer, err := bap.store.GetStorer(unit)
	if err != nil {
		return nil, err
	}

	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBlockProcessor) computeBlockStatus(storerUnit dataRetriever.UnitType, blockAPI *api.Block) (string, error) {
	nonceToByteSlice := bap.uint64ByteSliceConverter.ToByteSlice(blockAPI.Nonce)
	headerHash, err := bap.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return "", err
	}

	if hex.EncodeToString(headerHash) != blockAPI.Hash {
		return BlockStatusReverted, err
	}

	return BlockStatusOnChain, nil
}

func (bap *baseAPIBlockProcessor) computeStatusAndPutInBlock(blockAPI *api.Block, storerUnit dataRetriever.UnitType) (*api.Block, error) {
	blockStatus, err := bap.computeBlockStatus(storerUnit, blockAPI)
	if err != nil {
		return nil, err
	}

	blockAPI.Status = blockStatus

	return blockAPI, nil
}

func (bap *baseAPIBlockProcessor) getBlockHeaderHashAndBytesByRound(round uint64, blockUnitType dataRetriever.UnitType) (headerHash []byte, blockBytes []byte, err error) {
	roundToByteSlice := bap.uint64ByteSliceConverter.ToByteSlice(round)
	headerHash, err = bap.store.Get(dataRetriever.RoundHdrHashDataUnit, roundToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	blockBytes, err = bap.getFromStorer(blockUnitType, headerHash)
	if err != nil {
		return nil, nil, err
	}

	return headerHash, blockBytes, nil
}

func extractExecutedTxHashes(mbTxHashes [][]byte, firstProcessed, lastProcessed int32) [][]byte {
	invalidIndexes := firstProcessed < 0 || lastProcessed >= int32(len(mbTxHashes)) || firstProcessed > lastProcessed
	if invalidIndexes {
		log.Warn("extractExecutedTxHashes encountered invalid indices", "firstProcessed", firstProcessed,
			"lastProcessed", lastProcessed,
			"len(mbTxHashes)", len(mbTxHashes),
		)
		return mbTxHashes
	}

	return mbTxHashes[firstProcessed : lastProcessed+1]
}

func addScheduledInfoInBlock(header data.HeaderHandler, apiBlock *api.Block) {
	additionalData := header.GetAdditionalData()
	if check.IfNil(additionalData) {
		return
	}

	apiBlock.ScheduledData = &api.ScheduledData{
		ScheduledRootHash:        hex.EncodeToString(additionalData.GetScheduledRootHash()),
		ScheduledAccumulatedFees: bigIntToStr(additionalData.GetScheduledAccumulatedFees()),
		ScheduledDeveloperFees:   bigIntToStr(additionalData.GetScheduledDeveloperFees()),
		ScheduledGasProvided:     additionalData.GetScheduledGasProvided(),
		ScheduledGasPenalized:    additionalData.GetScheduledGasPenalized(),
		ScheduledGasRefunded:     additionalData.GetScheduledGasRefunded(),
	}
}

func bigIntToStr(value *big.Int) string {
	if value == nil {
		return "0"
	}

	return value.String()
}

func alteredAccountsMapToAPIResponse(alteredAccounts map[string]*alteredAccount.AlteredAccount, tokensFilter string) []*alteredAccount.AlteredAccount {
	response := make([]*alteredAccount.AlteredAccount, 0, len(alteredAccounts))

	for address, altAccount := range alteredAccounts {
		apiAlteredAccount := &alteredAccount.AlteredAccount{
			Address: address,
			Balance: altAccount.Balance,
			Nonce:   altAccount.Nonce,
		}

		if len(tokensFilter) > 0 {
			attachTokensToAlteredAccount(apiAlteredAccount, altAccount, tokensFilter)
		}

		response = append(response, apiAlteredAccount)
	}

	return response
}

func attachTokensToAlteredAccount(apiAlteredAccount *alteredAccount.AlteredAccount, altAccount *alteredAccount.AlteredAccount, tokensFilter string) {
	for _, token := range altAccount.Tokens {
		if !shouldAddTokenToResult(token.Identifier, tokensFilter) {
			continue
		}

		apiAlteredAccount.Tokens = append(apiAlteredAccount.Tokens, &alteredAccount.AccountTokenData{
			Identifier:     token.Identifier,
			Balance:        token.Balance,
			Nonce:          token.Nonce,
			Properties:     token.Properties,
			MetaData:       nil,
			AdditionalData: nil,
		})
	}
}

func shouldAddTokenToResult(tokenIdentifier string, tokensFilter string) bool {
	if shouldIncludeAllTokens(tokensFilter) {
		return true
	}

	return strings.Contains(tokensFilter, tokenIdentifier)
}

func shouldIncludeAllTokens(tokensFilter string) bool {
	return tokensFilter == "*" || tokensFilter == "all"
}

func (bap *baseAPIBlockProcessor) apiBlockToAlteredAccounts(apiBlock *api.Block, options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	blockHash, err := hex.DecodeString(apiBlock.Hash)
	if err != nil {
		return nil, err
	}

	var blockRootHash []byte
	blockRootHash, err = bap.scheduledTxsExecutionHandler.GetScheduledRootHashForHeaderWithEpoch(blockHash, apiBlock.Epoch)
	if err != nil {
		blockRootHash, err = hex.DecodeString(apiBlock.StateRootHash)
		if err != nil {
			return nil, err
		}
	}

	alteredAccountsOptions := shared.AlteredAccountsOptions{
		WithCustomAccountsRepository: true,
		AccountsRepository:           bap.accountsRepository,
		AccountQueryOptions: api.AccountQueryOptions{
			BlockHash:     blockHash,
			BlockNonce:    core.OptionalUint64{HasValue: true, Value: apiBlock.Nonce},
			BlockRootHash: blockRootHash,
			HintEpoch:     core.OptionalUint32{HasValue: true, Value: apiBlock.Epoch},
		},
	}

	// TODO: might refactor, so altered accounts component could only need a slice of addresses instead of a tx pool
	outportPool, err := bap.apiBlockToOutportPool(apiBlock)
	if err != nil {
		return nil, err
	}
	alteredAccounts, err := bap.alteredAccountsProvider.ExtractAlteredAccountsFromPool(outportPool, alteredAccountsOptions)
	if err != nil {
		return nil, err
	}

	alteredAccountsAPI := alteredAccountsMapToAPIResponse(alteredAccounts, options.TokensFilter)
	return alteredAccountsAPI, nil
}

func (bap *baseAPIBlockProcessor) apiBlockToOutportPool(apiBlock *api.Block) (*outport.TransactionPool, error) {
	pool := &outport.TransactionPool{
		Transactions:         make(map[string]*outport.TxInfo),
		SmartContractResults: make(map[string]*outport.SCRInfo),
		InvalidTxs:           make(map[string]*outport.TxInfo),
		Rewards:              make(map[string]*outport.RewardInfo),
		Logs:                 make([]*outport.LogData, 0),
	}

	var err error
	for _, miniBlock := range apiBlock.MiniBlocks {
		for _, tx := range miniBlock.Transactions {
			err = bap.addTxToPool(tx, pool)
			if err != nil {
				return nil, err
			}

			err = bap.addLogsToPool(tx, pool)
			if err != nil {
				return nil, err
			}
		}
	}

	return pool, nil
}

func (bap *baseAPIBlockProcessor) addLogsToPool(tx *transaction.ApiTransactionResult, pool *outport.TransactionPool) error {
	if tx.Logs == nil {
		return nil
	}

	logAddressBytes, err := bap.addressPubKeyConverter.Decode(tx.Logs.Address)
	if err != nil {
		return fmt.Errorf("error while decoding the log's address. address=%s, error=%s", tx.Logs.Address, err.Error())
	}

	logsEvents := make([]*transaction.Event, 0)
	for _, logEvent := range tx.Logs.Events {
		eventAddressBytes, err := bap.addressPubKeyConverter.Decode(logEvent.Address)
		if err != nil {
			return fmt.Errorf("error while decoding the event's address. address=%s, error=%s", logEvent.Address, err.Error())
		}

		logsEvents = append(logsEvents, &transaction.Event{
			Address:    eventAddressBytes,
			Identifier: []byte(logEvent.Identifier),
			Topics:     logEvent.Topics,
			Data:       logEvent.Data,
		})
	}

	pool.Logs = append(pool.Logs, &outport.LogData{
		TxHash: tx.Hash,
		Log: &transaction.Log{
			Address: logAddressBytes,
			Events:  logsEvents,
		},
	})

	return nil
}

func (bap *baseAPIBlockProcessor) addTxToPool(tx *transaction.ApiTransactionResult, pool *outport.TransactionPool) error {
	senderBytes, err := bap.addressPubKeyConverter.Decode(tx.Sender)
	if err != nil && tx.Type != string(transaction.TxTypeReward) {
		return fmt.Errorf("error while decoding the sender address. address=%s, error=%s", tx.Sender, err.Error())
	}
	receiverBytes, err := bap.addressPubKeyConverter.Decode(tx.Receiver)
	if err != nil {
		return fmt.Errorf("error while decoding the receiver address. address=%s, error=%s", tx.Receiver, err.Error())
	}

	txValueString := tx.Value
	if len(txValueString) == 0 {
		txValueString = "0"
	}
	txValue, ok := big.NewInt(0).SetString(txValueString, 10)
	if !ok {
		return fmt.Errorf("cannot convert tx value to big int. Value=%s", tx.Value)
	}

	switch tx.Type {
	case string(transaction.TxTypeNormal):
		pool.Transactions[tx.Hash] = &outport.TxInfo{
			Transaction: &transaction.Transaction{
				SndAddr: senderBytes,
				RcvAddr: receiverBytes,
				Value:   txValue,
			},
			FeeInfo: newFeeInfo(),
		}

	case string(transaction.TxTypeUnsigned):
		pool.SmartContractResults[tx.Hash] = &outport.SCRInfo{
			SmartContractResult: &smartContractResult.SmartContractResult{
				SndAddr: senderBytes,
				RcvAddr: receiverBytes,
				Value:   txValue,
			},
			FeeInfo: newFeeInfo(),
		}
	case string(transaction.TxTypeInvalid):
		pool.InvalidTxs[tx.Hash] = &outport.TxInfo{
			Transaction: &transaction.Transaction{
				SndAddr: senderBytes,
				// do not set the receiver since the cost is only on sender's side in case of invalid txs
				Value: txValue,
			},
			FeeInfo: newFeeInfo(),
		}

	case string(transaction.TxTypeReward):
		pool.Rewards[tx.Hash] = &outport.RewardInfo{
			Reward: &rewardTx.RewardTx{
				RcvAddr: receiverBytes,
				Value:   txValue,
			},
		}
	}

	return nil
}

func newFeeInfo() *outport.FeeInfo {
	return &outport.FeeInfo{
		GasUsed:        0,
		Fee:            big.NewInt(0),
		InitialPaidFee: big.NewInt(0),
	}
}

func createAlteredBlockHash(hash []byte) []byte {
	alteredHash := make([]byte, 0)
	alteredHash = append(alteredHash, hash...)
	alteredHash = append(alteredHash, []byte(common.GenesisStorageSuffix)...)

	return alteredHash
}
