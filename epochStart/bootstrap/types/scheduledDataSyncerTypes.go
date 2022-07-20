package types

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ScheduledDataSyncerCreateArgs holds the arguments to create a scheduled data syncer factory
type ScheduledDataSyncerCreateArgs struct {
	ScheduledTxsHandler  process.ScheduledTxsExecutionHandler
	HeadersSyncer        epochStart.HeadersByHashSyncer
	MiniBlocksSyncer     epochStart.MiniBlocksSyncHandler
	TxSyncer             update.TransactionsSyncHandler
	ScheduledEnableEpoch uint32
}

// ScheduledDataSyncerCreator is the interface implemented by the scheduled data syncer factory allowing to create
// scheduled data syncer instances
type ScheduledDataSyncerCreator interface {
	Create(args *ScheduledDataSyncerCreateArgs) (ScheduledDataSyncer, error)
	IsInterfaceNil() bool
}

// ScheduledDataSyncer interface allows to synchronize the correct headers and root hash with or without scheduled sc calls feature activated.
type ScheduledDataSyncer interface {
	UpdateSyncDataIfNeeded(notarizedShardHeader data.ShardHeaderHandler) (data.ShardHeaderHandler, map[string]data.HeaderHandler, error)
	GetRootHashToSync(notarizedShardHeader data.ShardHeaderHandler) []byte
	IsInterfaceNil() bool
}
