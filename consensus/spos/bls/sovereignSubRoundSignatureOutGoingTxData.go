package bls

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundSignatureOutGoingTxData struct {
	signingHandler consensus.SigningHandler
}

func NewSovereignSubRoundEndOutGoingTxData(signingHandler consensus.SigningHandler) (*sovereignSubRoundSignatureOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &sovereignSubRoundSignatureOutGoingTxData{
		signingHandler: signingHandler,
	}, nil
}

func (sr *sovereignSubRoundSignatureOutGoingTxData) CreateSignatureShare(
	header data.HeaderHandler,
	selfIndex uint16,
	selfPubKey []byte,
) ([]byte, error) {
	sovChainHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundSignatureOutGoingTxData.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMBHeader := sovChainHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMBHeader) {
		return make([]byte, 0), nil
	}

	return sr.signingHandler.CreateSignatureShareForPublicKey(
		outGoingMBHeader.GetOutGoingOperationsHash(),
		selfIndex,
		header.GetEpoch(),
		selfPubKey)
}

func (sr *sovereignSubRoundSignatureOutGoingTxData) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) {
	cnsMsg.SignatureShareOutGoingTxData = sigShare
}

func (sr *sovereignSubRoundSignatureOutGoingTxData) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	return sr.signingHandler.StoreSignatureShare(index, cnsMsg.SignatureShareOutGoingTxData)
}

func (sr *sovereignSubRoundSignatureOutGoingTxData) Identifier() string {
	return "sovereignSubRoundSignatureOutGoingTxData"
}

func (sr *sovereignSubRoundSignatureOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
