package consensus

import "github.com/multiversx/mx-chain-core-go/data"

// HeaderSigVerifierMock -
type HeaderSigVerifierMock struct {
	VerifyRandSeedAndLeaderSignatureCalled func(header data.HeaderHandler) error
	VerifySignatureCalled                  func(header data.HeaderHandler) error
	VerifyRandSeedCalled                   func(header data.HeaderHandler) error
	VerifyLeaderSignatureCalled            func(header data.HeaderHandler) error
	VerifySignatureForHashCalled           func(header data.HeaderHandler, hash []byte, pubkeysBitmap []byte, signature []byte) error
	VerifyHeaderProofCalled                func(proofHandler data.HeaderProofHandler) error
}

// VerifyRandSeed -
func (mock *HeaderSigVerifierMock) VerifyRandSeed(header data.HeaderHandler) error {
	if mock.VerifyRandSeedCalled != nil {
		return mock.VerifyRandSeedCalled(header)
	}

	return nil
}

// VerifyRandSeedAndLeaderSignature -
func (mock *HeaderSigVerifierMock) VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error {
	if mock.VerifyRandSeedAndLeaderSignatureCalled != nil {
		return mock.VerifyRandSeedAndLeaderSignatureCalled(header)
	}

	return nil
}

// VerifySignature -
func (mock *HeaderSigVerifierMock) VerifySignature(header data.HeaderHandler) error {
	if mock.VerifySignatureCalled != nil {
		return mock.VerifySignatureCalled(header)
	}

	return nil
}

// VerifyLeaderSignature -
func (mock *HeaderSigVerifierMock) VerifyLeaderSignature(header data.HeaderHandler) error {
	if mock.VerifyLeaderSignatureCalled != nil {
		return mock.VerifyLeaderSignatureCalled(header)
	}

	return nil
}

// VerifySignatureForHash -
func (mock *HeaderSigVerifierMock) VerifySignatureForHash(header data.HeaderHandler, hash []byte, pubkeysBitmap []byte, signature []byte) error {
	if mock.VerifySignatureForHashCalled != nil {
		return mock.VerifySignatureForHashCalled(header, hash, pubkeysBitmap, signature)
	}

	return nil
}

// VerifyHeaderProof -
func (mock *HeaderSigVerifierMock) VerifyHeaderProof(proofHandler data.HeaderProofHandler) error {
	if mock.VerifyHeaderProofCalled != nil {
		return mock.VerifyHeaderProofCalled(proofHandler)
	}

	return nil
}

// IsInterfaceNil -
func (mock *HeaderSigVerifierMock) IsInterfaceNil() bool {
	return mock == nil
}
