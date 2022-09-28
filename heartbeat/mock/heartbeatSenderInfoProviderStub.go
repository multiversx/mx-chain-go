package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// HeartbeatSenderInfoProviderStub -
type HeartbeatSenderInfoProviderStub struct {
	GetSenderInfoCalled func() (string, core.P2PPeerSubType, error)
}

// GetSenderInfo -
func (stub *HeartbeatSenderInfoProviderStub) GetSenderInfo() (string, core.P2PPeerSubType, error) {
	if stub.GetSenderInfoCalled != nil {
		return stub.GetSenderInfoCalled()
	}

	return "", 0, nil
}

// IsInterfaceNil -
func (stub *HeartbeatSenderInfoProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
