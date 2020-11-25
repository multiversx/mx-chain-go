package bls

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

type subroundSignature struct {
	*spos.Subround

	appStatusHandler core.AppStatusHandler
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	baseSubround *spos.Subround,
	extend func(subroundId int),
	appStatusHandler core.AppStatusHandler,
) (*subroundSignature, error) {
	err := checkNewSubroundSignatureParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}
	if check.IfNil(appStatusHandler){
		return nil, spos.ErrNilAppStatusHandler
	}

	srSignature := subroundSignature{
		Subround:         baseSubround,
		appStatusHandler: appStatusHandler,
	}
	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = extend

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
func (sr *subroundSignature) doSignatureJob() bool {
	if !sr.IsNodeInConsensusGroup(sr.SelfPubKey()) {
		return true
	}
	if !sr.CanDoSubroundJob(sr.Current()) {
		return false
	}

	signatureShare, err := sr.MultiSigner().CreateSignatureShare(sr.GetData(), nil)
	if err != nil {
		log.Debug("doSignatureJob.CreateSignatureShare", "error", err.Error())
		return false
	}

	isSelfLeader := sr.IsSelfLeaderInCurrentRound()

	if !isSelfLeader {
		//TODO: Analyze it is possible to send message only to leader with O(1) instead of O(n)
		cnsMsg := consensus.NewConsensusMessage(
			sr.GetData(),
			signatureShare,
			nil,
			nil,
			[]byte(sr.SelfPubKey()),
			nil,
			int(MtSignature),
			sr.Rounder().Index(),
			sr.ChainID(),
			nil,
			nil,
			nil,
			sr.CurrentPid(),
		)

		err = sr.BroadcastMessenger().BroadcastConsensusMessage(cnsMsg)
		if err != nil {
			log.Debug("doSignatureJob.BroadcastConsensusMessage", "error", err.Error())
			return false
		}

		log.Debug("step 2: signature has been sent")
	}

	err = sr.SetSelfJobDone(sr.Current(), true)
	if err != nil {
		log.Debug("doSignatureJob.SetSelfJobDone",
			"subround", sr.Name(),
			"error", err.Error())
		return false
	}

	if isSelfLeader {
		go sr.waitAllSignatures()
	}

	return true
}

// receivedSignature method is called when a signature is received through the signature channel.
// If the signature is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Signature
func (sr *subroundSignature) receivedSignature(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

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

	if !sr.IsSelfLeaderInCurrentRound() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), sr.Current()) {
		return false
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Debug("receivedSignature.ConsensusGroupIndex",
			"node", node,
			"error", err.Error())
		return false
	}

	currentMultiSigner := sr.MultiSigner()
	err = currentMultiSigner.StoreSignatureShare(uint16(index), cnsDta.SignatureShare)
	if err != nil {
		log.Debug("receivedSignature.StoreSignatureShare",
			"index", index,
			"error", err.Error())
		return false
	}

	err = sr.SetJobDone(node, sr.Current(), true)
	if err != nil {
		log.Debug("receivedSignature.SetJobDone",
			"node", node,
			"subround", sr.Name(),
			"error", err.Error())
		return false
	}

	sr.PeerHonestyHandler().ChangeScore(
		node,
		spos.GetConsensusTopicID(sr.ShardCoordinator()),
		spos.ValidatorPeerHonestyIncreaseFactor,
	)

	sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "signed")
	return true
}

// doSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "signed")

		return true
	}

	isSelfLeader := sr.IsSelfLeaderInCurrentRound()
	isSelfInConsensusGroup := sr.IsNodeInConsensusGroup(sr.SelfPubKey())

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
	isTimeOut := sr.remainingTime() <= 0

	isJobDoneByLeader := isSelfLeader && (areAllSignaturesCollected || (areSignaturesCollected && isTimeOut))
	isJobDoneByConsensusNode := !isSelfLeader && isSelfInConsensusGroup && sr.IsSelfJobDone(sr.Current())

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

		sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "signed")

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

	select {
	case sr.ConsensusChannel() <- true:
	default:
	}
}

func (sr *subroundSignature) remainingTime() time.Duration {
	startTime := sr.Rounder().TimeStamp()
	maxTime := time.Duration(float64(sr.StartTime()) + float64(sr.EndTime()-sr.StartTime())*waitingAllSigsMaxTimeThreshold)
	remainigTime := sr.Rounder().RemainingTime(startTime, maxTime)

	return remainigTime
}
