package extraSigners

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type baseSubRoundSignature struct {
	signingHandler consensus.SigningHandler
	mbType         block.OutGoingMBType
}

func newBaseSubRoundSignature(
	signingHandler consensus.SigningHandler,
	mbType block.OutGoingMBType,
) (*baseSubRoundSignature, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &baseSubRoundSignature{
		signingHandler: signingHandler,
		mbType:         mbType,
	}, nil
}

func (br *baseSubRoundSignature) createSignatureShare(
	header data.HeaderHandler,
	selfIndex uint16,
	selfPubKey []byte,
) ([]byte, error) {
	sovChainHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundSignatureOutGoingTxData.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMBHeader := sovChainHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMBHeader) {
		return make([]byte, 0), nil
	}

	return br.signingHandler.CreateSignatureShareForPublicKey(
		outGoingMBHeader.GetOutGoingOperationsHash(),
		selfIndex,
		header.GetEpoch(),
		selfPubKey)
}

func (br *baseSubRoundSignature) addSigShareToConsensusMessage(sigShare []byte, cnsMsg *consensus.Message) error {
	if cnsMsg == nil {
		return errors.ErrNilConsensusMessage
	}

	if len(sigShare) == 0 {
		return nil

	}

	if cnsMsg.ExtraSignatures == nil {
		cnsMsg.ExtraSignatures = make(map[string]*consensus.ExtraSignatureData)
	}

	keyStr := br.mbType.String()
	if _, found := cnsMsg.ExtraSignatures[keyStr]; !found {
		cnsMsg.ExtraSignatures[keyStr] = &consensus.ExtraSignatureData{}
	}

	cnsMsg.ExtraSignatures[keyStr].SignatureShareOutGoingTxData = sigShare
	return nil
}

func (br *baseSubRoundSignature) storeSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	if cnsMsg == nil {
		return errors.ErrNilConsensusMessage
	}

	if extraSigData, found := cnsMsg.ExtraSignatures[br.mbType.String()]; found {
		return br.signingHandler.StoreSignatureShare(index, extraSigData.SignatureShareOutGoingTxData)
	}

	return nil
}
