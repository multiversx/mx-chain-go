package bls

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block"
)

type sovereignSubRoundEnd struct {
	*subroundEndRoundV2
	outGoingOperationsPool block.OutGoingOperationsPool
	bridgeOpHandler        BridgeOperationsHandler
}

// NewSovereignSubRoundEndRound creates a new sovereign end subround
func NewSovereignSubRoundEndRound(
	subRoundEnd *subroundEndRoundV2,
	outGoingOperationsPool block.OutGoingOperationsPool,
	bridgeOpHandler BridgeOperationsHandler,
) (*sovereignSubRoundEnd, error) {
	if check.IfNil(subRoundEnd) {
		return nil, spos.ErrNilSubround
	}
	if check.IfNil(outGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(bridgeOpHandler) {
		return nil, errors.ErrNilBridgeOpHandler
	}

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

	sovHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob", "error", errors.ErrWrongTypeAssertion)
		return false
	}

	outGoingMBHeader := sovHeader.GetOutGoingMiniBlockHeaderHandler()
	if check.IfNil(outGoingMBHeader) {
		sr.sendUnconfirmedOperationsIfFound(ctx)
		return true
	}

	currBridgeData, err := sr.updateBridgeDataWithSignatures(outGoingMBHeader)
	if err != nil {
		log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob.updateBridgeDataWithSignatures", "error", err)
		return false
	}

	if !sr.isSelfLeader() {
		return true
	}

	outGoingOperations := sr.getAllOutGoingOperations(currBridgeData)
	go sr.sendOutGoingOperations(ctx, outGoingOperations)

	return true
}

func (sr *sovereignSubRoundEnd) sendUnconfirmedOperationsIfFound(ctx context.Context) {
	if !sr.isSelfLeader() {
		return
	}

	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) == 0 {
		return
	}

	log.Debug("found unconfirmed operations", "num unconfirmed operations", len(unconfirmedOperations))
	go sr.sendOutGoingOperations(ctx, unconfirmedOperations)
}

func (sr *sovereignSubRoundEnd) updateBridgeDataWithSignatures(
	outGoingMBHeader data.OutGoingMiniBlockHeaderHandler,
) (*sovereign.BridgeOutGoingData, error) {
	hash := outGoingMBHeader.GetOutGoingOperationsHash()
	currBridgeData := sr.outGoingOperationsPool.Get(hash)
	if currBridgeData == nil {
		return nil, fmt.Errorf("%w in sovereignSubRoundEnd.updateBridgeDataWithSignatures for hash: %s",
			errors.ErrOutGoingOperationsNotFound, hex.EncodeToString(hash))
	}

	currBridgeData.LeaderSignature = outGoingMBHeader.GetLeaderSignatureOutGoingOperations()
	currBridgeData.AggregatedSignature = outGoingMBHeader.GetAggregatedSignatureOutGoingOperations()

	sr.outGoingOperationsPool.Delete(hash)
	sr.outGoingOperationsPool.Add(currBridgeData)
	return currBridgeData, nil
}

func (sr *sovereignSubRoundEnd) isSelfLeader() bool {
	return sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
}

func (sr *sovereignSubRoundEnd) getAllOutGoingOperations(currentOperations *sovereign.BridgeOutGoingData) []*sovereign.BridgeOutGoingData {
	outGoingOperations := make([]*sovereign.BridgeOutGoingData, 0)
	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) != 0 {
		log.Debug("found unconfirmed operations", "num unconfirmed operations", len(unconfirmedOperations))
		outGoingOperations = append(outGoingOperations, unconfirmedOperations...)
	}

	log.Debug("current outgoing operations", "hash", currentOperations.Hash)
	return append(outGoingOperations, currentOperations)
}

func (sr *sovereignSubRoundEnd) sendOutGoingOperations(ctx context.Context, data []*sovereign.BridgeOutGoingData) {
	resp, err := sr.bridgeOpHandler.Send(ctx, &sovereign.BridgeOperations{
		Data: data,
	})
	if err != nil {
		log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob.bridgeOpHandler.Send", "error", err)
		return
	}

	log.Debug("sent outgoing operations", "hashes", resp.TxHashes)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundEnd) IsInterfaceNil() bool {
	return sr == nil
}
