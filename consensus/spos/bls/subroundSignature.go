package bls

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type subroundSignature struct {
	*spos.Subround

	mutExtraSigners      sync.RWMutex
	extraSigners         map[string]SubRoundExtraDataSignatureHandler
	getMessageToSignFunc func() []byte
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	baseSubround *spos.Subround,
	extend func(subroundId int),
) (*subroundSignature, error) {
	err := checkNewSubroundSignatureParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srSignature := subroundSignature{
		Subround:     baseSubround,
		extraSigners: make(map[string]SubRoundExtraDataSignatureHandler),
	}
	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = extend
	srSignature.getMessageToSignFunc = srSignature.getMessageToSign

	return &srSignature, nil
}

func checkNewSubroundSignatureParams(
	baseSubround *spos.Subround,
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}
	if baseSubround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// doSignatureJob method does the job of the subround Signature
func (sr *subroundSignature) doSignatureJob(_ context.Context) bool {
	if !sr.IsNodeInConsensusGroup(sr.SelfPubKey()) && !sr.IsMultiKeyInConsensusGroup() {
		return true
	}
	if !sr.CanDoSubroundJob(sr.Current()) {
		return false
	}
	if check.IfNil(sr.Header) {
		log.Error("doSignatureJob", "error", spos.ErrNilHeader)
		return false
	}

	isSelfLeader := sr.IsSelfLeaderInCurrentRound()

	if isSelfLeader || sr.IsNodeInConsensusGroup(sr.SelfPubKey()) {
		selfIndex, err := sr.SelfConsensusGroupIndex()
		if err != nil {
			log.Debug("doSignatureJob.SelfConsensusGroupIndex: not in consensus group")
			return false
		}

		processedHeaderHash := sr.getMessageToSignFunc()
		signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
			processedHeaderHash,
			uint16(selfIndex),
			sr.Header.GetEpoch(),
			[]byte(sr.SelfPubKey()),
		)
		if err != nil {
			log.Debug("doSignatureJob.CreateSignatureShareForPublicKey", "error", err.Error())
			return false
		}
		extraSigShares, err := sr.createExtraSignatureShares(uint16(selfIndex))
		if err != nil {
			log.Debug("doSignatureJob.extraSigShare.CreateSignatureShare", "error", err.Error())
			return false
		}

		if !isSelfLeader {
			ok := sr.createAndSendSignatureMessage(signatureShare, extraSigShares, []byte(sr.SelfPubKey()))
			if !ok {
				return false
			}
		}

		ok := sr.completeSignatureSubRound(sr.SelfPubKey(), selfIndex, processedHeaderHash, isSelfLeader)
		if !ok {
			return false
		}
	}

	return sr.doSignatureJobForManagedKeys()
}

func (sr *subroundSignature) createExtraSignatureShares(selfIndex uint16) (map[string][]byte, error) {
	ret := make(map[string][]byte)

	sr.mutExtraSigners.RLock()
	for id, extraSigner := range sr.extraSigners {
		extraSigShare, err := extraSigner.CreateSignatureShare(selfIndex)
		if err != nil {
			log.Debug("doSignatureJob.createExtraSignatureShares.CreateSignatureShare", "error", err.Error())
			return nil, err
		}

		ret[id] = extraSigShare
	}
	sr.mutExtraSigners.RUnlock()

	return ret, nil
}

func (sr *subroundSignature) createAndSendSignatureMessage(
	signatureShare []byte,
	extraSigShares map[string][]byte,
	pkBytes []byte,
) bool {
	// TODO: Analyze it is possible to send message only to leader with O(1) instead of O(n)
	cnsMsg := consensus.NewConsensusMessage(
		sr.GetData(),
		signatureShare,
		nil,
		nil,
		pkBytes,
		nil,
		int(MtSignature),
		sr.RoundHandler().Index(),
		sr.ChainID(),
		nil,
		nil,
		nil,
		sr.GetAssociatedPid(pkBytes),
		nil,
		sr.getProcessedHeaderHash(),
	)

	err := sr.addExtraSigSharesToConsensusMessage(extraSigShares, cnsMsg)
	if err != nil {
		log.Debug("createAndSendSignatureMessage.addExtraSigSharesToConsensusMessage",
			"error", err.Error(), "pk", pkBytes)
		return false
	}

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
	if err != nil {
		log.Debug("createAndSendSignatureMessage.BroadcastConsensusMessage",
			"error", err.Error(), "pk", pkBytes)
		return false
	}

	log.Debug("step 2: signature has been sent", "pk", pkBytes)

	return true
}

