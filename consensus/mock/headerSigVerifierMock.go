package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderSigVerifierMock struct {
	VerifyRandSeedCaller func(header data.HeaderHandler) error
}

func (hsvm *HeaderSigVerifierMock) VerifyRandSeed(header data.HeaderHandler) error {
	if hsvm.VerifyRandSeedCaller != nil {
		return hsvm.VerifyRandSeedCaller(header)
	}

	return nil
}

func (hsvm *HeaderSigVerifierMock) IsInterfaceNil() bool {
	return false
}
