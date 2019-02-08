package spos

import (
	"sync"
)

// RoundConsensus defines the data needed by spos to do the consensus in each round
type RoundConsensus struct {
	eligibleList         []string
	consensusGroup       []string
	consensusGroupSize   int
	selfPubKey           string
	validatorRoundStates map[string]*RoundState
	mut                  sync.RWMutex
}

// NewRoundConsensus creates a new RoundConsensus object
func NewRoundConsensus(
	eligibleList []string,
	consensusGroupSize int,
	selfId string,
) *RoundConsensus {

	rcns := RoundConsensus{
		eligibleList:       eligibleList,
		consensusGroupSize: consensusGroupSize,
		selfPubKey:         selfId,
	}

	rcns.validatorRoundStates = make(map[string]*RoundState)

	return &rcns
}

// ConsensusGroupIndex returns the index of given public key in the current consensus group
func (rcns *RoundConsensus) ConsensusGroupIndex(pubKey string) (int, error) {
	for i, pk := range rcns.consensusGroup {
		if pk == pubKey {
			return i, nil
		}
	}
	return 0, ErrSelfNotFoundInConsensus
}

// IndexSelfConsensusGroup returns the index of self public key in current consensus group
func (rcns *RoundConsensus) IndexSelfConsensusGroup() (int, error) {
	for i, pubKey := range rcns.consensusGroup {
		if pubKey == rcns.selfPubKey {
			return i, nil
		}
	}
	return 0, ErrSelfNotFoundInConsensus
}

// EligibleList returns the eligible list ID's
func (rcns *RoundConsensus) EligibleList() []string {
	return rcns.eligibleList
}

// SetEligibleList sets the consensus group ID's
func (rcns *RoundConsensus) SetEligibleList(eligibleList []string) {
	rcns.eligibleList = eligibleList
}

// ConsensusGroup returns the consensus group ID's
func (rcns *RoundConsensus) ConsensusGroup() []string {
	return rcns.consensusGroup
}

// SetConsensusGroup sets the consensus group ID's
func (rcns *RoundConsensus) SetConsensusGroup(consensusGroup []string) {
	rcns.consensusGroup = consensusGroup

	rcns.mut.Lock()

	rcns.validatorRoundStates = make(map[string]*RoundState)

	for i := 0; i < len(consensusGroup); i++ {
		rcns.validatorRoundStates[rcns.consensusGroup[i]] = NewRoundState()
	}

	rcns.mut.Unlock()
}

// ConsensusGroupSize returns the consensus group size
func (rcns *RoundConsensus) ConsensusGroupSize() int {
	return rcns.consensusGroupSize
}

// SetConsensusGroupSize sets the consensus group size
func (rcns *RoundConsensus) SetConsensusGroupSize(consensusGroudpSize int) {
	rcns.consensusGroupSize = consensusGroudpSize
}

// SelfPubKey returns selfPubKey ID
func (rcns *RoundConsensus) SelfPubKey() string {
	return rcns.selfPubKey
}

// SetSelfPubKey sets selfPubKey ID
func (rcns *RoundConsensus) SetSelfPubKey(selfPubKey string) {
	rcns.selfPubKey = selfPubKey
}

// GetJobDone returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rcns *RoundConsensus) GetJobDone(key string, subroundId int) (bool, error) {
	rcns.mut.RLock()
	roundState := rcns.validatorRoundStates[key]

	if roundState == nil {
		rcns.mut.RUnlock()
		return false, ErrInvalidKey
	}

	retcode := roundState.JobDone(subroundId)
	rcns.mut.RUnlock()

	return retcode, nil
}

// SetJobDone set the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rcns *RoundConsensus) SetJobDone(key string, subroundId int, value bool) error {
	rcns.mut.Lock()

	roundState := rcns.validatorRoundStates[key]

	if roundState == nil {
		rcns.mut.Unlock()
		return ErrInvalidKey
	}

	roundState.SetJobDone(subroundId, value)
	rcns.mut.Unlock()

	return nil
}

// GetSelfJobDone returns the self state of the action done in subround given by the subroundId parameter
func (rcns *RoundConsensus) GetSelfJobDone(subroundId int) (bool, error) {
	return rcns.GetJobDone(rcns.selfPubKey, subroundId)
}

// SetSelfJobDone set the self state of the action done in subround given by the subroundId parameter
func (rcns *RoundConsensus) SetSelfJobDone(subroundId int, value bool) error {
	return rcns.SetJobDone(rcns.selfPubKey, subroundId, value)
}

// IsNodeInConsensusGroup method checks if the node is part of the jobDone group of the current round
func (rcns *RoundConsensus) IsNodeInConsensusGroup(node string) bool {
	for i := 0; i < len(rcns.consensusGroup); i++ {
		if rcns.consensusGroup[i] == node {
			return true
		}
	}

	return false
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current jobDone group
// related to this subround
func (rcns *RoundConsensus) ComputeSize(subroundId int) int {
	n := 0

	for i := 0; i < len(rcns.consensusGroup); i++ {
		isJobDone, err := rcns.GetJobDone(rcns.consensusGroup[i], subroundId)

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

// ResetRoundState method resets the state of each node from the current jobDone group, regarding to the
// consensus validatorRoundStates
func (rcns *RoundConsensus) ResetRoundState() {
	for i := 0; i < len(rcns.consensusGroup); i++ {
		rcns.mut.Lock()
		roundState := rcns.validatorRoundStates[rcns.consensusGroup[i]]

		if roundState == nil {
			rcns.mut.Unlock()
			log.Error(ErrNilRoundState.Error())
			continue
		}

		roundState.ResetJobsDone()

		rcns.mut.Unlock()
	}
}
