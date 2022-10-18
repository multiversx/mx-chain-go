package blockAPI

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/shared/logging"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/outport/process"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts/shared"
	"github.com/ElrondNetwork/elrond-go/state"
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
	hasDbLookupExtensions    bool
	selfShardID              uint32
	emptyReceiptsHash        []byte
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	historyRepo              dblookupext.HistoryRepository
	hasher                   hashing.Hasher
	addressPubKeyConverter   core.PubkeyConverter
	txStatusComputer         transaction.StatusComputerHandler
	apiTransactionHandler    APITransactionHandler
	logsFacade               logsFacade
	receiptsRepository       receiptsRepository
	alteredAccountsProvider  process.AlteredAccountsProviderHandler
	accountsRepository       state.AccountsRepository
}

var log = logger.GetOrCreate("node/blockAPI")

func (bap *baseAPIBlockProcessor) getIntrashardMiniblocksFromReceiptsStorage(header data.HeaderHandler, headerHash []byte, options api.BlockQueryOptions) ([]*api.MiniBlock, error) {
	receiptsHolder, err := bap.receiptsRepository.LoadReceipts(header, headerHash)
	if err != nil {
		return nil, err
	}

	apiMiniblocks := make([]*api.MiniBlock, 0, len(receiptsHolder.GetMiniblocks()))
	for _, miniblock := range receiptsHolder.GetMiniblocks() {
		apiMiniblock, err := bap.convertMiniblockFromReceiptsStorageToApiMiniblock(miniblock, header.GetEpoch(), options)
		if err != nil {
			return nil, err
		}

		apiMiniblocks = append(apiMiniblocks, apiMiniblock)
	}

	return apiMiniblocks, nil
}

