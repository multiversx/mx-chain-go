package headerCheck

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignHeaderSigVerifier struct {
	singleSigVerifier crypto.SingleSigner
}

// NewSovereignHeaderSigVerifier creates a new sovereign header sig verifier for outgoing operations
func NewSovereignHeaderSigVerifier(singleSigVerifier crypto.SingleSigner) (*sovereignHeaderSigVerifier, error) {
	if check.IfNil(singleSigVerifier) {
		return nil, process.ErrNilSingleSigner
	}

	return &sovereignHeaderSigVerifier{
		singleSigVerifier: singleSigVerifier,
	}, nil
}

// VerifyAggregatedSignature verifies aggregated sig for outgoing operations
func (hsv *sovereignHeaderSigVerifier) VerifyAggregatedSignature(
	header data.HeaderHandler,
	multiSigVerifier crypto.MultiSigner,
	pubKeysSigners [][]byte,
) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.VerifyAggregatedSignature", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := multiSigVerifier.VerifyAggregatedSig(
			pubKeysSigners,
			outGoingMBHdr.GetOutGoingOperationsHash(),
			outGoingMBHdr.GetAggregatedSignatureOutGoingOperations(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// VerifyLeaderSignature verifies leader sig for outgoing operations
func (hsv *sovereignHeaderSigVerifier) VerifyLeaderSignature(
	header data.HeaderHandler,
	leaderPubKey crypto.PublicKey,
) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.VerifyLeaderSignature", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		leaderMsgToSign := append(
			outGoingMBHdr.GetOutGoingOperationsHash(),
			outGoingMBHdr.GetAggregatedSignatureOutGoingOperations()...)

		err := hsv.singleSigVerifier.Verify(
			leaderPubKey,
			leaderMsgToSign,
			outGoingMBHdr.GetLeaderSignatureOutGoingOperations())
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveLeaderSignature removes leader sig from outgoing operations
func (hsv *sovereignHeaderSigVerifier) RemoveLeaderSignature(header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.RemoveLeaderSignature", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := outGoingMBHdr.SetLeaderSignatureOutGoingOperations(nil)
		if err != nil {
			return err
		}

		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMBHdr)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAllSignatures removes aggregated + leader sig from outgoing operations
func (hsv *sovereignHeaderSigVerifier) RemoveAllSignatures(header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignHeaderSigVerifier.RemoveAllSignatures", errors.ErrWrongTypeAssertion)
	}

	for _, outGoingMBHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := outGoingMBHdr.SetAggregatedSignatureOutGoingOperations(nil)
		if err != nil {
			return err
		}

		err = outGoingMBHdr.SetLeaderSignatureOutGoingOperations(nil)
		if err != nil {
			return err
		}

		err = sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMBHdr)
		if err != nil {
			return err
		}
	}

	return nil
}

// Identifier returns the unique id of the header verifier
func (hsv *sovereignHeaderSigVerifier) Identifier() string {
	return "sovereignHeaderSigVerifier"
}

// IsInterfaceNil checks if the underlying pointer is nil
func (hsv *sovereignHeaderSigVerifier) IsInterfaceNil() bool {
	return hsv == nil
}
