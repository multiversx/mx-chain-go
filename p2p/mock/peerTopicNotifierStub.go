package mock

import "github.com/multiversx/mx-chain-core-go/core"

// PeerTopicNotifierStub -
type PeerTopicNotifierStub struct {
	NewPeerFoundCalled func(pid core.PeerID, topic string)
}

// NewPeerFound -
func (stub *PeerTopicNotifierStub) NewPeerFound(pid core.PeerID, topic string) {
	if stub.NewPeerFoundCalled != nil {
		stub.NewPeerFoundCalled(pid, topic)
	}
}

// IsInterfaceNil -
func (stub *PeerTopicNotifierStub) IsInterfaceNil() bool {
	return stub == nil
}
