package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderSigVerifierStub struct {
	VerifyRandSeedCaller func(header data.HeaderHandler) error
}

func (hsvm *HeaderSigVerifierStub) VerifyRandSeed(header data.HeaderHandler) error {
	if hsvm.VerifyRandSeedCaller != nil {
		return hsvm.VerifyRandSeedCaller(header)
	}

	return nil
}

func (hsvm *HeaderSigVerifierStub) IsInterfaceNil() bool {
	return false
}
