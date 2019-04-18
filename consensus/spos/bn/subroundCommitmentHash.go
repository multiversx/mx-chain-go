package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundCommitmentHash struct {
	*subround

	sendConsensusMessage func(*consensus.ConsensusMessage) bool
}

// NewSubroundCommitmentHash creates a subroundCommitmentHash object
func NewSubroundCommitmentHash(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundCommitmentHash, error) {

	err := checkNewSubroundCommitmentHashParams(
		subround,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srCommitmentHash := subroundCommitmentHash{
		subround,
		sendConsensusMessage,
	}

	srCommitmentHash.job = srCommitmentHash.doCommitmentHashJob
	srCommitmentHash.check = srCommitmentHash.doCommitmentHashConsensusCheck
	srCommitmentHash.extend = extend

	return &srCommitmentHash, nil
}

func checkNewSubroundCommitmentHashParams(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	containerValidator := spos.ConsensusContainerValidator{}
	err := containerValidator.ValidateConsensusDataContainer(subround.consensusDataContainer)

	return err
}

// doCommitmentHashJob method does the job of the commitment hash subround
func (sr *subroundCommitmentHash) doCommitmentHashJob() bool {
	if !sr.ConsensusState().CanDoSubroundJob(SrCommitmentHash) {
		return false
	}

	commitmentHash, err := sr.genCommitmentHash()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := consensus.NewConsensusMessage(
		sr.ConsensusState().Data,
		commitmentHash,
		[]byte(sr.ConsensusState().SelfPubKey()),
		nil,
		int(MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 2: commitment hash has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	err = sr.ConsensusState().SetSelfJobDone(SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedCommitmentHash method is called when a commitment hash is received through the commitment hash
// channel. If the commitment hash is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround CommitmentHash
func (sr *subroundCommitmentHash) receivedCommitmentHash(cnsDta *consensus.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.ConsensusState().IsConsensusDataSet() {
		return false
	}

	if !sr.ConsensusState().IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.ConsensusState().IsNodeInConsensusGroup(node) { // is NOT this node in the consensus group?
		return false
	}

	if !sr.ConsensusState().CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrCommitmentHash) {
		return false
	}

	// if this node is leader in this round and it already received 2/3 + 1 of commitment hashes
	// it will ignore any others received later
	if sr.ConsensusState().IsSelfLeaderInCurrentRound() {
		threshold := sr.ConsensusState().Threshold(SrCommitmentHash)
		if sr.isCommitmentHashReceived(threshold) {
			return false
		}
	}

	index, err := sr.ConsensusState().ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.MultiSigner().StoreCommitmentHash(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.ConsensusState().SetJobDone(node, SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.ConsensusState().Threshold(SrCommitmentHash)
	if !sr.ConsensusState().IsSelfLeaderInCurrentRound() {
		threshold = len(sr.ConsensusState().ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		n := sr.ConsensusState().ComputeSize(SrCommitmentHash)
		log.Info(fmt.Sprintf("%sStep 2: received %d from %d commitment hashes\n",
			sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusState().ConsensusGroup())))
	} else {
		threshold = sr.ConsensusState().Threshold(SrBitmap)
		if sr.commitmentHashesCollected(threshold) {
			n := sr.ConsensusState().ComputeSize(SrCommitmentHash)
			log.Info(fmt.Sprintf("%sStep 2: received %d from %d commitment hashes\n",
				sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusState().ConsensusGroup())))
		}
	}

	return true
}

// doCommitmentHashConsensusCheck method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (sr *subroundCommitmentHash) doCommitmentHashConsensusCheck() bool {
	if sr.ConsensusState().RoundCanceled {
		return false
	}

	if sr.ConsensusState().Status(SrCommitmentHash) == spos.SsFinished {
		return true
	}

	threshold := sr.ConsensusState().Threshold(SrCommitmentHash)
	if !sr.ConsensusState().IsSelfLeaderInCurrentRound() {
		threshold = len(sr.ConsensusState().ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 2: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.ConsensusState().SetStatus(SrCommitmentHash, spos.SsFinished)
		return true
	}

	threshold = sr.ConsensusState().Threshold(SrBitmap)
	if sr.commitmentHashesCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 2: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.ConsensusState().SetStatus(SrCommitmentHash, spos.SsFinished)
		return true
	}

	return false
}

// isCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (sr *subroundCommitmentHash) isCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		node := sr.ConsensusState().ConsensusGroup()[i]
		isJobDone, err := sr.ConsensusState().JobDone(node, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
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

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		node := sr.ConsensusState().ConsensusGroup()[i]
		isBitmapJobDone, err := sr.ConsensusState().JobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommHashJobDone, err := sr.ConsensusState().JobDone(node, SrCommitmentHash)

			if err != nil {
				log.Error(err.Error())
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
	_, commitment := sr.MultiSigner().CreateCommitment()

	selfIndex, err := sr.ConsensusState().SelfConsensusGroupIndex()

	if err != nil {
		return nil, err
	}

	commitmentHash := sr.Hasher().Compute(string(commitment))

	err = sr.MultiSigner().StoreCommitmentHash(uint16(selfIndex), commitmentHash)

	if err != nil {
		return nil, err
	}

	return commitmentHash, nil
}