func (sr *subroundSignature) addExtraSigSharesToConsensusMessage(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error {
	sr.mutExtraSigners.RLock()
	for id, extraSigShare := range extraSigShares {
		// this should never happen, but keep this sanity check anyway
		extraSigner, found := sr.extraSigners[id]
		if !found {
			return fmt.Errorf("extra signed not found for id=%s when trying to add extra sig share to consensus msg", id)
		}

		extraSigner.AddSigShareToConsensusMessage(extraSigShare, cnsMsg)
	}
	sr.mutExtraSigners.RUnlock()

	return nil
}

func (sr *subroundSignature) completeSignatureSubRound(
	pk string,
	index int,
	processedHeaderHash []byte,
	shouldWaitForAllSigsAsync bool,
) bool {
	err := sr.SetJobDone(pk, sr.Current(), true)
	if err != nil {
		log.Debug("doSignatureJob.SetSelfJobDone",
			"subround", sr.Name(),
			"error", err.Error(),
			"pk", []byte(pk),
		)
		return false
	}

	if shouldWaitForAllSigsAsync {
		if sr.EnableEpochHandler().IsConsensusModelV2Enabled() {
			sr.AddProcessedHeadersHashes(processedHeaderHash, index)
		}

		go sr.waitAllSignatures()
	}

	return true
}

func (sr *subroundSignature) getProcessedHeaderHash() []byte {
	if sr.EnableEpochHandler().IsConsensusModelV2Enabled() {
		return sr.getMessageToSignFunc()
	}

	return nil
}

// receivedSignature method is called when a signature is received through the signature channel.
// If the signature is valid, then the jobDone map corresponding to the node which sent it,
// is set on true for the subround Signature
func (sr *subroundSignature) receivedSignature(_ context.Context, cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)
	pkForLogs := core.GetTrimmedPk(hex.EncodeToString(cnsDta.PubKey))

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsNodeInConsensusGroup(node) {
		sr.PeerHonestyHandler().ChangeScore(
			node,
			spos.GetConsensusTopicID(sr.ShardCoordinator()),
			spos.ValidatorPeerHonestyDecreaseFactor,
		)

		return false
	}

	if !sr.IsSelfLeaderInCurrentRound() && !sr.IsMultiKeyLeaderInCurrentRound() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.HeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.RoundHandler().Index(), sr.Current()) {
		return false
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Debug("receivedSignature.ConsensusGroupIndex",
			"node", pkForLogs,
			"error", err.Error())
		return false
	}

	err = sr.SigningHandler().StoreSignatureShare(uint16(index), cnsDta.SignatureShare)
	if err != nil {
		log.Debug("receivedSignature.StoreSignatureShare",
			"node", pkForLogs,
			"index", index,
			"error", err.Error())
		return false
	}

	err = sr.storeExtraSignatureShare(uint16(index), cnsDta)
	if err != nil {
		log.Debug("receivedSignature.storeExtraSignatureShare",
			"node", pkForLogs,
			"index", index,
			"error", err.Error())
		return false
	}

	err = sr.SetJobDone(node, sr.Current(), true)
	if err != nil {
		log.Debug("receivedSignature.SetJobDone",
			"node", pkForLogs,
			"subround", sr.Name(),
			"error", err.Error())
		return false
	}

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.ValidatorPeerHonestyIncreaseFactor,
	)

	if sr.EnableEpochHandler().IsConsensusModelV2Enabled() {
		sr.AddProcessedHeadersHashes(cnsDta.ProcessedHeaderHash, index)
	}

	sr.AppStatusHandler().SetStringValue(common.MetricConsensusRoundState, "signed")
	return true
}

func (sr *subroundSignature) storeExtraSignatureShare(index uint16, cnsMsg *consensus.Message) error {
	sr.mutExtraSigners.RLock()
	defer sr.mutExtraSigners.RUnlock()

	for _, extraSigner := range sr.extraSigners {
		err := extraSigner.StoreSignatureShare(index, cnsMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

// doSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		sr.AppStatusHandler().SetStringValue(common.MetricConsensusRoundState, "signed")

		return true
	}

	isSelfLeader := sr.IsSelfLeaderInCurrentRound() || sr.IsMultiKeyLeaderInCurrentRound()
	isSelfInConsensusGroup := sr.IsNodeInConsensusGroup(sr.SelfPubKey()) || sr.IsMultiKeyInConsensusGroup()

	threshold := sr.Threshold(sr.Current())
	if sr.FallbackHeaderValidator().ShouldApplyFallbackValidation(sr.Header) {
		threshold = sr.FallbackThreshold(sr.Current())
		log.Warn("subroundSignature.doSignatureConsensusCheck: fallback validation has been applied",
			"minimum number of signatures required", threshold,
			"actual number of signatures received", sr.getNumOfSignaturesCollected(),
		)
	}

	areSignaturesCollected, numSigs := sr.areSignaturesCollected(threshold)
	areAllSignaturesCollected := numSigs == sr.ConsensusGroupSize()

	isJobDoneByLeader := isSelfLeader && (areAllSignaturesCollected || (areSignaturesCollected && sr.WaitingAllSignaturesTimeOut))

	selfJobDone := true
	if sr.IsNodeInConsensusGroup(sr.SelfPubKey()) {
		selfJobDone = sr.IsSelfJobDone(sr.Current())
	}
	multiKeyJobDone := true
	if sr.IsMultiKeyInConsensusGroup() {
		multiKeyJobDone = sr.IsMultiKeyJobDone(sr.Current())
	}
	isJobDoneByConsensusNode := !isSelfLeader && isSelfInConsensusGroup && selfJobDone && multiKeyJobDone

	isSubroundFinished := !isSelfInConsensusGroup || isJobDoneByConsensusNode || isJobDoneByLeader

	if isSubroundFinished {
		if isSelfLeader {
			log.Debug("step 2: signatures",
				"received", numSigs,
				"total", len(sr.ConsensusGroup()))
		}

		log.Debug("step 2: subround has been finished",
			"subround", sr.Name())
		sr.SetStatus(sr.Current(), spos.SsFinished)

		sr.AppStatusHandler().SetStringValue(common.MetricConsensusRoundState, "signed")

		return true
	}

	return false
}

// areSignaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are more than the necessary given threshold
func (sr *subroundSignature) areSignaturesCollected(threshold int) (bool, int) {
	n := sr.getNumOfSignaturesCollected()
	return n >= threshold, n
}

func (sr *subroundSignature) getNumOfSignaturesCollected() int {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]

		isSignJobDone, err := sr.JobDone(node, sr.Current())
		if err != nil {
			log.Debug("getNumOfSignaturesCollected.JobDone",
				"node", node,
				"subround", sr.Name(),
				"error", err.Error())
			continue
		}

		if isSignJobDone {
			n++
		}
	}

	return n
}

