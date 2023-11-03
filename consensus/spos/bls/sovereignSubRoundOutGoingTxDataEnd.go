package bls

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundOutGoingTxDataEnd struct {
	signingHandler consensus.SigningHandler
}

func NewSovereignSubRoundOutGoingTxDataEnd(
	signingHandler consensus.SigningHandler,
) (*sovereignSubRoundOutGoingTxDataEnd, error) {
	return &sovereignSubRoundOutGoingTxDataEnd{
		signingHandler: signingHandler,
	}, nil
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataEnd.SeAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()

	return sr.signingHandler.Verify(outGoingMb.GetOutGoingOperationsHash(), bitmap, header.GetEpoch())
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) AggregateSignatures(bitmap []byte, epoch uint32) ([]byte, error) {
	sig, err := sr.signingHandler.AggregateSigs(bitmap, epoch)
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

func (sr *sovereignSubRoundOutGoingTxDataEnd) IsInterfaceNil() bool {
	return sr == nil
}
