package mock

import "github.com/ElrondNetwork/elrond-go/core/indexer/workItems"

// DispatcherMock -
type DispatcherMock struct {
	StartIndexDataCalled func()
	CloseCalled          func() error
	AddCalled            func(item workItems.WorkItemHandler)
}

// StartIndexData -
func (dm *DispatcherMock) StartIndexData() {
	if dm.StartIndexDataCalled != nil {
		dm.StartIndexDataCalled()
	}
}

// Close -
func (dm *DispatcherMock) Close() error {
	if dm.CloseCalled != nil {
		return dm.CloseCalled()
	}
	return nil
}

// Add -
func (dm *DispatcherMock) Add(item workItems.WorkItemHandler) {
	if dm.AddCalled != nil {
		dm.AddCalled(item)
	}
}
