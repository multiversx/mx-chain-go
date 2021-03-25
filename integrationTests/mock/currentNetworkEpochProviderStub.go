package mock

import "github.com/ElrondNetwork/elrond-go/data"

// CurrentNetworkEpochProviderStub -
type CurrentNetworkEpochProviderStub struct {
	SetNetworkEpochAtBootstrapCalled func(epoch uint32)
	EpochIsActiveInNetworkCalled     func(epoch uint32) bool
	CurrentEpochCalled               func() uint32
	EpochChangedCalled               func(header data.HeaderHandler)
	EpochStartActionCalled           func(header data.HeaderHandler)
	EpochStartPrepareCalled          func(header data.HeaderHandler, body data.BodyHandler)
	NotifyOrderCalled                func() uint32
}

// EpochIsActiveInNetwork -
func (cneps *CurrentNetworkEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	if cneps.EpochIsActiveInNetworkCalled != nil {
		return cneps.EpochIsActiveInNetworkCalled(epoch)
	}

	return true
}

// EpochStartAction -
func (cneps *CurrentNetworkEpochProviderStub) EpochStartAction(header data.HeaderHandler) {
	if cneps.EpochStartActionCalled != nil {
		cneps.EpochStartActionCalled(header)
	}
}

// EpochStartPrepare does nothing
func (cneps *CurrentNetworkEpochProviderStub) EpochStartPrepare(header data.HeaderHandler, body data.BodyHandler) {
	if cneps.EpochStartPrepareCalled != nil {
		cneps.EpochStartPrepareCalled(header, body)
	}
}

// NotifyOrder will return the core.CurrentNetworkEpochProvider value
func (cneps *CurrentNetworkEpochProviderStub) NotifyOrder() uint32 {
	if cneps.NotifyOrderCalled != nil {
		return cneps.NotifyOrderCalled()
	}

	return 0
}

// IsInterfaceNil -
func (cneps *CurrentNetworkEpochProviderStub) IsInterfaceNil() bool {
	return cneps == nil
}
