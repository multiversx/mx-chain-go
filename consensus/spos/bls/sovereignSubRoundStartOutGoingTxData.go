package bls

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type sovereignSubRoundStartOutGoingTxData struct {
	signingHandler consensus.SigningHandler
}

// NewSovereignSubRoundStartOutGoingTxData creates a new signer for sovereign outgoing tx data in start sub round
func NewSovereignSubRoundStartOutGoingTxData(signingHandler consensus.SigningHandler) (*sovereignSubRoundStartOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &sovereignSubRoundStartOutGoingTxData{
		signingHandler: signingHandler,
	}, nil
}

// Reset resets the pub keys
func (sr *sovereignSubRoundStartOutGoingTxData) Reset(pubKeys []string) error {
	return sr.signingHandler.Reset(pubKeys)
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundStartOutGoingTxData) Identifier() string {
	return "sovereignSubRoundStartOutGoingTxData"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundStartOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
