package bls

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block"
)

type sovereignSubRoundEnd struct {
	*subroundEndRoundV2
	outGoingOperationsPool block.OutGoingOperationsPool
	bridgeOpHandler        BridgeOperationsHandler
}

func NewSovereignSubRoundEndRound(
	subRoundEnd *subroundEndRoundV2,
	outGoingOperationsPool block.OutGoingOperationsPool,
	bridgeOpHandler BridgeOperationsHandler,
) (*sovereignSubRoundEnd, error) {
	sr := &sovereignSubRoundEnd{
		subroundEndRoundV2:     subRoundEnd,
		outGoingOperationsPool: outGoingOperationsPool,
		bridgeOpHandler:        bridgeOpHandler,
	}

	sr.Job = sr.doSovereignEndRoundJob

	return sr, nil
}

func (sr *sovereignSubRoundEnd) doSovereignEndRoundJob(ctx context.Context) bool {
	success := sr.subroundEndRoundV2.doEndRoundJob(ctx)
	if !success {
		return false
	}

	if !sr.isSelfLeader() {
		return true
	}

	sovHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Error("%w in sovereignSubRoundEndOutGoingTxData.HaveConsensusHeaderWithFullInfo", errors.ErrWrongTypeAssertion)
		return false
	}

	outGoingMBHeader := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMBHeader) {
		return true
	}

	err := sr.bridgeOpHandler.Send(ctx, &sovereign.BridgeOperations{
		Data: sr.getOutGoingOperations(outGoingMBHeader),
	})

	if err != nil {
		return false
	}

	return true
}

func (sr *sovereignSubRoundEnd) getOutGoingOperations(outGoingMBHeader data.OutGoingMiniBlockHeaderHandler) []*sovereign.BridgeOutGoingData {
	outGoingOperations := make([]*sovereign.BridgeOutGoingData, 0)

	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) != 0 {
		log.Debug("found unconfirmed operations", "num unconfirmed operations", len(unconfirmedOperations))
		outGoingOperations = append(outGoingOperations, unconfirmedOperations...)
	}

	hash := outGoingMBHeader.GetOutGoingOperationsHash()
	currentOperations := sr.outGoingOperationsPool.Get(hash)
	return append(outGoingOperations, currentOperations)
}
