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

func NewSovereignSubRoundOutGoingTxDataEnd(
	subRound *spos.Subround,
	signingHandler consensus.SigningHandler,
) (*sovereignSubRoundOutGoingTxDataEnd, error) {
	return &sovereignSubRoundOutGoingTxDataEnd{
		Subround:       subRound,
		signingHandler: signingHandler,
	}, nil
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
	sig, err := sr.signingHandler.AggregateSigs(bitmap, sr.Header.GetEpoch())
	if err != nil {
		return nil, err
	}

	err = sr.signingHandler.SetAggregatedSig(sig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) AddAggregatedSignature(aggregatedSig []byte, cnsMsg *consensus.Message) error {

	/*
		sovHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
		if !castOk {
			return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
		}

		if check.IfNil(sovHeader.GetOutGoingMiniBlockHeaderHandler()) {
			return nil
		}

		outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
		err := outGoingMb.SetAggregatedSignatureOutGoingOperations(aggregatedSig)
		if err != nil {
			return err
		}
	*/
	cnsMsg.AggregatedSignatureOutGoingTxData = aggregatedSig

	return nil
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) Identifier() string {
	return "sovereignSubRoundOutGoingTxDataEnd"
}
