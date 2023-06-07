package guardianMocks

import (
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// GuardedAccountHandlerStub -
type GuardedAccountHandlerStub struct {
	GetActiveGuardianCalled      func(handler vmcommon.UserAccountHandler) ([]byte, error)
	SetGuardianCalled            func(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte, guardianServiceUID []byte) error
	HasPendingGuardianCalled     func(uah common.UserAccountHandler) bool
	HasActiveGuardianCalled      func(uah common.UserAccountHandler) bool
	CleanOtherThanActiveCalled   func(uah vmcommon.UserAccountHandler)
	GetConfiguredGuardiansCalled func(uah common.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error)
}

// GetActiveGuardian -
func (gahs *GuardedAccountHandlerStub) GetActiveGuardian(handler vmcommon.UserAccountHandler) ([]byte, error) {
	if gahs.GetActiveGuardianCalled != nil {
		return gahs.GetActiveGuardianCalled(handler)
	}
	return nil, nil
}

// HasActiveGuardian -
func (gahs *GuardedAccountHandlerStub) HasActiveGuardian(uah common.UserAccountHandler) bool {
	if gahs.HasActiveGuardianCalled != nil {
		return gahs.HasActiveGuardianCalled(uah)
	}
	return false
}

// HasPendingGuardian -
func (gahs *GuardedAccountHandlerStub) HasPendingGuardian(uah common.UserAccountHandler) bool {
	if gahs.HasPendingGuardianCalled != nil {
		return gahs.HasPendingGuardianCalled(uah)
	}
	return false
}

// SetGuardian -
func (gahs *GuardedAccountHandlerStub) SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte, guardianServiceUID []byte) error {
	if gahs.SetGuardianCalled != nil {
		return gahs.SetGuardianCalled(uah, guardianAddress, txGuardianAddress, guardianServiceUID)
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
func (gahs *GuardedAccountHandlerStub) GetConfiguredGuardians(uah common.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
	if gahs.GetConfiguredGuardiansCalled != nil {
		return gahs.GetConfiguredGuardiansCalled(uah)
	}
	return nil, nil, nil
}

// IsInterfaceNil -
func (gahs *GuardedAccountHandlerStub) IsInterfaceNil() bool {
	return gahs == nil
}
