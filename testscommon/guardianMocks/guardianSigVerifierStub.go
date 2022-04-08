package guardianMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// GuardianSigVerifierStub -
type GuardianSigVerifierStub struct {
	VerifyGuardianSignatureCalled func(account data.UserAccountHandler, inTx process.InterceptedTransactionHandler) error
}

// VerifyGuardianSignature -
func (gsvs *GuardianSigVerifierStub) VerifyGuardianSignature(account data.UserAccountHandler, inTx process.InterceptedTransactionHandler) error {
	if gsvs.VerifyGuardianSignatureCalled != nil {
		return gsvs.VerifyGuardianSignatureCalled(account, inTx)
	}
	return nil
}

// IsInterfaceNil -
func (gsvs *GuardianSigVerifierStub) IsInterfaceNil() bool {
	return gsvs == nil
}
