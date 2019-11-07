package bn

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

type subroundCommitmentHash struct {
	*spos.Subround
}

// NewSubroundCommitmentHash creates a subroundCommitmentHash object
func NewSubroundCommitmentHash(
	baseSubround *spos.Subround,
	extend func(subroundId int),
) (*subroundCommitmentHash, error) {
	err := checkNewSubroundCommitmentHashParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srCommitmentHash := subroundCommitmentHash{
		baseSubround,
	}
	srCommitmentHash.Job = srCommitmentHash.doCommitmentHashJob
	srCommitmentHash.Check = srCommitmentHash.doCommitmentHashConsensusCheck
	srCommitmentHash.Extend = extend

	return &srCommitmentHash, nil
}

func checkNewSubroundCommitmentHashParams(
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

// doCommitmentHashJob method does the job of the subround CommitmentHash
func (sr *subroundCommitmentHash) doCommitmentHashJob() bool {
	if !sr.CanDoSubroundJob(SrCommitmentHash) {
		return false
	}

	commitmentHash, err := sr.genCommitmentHash()
	if err != nil {
		log.Debug("genCommitmentHash", "type", "spos/bn", "error", err.Error())
		return false
	}

	msg := consensus.NewConsensusMessage(
		sr.Data,
		commitmentHash,
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
	if err != nil {
		log.Debug("BroadcastConsensusMessage", "type", "spos/bn", "error", err.Error())
		return false
	}

	log.Debug("step 2: commitment hash has been sent",
		"type", "spos/bn",
		"time [s]", sr.SyncTimer().FormattedCurrentTime())

	err = sr.SetSelfJobDone(SrCommitmentHash, true)
	if err != nil {
		log.Debug("SetSelfJobDone", "type", "spos/bn", "error", err.Error())
		return false
	}

	return true
}

// receivedCommitmentHash method is called when a commitment hash is received through the commitment hash
// channel. If the commitment hash is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround CommitmentHash
func (sr *subroundCommitmentHash) receivedCommitmentHash(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.IsNodeInConsensusGroup(node) { // is NOT this node in the consensus group?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrCommitmentHash) {
		return false
	}

	// if this node is leader in this round and it already received 2/3 + 1 of commitment hashes
	// it will ignore any others received later
	if sr.IsSelfLeaderInCurrentRound() {
		threshold := sr.Threshold(SrCommitmentHash)
		if sr.isCommitmentHashReceived(threshold) {
			return false
		}
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Debug("ConsensusGroupIndex", "type", "spos/bn", "error", err.Error())
		return false
	}

	currentMultiSigner, err := getBnMultiSigner(sr.MultiSigner())
	if err != nil {
		log.Debug("getBnMultiSigner", "type", "spos/bn", "error", err.Error())
		return false
	}

	err = currentMultiSigner.StoreCommitmentHash(uint16(index), cnsDta.SubRoundData)
	if err != nil {
		log.Debug("StoreCommitmentHash", "type", "spos/bn", "error", err.Error())
		return false
	}

	err = sr.SetJobDone(node, SrCommitmentHash, true)
	if err != nil {
		log.Debug("SetJobDone", "type", "spos/bn", "error", err.Error())
		return false
	}

	threshold := sr.Threshold(SrCommitmentHash)
	if !sr.IsSelfLeaderInCurrentRound() {
		threshold = len(sr.ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		n := sr.ComputeSize(SrCommitmentHash)
		log.Debug("step 2: received commitment hashes",
			"type", "spos/bn",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"received", n,
			"total", len(sr.ConsensusGroup()))
	} else {
		threshold = sr.Threshold(SrBitmap)
		if sr.commitmentHashesCollected(threshold) {
			n := sr.ComputeSize(SrCommitmentHash)
			log.Debug("step 2: received commitment hashes",
				"type", "spos/bn",
				"time [s]", sr.SyncTimer().FormattedCurrentTime(),
				"received", n,
				"total", len(sr.ConsensusGroup()))
		}
	}

	return true
}

// doCommitmentHashConsensusCheck method checks if the consensus in the subround CommitmentHash is achieved
func (sr *subroundCommitmentHash) doCommitmentHashConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrCommitmentHash) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrCommitmentHash)
	if !sr.IsSelfLeaderInCurrentRound() {
		threshold = len(sr.ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		log.Debug("step 2: subround has been finished",
			"type", "spos/bn",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"subround", sr.Name())
		sr.SetStatus(SrCommitmentHash, spos.SsFinished)
		return true
	}

	threshold = sr.Threshold(SrBitmap)
	if sr.commitmentHashesCollected(threshold) {
		log.Debug("step 2: subround has been finished",
			"type", "spos/bn",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"subround", sr.Name())
		sr.SetStatus(SrCommitmentHash, spos.SsFinished)
		return true
	}

	return false
}

// isCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (sr *subroundCommitmentHash) isCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isJobDone, err := sr.JobDone(node, SrCommitmentHash)
		if err != nil {
			log.Debug("JobDone SrCommitmentHash", "type", "spos/bn", "error", err.Error())
			continue
		}

		if isJobDone {
			n++
		}
	}

	return n >= threshold
}

// commitmentHashesCollected method checks if the commitment hashes received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundCommitmentHash) commitmentHashesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isBitmapJobDone, err := sr.JobDone(node, SrBitmap)
		if err != nil {
			log.Debug("JobDone SrBitmap", "type", "spos/bn", "error", err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommHashJobDone, err := sr.JobDone(node, SrCommitmentHash)
			if err != nil {
				log.Debug("JobDone SrCommitmentHash", "type", "spos/bn", "error", err.Error())
				continue
			}

			if !isCommHashJobDone {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

func (sr *subroundCommitmentHash) genCommitmentHash() ([]byte, error) {
	currentMultiSigner, err := getBnMultiSigner(sr.MultiSigner())
	if err != nil {
		return nil, err
	}

	_, commitment := currentMultiSigner.CreateCommitment()

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		return nil, err
	}

	commitmentHash := sr.Hasher().Compute(string(commitment))
	err = currentMultiSigner.StoreCommitmentHash(uint16(selfIndex), commitmentHash)
	if err != nil {
		return nil, err
	}

	return commitmentHash, nil
}
