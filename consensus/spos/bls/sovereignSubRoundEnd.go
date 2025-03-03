package bls

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type sovereignSubRoundEnd struct {
	*subroundEndRoundV2
	outGoingOperationsPool OutGoingOperationsPool
	bridgeOpHandler        BridgeOperationsHandler
}

// NewSovereignSubRoundEndRound creates a new sovereign end subround
func NewSovereignSubRoundEndRound(
	subRoundEnd *subroundEndRoundV2,
	outGoingOperationsPool OutGoingOperationsPool,
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

func (sr *sovereignSubRoundEnd) receivedBlockHeaderFinalInfo(ctx context.Context, cnsDta *consensus.Message) bool {
	success := sr.subroundEndRoundV2.receivedBlockHeaderFinalInfo(ctx, cnsDta)
	if !success {
		return false
	}

	// TODO: MX-15502 once we have ZKProofs included in blocks for leaders which have resent the unconfirmed
	// outgoing operation we should also call resetOutGoingOpTimer here for consensus participants
	return sr.updateOutGoingPoolIfNeeded(cnsDta) == nil
}

func (sr *sovereignSubRoundEnd) updateOutGoingPoolIfNeeded(cnsDta *consensus.Message) error {
	sovHeader, castOk := sr.Header.(data.SovereignChainHeaderHandler)
	if !castOk {
		log.Error("sovereignSubRoundEnd.updateOutGoingPoolIfNeeded", "error", errors.ErrWrongTypeAssertion)
		return errors.ErrWrongTypeAssertion
	}

	for _, outGoingMbHdr := range sovHeader.GetOutGoingMiniBlockHeaderHandlers() {
		err := sr.updatePoolForOutGoingMiniBlock(outGoingMbHdr, cnsDta)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sr *sovereignSubRoundEnd) updatePoolForOutGoingMiniBlock(
	outGoingMBHeader data.OutGoingMiniBlockHeaderHandler,
	cnsDta *consensus.Message,
) error {
	mbType := block.OutGoingMBType(outGoingMBHeader.GetOutGoingMBTypeInt32()).String()
	extraSigData, found := cnsDta.ExtraSignatures[mbType]
	if !found {
		return fmt.Errorf("%w for type %s", ErrExtraSigShareDataNotFound, mbType)
	}

	err := outGoingMBHeader.SetAggregatedSignatureOutGoingOperations(extraSigData.AggregatedSignatureOutGoingTxData)
	if err != nil {
		log.Error("sovereignSubRoundEnd.updatePoolForOutGoingMiniBlock.SetAggregatedSignatureOutGoingOperations", "error", err)
		return err
	}

	err = outGoingMBHeader.SetLeaderSignatureOutGoingOperations(extraSigData.LeaderSignatureOutGoingTxData)
	if err != nil {
		log.Error("sovereignSubRoundEnd.updatePoolForOutGoingMiniBlock.SetLeaderSignatureOutGoingOperations", "error", err)
		return err
	}

	log.Debug("step 3.1: block header final info has been received with outgoing mb",
		"LeaderSignatureOutGoingTxData", extraSigData.LeaderSignatureOutGoingTxData,
		"AggregatedSignatureOutGoingTxData", extraSigData.AggregatedSignatureOutGoingTxData,
		"type", mbType,
	)

	_, err = sr.updateBridgeDataWithSignatures(outGoingMBHeader, cnsDta.PubKeysBitmap)
	if err != nil {
		log.Error("sovereignSubRoundEnd.updatePoolForOutGoingMiniBlock.updateBridgeDataWithSignatures", "error", err)
		return err
	}

	return nil
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

	outGoingMBHeaders := sovHeader.GetOutGoingMiniBlockHeaderHandlers()
	if len(outGoingMBHeaders) == 0 {
		sr.sendUnconfirmedOperationsIfFound(ctx)
		return true
	}

	currentOperations := make([]*sovereign.BridgeOutGoingData, len(outGoingMBHeaders))
	for idx, outGoingMBHdr := range outGoingMBHeaders {
		currBridgeData, err := sr.updateBridgeDataWithSignatures(outGoingMBHdr, sovHeader.GetPubKeysBitmap())
		if err != nil {
			log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob.updateBridgeDataWithSignatures", "error", err)
			return false
		}
		currentOperations[idx] = currBridgeData
	}

	if !sr.isSelfLeader() {
		return true
	}

	outGoingOperations := sr.getAllOutGoingOperations(currentOperations)
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
	outGoingMBHeader data.OutGoingMiniBlockHeaderHandler, pubKeysBitmap []byte,
) (*sovereign.BridgeOutGoingData, error) {
	hash := outGoingMBHeader.GetOutGoingOperationsHash()
	currBridgeData := sr.outGoingOperationsPool.Get(hash)
	if currBridgeData == nil {
		return nil, fmt.Errorf("%w in sovereignSubRoundEnd.updateBridgeDataWithSignatures for hash: %s",
			errors.ErrOutGoingOperationsNotFound, hex.EncodeToString(hash))
	}

	currBridgeData.LeaderSignature = outGoingMBHeader.GetLeaderSignatureOutGoingOperations()
	currBridgeData.AggregatedSignature = outGoingMBHeader.GetAggregatedSignatureOutGoingOperations()
	currBridgeData.PubKeysBitmap = pubKeysBitmap

	sr.outGoingOperationsPool.Delete(hash)
	sr.outGoingOperationsPool.Add(currBridgeData)
	return currBridgeData, nil
}

func (sr *sovereignSubRoundEnd) isSelfLeader() bool {
	return sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
}

func (sr *sovereignSubRoundEnd) getAllOutGoingOperations(currentOperations []*sovereign.BridgeOutGoingData) []*sovereign.BridgeOutGoingData {
	outGoingOperations := make([]*sovereign.BridgeOutGoingData, 0)
	unconfirmedOperations := sr.outGoingOperationsPool.GetUnconfirmedOperations()
	if len(unconfirmedOperations) != 0 {
		log.Debug("found unconfirmed operations", "num unconfirmed operations", len(unconfirmedOperations))
		outGoingOperations = append(unconfirmedOperations, outGoingOperations...)
	}

	log.Debug("current outgoing operations", "num", len(currentOperations))
	return append(outGoingOperations, currentOperations...)
}

func (sr *sovereignSubRoundEnd) sendOutGoingOperations(ctx context.Context, data []*sovereign.BridgeOutGoingData) {
	resp, err := sr.bridgeOpHandler.Send(ctx, &sovereign.BridgeOperations{
		Data: data,
	})
	if err != nil {
		log.Error("sovereignSubRoundEnd.doSovereignEndRoundJob.bridgeOpHandler.Send", "error", err)
		return
	}

	sr.resetOutGoingOpTimer(data)
	log.Debug("sent outgoing operations", "hashes", resp.TxHashes)
}

func (sr *sovereignSubRoundEnd) resetOutGoingOpTimer(data []*sovereign.BridgeOutGoingData) {
	hashes := make([][]byte, len(data))
	for idx, dta := range data {
		hashes[idx] = dta.Hash
	}

	sr.outGoingOperationsPool.ResetTimer(hashes)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sr *sovereignSubRoundEnd) IsInterfaceNil() bool {
	return sr == nil
}
