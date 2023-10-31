package bls

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundOutGoingTxDataSignature struct {
	*spos.Subround

	signingHandler consensus.SigningHandler
}

func NewSovereignSubRoundOutGoingTxDataSignature(
	subRound *spos.Subround,
	signingHandler consensus.SigningHandler,
) (*sovereignSubRoundOutGoingTxDataSignature, error) {
	return &sovereignSubRoundOutGoingTxDataSignature{
		Subround:       subRound,
		signingHandler: signingHandler,
	}, nil
}

func (sr *sovereignSubRoundOutGoingTxDataSignature) CreateSignatureShare(selfIndex uint16) ([]byte, error) {
	sovChainHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	if len(sovChainHeader.GetOutGoingOperationHashes()) == 0 {
		return make([]byte, 0), nil
	}

	return sr.signingHandler.CreateSignatureShareForPublicKey(
		sovChainHeader.GetOutGoingOperationHashes()[0],
		selfIndex,
		sr.Header.GetEpoch(),
		[]byte(sr.SelfPubKey()))
}

func (sr *sovereignSubRoundOutGoingTxDataSignature) AddSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) {
	cnsMsg.SignatureShareOutGoingTxData = sigShare
}

func (sr *sovereignSubRoundOutGoingTxDataSignature) StoreSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	return sr.signingHandler.StoreSignatureShare(index, cnsMsg.SignatureShareOutGoingTxData)
}

func (sr *sovereignSubRoundOutGoingTxDataSignature) Identifier() string {
	return "sovereignSubRoundOutGoingTxDataSignature"
}

func (sr *sovereignSubRoundOutGoingTxDataSignature) IsInterfaceNil() bool {
	return sr == nil
}
