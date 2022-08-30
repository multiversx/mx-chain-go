package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// MessengerStub -
type MessengerStub struct {
	BroadcastWithPrivateKeyCalled func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastCalled               func(topic string, buff []byte)
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// BroadcastWithPrivateKey -
func (ms *MessengerStub) BroadcastWithPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if ms.BroadcastWithPrivateKeyCalled != nil {
		ms.BroadcastWithPrivateKeyCalled(topic, buff, pid, skBytes)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
