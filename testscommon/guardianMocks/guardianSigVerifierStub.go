package guardianMocks

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
)

// GuardianSigVerifierStub -
type GuardianSigVerifierStub struct {
	VerifyGuardianSignatureCalled func(inTx process.InterceptedTransactionHandler) error
	HasPendingGuardianCalled      func(uah state.UserAccountHandler) bool
}

// VerifyGuardianSignature -
func (gsvs *GuardianSigVerifierStub) VerifyGuardianSignature(inTx process.InterceptedTransactionHandler) error {
	if gsvs.VerifyGuardianSignatureCalled != nil {
		return gsvs.VerifyGuardianSignatureCalled(inTx)
	}
	return nil
}

// HasPendingGuardian -
func (gsvs *GuardianSigVerifierStub) HasPendingGuardian(uah state.UserAccountHandler) bool {
	if gsvs.HasPendingGuardianCalled != nil {
		return gsvs.HasPendingGuardianCalled(uah)
	}
	return false
}

// IsInterfaceNil -
func (gsvs *GuardianSigVerifierStub) IsInterfaceNil() bool {
	return gsvs == nil
}
