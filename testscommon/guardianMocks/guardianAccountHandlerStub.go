package guardianMocks

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// GuardedAccountHandlerStub -
type GuardedAccountHandlerStub struct {
	GetActiveGuardianCalled func(handler vmcommon.UserAccountHandler) ([]byte, error)
	SetGuardianCalled       func(uah vmcommon.UserAccountHandler, guardianAddress []byte) error
}

// GetActiveGuardian -
func (gahs *GuardedAccountHandlerStub) GetActiveGuardian(handler vmcommon.UserAccountHandler) ([]byte, error) {
	if gahs.GetActiveGuardianCalled != nil {
		return gahs.GetActiveGuardianCalled(handler)
	}
	return nil, nil
}

// SetGuardian -
func (gahs *GuardedAccountHandlerStub) SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte) error {
	if gahs.SetGuardianCalled != nil {
		return gahs.SetGuardianCalled(uah, guardianAddress)
	}
	return nil
}

// IsInterfaceNil -
func (gahs *GuardedAccountHandlerStub) IsInterfaceNil() bool {
	return gahs == nil
}
