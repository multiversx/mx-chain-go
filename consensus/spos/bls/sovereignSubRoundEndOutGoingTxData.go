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
	*spos.Subround
	signingHandler consensus.SigningHandler
}

func NewSovereignSubRoundEndOutGoingTxData(
	subRound *spos.Subround,
	signingHandler consensus.SigningHandler,
) (*sovereignSubRoundEndOutGoingTxData, error) {
	return &sovereignSubRoundEndOutGoingTxData{
		Subround:       subRound,
		signingHandler: signingHandler,
	}, nil
}

func (sr *sovereignSubRoundEndOutGoingTxData) CreateSignatureShare(
	header data.HeaderHandler,
	selfIndex uint16,
	selfPubKey []byte,
) ([]byte, error) {
	sovChainHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMBHeader := sovChainHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMBHeader) {
		return make([]byte, 0), nil
	}

	err := sr.signingHandler.Reset(sr.ConsensusGroup())
	if err != nil {
		return nil, err
	}

	return sr.signingHandler.CreateSignatureShareForPublicKey(
		outGoingMBHeader.GetOutGoingOperationsHash(),
		selfIndex,
		header.GetEpoch(),
		selfPubKey)
}

func (sr *sovereignSubRoundEndOutGoingTxData) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) {
	cnsMsg.SignatureShareOutGoingTxData = sigShare
}

func (sr *sovereignSubRoundEndOutGoingTxData) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	return sr.signingHandler.StoreSignatureShare(index, cnsMsg.SignatureShareOutGoingTxData)
}

func (sr *sovereignSubRoundEndOutGoingTxData) Identifier() string {
	return "sovereignSubRoundEndOutGoingTxData"
}

func (sr *sovereignSubRoundEndOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
