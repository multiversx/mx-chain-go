package headerCheck

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignHeaderSigVerifier struct {
	singleSigVerifier crypto.SingleSigner
}

func (hsv *sovereignHeaderSigVerifier) VerifySignature(
	header data.HeaderHandler,
	multiSigVerifier crypto.MultiSigner,
	pubKeysSigners [][]byte,
) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMBHeader := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMBHeader) {
		return nil
	}

	return multiSigVerifier.VerifyAggregatedSig(
		pubKeysSigners,
		outGoingMBHeader.GetOutGoingOperationsHash(),
		outGoingMBHeader.GetAggregatedSignatureOutGoingOperations(),
	)
}

func (hsv *sovereignHeaderSigVerifier) VerifyLeaderSignature(
	headerWithoutSignatures data.HeaderHandler,
	leaderPubKey crypto.PublicKey,
	headerWithSignatures data.HeaderHandler,
) error {
	sovHeader, castOk := headerWithoutSignatures.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()

	leaderMsgToSign := append(
		outGoingMb.GetOutGoingOperationsHash(),
		outGoingMb.GetAggregatedSignatureOutGoingOperations()...)

	return hsv.singleSigVerifier.Verify(
		leaderPubKey,
		leaderMsgToSign,
		headerWithSignatures.(data.SovereignChainHeaderHandler).GetOutGoingMiniBlockHeaderHandler().GetLeaderSignatureOutGoingOperations())
}

func (hsv *sovereignHeaderSigVerifier) RemoveLeaderSig(header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	outGoingMb.SetLeaderSignatureOutGoingOperations(nil)
	sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)

	return nil
}

func (hsv *sovereignHeaderSigVerifier) RemoveAllSignatures(headerCopy data.HeaderHandler) error {
	sovHeader, castOk := headerCopy.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	outGoingMb.SetAggregatedSignatureOutGoingOperations(nil)
	outGoingMb.SetLeaderSignatureOutGoingOperations(nil)
	sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)

	return nil
}
