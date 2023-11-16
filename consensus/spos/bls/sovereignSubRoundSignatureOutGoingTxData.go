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

// NewSovereignSubRoundSignatureOutGoingTxData creates a new signer for sovereign outgoing tx data in signature sub round
func NewSovereignSubRoundSignatureOutGoingTxData(signingHandler consensus.SigningHandler) (*sovereignSubRoundSignatureOutGoingTxData, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &sovereignSubRoundSignatureOutGoingTxData{
		signingHandler: signingHandler,
	}, nil
}

// CreateSignatureShare creates a signature share for outgoing tx hash, if exists
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

// AddSigShareToConsensusMessage adds the provided sig share for outgoing tx data to the consensus message
func (sr *sovereignSubRoundSignatureOutGoingTxData) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) {
	if len(sigShare) != 0 {
		cnsMsg.SignatureShareOutGoingTxData = sigShare
	}
}

// StoreSignatureShare stores the provided sig share for outgoing tx data from the consensus message
func (sr *sovereignSubRoundSignatureOutGoingTxData) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	if cnsMsg != nil && len(cnsMsg.SignatureShareOutGoingTxData) != 0 {
		return sr.signingHandler.StoreSignatureShare(index, cnsMsg.SignatureShareOutGoingTxData)
	}

	return nil
}

// Identifier returns the unique id of the signer
func (sr *sovereignSubRoundSignatureOutGoingTxData) Identifier() string {
	return "sovereignSubRoundSignatureOutGoingTxData"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundSignatureOutGoingTxData) IsInterfaceNil() bool {
	return sr == nil
}
