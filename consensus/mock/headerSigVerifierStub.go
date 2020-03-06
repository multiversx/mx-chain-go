package mock

import "github.com/ElrondNetwork/elrond-go/data"

// HeaderSigVerifierStub -
type HeaderSigVerifierStub struct {
	VerifyRandSeedCaller func(header data.HeaderHandler) error
}

// VerifyRandSeed -
func (hsvm *HeaderSigVerifierStub) VerifyRandSeed(header data.HeaderHandler) error {
	if hsvm.VerifyRandSeedCaller != nil {
		return hsvm.VerifyRandSeedCaller(header)
	}

	return nil
}

// IsInterfaceNil -
func (hsvm *HeaderSigVerifierStub) IsInterfaceNil() bool {
	return hsvm == nil
}
