package extraSigners

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/consensus"
)

type sovereignSubRoundSignatureOutGoingTxData struct {
	*baseSubRoundSignature
}

// NewSovereignSubRoundSignatureOutGoingTxData creates a new signer for sovereign outgoing tx data in signature sub round
func NewSovereignSubRoundSignatureOutGoingTxData(signingHandler consensus.SigningHandler) (*sovereignSubRoundSignatureOutGoingTxData, error) {
	baseHandler, err := newBaseSubRoundSignature(signingHandler, block.OutGoingMbTx)
	if err != nil {
		return nil, err
	}

	return &sovereignSubRoundSignatureOutGoingTxData{
		baseSubRoundSignature: baseHandler,
	}, nil
}

// CreateSignatureShare creates a signature share for outgoing tx hash, if exists
func (sr *sovereignSubRoundSignatureOutGoingTxData) CreateSignatureShare(
	header data.HeaderHandler,
	selfIndex uint16,
	selfPubKey []byte,
) ([]byte, error) {
	return sr.createSignatureShare(header, selfIndex, selfPubKey)
}

// AddSigShareToConsensusMessage adds the provided sig share for outgoing tx data to the consensus message
func (sr *sovereignSubRoundSignatureOutGoingTxData) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) error {
	return sr.addSigShareToConsensusMessage(sigShare, cnsMsg)
}

// StoreSignatureShare stores the provided sig share for outgoing tx data from the consensus message
func (sr *sovereignSubRoundSignatureOutGoingTxData) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	return sr.storeSignatureShare(index, cnsMsg)
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundSignatureOutGoingTxData) Identifier() string {
	return "sovereignSubRoundSignatureOutGoingTxData"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundSignatureOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
