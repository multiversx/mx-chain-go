package mock

import "github.com/multiversx/mx-chain-core-go/core"

// MessengerStub -
type MessengerStub struct {
	BroadcastUsingPrivateKeyCalled func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	BroadcastCalled                func(topic string, buff []byte)
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// BroadcastUsingPrivateKey -
func (ms *MessengerStub) BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if ms.BroadcastUsingPrivateKeyCalled != nil {
		ms.BroadcastUsingPrivateKeyCalled(topic, buff, pid, skBytes)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
