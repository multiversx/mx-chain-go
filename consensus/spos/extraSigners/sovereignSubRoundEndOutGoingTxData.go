package extraSigners

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/consensus"
)

var log = logger.GetOrCreate("extra-signers")

type sovereignSubRoundEndOutGoingTxData struct {
	*baseSubRoundEnd
}

// NewSovereignSubRoundEndOutGoingTxData creates a new signer for sovereign outgoing tx data in end sub round
func NewSovereignSubRoundEndOutGoingTxData(
	signingHandler consensus.SigningHandler,
) (*sovereignSubRoundEndOutGoingTxData, error) {
	baseHandler, err := newBaseSubRoundEnd(signingHandler, block.OutGoingMbTx)
	if err != nil {
		return nil, err
	}

	return &sovereignSubRoundEndOutGoingTxData{
		baseSubRoundEnd: baseHandler,
	}, nil
}

// VerifyAggregatedSignatures verifies outgoing tx aggregated signatures from provided header
func (sr *sovereignSubRoundEndOutGoingTxData) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	return sr.verifyAggregatedSignatures(bitmap, header)
}

// AggregateAndSetSignatures aggregates and sets signatures for outgoing tx data
func (sr *sovereignSubRoundEndOutGoingTxData) AggregateAndSetSignatures(bitmap []byte, header data.HeaderHandler) ([]byte, error) {
	return sr.aggregateAndSetSignatures(bitmap, header)
}

// SetAggregatedSignatureInHeader sets aggregated signature for outgoing tx in header
func (sr *sovereignSubRoundEndOutGoingTxData) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	return sr.setAggregatedSignatureInHeader(header, aggregatedSig)
}

// SignAndSetLeaderSignature signs and sets leader signature for outgoing tx in header
func (sr *sovereignSubRoundEndOutGoingTxData) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	return sr.signAndSetLeaderSignature(header, leaderPubKey)
}

// SetConsensusDataInHeader sets aggregated and leader signature in header with provided data from consensus message
func (sr *sovereignSubRoundEndOutGoingTxData) SetConsensusDataInHeader(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	return sr.setConsensusDataInHeader(header, cnsMsg)
}

// AddLeaderAndAggregatedSignatures adds aggregated and leader signature in consensus message with provided data from header
func (sr *sovereignSubRoundEndOutGoingTxData) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	return sr.addLeaderAndAggregatedSignatures(header, cnsMsg)
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundEndOutGoingTxData) Identifier() string {
	return "sovereignSubRoundEndOutGoingTxData"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundEndOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
