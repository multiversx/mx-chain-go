package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
)

func checkGoRoutinesThrottler(
	ctx context.Context,
	signatureThrottler core.Throttler,
) error {
	for {
		if signatureThrottler.CanProcess() {
			break
		}

		select {
		case <-time.After(timeSpentBetweenChecks):
			continue
		case <-ctx.Done():
			return fmt.Errorf("%w while checking the throttler", spos.ErrTimeIsOut)
		}
	}

	return nil
}

// computeSignatureThreshold returns the signature threshold for the given header, together
// with whether fallback validation applies; shared by end round and evidence capture
func computeSignatureThreshold(sr *spos.Subround, header data.HeaderHandler) (int, bool) {
	isTransitionBlock := common.IsEpochChangeBlockForFlagActivation(header, sr.EnableEpochsHandler(), common.AndromedaFlag)

	threshold := sr.Threshold(bls.SrSignature)
	if isTransitionBlock {
		threshold = core.GetPBFTThreshold(sr.ConsensusGroupSize())
	}

	fallbackApplied := sr.FallbackHeaderValidator().ShouldApplyFallbackValidation(header)
	if fallbackApplied {
		threshold = sr.FallbackThreshold(bls.SrSignature)
		if isTransitionBlock {
			threshold = core.GetPBFTFallbackThreshold(sr.ConsensusGroupSize())
		}
	}

	return threshold, fallbackApplied
}
