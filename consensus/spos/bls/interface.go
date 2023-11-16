package bls

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundStartExtraSignersHolder manages extra signers during start subround in a consensus process
type SubRoundStartExtraSignersHolder interface {
	Reset(pubKeys []string) error
	RegisterExtraSigningHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error
	IsInterfaceNil() bool
}

// SubRoundSignatureExtraSignersHolder manages extra signers during signing subround in a consensus process
type SubRoundSignatureExtraSignersHolder interface {
	CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error)
	AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error
	StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error
	RegisterExtraSigningHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error
	IsInterfaceNil() bool
}

// SubRoundEndExtraSignersHolder manages extra signers during end subround in a consensus process
type SubRoundEndExtraSignersHolder interface {
	AggregateSignatures(bitmap []byte, header data.HeaderHandler) (map[string][]byte, error)
	AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error
	SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSigs map[string][]byte) error
	VerifyAggregatedSignatures(header data.HeaderHandler, bitmap []byte) error
	HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error
	RegisterExtraSigningHandler(extraSigner consensus.SubRoundEndExtraSignatureHandler) error
	IsInterfaceNil() bool
}
