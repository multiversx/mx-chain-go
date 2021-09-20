package bootstrap

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	factoryDisabled "github.com/ElrondNetwork/elrond-go/factory/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	updateSync "github.com/ElrondNetwork/elrond-go/update/sync"
)

type startInEpochWithScheduledDataSyncer struct {
	scheduledTxsHandler       process.ScheduledTxsExecutionHandler
	scheduledHeadersSyncer    epochStart.HeadersByHashSyncer
	scheduledMiniBlocksSyncer epochStart.PendingMiniBlocksSyncHandler
	txSyncer                  update.TransactionsSyncHandler
	scheduledEnableEpoch      uint32
}

func NewStartInEpochShardHeaderDataSyncerWithScheduled(
	scheduledSCRsStorer storage.Storer,
	dataPool dataRetriever.PoolsHolder,
	marshaller marshal.Marshalizer,
	requestHandler process.RequestHandler,
	scheduledEnableEpoch uint32,
) (*startInEpochWithScheduledDataSyncer, error) {

	if check.IfNil(scheduledSCRsStorer) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(dataPool) {
		return nil, epochStart.ErrNilDataPoolsHolder
	}
	if check.IfNil(marshaller) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(requestHandler) {
		return nil, epochStart.ErrNilRequestHandler
	}

	scheduledTxsHandler, err := preprocess.NewScheduledTxsExecution(
		&factoryDisabled.TxProcessor{},
		&factoryDisabled.TxCoordinator{},
		scheduledSCRsStorer,
		marshaller,
	)
	if err != nil {
		return nil, err
	}

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
		scheduledTxsHandler:       scheduledTxsHandler,
		scheduledMiniBlocksSyncer: miniBlocksSyncer,
		scheduledHeadersSyncer:    headersSyncer,
		txSyncer:                  txSyncer,
		scheduledEnableEpoch:      scheduledEnableEpoch,
	}, nil
}

func (ses *startInEpochWithScheduledDataSyncer) updateSyncDataIfNeeded(
	notarizedShardHeader data.ShardHeaderHandler,
) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
	if ses.scheduledEnableEpoch > notarizedShardHeader.GetEpoch() {
		return notarizedShardHeader, nil, nil
	}

	headerToBeProcessed, headers, err := ses.getRequiredHeaderByHash(notarizedShardHeader)
	if err != nil {
		return nil, nil, err
	}

	allMiniBlocks, err := ses.getMiniBlocks(notarizedShardHeader)
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

	return headerToBeProcessed, headers, nil
}

func (ses *startInEpochWithScheduledDataSyncer) getRequiredHeaderByHash(
	notarizedShardHeader data.ShardHeaderHandler,
) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error) {
	shardIDs, hashesToRequest := getShardIdAndHashesForIncludedMetaBlocks(notarizedShardHeader)

	shardIDs = append(shardIDs, notarizedShardHeader.GetShardID())
	hashesToRequest = append(hashesToRequest, notarizedShardHeader.GetPrevHash())

	headers, err := ses.syncHeaders(shardIDs, hashesToRequest)
	if err != nil {
		return nil, nil, err
	}

	headerToBeProcessed, ok := headers[string(notarizedShardHeader.GetPrevHash())].(data.ShardHeaderHandler)
	if !ok {
		return nil, nil, epochStart.ErrMissingHeader
	}

	shardIDs, hashesToRequest = getShardIdAndHashesForIncludedMetaBlocks(headerToBeProcessed)
	prevHeaders, err := ses.syncHeaders(shardIDs, hashesToRequest)
	if err != nil {
		return nil, nil, err
	}

	for hash, hdr := range prevHeaders {
		headers[hash] = hdr
	}

	// get also the previous meta block to the first used meta block, which should should be completed
	if len(prevHeaders) > 0 {
		var prevPrevHeaders map[string]data.HeaderHandler
		shardIDs = []uint32{core.MetachainShardId}
		hashesToRequest = [][]byte{prevHeaders[string(hashesToRequest[0])].GetPrevHash()}
		prevPrevHeaders, err = ses.syncHeaders(shardIDs, hashesToRequest)
		if err != nil {
			return nil, nil, err
		}
		for hash, hdr := range prevPrevHeaders {
			headers[hash] = hdr
		}
	}

	return headerToBeProcessed, headers, nil
}

