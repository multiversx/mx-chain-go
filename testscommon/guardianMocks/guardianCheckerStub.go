package guardianMocks

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// GuardianCheckerStub -
type GuardianCheckerStub struct {
	GetActiveGuardianCalled func(handler vmcommon.UserAccountHandler) ([]byte, error)
}

// GetActiveGuardian -
func (gcs *GuardianCheckerStub) GetActiveGuardian(handler vmcommon.UserAccountHandler) ([]byte, error) {
	if gcs.GetActiveGuardianCalled != nil {
		return gcs.GetActiveGuardianCalled(handler)
	}

	return nil, nil
}

// IsInterfaceNil -
func (gcs *GuardianCheckerStub) IsInterfaceNil() bool {
	return gcs == nil
}
