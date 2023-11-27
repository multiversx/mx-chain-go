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

// NewSovereignSubRoundEndOutGoingTxData creates a new signer for sovereign outgoing tx data in end sub round
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

// VerifyAggregatedSignatures verifies outgoing tx aggregated signatures from provided header
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

// AggregateAndSetSignatures aggregates and sets signatures for outgoing tx data
func (sr *sovereignSubRoundEndOutGoingTxData) AggregateAndSetSignatures(bitmap []byte, header data.HeaderHandler) ([]byte, error) {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil, nil
	}

	sig, err := sr.signingHandler.AggregateSigs(bitmap, header.GetEpoch())
	if err != nil {
		return nil, err
	}

	err = sr.signingHandler.SetAggregatedSig(sig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// SetAggregatedSignatureInHeader sets aggregated signature for outgoing tx in header
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

// SignAndSetLeaderSignature signs and sets leader signature for outgoing tx in header
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

// SetConsensusDataInHeader sets aggregated and leader signature in header with provided data from consensus message
func (sr *sovereignSubRoundEndOutGoingTxData) SetConsensusDataInHeader(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetConsensusDataInHeader", errors.ErrWrongTypeAssertion)
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

// AddLeaderAndAggregatedSignatures adds aggregated and leader signature in consensus message with provided data from header
func (sr *sovereignSubRoundEndOutGoingTxData) AddLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetConsensusDataInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMb) {
		return nil
	}

	cnsMsg.AggregatedSignatureOutGoingTxData = outGoingMb.GetAggregatedSignatureOutGoingOperations()
	cnsMsg.LeaderSignatureOutGoingTxData = outGoingMb.GetLeaderSignatureOutGoingOperations()

	return nil
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundEndOutGoingTxData) Identifier() string {
	return "sovereignSubRoundEndOutGoingTxData"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundEndOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