func (sr *subroundSignature) waitAllSignatures() {
	remainingTime := sr.remainingTime()
	time.Sleep(remainingTime)

	if sr.IsSubroundFinished(sr.Current()) {
		return
	}

	sr.WaitingAllSignaturesTimeOut = true

	select {
	case sr.ConsensusChannel() <- true:
	default:
	}
}

func (sr *subroundSignature) remainingTime() time.Duration {
	startTime := sr.RoundHandler().TimeStamp()
	maxTime := time.Duration(float64(sr.StartTime()) + float64(sr.EndTime()-sr.StartTime())*waitingAllSigsMaxTimeThreshold)
	remainigTime := sr.RoundHandler().RemainingTime(startTime, maxTime)

	return remainigTime
}

func (sr *subroundSignature) doSignatureJobForManagedKeys() bool {
	isMultiKeyLeader := sr.IsMultiKeyLeaderInCurrentRound()

	numMultiKeysSignaturesSent := 0
	for idx, pk := range sr.ConsensusGroup() {
		pkBytes := []byte(pk)
		if sr.IsJobDone(pk, sr.Current()) {
			continue
		}
		if !sr.IsKeyManagedByCurrentNode(pkBytes) {
			continue
		}

		selfIndex, err := sr.ConsensusGroupIndex(pk)
		if err != nil {
			log.Warn("doSignatureJobForManagedKeys: index not found", "pk", pkBytes)
			continue
		}

		processedHeaderHash := sr.getMessageToSignFunc()
		signatureShare, err := sr.SigningHandler().CreateSignatureShareForPublicKey(
			processedHeaderHash,
			uint16(selfIndex),
			sr.Header.GetEpoch(),
			pkBytes,
		)
		if err != nil {
			log.Debug("doSignatureJobForManagedKeys.CreateSignatureShareForPublicKey", "error", err.Error())
			return false
		}

		extraSigShare, err := sr.createExtraSignatureShares(uint16(selfIndex))
		if err != nil {
			log.Debug("doSignatureJobForManagedKeys.extraSigShare.CreateSignatureShare", "error", err.Error())
			return false
		}

		if !isMultiKeyLeader {
			ok := sr.createAndSendSignatureMessage(signatureShare, extraSigShare, pkBytes)
			if !ok {
				return false
			}

			numMultiKeysSignaturesSent++
		}

		isLeader := idx == spos.IndexOfLeaderInConsensusGroup
		ok := sr.completeSignatureSubRound(pk, selfIndex, processedHeaderHash, isLeader)
		if !ok {
			return false
		}
	}

	if numMultiKeysSignaturesSent > 0 {
		log.Debug("step 2: multi keys signatures have been sent", "num", numMultiKeysSignaturesSent)
	}

	return true
}

func (sr *subroundSignature) RegisterExtraSubRoundSigner(extraSigner SubRoundExtraDataSignatureHandler) error {
	if check.IfNil(extraSigner) {
		return errors.ErrNilExtraSubRoundSigner
	}

	id := extraSigner.Identifier()
	log.Debug("subroundSignature.RegisterExtraSubRoundSigner", "identifier", id)

	sr.mutExtraSigners.Lock()
	sr.extraSigners[id] = extraSigner
	sr.mutExtraSigners.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *subroundSignature) IsInterfaceNil() bool {
	return sr == nil
}

func (sr *subroundSignature) getMessageToSign() []byte {
	return sr.GetData()
}
