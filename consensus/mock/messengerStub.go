package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// MessengerStub -
type MessengerStub struct {
	BroadcastWithSkCalled func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastCalled       func(topic string, buff []byte)
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// BroadcastWithSk -
func (ms *MessengerStub) BroadcastWithSk(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if ms.BroadcastWithSkCalled != nil {
		ms.BroadcastWithSkCalled(topic, buff, pid, skBytes)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
