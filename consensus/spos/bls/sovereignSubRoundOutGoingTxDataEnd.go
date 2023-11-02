package bls

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundOutGoingTxDataEnd struct {
	*spos.Subround

	signingHandler consensus.SigningHandler
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) VerifyFinalBlockSignatures(cnsDta *consensus.Message) error {
	if check.IfNil(sr.Header) {
		return spos.ErrNilHeader
	}

	sovHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	if check.IfNil(sovHeader.GetOutGoingMiniBlockHeaderHandler()) {
		return nil
	}

	return nil
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) AggregateSignatures(bitmap []byte) ([]byte, error) {
	return sr.signingHandler.AggregateSigs(bitmap, sr.Header.GetEpoch())
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) AddAggregatedSignature(aggregatedSig []byte, cnsMsg *consensus.Message) {
	cnsMsg.AggregateSignature = aggregatedSig
}
