package bootstrap

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
	updateSync "github.com/ElrondNetwork/elrond-go/update/sync"
)

type startInEpochWithScheduledDataSyncer struct {
	scheduledHeadersSyncer    epochStart.HeadersByHashSyncer
	scheduledMiniBlocksSyncer epochStart.PendingMiniBlocksSyncHandler
	txSyncer                  update.TransactionsSyncHandler
	scheduledEnableEpoch      uint32
}

func NewStartInEpochShardHeaderDataSyncerWithScheduled(
	dataPool dataRetriever.PoolsHolder,
	marshaller marshal.Marshalizer,
	requestHandler process.RequestHandler,
	scheduledEnableEpoch uint32,
) (*startInEpochWithScheduledDataSyncer, error) {
	syncMiniBlocksArgs := updateSync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          dataPool.MiniBlocks(),
		Marshalizer:    marshaller,
		RequestHandler: requestHandler,
	}
	miniBlocksSyncer, err := updateSync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
	if err != nil {
		return nil, err
	}

	syncMissingHeadersArgs := updateSync.ArgsNewMissingHeadersByHashSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          dataPool.Headers(),
		Marshalizer:    marshaller,
		RequestHandler: requestHandler,
	}

	headersSyncer, err := updateSync.NewMissingheadersByHashSyncer(syncMissingHeadersArgs)
	if err != nil {
		return nil, err
	}

	syncTxsArgs := updateSync.ArgsNewTransactionsSyncer{
		DataPools:      dataPool,
		Storages:       dataRetriever.NewChainStorer(),
		Marshalizer:    marshaller,
		RequestHandler: requestHandler,
	}

	txSyncer, err := updateSync.NewTransactionsSyncer(syncTxsArgs)
	if err != nil {
		return nil, err
	}

	return &startInEpochWithScheduledDataSyncer{
		scheduledMiniBlocksSyncer: miniBlocksSyncer,
		scheduledHeadersSyncer:    headersSyncer,
		txSyncer:                  txSyncer,
		scheduledEnableEpoch:      scheduledEnableEpoch,
	}, nil
}

func (ses *startInEpochWithScheduledDataSyncer) updateSyncDataIfNeeded(
	notarizedShardHeader data.ShardHeaderHandler,
	pendingMiniBlocks map[string]*block.MiniBlock,
) (data.ShardHeaderHandler, map[string]*block.MiniBlock, error) {
	if ses.scheduledEnableEpoch > notarizedShardHeader.GetEpoch() {
		return notarizedShardHeader, pendingMiniBlocks, nil
	}

	headerToBeProcessed, err := ses.getRequiredHeaderByHash(notarizedShardHeader)
	if err != nil {
		return nil, nil, err
	}

	allMiniBlocks, err := ses.getMiniBlocks(notarizedShardHeader)
	if err != nil {
		return nil, nil, err
	}

	updatedPendingMiniBlocks, err := ses.getPendingMiniBlocks(
		notarizedShardHeader,
		pendingMiniBlocks,
		allMiniBlocks,
	)
	if err != nil {
		return nil, nil, err
	}

	err = ses.prepareScheduledSCRs(
		headerToBeProcessed,
		notarizedShardHeader,
		allMiniBlocks,
	)
	if err != nil {
		return nil, nil, err
	}

	return headerToBeProcessed, updatedPendingMiniBlocks, nil
}

