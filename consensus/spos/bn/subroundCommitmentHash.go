package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundCommitmentHash struct {
	*subround

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundCommitmentHash creates a subroundCommitmentHash object
func NewSubroundCommitmentHash(
	subround *subround,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
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
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	return nil
}

// doCommitmentHashJob method does the job of the commitment hash subround
func (sr *subroundCommitmentHash) doCommitmentHashJob() bool {
	if !sr.GetConsensusState().CanDoSubroundJob(SrCommitmentHash) {
		return false
	}

	commitmentHash, err := sr.genCommitmentHash()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := spos.NewConsensusMessage(
		sr.GetConsensusState().Data,
		commitmentHash,
		[]byte(sr.GetConsensusState().SelfPubKey()),
		nil,
		int(MtCommitmentHash),
		uint64(sr.GetRounder().TimeStamp().Unix()),
		sr.GetRounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 2: commitment hash has been sent\n", sr.GetSyncTimer().FormattedCurrentTime()))

	err = sr.GetConsensusState().SetSelfJobDone(SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedCommitmentHash method is called when a commitment hash is received through the commitment hash
// channel. If the commitment hash is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround CommitmentHash
func (sr *subroundCommitmentHash) receivedCommitmentHash(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.GetConsensusState().IsConsensusDataSet() {
		return false
	}

	if !sr.GetConsensusState().IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.GetConsensusState().IsNodeInConsensusGroup(node) { // is NOT this node in the consensus group?
		return false
	}

	if !sr.GetConsensusState().CanProcessReceivedMessage(cnsDta, sr.GetRounder().Index(), SrCommitmentHash) {
		return false
	}

	// if this node is leader in this round and it already received 2/3 + 1 of commitment hashes
	// it will ignore any others received later
	if sr.GetConsensusState().IsSelfLeaderInCurrentRound() {
		threshold := sr.GetConsensusState().Threshold(SrCommitmentHash)
		if sr.isCommitmentHashReceived(threshold) {
			return false
		}
	}

	index, err := sr.GetConsensusState().ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.GetMultiSigner().StoreCommitmentHash(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.GetConsensusState().SetJobDone(node, SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.GetConsensusState().Threshold(SrCommitmentHash)
	if !sr.GetConsensusState().IsSelfLeaderInCurrentRound() {
		threshold = len(sr.GetConsensusState().ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		n := sr.GetConsensusState().ComputeSize(SrCommitmentHash)
		log.Info(fmt.Sprintf("%sStep 2: received %d from %d commitment hashes\n",
			sr.GetSyncTimer().FormattedCurrentTime(), n, len(sr.GetConsensusState().ConsensusGroup())))
	} else {
		threshold = sr.GetConsensusState().Threshold(SrBitmap)
		if sr.commitmentHashesCollected(threshold) {
			n := sr.GetConsensusState().ComputeSize(SrCommitmentHash)
			log.Info(fmt.Sprintf("%sStep 2: received %d from %d commitment hashes\n",
				sr.GetSyncTimer().FormattedCurrentTime(), n, len(sr.GetConsensusState().ConsensusGroup())))
		}
	}

	return true
}

// doCommitmentHashConsensusCheck method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (sr *subroundCommitmentHash) doCommitmentHashConsensusCheck() bool {
	if sr.GetConsensusState().RoundCanceled {
		return false
	}

	if sr.GetConsensusState().Status(SrCommitmentHash) == spos.SsFinished {
		return true
	}

	threshold := sr.GetConsensusState().Threshold(SrCommitmentHash)
	if !sr.GetConsensusState().IsSelfLeaderInCurrentRound() {
		threshold = len(sr.GetConsensusState().ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 2: subround %s has been finished\n", sr.GetSyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.GetConsensusState().SetStatus(SrCommitmentHash, spos.SsFinished)
		return true
	}

	threshold = sr.GetConsensusState().Threshold(SrBitmap)
	if sr.commitmentHashesCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 2: subround %s has been finished\n", sr.GetSyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.GetConsensusState().SetStatus(SrCommitmentHash, spos.SsFinished)
		return true
	}

	return false
}

// isCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (sr *subroundCommitmentHash) isCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		node := sr.GetConsensusState().ConsensusGroup()[i]
		isJobDone, err := sr.GetConsensusState().JobDone(node, SrCommitmentHash)

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

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		node := sr.GetConsensusState().ConsensusGroup()[i]
		isBitmapJobDone, err := sr.GetConsensusState().JobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommHashJobDone, err := sr.GetConsensusState().JobDone(node, SrCommitmentHash)

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
	_, commitment := sr.GetMultiSigner().CreateCommitment()

	selfIndex, err := sr.GetConsensusState().SelfConsensusGroupIndex()

	if err != nil {
		return nil, err
	}

	commitmentHash := sr.GetHasher().Compute(string(commitment))

	err = sr.GetMultiSigner().StoreCommitmentHash(uint16(selfIndex), commitmentHash)

	if err != nil {
		return nil, err
	}

	return commitmentHash, nil
}
