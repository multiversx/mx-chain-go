package scheduledDataSyncer

import "github.com/multiversx/mx-chain-go/epochStart/bootstrap/types"

// ScheduledSyncerFactoryStub -
type ScheduledSyncerFactoryStub struct {
	CreateCalled func(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error)
}

// Create -
func (sdscs *ScheduledSyncerFactoryStub) Create(args *types.ScheduledDataSyncerCreateArgs) (types.ScheduledDataSyncer, error) {
	if sdscs.CreateCalled != nil {
		return sdscs.CreateCalled(args)
	}
	return &ScheduledSyncerStub{}, nil
}

// IsInterfaceNil -
func (sdscs *ScheduledSyncerFactoryStub) IsInterfaceNil() bool {
	return sdscs == nil
}
