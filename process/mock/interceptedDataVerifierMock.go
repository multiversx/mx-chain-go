package mock

import (
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// InterceptedDataVerifierMock -
type InterceptedDataVerifierMock struct {
	VerifyCalled       func(interceptedData process.InterceptedData, topic string, broadcastMethod p2p.BroadcastMethod) error
	MarkVerifiedCalled func(interceptedData process.InterceptedData, broadcastMethod p2p.BroadcastMethod)
}

// Verify -
func (idv *InterceptedDataVerifierMock) Verify(interceptedData process.InterceptedData, topic string, broadcastMethod p2p.BroadcastMethod) error {
	if idv.VerifyCalled != nil {
		return idv.VerifyCalled(interceptedData, topic, broadcastMethod)
	}

	return nil
}

// MarkVerified -
func (idv *InterceptedDataVerifierMock) MarkVerified(interceptedData process.InterceptedData, broadcastMethod p2p.BroadcastMethod) {
	if idv.MarkVerifiedCalled != nil {
		idv.MarkVerifiedCalled(interceptedData, broadcastMethod)
	}
}

// IsInterfaceNil -
func (idv *InterceptedDataVerifierMock) IsInterfaceNil() bool {
	return idv == nil
}
