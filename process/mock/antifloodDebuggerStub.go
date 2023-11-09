package mock

import "github.com/multiversx/mx-chain-core-go/core"

// AntifloodDebuggerStub -
type AntifloodDebuggerStub struct {
	AddDataCalled func(pid core.PeerID, topic string, numRejected uint32, sizeRejected uint64, sequence []byte, isBlacklisted bool)
	CloseCalled   func() error
}

// AddData -
func (ads *AntifloodDebuggerStub) AddData(pid core.PeerID, topic string, numRejected uint32, sizeRejected uint64, sequence []byte, isBlacklisted bool) {
	if ads.AddDataCalled != nil {
		ads.AddDataCalled(pid, topic, numRejected, sizeRejected, sequence, isBlacklisted)
	}
}

// Close -
func (ads *AntifloodDebuggerStub) Close() error {
	if ads.CloseCalled != nil {
		return ads.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (ads *AntifloodDebuggerStub) IsInterfaceNil() bool {
	return ads == nil
}
