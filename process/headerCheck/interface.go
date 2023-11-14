package headerCheck

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-crypto-go"
)

type ExtraHeaderSigVerifierHandler interface {
	VerifyAggregatedSignature(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error
	VerifyLeaderSignature(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error
	RemoveLeaderSignature(header data.HeaderHandler) error
	RemoveAllSignatures(header data.HeaderHandler) error
	Identifier() string
	IsInterfaceNil() bool
}

type ExtraHeaderSigVerifierHolder interface {
	VerifyAggregatedSignature(header data.HeaderHandler, multiSigVerifier crypto.MultiSigner, pubKeysSigners [][]byte) error
	VerifyLeaderSignature(header data.HeaderHandler, leaderPubKey crypto.PublicKey) error
	RemoveLeaderSignature(header data.HeaderHandler) error
	RemoveAllSignatures(header data.HeaderHandler) error
	RegisterExtraHeaderSigVerifier(extraVerifier ExtraHeaderSigVerifierHandler) error
	IsInterfaceNil() bool
}
