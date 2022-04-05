package guardianMocks

import "github.com/ElrondNetwork/elrond-go-core/data"

// GuardianCheckerStub -
type GuardianCheckerStub struct {
	GetActiveGuardianCalled func(handler data.UserAccountHandler) ([]byte, error)
}

// GetActiveGuardian -
func (gcs *GuardianCheckerStub) GetActiveGuardian(handler data.UserAccountHandler) ([]byte, error) {
	if gcs.GetActiveGuardianCalled != nil {
		return gcs.GetActiveGuardianCalled(handler)
	}

	return nil, nil
}

// IsInterfaceNil -
func (gcs *GuardianCheckerStub) IsInterfaceNil() bool {
	return gcs == nil
}
