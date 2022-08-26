package guardianMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data/guardians"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// GuardedAccountHandlerStub -
type GuardedAccountHandlerStub struct {
	GetActiveGuardianCalled      func(handler vmcommon.UserAccountHandler) ([]byte, error)
	SetGuardianCalled            func(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte) error
	HasPendingGuardianCalled     func(uah state.UserAccountHandler) bool
	HasActiveGuardianCalled      func(uah state.UserAccountHandler) bool
	CleanOtherThanActiveCalled   func(uah vmcommon.UserAccountHandler)
	GetConfiguredGuardiansCalled func(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error)
}

// GetActiveGuardian -
func (gahs *GuardedAccountHandlerStub) GetActiveGuardian(handler vmcommon.UserAccountHandler) ([]byte, error) {
	if gahs.GetActiveGuardianCalled != nil {
		return gahs.GetActiveGuardianCalled(handler)
	}
	return nil, nil
}

// HasActiveGuardian -
func (gahs *GuardedAccountHandlerStub) HasActiveGuardian(uah state.UserAccountHandler) bool {
	if gahs.HasActiveGuardianCalled != nil {
		return gahs.HasActiveGuardianCalled(uah)
	}
	return false
}

// HasPendingGuardian -
func (gahs *GuardedAccountHandlerStub) HasPendingGuardian(uah state.UserAccountHandler) bool {
	if gahs.HasPendingGuardianCalled != nil {
		return gahs.HasPendingGuardianCalled(uah)
	}
	return false
}

// SetGuardian -
func (gahs *GuardedAccountHandlerStub) SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte) error {
	if gahs.SetGuardianCalled != nil {
		return gahs.SetGuardianCalled(uah, guardianAddress, txGuardianAddress)
	}
	return nil
}

// CleanOtherThanActive -
func (gahs *GuardedAccountHandlerStub) CleanOtherThanActive(uah vmcommon.UserAccountHandler) {
	if gahs.CleanOtherThanActiveCalled != nil {
		gahs.CleanOtherThanActiveCalled(uah)
	}
}

// GetConfiguredGuardians -
func (gahs *GuardedAccountHandlerStub) GetConfiguredGuardians(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
	if gahs.GetConfiguredGuardiansCalled != nil {
		return gahs.GetConfiguredGuardiansCalled(uah)
	}
	return nil, nil, nil
}

// IsInterfaceNil -
func (gahs *GuardedAccountHandlerStub) IsInterfaceNil() bool {
	return gahs == nil
}
