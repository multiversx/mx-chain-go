package extraSigners

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type sovereignSubRoundStartOutGoingTxData struct {
	signingHandler consensus.SigningHandler
	mbType         block.OutGoingMBType
}

// NewSovereignSubRoundStartExtraSigner creates a new signer for sovereign outgoing mb in start subround
func NewSovereignSubRoundStartExtraSigner(
	signingHandler consensus.SigningHandler,
	mbType block.OutGoingMBType,
) (*sovereignSubRoundStartOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &sovereignSubRoundStartOutGoingTxData{
		signingHandler: signingHandler,
		mbType:         mbType,
	}, nil
}

// Reset resets the pub keys
func (sr *sovereignSubRoundStartOutGoingTxData) Reset(pubKeys []string) error {
	return sr.signingHandler.Reset(pubKeys)
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundStartOutGoingTxData) Identifier() string {
	return sr.mbType.String()
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundStartOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
