package bls

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

type SubRoundExtraDataSignatureHandler interface {
	CreateSignatureShare(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) ([]byte, error)
	AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error
	Identifier() string
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
