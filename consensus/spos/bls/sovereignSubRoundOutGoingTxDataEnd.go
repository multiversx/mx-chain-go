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

func (sr *sovereignSubRoundOutGoingTxDataEnd) SeAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataEnd.SeAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	err := outGoingMb.SetAggregatedSignatureOutGoingOperations(aggregatedSig)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataEnd.SeAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	leaderMsgToSign := append(
		outGoingMb.GetOutGoingOperationsHash(),
		outGoingMb.GetAggregatedSignatureOutGoingOperations()...)

	leaderSig, err := sr.signingHandler.CreateSignatureForPublicKey(leaderMsgToSign, leaderPubKey)
	if err != nil {
		return err
	}

	err = outGoingMb.SetLeaderSignatureOutGoingOperations(leaderSig)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataEnd.HaveConsensusHeaderWithFullInfo", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	err := outGoingMb.SetAggregatedSignatureOutGoingOperations(cnsMsg.AggregatedSignatureOutGoingTxData)
	if err != nil {
		return err
	}
	err = outGoingMb.SetLeaderSignatureOutGoingOperations(cnsMsg.LeaderSignatureOutGoingTxData)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataEnd.HaveConsensusHeaderWithFullInfo", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()

	cnsMsg.AggregatedSignatureOutGoingTxData = outGoingMb.GetAggregatedSignatureOutGoingOperations()
	cnsMsg.LeaderSignatureOutGoingTxData = outGoingMb.GetLeaderSignatureOutGoingOperations()

	return nil
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) Identifier() string {
	return "sovereignSubRoundOutGoingTxDataEnd"
}
