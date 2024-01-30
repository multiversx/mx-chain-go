package bls

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
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

// ExtraSignersHolder manages all extra signer holders
type ExtraSignersHolder interface {
	GetSubRoundStartExtraSignersHolder() SubRoundStartExtraSignersHolder
	GetSubRoundSignatureExtraSignersHolder() SubRoundSignatureExtraSignersHolder
	GetSubRoundEndExtraSignersHolder() SubRoundEndExtraSignersHolder
	IsInterfaceNil() bool
}

// SubRoundEndV2Creator should create an end subround v2 and add it to the consensus core chronology
type SubRoundEndV2Creator interface {
	CreateAndAddSubRoundEnd(
		subroundEndRoundInstance *subroundEndRound,
		worker spos.WorkerHandler,
		consensusCore spos.ConsensusCoreHandler,
	) error
	IsInterfaceNil() bool
}

// BridgeOperationsHandler handles sending outgoing txs from sovereign to main chain
type BridgeOperationsHandler interface {
	Send(ctx context.Context, data *sovereign.BridgeOperations) (*sovereign.BridgeOperationsResponse, error)
	IsInterfaceNil() bool
}
