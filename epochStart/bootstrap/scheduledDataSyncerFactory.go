package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/types"
)

type ScheduledDataSyncerFactory struct {}

func NewScheduledDataSyncerFactory() *ScheduledDataSyncerFactory {
	return &ScheduledDataSyncerFactory{}
}

func (sdsf *ScheduledDataSyncerFactory) Create(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
	return newStartInEpochShardHeaderDataSyncerWithScheduled(
		args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, args.ScheduledEnableEpoch)
}

func (sdsf *ScheduledDataSyncerFactory) IsInterfaceNil() bool {
	return sdsf == nil
}
