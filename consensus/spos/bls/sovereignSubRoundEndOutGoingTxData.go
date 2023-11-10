package bls

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundEndOutGoingTxData struct {
	signingHandler consensus.SigningHandler
}

func NewSovereignSubRoundEndOutGoingTxData(
	signingHandler consensus.SigningHandler,
) (*sovereignSubRoundEndOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &sovereignSubRoundEndOutGoingTxData{
		signingHandler: signingHandler,
	}, nil
}

func (sr *sovereignSubRoundEndOutGoingTxData) VerifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil
	}

	return sr.signingHandler.Verify(outGoingMb.GetOutGoingOperationsHash(), bitmap, header.GetEpoch())
}

func (sr *sovereignSubRoundEndOutGoingTxData) AggregateSignatures(bitmap []byte, epoch uint32) ([]byte, error) {
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

func (sr *sovereignSubRoundEndOutGoingTxData) SetAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil
	}

	err := outGoingMb.SetAggregatedSignatureOutGoingOperations(aggregatedSig)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (sr *sovereignSubRoundEndOutGoingTxData) SignAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil
	}

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

func (sr *sovereignSubRoundEndOutGoingTxData) HaveConsensusHeaderWithFullInfo(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.HaveConsensusHeaderWithFullInfo", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil
	}

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

func (sr *sovereignSubRoundEndOutGoingTxData) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.HaveConsensusHeaderWithFullInfo", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil
	}

	cnsMsg.AggregatedSignatureOutGoingTxData = outGoingMb.GetAggregatedSignatureOutGoingOperations()
	cnsMsg.LeaderSignatureOutGoingTxData = outGoingMb.GetLeaderSignatureOutGoingOperations()

	return nil
}

func (sr *sovereignSubRoundEndOutGoingTxData) Identifier() string {
	return "sovereignSubRoundEndOutGoingTxData"
}

func (sr *sovereignSubRoundEndOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
