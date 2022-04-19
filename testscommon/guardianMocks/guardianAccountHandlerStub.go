package guardianMocks

import "github.com/ElrondNetwork/elrond-go-core/data"

// GuardedAccountHandlerStub -
type GuardedAccountHandlerStub struct {
	GetActiveGuardianCalled func(handler data.UserAccountHandler) ([]byte, error)
	SetGuardianCalled       func(uah data.UserAccountHandler, guardianAddress []byte) error
}

// GetActiveGuardian -
func (gahs *GuardedAccountHandlerStub) GetActiveGuardian(handler data.UserAccountHandler) ([]byte, error) {
	if gahs.GetActiveGuardianCalled != nil {
		return gahs.GetActiveGuardianCalled(handler)
	}
	return nil, nil
}

// SetGuardian -
func (gahs *GuardedAccountHandlerStub) SetGuardian(uah data.UserAccountHandler, guardianAddress []byte) error {
	if gahs.SetGuardianCalled != nil {
		return gahs.SetGuardianCalled(uah, guardianAddress)
	}
	return nil
}

// IsInterfaceNil -
func (gahs *GuardedAccountHandlerStub) IsInterfaceNil() bool {
	return gahs == nil
}
