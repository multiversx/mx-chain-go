package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
)

// KeysHandlerStub -
type KeysHandlerStub struct {
	GetHandledPrivateKeyCalled                   func(pkBytes []byte) crypto.PrivateKey
	GetP2PIdentityCalled                         func(pkBytes []byte) ([]byte, core.PeerID, error)
	IsKeyManagedByCurrentNodeCalled              func(pkBytes []byte) bool
	IncrementRoundsWithoutReceivedMessagesCalled func(pkBytes []byte)
	GetAssociatedPidCalled                       func(pkBytes []byte) core.PeerID
	IsOriginalPublicKeyOfTheNodeCalled           func(pkBytes []byte) bool
	ResetRoundsWithoutReceivedMessagesCalled     func(pkBytes []byte, pid core.PeerID)
}

// GetHandledPrivateKey -
func (stub *KeysHandlerStub) GetHandledPrivateKey(pkBytes []byte) crypto.PrivateKey {
	if stub.GetHandledPrivateKeyCalled != nil {
		return stub.GetHandledPrivateKeyCalled(pkBytes)
	}

	return &cryptoMocks.PrivateKeyStub{}
}

// GetP2PIdentity -
func (stub *KeysHandlerStub) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	if stub.GetP2PIdentityCalled != nil {
		return stub.GetP2PIdentityCalled(pkBytes)
	}

	return make([]byte, 0), "", nil
}

// IsKeyManagedByCurrentNode -
func (stub *KeysHandlerStub) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	if stub.IsKeyManagedByCurrentNodeCalled != nil {
		return stub.IsKeyManagedByCurrentNodeCalled(pkBytes)
	}

	return false
}

// IncrementRoundsWithoutReceivedMessages -
func (stub *KeysHandlerStub) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
	if stub.IncrementRoundsWithoutReceivedMessagesCalled != nil {
		stub.IncrementRoundsWithoutReceivedMessagesCalled(pkBytes)
	}
}

// GetAssociatedPid -
func (stub *KeysHandlerStub) GetAssociatedPid(pkBytes []byte) core.PeerID {
	if stub.GetAssociatedPidCalled != nil {
		return stub.GetAssociatedPidCalled(pkBytes)
	}

	return ""
}

// IsOriginalPublicKeyOfTheNode -
func (stub *KeysHandlerStub) IsOriginalPublicKeyOfTheNode(pkBytes []byte) bool {
	if stub.IsOriginalPublicKeyOfTheNodeCalled != nil {
		return stub.IsOriginalPublicKeyOfTheNodeCalled(pkBytes)
	}

	return true
}

// ResetRoundsWithoutReceivedMessages -
func (stub *KeysHandlerStub) ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID) {
	if stub.ResetRoundsWithoutReceivedMessagesCalled != nil {
		stub.ResetRoundsWithoutReceivedMessagesCalled(pkBytes, pid)
	}
}

// IsInterfaceNil -
func (stub *KeysHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
