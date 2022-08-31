package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// MessengerStub -
type MessengerStub struct {
	IDCalled                       func() core.PeerID
	BroadcastCalled                func(topic string, buff []byte)
	BroadcastUsingPrivateKeyCalled func(topic string, buff []byte, pid core.PeerID, skBytes []byte)
	SignCalled                     func(payload []byte) ([]byte, error)
	SignUsingPrivateKeyCalled      func(skBytes []byte, payload []byte) ([]byte, error)
	VerifyCalled                   func(payload []byte, pid core.PeerID, signature []byte) error
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	if ms.BroadcastCalled != nil {
		ms.BroadcastCalled(topic, buff)
	}
}

// BroadcastUsingPrivateKey -
func (ms *MessengerStub) BroadcastUsingPrivateKey(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
	if ms.BroadcastUsingPrivateKeyCalled != nil {
		ms.BroadcastUsingPrivateKeyCalled(topic, buff, pid, skBytes)
	}
}

// Sign -
func (ms *MessengerStub) Sign(payload []byte) ([]byte, error) {
	if ms.SignCalled != nil {
		return ms.SignCalled(payload)
	}

	return make([]byte, 0), nil
}

// SignUsingPrivateKey -
func (ms *MessengerStub) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	if ms.SignUsingPrivateKeyCalled != nil {
		return ms.SignUsingPrivateKeyCalled(skBytes, payload)
	}

	return make([]byte, 0), nil
}

// Verify -
func (ms *MessengerStub) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	if ms.VerifyCalled != nil {
		return ms.VerifyCalled(payload, pid, signature)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