func getShardIdAndHashesForIncludedMetaBlocks(notarizedShardHeader data.ShardHeaderHandler) ([]uint32, [][]byte) {
	shardIDs := make([]uint32, 0)
	hashesToRequest := make([][]byte, 0)

	// if there were processed metaBlocks in the notarized header, these need to be synced so they can get again processed
	metaBlockHashes := notarizedShardHeader.GetMetaBlockHashes()
	for i := range metaBlockHashes {
		shardIDs = append(shardIDs, core.MetachainShardId)
		hashesToRequest = append(hashesToRequest, metaBlockHashes[i])
	}
	return shardIDs, hashesToRequest
}

func (ses *startInEpochWithScheduledDataSyncer) syncHeaders(
	shardIDs []uint32,
	hashesToRequest [][]byte,
) (map[string]data.HeaderHandler, error) {
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
	return headers, nil
}

func (ses *startInEpochWithScheduledDataSyncer) getMiniBlocks(
	notarizedShardHeader data.ShardHeaderHandler,
) (map[string]*block.MiniBlock, error) {
	processedMiniBlockHeaders := notarizedShardHeader.GetMiniBlockHeaderHandlers()
	return ses.getRequiredMiniBlocksByMbHeader(processedMiniBlockHeaders)
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
	scheduledTxHashes, err := ses.getScheduledTransactionHashes(prevHeader)
	if err != nil {
		return err
	}

	allTxs, err := ses.getAllTransactionsForMiniBlocks(miniBlocks, header.GetEpoch())
	if err != nil {
		return err
	}

	scheduledSCRs, err := ses.filterScheduledSCRs(scheduledTxHashes, allTxs)
	if err != nil {
		return err
	}

	additionalData := header.GetAdditionalData()
	if additionalData != nil {
		ses.saveScheduledSCRs(scheduledSCRs, header.GetAdditionalData().GetScheduledRootHash(), header.GetPrevHash())
	}

	return nil
}

func (ses *startInEpochWithScheduledDataSyncer) filterScheduledSCRs(
	scheduledTxHashes map[string]struct{},
	allTxs map[string]data.TransactionHandler,
) (map[string]data.TransactionHandler, error) {

	scheduledSCRs := make(map[string]data.TransactionHandler)
	for txHash := range allTxs {
		scr, ok := allTxs[txHash].(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		_, isScheduledSCR := scheduledTxHashes[string(scr.PrevTxHash)]
		if isScheduledSCR {
			scheduledSCRs[txHash] = allTxs[txHash]
		}
	}

	return scheduledSCRs, nil
}

func (ses *startInEpochWithScheduledDataSyncer) saveScheduledSCRs(
	scheduledSCRs map[string]data.TransactionHandler,
	scheduledRootHash []byte,
	headerHash []byte,
) {
	if scheduledRootHash == nil {
		return
	}

	log.Debug("startInEpochWithScheduledDataSyncer.saveScheduledSCRs",
		"headerHash", headerHash,
		"scheduledRootHash", scheduledRootHash,
	)
	// prepare the scheduledSCRs in the form of map[block.Type][]data.TransactionHandler
	// the order should not matter, as the processing is done after sorting by scr hash
	mapScheduledSCRs := make(map[block.Type][]data.TransactionHandler)
	scheduledSCRsList := make([]data.TransactionHandler, 0, len(scheduledSCRs))
	for scrHash := range scheduledSCRs {
		scheduledSCRsList = append(scheduledSCRsList, scheduledSCRs[scrHash])
	}
	mapScheduledSCRs[block.TxBlock] = scheduledSCRsList
	ses.scheduledTxsHandler.SaveState(headerHash, scheduledRootHash, mapScheduledSCRs)
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
		if len(reserved) > 0 && reserved[0] == byte(block.Scheduled) {
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
