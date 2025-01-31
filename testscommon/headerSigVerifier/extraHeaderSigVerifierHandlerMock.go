package headerSigVerifier

import (
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// ExtraHeaderSigVerifierHandlerMock -
type ExtraHeaderSigVerifierHandlerMock struct {
	VerifyAggregatedSignatureCalled func(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error
	VerifyLeaderSignatureCalled     func(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error
	RemoveLeaderSignatureCalled     func(header data.HeaderHandler) error
	RemoveAllSignaturesCalled       func(header data.HeaderHandler) error
	IdentifierCalled                func() string
}

// VerifyAggregatedSignature -
func (mock *ExtraHeaderSigVerifierHandlerMock) VerifyAggregatedSignature(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error {
	if mock.VerifyAggregatedSignatureCalled != nil {
		return mock.VerifyAggregatedSignatureCalled(header, multiSigVerifier, pubKeysSigners)
	}
	return nil
}

// VerifyLeaderSignature -
func (mock *ExtraHeaderSigVerifierHandlerMock) VerifyLeaderSignature(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error {
	if mock.VerifyLeaderSignatureCalled != nil {
		return mock.VerifyLeaderSignatureCalled(header, leaderPubKey)
	}
	return nil
}

// RemoveLeaderSignature -
func (mock *ExtraHeaderSigVerifierHandlerMock) RemoveLeaderSignature(header data.HeaderHandler) error {
	if mock.RemoveLeaderSignatureCalled != nil {
		return mock.RemoveLeaderSignatureCalled(header)
	}
	return nil
}

// RemoveAllSignatures -
func (mock *ExtraHeaderSigVerifierHandlerMock) RemoveAllSignatures(header data.HeaderHandler) error {
	if mock.RemoveAllSignaturesCalled != nil {
		return mock.RemoveAllSignaturesCalled(header)
	}
	return nil
}

// Identifier -
func (mock *ExtraHeaderSigVerifierHandlerMock) Identifier() string {
	if mock.IdentifierCalled != nil {
		return mock.IdentifierCalled()
	}
	return ""
}

// IsInterfaceNil -
func (mock *ExtraHeaderSigVerifierHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
