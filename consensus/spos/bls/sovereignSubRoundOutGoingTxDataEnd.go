package bls

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignSubRoundOutGoingTxDataEnd struct {
	*spos.Subround
}

func (sr *sovereignSubRoundOutGoingTxDataEnd) VerifyFinalBlockSignatures(cnsDta *consensus.Message) error {
	if check.IfNil(sr.Header) {
		return spos.ErrNilHeader
	}

	sovHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundOutGoingTxDataSignature.CreateSignatureShare", errors.ErrWrongTypeAssertion)
	}

	if len(sovHeader.GetOutGoingOperationHashes()) == 0 {
		return nil
	}

	return nil
}