func (bap *baseAPIBlockProcessor) convertMiniblockFromReceiptsStorageToApiMiniblock(miniblock *block.MiniBlock, epoch uint32, options api.BlockQueryOptions) (*api.MiniBlock, error) {
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

		err = bap.getAndAttachTxsToMbByEpoch(mbHash, miniblock, epoch, miniblockAPI, firstProcessed, lastProcessed, options)
		if err != nil {
			return nil, err
		}
	}

	return miniblockAPI, nil
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMb(
	mbHeader data.MiniBlockHeaderHandler,
	epoch uint32,
	apiMiniblock *api.MiniBlock,
	options api.BlockQueryOptions,
) error {
	miniblockHash := mbHeader.GetHash()
	miniBlock, err := bap.getMiniblockByHashAndEpoch(miniblockHash, epoch)
	if err != nil {
		return err
	}

	firstProcessed := mbHeader.GetIndexOfFirstTxProcessed()
	lastProcessed := mbHeader.GetIndexOfLastTxProcessed()
	return bap.getAndAttachTxsToMbByEpoch(miniblockHash, miniBlock, epoch, apiMiniblock, firstProcessed, lastProcessed, options)
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
	epoch uint32,
	apiMiniblock *api.MiniBlock,
	firstProcessedTxIndex int32,
	lastProcessedTxIndex int32,
	options api.BlockQueryOptions,
) error {
	var err error

	switch miniBlock.Type {
	case block.TxBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeNormal, dataRetriever.TransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.RewardsBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeReward, dataRetriever.RewardTransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.SmartContractResultBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeUnsigned, dataRetriever.UnsignedTransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.InvalidBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeInvalid, dataRetriever.TransactionUnit, firstProcessedTxIndex, lastProcessedTxIndex)
	case block.ReceiptBlock:
		apiMiniblock.Receipts, err = bap.getReceiptsFromMiniblock(miniBlock, epoch)
	}

	if err != nil {
		return err
	}

	if options.WithLogs {
		err = bap.logsFacade.IncludeLogsInTransactions(apiMiniblock.Transactions, miniBlock.TxHashes, epoch)
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
	epoch uint32,
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
	marshalledTxs, err := storer.GetBulkFromEpoch(executedTxHashes, epoch)
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
		tx.Epoch = epoch
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

func alteredAccountsMapToAPIResponse(alteredAccounts map[string]*outport.AlteredAccount, tokensFilter string, withMetadata bool) []*outport.AlteredAccount {
	response := make([]*outport.AlteredAccount, 0, len(alteredAccounts))

	for address, altAccount := range alteredAccounts {
		apiAlteredAccount := &outport.AlteredAccount{
			Address: address,
			Balance: altAccount.Balance,
		}

		if len(tokensFilter) > 0 {
			attachTokensToAlteredAccount(apiAlteredAccount, altAccount, tokensFilter, withMetadata)
		}

		response = append(response, apiAlteredAccount)
	}

	return response
}

func attachTokensToAlteredAccount(apiAlteredAccount *outport.AlteredAccount, altAccount *outport.AlteredAccount, tokensFilter string, withMetadata bool) {
	for _, token := range altAccount.Tokens {
		if !shouldAddTokenToResult(token.Identifier, tokensFilter) {
			continue
		}
		if withMetadata {
			apiAlteredAccount.Tokens = append(apiAlteredAccount.Tokens, token)
			continue
		}

		apiAlteredAccount.Tokens = append(apiAlteredAccount.Tokens, &outport.AccountTokenData{
			Identifier: token.Identifier,
			Balance:    token.Balance,
			Nonce:      token.Nonce,
			Properties: token.Properties,
			MetaData:   nil,
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

func (bap *baseAPIBlockProcessor) apiBlockToAlteredAccounts(apiBlock *api.Block, options api.GetAlteredAccountsForBlockOptions) ([]*outport.AlteredAccount, error) {
	alteredAccountsOptions := shared.AlteredAccountsOptions{
		WithCustomAccountsRepository: true,
		AccountsRepository:           bap.accountsRepository,
		// TODO: AccountQueryOptions could be used like options.WithBlockNonce(..) instead of thinking what to provide

		// send the block nonce as it guarantees the opening of the storer for the right epoch. Sending the block hash
		// would be more optimal, but there is no link between a root hash and a block, which can result in the endpoint
		// not working
		AccountQueryOptions: api.AccountQueryOptions{
			BlockNonce: core.OptionalUint64{
				HasValue: true,
				Value:    apiBlock.Nonce,
			},
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

	alteredAccountsAPI := alteredAccountsMapToAPIResponse(alteredAccounts, options.TokensFilter, options.WithMetadata)
	return alteredAccountsAPI, nil
}

func (bap *baseAPIBlockProcessor) apiBlockToOutportPool(apiBlock *api.Block) (*outport.Pool, error) {
	pool := &outport.Pool{
		Txs:     make(map[string]data.TransactionHandlerWithGasUsedAndFee),
		Scrs:    make(map[string]data.TransactionHandlerWithGasUsedAndFee),
		Invalid: make(map[string]data.TransactionHandlerWithGasUsedAndFee),
		Rewards: make(map[string]data.TransactionHandlerWithGasUsedAndFee),
		Logs:    make([]*data.LogData, 0),
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

func (bap *baseAPIBlockProcessor) addLogsToPool(tx *transaction.ApiTransactionResult, pool *outport.Pool) error {
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

	pool.Logs = append(pool.Logs, &data.LogData{
		LogHandler: &transaction.Log{
			Address: logAddressBytes,
			Events:  logsEvents,
		},
		TxHash: tx.Hash,
	})

	return nil
}

func (bap *baseAPIBlockProcessor) addTxToPool(tx *transaction.ApiTransactionResult, pool *outport.Pool) error {
	senderBytes, err := bap.addressPubKeyConverter.Decode(tx.Sender)
	if err != nil && tx.Type != string(transaction.TxTypeReward) {
		return fmt.Errorf("error while decoding the sender address. address=%s, error=%s", tx.Sender, err.Error())
	}
	receiverBytes, err := bap.addressPubKeyConverter.Decode(tx.Receiver)
	if err != nil {
		return fmt.Errorf("error while decoding the receiver address. address=%s, error=%s", tx.Receiver, err.Error())
	}

	zeroBigInt := big.NewInt(0)

	switch tx.Type {
	case string(transaction.TxTypeNormal):
		pool.Txs[tx.Hash] = outport.NewTransactionHandlerWithGasAndFee(
			&transaction.Transaction{
				SndAddr: senderBytes,
				RcvAddr: receiverBytes,
			},
			0,
			zeroBigInt,
		)
	case string(transaction.TxTypeUnsigned):
		pool.Scrs[tx.Hash] = outport.NewTransactionHandlerWithGasAndFee(
			&smartContractResult.SmartContractResult{
				SndAddr: senderBytes,
				RcvAddr: receiverBytes,
			},
			0,
			zeroBigInt,
		)
	case string(transaction.TxTypeInvalid):
		pool.Invalid[tx.Hash] = outport.NewTransactionHandlerWithGasAndFee(
			&transaction.Transaction{
				SndAddr: senderBytes,
				// do not set the receiver since the cost is only on sender's side in case of invalid txs
			},
			0,
			zeroBigInt,
		)
	case string(transaction.TxTypeReward):
		pool.Rewards[tx.Hash] = outport.NewTransactionHandlerWithGasAndFee(
			&rewardTx.RewardTx{
				RcvAddr: receiverBytes,
			},
			0,
			zeroBigInt,
		)
	}

	return nil
}
