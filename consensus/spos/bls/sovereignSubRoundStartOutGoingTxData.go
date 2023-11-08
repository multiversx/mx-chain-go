package bls

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type sovereignSubRoundStartOutGoingTxData struct {
	signingHandler consensus.SigningHandler
}

func NewSovereignSubRoundStartOutGoingTxData(signingHandler consensus.SigningHandler) (*sovereignSubRoundStartOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &sovereignSubRoundStartOutGoingTxData{
		signingHandler: signingHandler,
	}, nil
}

func (sr *sovereignSubRoundStartOutGoingTxData) Reset(pubKeys []string) error {
	return sr.signingHandler.Reset(pubKeys)
}

func (sr *sovereignSubRoundStartOutGoingTxData) Identifier() string {
	return "sovereignSubRoundStartOutGoingTxData"
}

func (sr *sovereignSubRoundStartOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
