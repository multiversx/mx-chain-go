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

func NewSovereignSubRoundEndRound(subRoundEnd *subroundEndRoundV2, outGoingOperationsPool block.OutGoingOperationsPool) (*sovereignSubRoundEnd, error) {
	sr := &sovereignSubRoundEnd{
		subroundEndRoundV2:     subRoundEnd,
		outGoingOperationsPool: outGoingOperationsPool,
	}

	sr.Job = sr.doSovereignBlockJob

	return &sovereignSubRoundEnd{}, nil
}

func (sr *sovereignSubRoundEnd) doSovereignBlockJob(ctx context.Context) bool {
	success := sr.subroundEndRoundV2.doEndRoundJob(ctx)
	if !success {
		return false
	}

	if !sr.IsSelfLeaderInCurrentRound() && !sr.IsMultiKeyLeaderInCurrentRound() {
		if sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup() {
			return true
		}
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

	outGoingOperations := make([]*sovereign.BridgeOutGoingData, 0)

	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) != 0 {
		outGoingOperations = append(outGoingOperations, unconfirmedOperations...)
	}

	hash := outGoingMBHeader.GetOutGoingOperationsHash()
	currentOperations := sr.outGoingOperationsPool.Get(hash)
	outGoingOperations = append(outGoingOperations, currentOperations)

	err := sr.bridgeOpHandler.Send(ctx, &sovereign.BridgeOperations{
		Data: outGoingOperations,
	})
	if err != nil {
		return false
	}

	return true
}
