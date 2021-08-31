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
	"github.com/ElrondNetwork/elrond-go/storage"
	updateSync "github.com/ElrondNetwork/elrond-go/update/sync"
)

type startInEpochWithScheduled struct {
	headersSyncer        epochStart.HeadersByHashSyncer
	miniBlocksSyncer     epochStart.PendingMiniBlocksSyncHandler
	scheduledEnableEpoch uint32
}

func NewStartInEpochWithScheduledSyncer(
	miniBlocksPool storage.Cacher,
	headersPool dataRetriever.HeadersPool,
	marshaller marshal.Marshalizer,
	requestHandler process.RequestHandler,
	scheduledEnableEpoch uint32,
) (*startInEpochWithScheduled, error) {
	syncMiniBlocksArgs := updateSync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          miniBlocksPool,
		Marshalizer:    marshaller,
		RequestHandler: requestHandler,
	}
	miniBlocksSyncer, err := updateSync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
	if err != nil {
		return nil, err
	}

	syncMissingHeadersArgs := updateSync.ArgsNewMissingHeadersByHashSyncer{
		Storage:        disabled.CreateMemUnit(),
		Cache:          headersPool,
		Marshalizer:    marshaller,
		RequestHandler: requestHandler,
	}

	headersSyncer, err := updateSync.NewMissingheadersByHashSyncer(syncMissingHeadersArgs)
	if err != nil {
		return nil, err
	}

	return &startInEpochWithScheduled{
		miniBlocksSyncer:     miniBlocksSyncer,
		headersSyncer:        headersSyncer,
		scheduledEnableEpoch: scheduledEnableEpoch,
	}, nil
}

func (ses *startInEpochWithScheduled) getShardHeaderAndPendingMiniBlocksToBeProcessed(
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

	updatedPendingMiniBlocks, err := ses.getPendingMiniBlocks(
		notarizedShardHeader,
		pendingMiniBlocks,
	)
	if err != nil {
		return nil, nil, err
	}

	return headerToBeProcessed, updatedPendingMiniBlocks, nil
}

func (ses *startInEpochWithScheduled) getRequiredHeaderByHash(notarizedShardHeader data.ShardHeaderHandler, ) (data.ShardHeaderHandler, error) {
	shardIDs := []uint32{
		notarizedShardHeader.GetShardID(),
	}
	hashesToRequest := [][]byte{
		notarizedShardHeader.GetPrevHash(),
	}

	ses.headersSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := ses.headersSyncer.SyncMissingHeadersByHash(shardIDs, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	headers, err := ses.headersSyncer.GetHeaders()
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

func (ses *startInEpochWithScheduled) getPendingMiniBlocks(
	notarizedShardHeader data.ShardHeaderHandler,
	pendingMiniBlocks map[string]*block.MiniBlock,
) (map[string]*block.MiniBlock, error) {
	previousPendingMiniBlocks := copyPendingMiniBlocksMap(pendingMiniBlocks)
	processedMiniBlockHeaders := notarizedShardHeader.GetMiniBlockHeaderHandlers()
	ownShardID := notarizedShardHeader.GetShardID()
	previousPendingMbHeaders := make([]data.MiniBlockHeaderHandler, 0)

	for i, mbHeader := range processedMiniBlockHeaders {
		if mbHeader.GetReceiverShardID() != ownShardID {
			continue
		}
		if mbHeader.GetSenderShardID() == ownShardID {
			continue
		}
		previousPendingMbHeaders = append(previousPendingMbHeaders, processedMiniBlockHeaders[i])
	}

	processedPendingMbs, err := ses.getRequiredMiniBlocksByMbHeader(previousPendingMbHeaders)
	if err != nil {
		return nil, err
	}

	for i := range processedPendingMbs {
		previousPendingMiniBlocks[i] = processedPendingMbs[i]
	}

	return previousPendingMiniBlocks, nil
}

func (ses *startInEpochWithScheduled) getRequiredMiniBlocksByMbHeader(
	mbHeaders []data.MiniBlockHeaderHandler,
) (map[string]*block.MiniBlock, error) {
	ses.miniBlocksSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := ses.miniBlocksSyncer.SyncPendingMiniBlocks(mbHeaders, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	return ses.miniBlocksSyncer.GetMiniBlocks()
}

func (ses *startInEpochWithScheduled) getRootHashToSync(notarizedShardHeader data.ShardHeaderHandler) []byte {
	if ses.scheduledEnableEpoch > notarizedShardHeader.GetEpoch() {
		return notarizedShardHeader.GetRootHash()
	}

	additionalData := notarizedShardHeader.GetAdditionalData()
	if additionalData != nil {
		return additionalData.GetScheduledRootHash()
	}

	return notarizedShardHeader.GetRootHash()
}
