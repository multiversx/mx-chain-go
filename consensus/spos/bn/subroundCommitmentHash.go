package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
)

type subroundCommitmentHash struct {
	*subround

	consensusState *spos.ConsensusState
	hasher         hashing.Hasher
	multiSigner    crypto.MultiSigner
	rounder        consensus.Rounder
	syncTimer      ntp.SyncTimer

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundCommitmentHash creates a subroundCommitmentHash object
func NewSubroundCommitmentHash(
	subround *subround,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundCommitmentHash, error) {

	err := checkNewSubroundCommitmentHashParams(
		subround,
		consensusState,
		hasher,
		multiSigner,
		rounder,
		syncTimer,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srCommitmentHash := subroundCommitmentHash{
		subround,
		consensusState,
		hasher,
		multiSigner,
		rounder,
		syncTimer,
		sendConsensusMessage,
	}

	srCommitmentHash.job = srCommitmentHash.doCommitmentHashJob
	srCommitmentHash.check = srCommitmentHash.doCommitmentHashConsensusCheck
	srCommitmentHash.extend = extend

	return &srCommitmentHash, nil
}

func checkNewSubroundCommitmentHashParams(
	subround *subround,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
	}

	if hasher == nil {
		return spos.ErrNilHasher
	}

	if multiSigner == nil {
		return spos.ErrNilMultiSigner
	}

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	return nil
}

// doCommitmentHashJob method does the job of the commitment hash subround
func (sr *subroundCommitmentHash) doCommitmentHashJob() bool {
	if !sr.consensusState.CanDoSubroundJob(SrCommitmentHash) {
		return false
	}

	commitmentHash, err := sr.genCommitmentHash()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := spos.NewConsensusMessage(
		sr.consensusState.Data,
		commitmentHash,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtCommitmentHash),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 2: commitment hash has been sent\n", sr.syncTimer.FormattedCurrentTime()))

	err = sr.consensusState.SetSelfJobDone(SrCommitmentHash, true)

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

	if !sr.consensusState.IsConsensusDataSet() {
		return false
	}

	if !sr.consensusState.IsNodeInConsensusGroup(node) { // is NOT this node in the consensus group?
		return false
	}

	if !sr.consensusState.CanProcessReceivedMessage(cnsDta, sr.rounder.Index(), SrCommitmentHash) {
		return false
	}

	// if this node is leader in this round and it already received 2/3 + 1 of commitment hashes
	// it will ignore any others received later
	if sr.consensusState.IsSelfLeaderInCurrentRound() {
		threshold := sr.consensusState.Threshold(SrCommitmentHash)
		if sr.isCommitmentHashReceived(threshold) {
			return false
		}
	}

	index, err := sr.consensusState.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.multiSigner.StoreCommitmentHash(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.consensusState.SetJobDone(node, SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// doCommitmentHashConsensusCheck method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (sr *subroundCommitmentHash) doCommitmentHashConsensusCheck() bool {
	if sr.consensusState.RoundCanceled {
		return false
	}

	if sr.consensusState.Status(SrCommitmentHash) == spos.SsFinished {
		return true
	}

	threshold := sr.consensusState.Threshold(SrCommitmentHash)

	if !sr.consensusState.IsSelfLeaderInCurrentRound() {
		threshold = len(sr.consensusState.ConsensusGroup())
	}

	if sr.isCommitmentHashReceived(threshold) {
		n := sr.consensusState.ComputeSize(SrCommitmentHash)

		if n == len(sr.consensusState.ConsensusGroup()) {
			log.Info(fmt.Sprintf("%sStep 2: received all (%d from %d) commitment hashes\n",
				sr.syncTimer.FormattedCurrentTime(), n, len(sr.consensusState.ConsensusGroup())))
		} else {
			log.Info(fmt.Sprintf("%sStep 2: received %d from %d commitment hashes, which are enough\n",
				sr.syncTimer.FormattedCurrentTime(), n, len(sr.consensusState.ConsensusGroup())))
		}

		log.Info(fmt.Sprintf("%sStep 2: subround %s has been finished\n", sr.syncTimer.FormattedCurrentTime(), sr.Name()))

		sr.consensusState.SetStatus(SrCommitmentHash, spos.SsFinished)

		return true
	}

	threshold = sr.consensusState.Threshold(SrBitmap)

	if sr.commitmentHashesCollected(threshold) {
		n := sr.consensusState.ComputeSize(SrCommitmentHash)

		if n == len(sr.consensusState.ConsensusGroup()) {
			log.Info(fmt.Sprintf("%sStep 2: received all (%d from %d) commitment hashes\n",
				sr.syncTimer.FormattedCurrentTime(), n, len(sr.consensusState.ConsensusGroup())))
		} else {
			log.Info(fmt.Sprintf("%sStep 2: received %d from %d commitment hashes, which are enough\n",
				sr.syncTimer.FormattedCurrentTime(), n, len(sr.consensusState.ConsensusGroup())))
		}

		log.Info(fmt.Sprintf("%sStep 2: subround %s has been finished\n", sr.syncTimer.FormattedCurrentTime(), sr.Name()))

		sr.consensusState.SetStatus(SrCommitmentHash, spos.SsFinished)

		return true
	}

	return false
}

// isCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (sr *subroundCommitmentHash) isCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		node := sr.consensusState.ConsensusGroup()[i]
		isJobDone, err := sr.consensusState.JobDone(node, SrCommitmentHash)

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

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		node := sr.consensusState.ConsensusGroup()[i]
		isBitmapJobDone, err := sr.consensusState.JobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommHashJobDone, err := sr.consensusState.JobDone(node, SrCommitmentHash)

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
	_, commitment := sr.multiSigner.CreateCommitment()

	selfIndex, err := sr.consensusState.SelfConsensusGroupIndex()

	if err != nil {
		return nil, err
	}

	commitmentHash := sr.hasher.Compute(string(commitment))

	err = sr.multiSigner.StoreCommitmentHash(uint16(selfIndex), commitmentHash)

	if err != nil {
		return nil, err
	}

	return commitmentHash, nil
}
