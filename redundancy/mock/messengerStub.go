package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// MessengerStub -
type MessengerStub struct {
	IDCalled func() core.PeerID
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
