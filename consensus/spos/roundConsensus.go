package spos

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// RoundConsensus defines the data needed by spos to do the consensus in each round
type RoundConsensus struct {
	consensusGroup       []string
	selfPubKey           string
	validatorRoundStates map[string]*RoundState
	mut                  sync.RWMutex
}

// ConsensusGroup returns the consensus group ID's
func (rCns *RoundConsensus) ConsensusGroup() []string {
	return rCns.consensusGroup
}

// ConsensusGroupIndex returns the index of given public key in the current consensus group
func (rCns *RoundConsensus) ConsensusGroupIndex(pubKey string) (int, error) {
	for i, pk := range rCns.consensusGroup {
		if pk == pubKey {
			return i, nil
		}
	}
	return 0, ErrSelfNotFoundInConsensus
}

// IndexSelfConsensusGroup returns the index of self public key in current consensus group
func (rCns *RoundConsensus) IndexSelfConsensusGroup() (int, error) {
	for i, pubKey := range rCns.consensusGroup {
		if pubKey == rCns.selfPubKey {
			return i, nil
		}
	}
	return 0, ErrSelfNotFoundInConsensus
}

// SetConsensusGroup sets the consensus group ID's
func (rCns *RoundConsensus) SetConsensusGroup(consensusGroup []string) {
	rCns.consensusGroup = consensusGroup

	rCns.mut.Lock()

	rCns.validatorRoundStates = make(map[string]*RoundState)

	for i := 0; i < len(consensusGroup); i++ {
		rCns.validatorRoundStates[rCns.consensusGroup[i]] = NewRoundState()
	}

	rCns.mut.Unlock()
}

// SelfPubKey returns selfPubKey ID
func (rCns *RoundConsensus) SelfPubKey() string {
	return rCns.selfPubKey
}

// SetSelfPubKey sets selfPubKey ID
func (rCns *RoundConsensus) SetSelfPubKey(selfPubKey string) {
	rCns.selfPubKey = selfPubKey
}

// NewRoundConsensus creates a new RoundConsensus object
func NewRoundConsensus(
	consensusGroup []string,
	selfId string,
) *RoundConsensus {

	v := RoundConsensus{
		consensusGroup: consensusGroup,
		selfPubKey:     selfId,
	}

	v.validatorRoundStates = make(map[string]*RoundState)

	for i := 0; i < len(consensusGroup); i++ {
		v.validatorRoundStates[v.consensusGroup[i]] = NewRoundState()
	}

	return &v
}

// GetJobDone returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rCns *RoundConsensus) GetJobDone(key string, subroundId chronology.SubroundId) (bool, error) {
	rCns.mut.RLock()
	roundState := rCns.validatorRoundStates[key]

	if roundState == nil {
		rCns.mut.RUnlock()
		return false, ErrInvalidKey
	}

	retcode := roundState.JobDone(subroundId)
	rCns.mut.RUnlock()

	return retcode, nil
}

// GetSelfJobDone returns the self state of the action done in subround given by the subroundId parameter
func (rCns *RoundConsensus) GetSelfJobDone(subroundId chronology.SubroundId) (bool, error) {
	return rCns.GetJobDone(rCns.selfPubKey, subroundId)
}

// SetJobDone set the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rCns *RoundConsensus) SetJobDone(key string, subroundId chronology.SubroundId, value bool) error {
	rCns.mut.Lock()

	roundState := rCns.validatorRoundStates[key]

	if roundState == nil {
		rCns.mut.Unlock()
		return ErrInvalidKey
	}

	roundState.SetJobDone(subroundId, value)
	rCns.mut.Unlock()

	return nil
}

// SetSelfJobDone set the self state of the action done in subround given by the subroundId parameter
func (rCns *RoundConsensus) SetSelfJobDone(subroundId chronology.SubroundId, value bool) error {
	return rCns.SetJobDone(rCns.selfPubKey, subroundId, value)
}

// ResetRoundState method resets the state of each node from the current jobDone group, regarding to the
// consensus validatorRoundStates
func (rCns *RoundConsensus) ResetRoundState() {
	for i := 0; i < len(rCns.consensusGroup); i++ {
		rCns.mut.Lock()
		roundState := rCns.validatorRoundStates[rCns.consensusGroup[i]]

		if roundState == nil {
			rCns.mut.Unlock()
			log.Error(ErrNilRoundState.Error())
			continue
		}

		roundState.ResetJobsDone()

		rCns.mut.Unlock()
	}
}

// IsValidatorInBitmap method checks if the node is part of the bitmap received from leader
func (rCns *RoundConsensus) IsValidatorInBitmap(validator string) bool {
	isJobDone, err := rCns.GetJobDone(validator, SrBitmap)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return isJobDone
}

// IsSelfInBitmap method checks if the current node is part of the bitmap received from leader
func (rCns *RoundConsensus) IsSelfInBitmap() bool {
	return rCns.IsValidatorInBitmap(rCns.selfPubKey)
}

// IsNodeInConsensusGroup method checks if the node is part of the jobDone group of the current round
func (rCns *RoundConsensus) IsNodeInConsensusGroup(node string) bool {
	for i := 0; i < len(rCns.consensusGroup); i++ {
		if rCns.consensusGroup[i] == node {
			return true
		}
	}

	return false
}

// IsBlockReceived method checks if the block was received from the leader in the current round
func (rCns *RoundConsensus) IsBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		isJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrBlock)

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

// IsCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (rCns *RoundConsensus) IsCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		isJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrCommitmentHash)

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

// CommitmentHashesCollected method checks if the commitment hashes received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (rCns *RoundConsensus) CommitmentHashesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		isBitmapJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommHashJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrCommitmentHash)

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

// CommitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (rCns *RoundConsensus) CommitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		isBitmapJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrCommitment)

			if err != nil {
				log.Error(err.Error())
				continue
			}

			if !isCommJobDone {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// SignaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (rCns *RoundConsensus) SignaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		isBitmapJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {

			isSignJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], SrSignature)

			if err != nil {
				log.Error(err.Error())
				continue
			}

			if !isSignJobDone {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current jobDone group
// related to this subround
func (rCns *RoundConsensus) ComputeSize(subroundId chronology.SubroundId) int {
	n := 0

	for i := 0; i < len(rCns.consensusGroup); i++ {
		isJobDone, err := rCns.GetJobDone(rCns.consensusGroup[i], subroundId)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobDone {
			n++
		}
	}

	return n
}
