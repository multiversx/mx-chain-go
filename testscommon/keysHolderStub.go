package testscommon

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// KeysHolderStub -
type KeysHolderStub struct {
	AddVirtualPeerCalled                         func(privateKeyBytes []byte) error
	GetPrivateKeyCalled                          func(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentityCalled                         func(pkBytes []byte) ([]byte, core.PeerID, error)
	GetMachineIDCalled                           func(pkBytes []byte) (string, error)
	IncrementRoundsWithoutReceivedMessagesCalled func(pkBytes []byte) error
	ResetRoundsWithoutReceivedMessagesCalled     func(pkBytes []byte) error
	GetManagedKeysByCurrentNodeCalled            func() map[string]crypto.PrivateKey
	IsKeyManagedByCurrentNodeCalled              func(pkBytes []byte) bool
	IsKeyRegisteredCalled                        func(pkBytes []byte) bool
	IsPidManagedByCurrentNodeCalled              func(pid core.PeerID) bool
	IsKeyValidatorCalled                         func(pkBytes []byte) (bool, error)
	SetValidatorStateCalled                      func(pkBytes []byte, state bool) error
	GetNextPeerAuthenticationTimeCalled          func(pkBytes []byte) (time.Time, error)
	SetNextPeerAuthenticationTimeCalled          func(pkBytes []byte, nextTime time.Time) error
}

// AddVirtualPeer -
func (stub *KeysHolderStub) AddVirtualPeer(privateKeyBytes []byte) error {
	if stub.AddVirtualPeerCalled != nil {
		return stub.AddVirtualPeerCalled(privateKeyBytes)
	}
	return nil
}

// GetPrivateKey -
func (stub *KeysHolderStub) GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error) {
	if stub.GetPrivateKeyCalled != nil {
		return stub.GetPrivateKeyCalled(pkBytes)
	}
	return nil, nil
}

// GetP2PIdentity -
func (stub *KeysHolderStub) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	if stub.GetP2PIdentityCalled != nil {
		return stub.GetP2PIdentityCalled(pkBytes)
	}
	return nil, "", nil
}

// GetMachineID -
func (stub *KeysHolderStub) GetMachineID(pkBytes []byte) (string, error) {
	if stub.GetMachineIDCalled != nil {
		return stub.GetMachineIDCalled(pkBytes)
	}
	return "", nil
}

// IncrementRoundsWithoutReceivedMessages -
func (stub *KeysHolderStub) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) error {
	if stub.IncrementRoundsWithoutReceivedMessagesCalled != nil {
		return stub.IncrementRoundsWithoutReceivedMessagesCalled(pkBytes)
	}
	return nil
}

// ResetRoundsWithoutReceivedMessages -
func (stub *KeysHolderStub) ResetRoundsWithoutReceivedMessages(pkBytes []byte) error {
	if stub.ResetRoundsWithoutReceivedMessagesCalled != nil {
		return stub.ResetRoundsWithoutReceivedMessagesCalled(pkBytes)
	}
	return nil
}

// GetManagedKeysByCurrentNode -
func (stub *KeysHolderStub) GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey {
	if stub.GetManagedKeysByCurrentNodeCalled != nil {
		return stub.GetManagedKeysByCurrentNodeCalled()
	}
	return nil
}

// IsKeyManagedByCurrentNode -
func (stub *KeysHolderStub) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	if stub.IsKeyManagedByCurrentNodeCalled != nil {
		return stub.IsKeyManagedByCurrentNodeCalled(pkBytes)
	}
	return false
}

// IsKeyRegistered -
func (stub *KeysHolderStub) IsKeyRegistered(pkBytes []byte) bool {
	if stub.IsKeyRegisteredCalled != nil {
		return stub.IsKeyRegisteredCalled(pkBytes)
	}
	return false
}

// IsPidManagedByCurrentNode -
func (stub *KeysHolderStub) IsPidManagedByCurrentNode(pid core.PeerID) bool {
	if stub.IsPidManagedByCurrentNodeCalled != nil {
		return stub.IsPidManagedByCurrentNodeCalled(pid)
	}
	return false
}

// IsKeyValidator -
func (stub *KeysHolderStub) IsKeyValidator(pkBytes []byte) (bool, error) {
	if stub.IsKeyValidatorCalled != nil {
		return stub.IsKeyValidatorCalled(pkBytes)
	}
	return false, nil
}

// SetValidatorState -
func (stub *KeysHolderStub) SetValidatorState(pkBytes []byte, state bool) error {
	if stub.SetValidatorStateCalled != nil {
		return stub.SetValidatorStateCalled(pkBytes, state)
	}
	return nil
}

// GetNextPeerAuthenticationTime -
func (stub *KeysHolderStub) GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error) {
	if stub.GetNextPeerAuthenticationTimeCalled != nil {
		return stub.GetNextPeerAuthenticationTimeCalled(pkBytes)
	}
	return time.Time{}, nil
}

// SetNextPeerAuthenticationTime -
func (stub *KeysHolderStub) SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time) error {
	if stub.SetNextPeerAuthenticationTimeCalled != nil {
		return stub.SetNextPeerAuthenticationTimeCalled(pkBytes, nextTime)
	}
	return nil
}

// IsInterfaceNil -
func (stub *KeysHolderStub) IsInterfaceNil() bool {
	return stub == nil
}
