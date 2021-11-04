package types

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)
type ScheduledDataSyncerCreateArgs struct {
	ScheduledTxsHandler  process.ScheduledTxsExecutionHandler
	HeadersSyncer        epochStart.HeadersByHashSyncer
	MiniBlocksSyncer     epochStart.PendingMiniBlocksSyncHandler
	TxSyncer             update.TransactionsSyncHandler
	ScheduledEnableEpoch uint32
}

type ScheduledDataSyncerCreator interface {
	Create(args *ScheduledDataSyncerCreateArgs) (ScheduledDataSyncer, error)
	IsInterfaceNil() bool
}

type ScheduledDataSyncer interface {
	UpdateSyncDataIfNeeded(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error)
	GetRootHashToSync(notarizedShardHeader data.ShardHeaderHandler) []byte
	IsInterfaceNil() bool
}
