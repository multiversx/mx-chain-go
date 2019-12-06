package mock

import "github.com/ElrondNetwork/elrond-go/data"

type HeaderSigVerifierMock struct {
	VerifyRandSeedAndLeaderSignatureCalled func(header data.HeaderHandler) error
	VerifySignatureCalled                  func(header data.HeaderHandler) error
}

func (hsvm *HeaderSigVerifierMock) VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error {
	if hsvm.VerifyRandSeedAndLeaderSignatureCalled != nil {
		return hsvm.VerifyRandSeedAndLeaderSignatureCalled(header)
	}

	return nil
}

func (hsvm *HeaderSigVerifierMock) VerifySignature(header data.HeaderHandler) error {
	if hsvm.VerifySignatureCalled != nil {
		return hsvm.VerifySignatureCalled(header)
	}

	return nil
}

func (hsvm *HeaderSigVerifierMock) IsInterfaceNil() bool {
	return false
}
