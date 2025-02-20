package headerSigVerifier

import (
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/process"
)

// ExtraHeaderSigVerifierHolderMock -
type ExtraHeaderSigVerifierHolderMock struct {
	VerifyAggregatedSignatureCalled      func(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error
	VerifyLeaderSignatureCalled          func(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error
	RemoveLeaderSignatureCalled          func(header data.HeaderHandler) error
	RemoveAllSignaturesCalled            func(header data.HeaderHandler) error
	RegisterExtraHeaderSigVerifierCalled func(extraVerifier process.ExtraHeaderSigVerifierHandler) error
}

// VerifyAggregatedSignature -
func (mock *ExtraHeaderSigVerifierHolderMock) VerifyAggregatedSignature(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error {
	if mock.VerifyAggregatedSignatureCalled != nil {
		return mock.VerifyAggregatedSignatureCalled(header, multiSigVerifier, pubKeysSigners)
	}
	return nil
}

// VerifyLeaderSignature -
func (mock *ExtraHeaderSigVerifierHolderMock) VerifyLeaderSignature(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error {
	if mock.VerifyLeaderSignatureCalled != nil {
		return mock.VerifyLeaderSignatureCalled(header, leaderPubKey)
	}
	return nil
}

// RemoveLeaderSignature -
func (mock *ExtraHeaderSigVerifierHolderMock) RemoveLeaderSignature(header data.HeaderHandler) error {
	if mock.RemoveLeaderSignatureCalled != nil {
		return mock.RemoveLeaderSignatureCalled(header)
	}
	return nil
}

// RemoveAllSignatures -
func (mock *ExtraHeaderSigVerifierHolderMock) RemoveAllSignatures(header data.HeaderHandler) error {
	if mock.RemoveAllSignaturesCalled != nil {
		return mock.RemoveAllSignaturesCalled(header)
	}
	return nil
}

// RegisterExtraHeaderSigVerifier -
func (mock *ExtraHeaderSigVerifierHolderMock) RegisterExtraHeaderSigVerifier(extraVerifier process.ExtraHeaderSigVerifierHandler) error {
	if mock.RegisterExtraHeaderSigVerifierCalled != nil {
		return mock.RegisterExtraHeaderSigVerifierCalled(extraVerifier)
	}
	return nil
}

// IsInterfaceNil -
func (mock *ExtraHeaderSigVerifierHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