func (ses *startInEpochWithScheduledDataSyncer) getRequiredHeaderByHash(notarizedShardHeader data.ShardHeaderHandler, ) (data.ShardHeaderHandler, error) {
	shardIDs := []uint32{
		notarizedShardHeader.GetShardID(),
	}
	hashesToRequest := [][]byte{
		notarizedShardHeader.GetPrevHash(),
	}

	ses.scheduledHeadersSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := ses.scheduledHeadersSyncer.SyncMissingHeadersByHash(shardIDs, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	headers, err := ses.scheduledHeadersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	headerToBeProcessed, ok := headers[string(notarizedShardHeader.GetPrevHash())].(data.ShardHeaderHandler)
	if !ok {
		return nil, epochStart.ErrMissingHeader
	}

	return headerToBeProcessed, nil
}

func copyPendingMiniBlocksMap(pendingMiniBlocks map[string]*block.MiniBlock) map[string]*block.MiniBlock {
	result := make(map[string]*block.MiniBlock)
	for i := range pendingMiniBlocks {
		result[i] = pendingMiniBlocks[i]
	}
	return result
}

func (ses *startInEpochWithScheduledDataSyncer) getMiniBlocks(
	notarizedShardHeader data.ShardHeaderHandler,
) (map[string]*block.MiniBlock, error) {
	processedMiniBlockHeaders := notarizedShardHeader.GetMiniBlockHeaderHandlers()
	return ses.getRequiredMiniBlocksByMbHeader(processedMiniBlockHeaders)
}

func (ses *startInEpochWithScheduledDataSyncer) getPendingMiniBlocks(
	notarizedShardHeader data.ShardHeaderHandler,
	pendingMiniBlocks map[string]*block.MiniBlock,
	allMiniBlocks map[string]*block.MiniBlock,
) (map[string]*block.MiniBlock, error) {
	previousPendingMiniBlocks := copyPendingMiniBlocksMap(pendingMiniBlocks)
	processedMiniBlockHeaders := notarizedShardHeader.GetMiniBlockHeaderHandlers()
	ownShardID := notarizedShardHeader.GetShardID()

	for _, mbHeader := range processedMiniBlockHeaders {
		if mbHeader.GetReceiverShardID() != ownShardID {
			continue
		}
		if mbHeader.GetSenderShardID() == ownShardID {
			continue
		}

		mbHash := mbHeader.GetHash()
		mb, ok := allMiniBlocks[string(mbHash)]
		if !ok {
			return nil, epochStart.ErrMissingMiniBlock
		}

		previousPendingMiniBlocks[string(mbHash)] = mb
	}

	return previousPendingMiniBlocks, nil
}

func (ses *startInEpochWithScheduledDataSyncer) getRequiredMiniBlocksByMbHeader(
	mbHeaders []data.MiniBlockHeaderHandler,
) (map[string]*block.MiniBlock, error) {
	ses.scheduledMiniBlocksSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := ses.scheduledMiniBlocksSyncer.SyncPendingMiniBlocks(mbHeaders, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	return ses.scheduledMiniBlocksSyncer.GetMiniBlocks()
}

func (ses *startInEpochWithScheduledDataSyncer) getRootHashToSync(notarizedShardHeader data.ShardHeaderHandler) []byte {
	if ses.scheduledEnableEpoch > notarizedShardHeader.GetEpoch() {
		return notarizedShardHeader.GetRootHash()
	}

	additionalData := notarizedShardHeader.GetAdditionalData()
	if additionalData != nil {
		return additionalData.GetScheduledRootHash()
	}

	return notarizedShardHeader.GetRootHash()
}

func (ses *startInEpochWithScheduledDataSyncer) prepareScheduledSCRs(
	prevHeader data.HeaderHandler,
	header data.HeaderHandler,
	miniBlocks map[string]*block.MiniBlock,
) error {
	// 1. get from the previous block all scheduled transactions hashes
	// 2. sync all transactions from notarized block
	// 3. for each synced transaction that is a smart contract result, check if prev transaction is in the set from 1.
	//     if found in the set, then it is a scheduled SCR, so it needs to be saved with the notarized header sch root hash

	// 1.
	scheduledTxHashes, err := ses.getScheduledTransactionHashes(prevHeader)
	if err != nil {
		return err
	}

	// 2.
	allTxs, err := ses.getAllTransactionsForMiniBlocks(miniBlocks, header.GetEpoch())
	if err != nil {
		return err
	}

	// 3.1
	scheduledSCRs, err := ses.filterScheduledSCRs(scheduledTxHashes, allTxs)
	if err != nil {
		return err
	}

	// 3.2
	return ses.saveScheduledSCRs(scheduledSCRs, header.GetAdditionalData().GetScheduledRootHash(), header.GetPrevHash())
}

func (ses *startInEpochWithScheduledDataSyncer) filterScheduledSCRs(
	scheduledTxHashes map[string]struct{},
	allTxs map[string]data.TransactionHandler,
) (map[string]data.TransactionHandler, error) {

	return nil, nil
}

func (ses *startInEpochWithScheduledDataSyncer) saveScheduledSCRs(
	scheduledSCRs map[string]data.TransactionHandler,
	scheduledRootHash []byte,
	headerHash []byte,
) error {
	//// mapScheduledSCRs  map[block.Type][]data.TransactionHandler

	return nil
}

func (ses *startInEpochWithScheduledDataSyncer) getAllTransactionsForMiniBlocks(miniBlocks map[string]*block.MiniBlock, epoch uint32) (map[string]data.TransactionHandler, error) {
	ctx := context.Background()
	err := ses.txSyncer.SyncTransactionsFor(miniBlocks, epoch, ctx)
	if err != nil {
		return nil, err
	}

	return ses.txSyncer.GetTransactions()
}

func (ses *startInEpochWithScheduledDataSyncer) getScheduledMiniBlockHeaders(header data.HeaderHandler) []data.MiniBlockHeaderHandler {
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	schMiniBlockHashes := make([]data.MiniBlockHeaderHandler, 0)

	for i, mbh := range miniBlockHeaders {
		reserved := mbh.GetReserved()
		if len(reserved) > 0 && reserved[0] == byte(block.ScheduledBlock) {
			schMiniBlockHashes = append(schMiniBlockHashes, miniBlockHeaders[i])
		}
	}

	return schMiniBlockHashes
}

func (ses *startInEpochWithScheduledDataSyncer) getScheduledTransactionHashes(header data.HeaderHandler) (map[string]struct{}, error) {
	miniBlockHeaders := ses.getScheduledMiniBlockHeaders(header)

	miniBlocks, err := ses.getRequiredMiniBlocksByMbHeader(miniBlockHeaders)
	if err != nil {
		return nil, err
	}

	scheduledTxs := make(map[string]struct{})
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			scheduledTxs[string(txHash)] = struct{}{}
		}
	}

	return scheduledTxs, nil
}
