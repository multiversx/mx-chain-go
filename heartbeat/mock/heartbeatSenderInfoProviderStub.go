package mock

import "github.com/multiversx/mx-chain-core-go/core"

// HeartbeatSenderInfoProviderStub -
type HeartbeatSenderInfoProviderStub struct {
	GetCurrentNodeTypeCalled func() (string, core.P2PPeerSubType, error)
}

// GetCurrentNodeType -
func (stub *HeartbeatSenderInfoProviderStub) GetCurrentNodeType() (string, core.P2PPeerSubType, error) {
	if stub.GetCurrentNodeTypeCalled != nil {
		return stub.GetCurrentNodeTypeCalled()
	}

	return "", 0, nil
}

// IsInterfaceNil -
func (stub *HeartbeatSenderInfoProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
