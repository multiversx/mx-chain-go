package bls

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// SubRoundStartExtraSignersHolder manages extra signers during start subround in a consensus process
type SubRoundStartExtraSignersHolder interface {
	Reset(pubKeys []string) error
	RegisterExtraSingingHandler(extraSigner consensus.SubRoundStartExtraSignatureHandler) error
	IsInterfaceNil() bool
}

// SubRoundSignatureExtraSignersHolder manages extra signers during signing subround in a consensus process
type SubRoundSignatureExtraSignersHolder interface {
	CreateExtraSignatureShares(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error)
	AddExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error
	StoreExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error
	RegisterExtraSingingHandler(extraSigner consensus.SubRoundSignatureExtraSignatureHandler) error
	IsInterfaceNil() bool
}

type SubRoundEndExtraSignersHolder interface {
	AggregateSignatures(bitmap []byte, epoch uint32) (map[string][]byte, error)
	AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error
	SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSigs map[string][]byte) error
	VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error
	HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error
	RegisterExtraEndRoundSigAggregatorHandler(extraSignatureAggregator SubRoundEndExtraSignatureAggregatorHandler) error
	IsInterfaceNil() bool
}

type SubRoundEndExtraSignatureAggregatorHandler interface {
	AggregateSignatures(bitmap []byte, epoch uint32) ([]byte, error)
	AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error
	SeAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error
	HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error
	VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error
	Identifier() string
	IsInterfaceNil() bool
}
