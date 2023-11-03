package bls

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

type SubRoundExtraDataSignatureHandler interface {
	CreateSignatureShare(selfIndex uint16) ([]byte, error)
	AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message)
	StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error
	Identifier() string
	IsInterfaceNil() bool
}

type SubRoundEndExtraSignatureAggregatorHandler interface {
	VerifyFinalBlockSignatures(cnsDta *consensus.Message) error
	AggregateSignatures(bitmap []byte) ([]byte, error)
	AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error
	SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error
	SeAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error
	HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error
	Identifier() string
	IsInterfaceNil() bool
}
