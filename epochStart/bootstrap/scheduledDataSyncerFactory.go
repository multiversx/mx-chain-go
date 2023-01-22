package bootstrap

import (
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/types"
)

// ScheduledDataSyncerFactory is a factory for the scheduled data syncer
type ScheduledDataSyncerFactory struct{}

// NewScheduledDataSyncerFactory creates a factory instance
func NewScheduledDataSyncerFactory() *ScheduledDataSyncerFactory {
	return &ScheduledDataSyncerFactory{}
}

// Create creates a scheduled data syncer
func (sdsf *ScheduledDataSyncerFactory) Create(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
	return newStartInEpochShardHeaderDataSyncerWithScheduled(
		args.ScheduledTxsHandler, args.HeadersSyncer, args.MiniBlocksSyncer, args.TxSyncer, args.ScheduledEnableEpoch)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (sdsf *ScheduledDataSyncerFactory) IsInterfaceNil() bool {
	return sdsf == nil
}
