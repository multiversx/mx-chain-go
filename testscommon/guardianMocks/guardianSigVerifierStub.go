package guardianMocks

import (
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// GuardianSigVerifierStub -
type GuardianSigVerifierStub struct {
	VerifyGuardianSignatureCalled func(account vmcommon.UserAccountHandler, inTx process.InterceptedTransactionHandler) error
}

// VerifyGuardianSignature -
func (gsvs *GuardianSigVerifierStub) VerifyGuardianSignature(account vmcommon.UserAccountHandler, inTx process.InterceptedTransactionHandler) error {
	if gsvs.VerifyGuardianSignatureCalled != nil {
		return gsvs.VerifyGuardianSignatureCalled(account, inTx)
	}
	return nil
}

// IsInterfaceNil -
func (gsvs *GuardianSigVerifierStub) IsInterfaceNil() bool {
	return gsvs == nil
}
