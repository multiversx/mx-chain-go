package testscommon

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// ManagedPeersHolderStub -
type ManagedPeersHolderStub struct {
	AddManagedPeerCalled                         func(privateKeyBytes []byte) error
	GetPrivateKeyCalled                          func(pkBytes []byte) (crypto.PrivateKey, error)
	GetP2PIdentityCalled                         func(pkBytes []byte) ([]byte, core.PeerID, error)
	GetMachineIDCalled                           func(pkBytes []byte) (string, error)
	GetNameAndIdentityCalled                     func(pkBytes []byte) (string, string, error)
	IncrementRoundsWithoutReceivedMessagesCalled func(pkBytes []byte)
	ResetRoundsWithoutReceivedMessagesCalled     func(pkBytes []byte, pid core.PeerID)
	GetManagedKeysByCurrentNodeCalled            func() map[string]crypto.PrivateKey
	IsKeyManagedByCurrentNodeCalled              func(pkBytes []byte) bool
	IsKeyRegisteredCalled                        func(pkBytes []byte) bool
	IsPidManagedByCurrentNodeCalled              func(pid core.PeerID) bool
	IsKeyValidatorCalled                         func(pkBytes []byte) bool
	SetValidatorStateCalled                      func(pkBytes []byte, state bool)
	GetNextPeerAuthenticationTimeCalled          func(pkBytes []byte) (time.Time, error)
	SetNextPeerAuthenticationTimeCalled          func(pkBytes []byte, nextTime time.Time)
	IsMultiKeyModeCalled                         func() bool
	GetRedundancyStepInReasonCalled              func() string
}

// AddManagedPeer -
func (stub *ManagedPeersHolderStub) AddManagedPeer(privateKeyBytes []byte) error {
	if stub.AddManagedPeerCalled != nil {
		return stub.AddManagedPeerCalled(privateKeyBytes)
	}
	return nil
}

// GetPrivateKey -
func (stub *ManagedPeersHolderStub) GetPrivateKey(pkBytes []byte) (crypto.PrivateKey, error) {
	if stub.GetPrivateKeyCalled != nil {
		return stub.GetPrivateKeyCalled(pkBytes)
	}
	return nil, nil
}

// GetP2PIdentity -
func (stub *ManagedPeersHolderStub) GetP2PIdentity(pkBytes []byte) ([]byte, core.PeerID, error) {
	if stub.GetP2PIdentityCalled != nil {
		return stub.GetP2PIdentityCalled(pkBytes)
	}
	return nil, "", nil
}

// GetMachineID -
func (stub *ManagedPeersHolderStub) GetMachineID(pkBytes []byte) (string, error) {
	if stub.GetMachineIDCalled != nil {
		return stub.GetMachineIDCalled(pkBytes)
	}
	return "", nil
}

// GetNameAndIdentity -
func (stub *ManagedPeersHolderStub) GetNameAndIdentity(pkBytes []byte) (string, string, error) {
	if stub.GetNameAndIdentityCalled != nil {
		return stub.GetNameAndIdentityCalled(pkBytes)
	}
	return "", "", nil
}

// IncrementRoundsWithoutReceivedMessages -
func (stub *ManagedPeersHolderStub) IncrementRoundsWithoutReceivedMessages(pkBytes []byte) {
	if stub.IncrementRoundsWithoutReceivedMessagesCalled != nil {
		stub.IncrementRoundsWithoutReceivedMessagesCalled(pkBytes)
	}
}

// ResetRoundsWithoutReceivedMessages -
func (stub *ManagedPeersHolderStub) ResetRoundsWithoutReceivedMessages(pkBytes []byte, pid core.PeerID) {
	if stub.ResetRoundsWithoutReceivedMessagesCalled != nil {
		stub.ResetRoundsWithoutReceivedMessagesCalled(pkBytes, pid)
	}
}

// GetManagedKeysByCurrentNode -
func (stub *ManagedPeersHolderStub) GetManagedKeysByCurrentNode() map[string]crypto.PrivateKey {
	if stub.GetManagedKeysByCurrentNodeCalled != nil {
		return stub.GetManagedKeysByCurrentNodeCalled()
	}
	return nil
}

// IsKeyManagedByCurrentNode -
func (stub *ManagedPeersHolderStub) IsKeyManagedByCurrentNode(pkBytes []byte) bool {
	if stub.IsKeyManagedByCurrentNodeCalled != nil {
		return stub.IsKeyManagedByCurrentNodeCalled(pkBytes)
	}
	return false
}

// IsKeyRegistered -
func (stub *ManagedPeersHolderStub) IsKeyRegistered(pkBytes []byte) bool {
	if stub.IsKeyRegisteredCalled != nil {
		return stub.IsKeyRegisteredCalled(pkBytes)
	}
	return false
}

// IsPidManagedByCurrentNode -
func (stub *ManagedPeersHolderStub) IsPidManagedByCurrentNode(pid core.PeerID) bool {
	if stub.IsPidManagedByCurrentNodeCalled != nil {
		return stub.IsPidManagedByCurrentNodeCalled(pid)
	}
	return false
}

// IsKeyValidator -
func (stub *ManagedPeersHolderStub) IsKeyValidator(pkBytes []byte) bool {
	if stub.IsKeyValidatorCalled != nil {
		return stub.IsKeyValidatorCalled(pkBytes)
	}
	return false
}

// SetValidatorState -
func (stub *ManagedPeersHolderStub) SetValidatorState(pkBytes []byte, state bool) {
	if stub.SetValidatorStateCalled != nil {
		stub.SetValidatorStateCalled(pkBytes, state)
	}
}

// GetNextPeerAuthenticationTime -
func (stub *ManagedPeersHolderStub) GetNextPeerAuthenticationTime(pkBytes []byte) (time.Time, error) {
	if stub.GetNextPeerAuthenticationTimeCalled != nil {
		return stub.GetNextPeerAuthenticationTimeCalled(pkBytes)
	}
	return time.Time{}, nil
}

// SetNextPeerAuthenticationTime -
func (stub *ManagedPeersHolderStub) SetNextPeerAuthenticationTime(pkBytes []byte, nextTime time.Time) {
	if stub.SetNextPeerAuthenticationTimeCalled != nil {
		stub.SetNextPeerAuthenticationTimeCalled(pkBytes, nextTime)
	}
}

// IsMultiKeyMode -
func (stub *ManagedPeersHolderStub) IsMultiKeyMode() bool {
	if stub.IsMultiKeyModeCalled != nil {
		return stub.IsMultiKeyModeCalled()
	}
	return false
}

// GetRedundancyStepInReason -
func (stub *ManagedPeersHolderStub) GetRedundancyStepInReason() string {
	if stub.GetRedundancyStepInReasonCalled != nil {
		return stub.GetRedundancyStepInReasonCalled()
	}

	return ""
}

// IsInterfaceNil -
func (stub *ManagedPeersHolderStub) IsInterfaceNil() bool {
	return stub == nil
}
