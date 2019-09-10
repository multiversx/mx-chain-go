package spos

import (
    "sync"
)

// roundConsensus defines the data needed by spos to do the consensus in each round
type roundConsensus struct {
    eligibleList         []string
    consensusGroup       []string
    consensusGroupSize   int
    selfPubKey           string
    validatorRoundStates map[string]*roundState
    mut                  sync.RWMutex
}

// NewRoundConsensus creates a new roundConsensus object
func NewRoundConsensus(
    eligibleList []string,
    consensusGroupSize int,
    selfId string,
) *roundConsensus {

    rcns := roundConsensus{
        eligibleList:       eligibleList,
        consensusGroupSize: consensusGroupSize,
        selfPubKey:         selfId,
    }

    rcns.validatorRoundStates = make(map[string]*roundState)

    return &rcns
}

// ConsensusGroupIndex returns the index of given public key in the current consensus group
func (rcns *roundConsensus) ConsensusGroupIndex(pubKey string) (int, error) {
    for i, pk := range rcns.consensusGroup {
        if pk == pubKey {
            return i, nil
        }
    }
    return 0, ErrNotFoundInConsensus
}

// SelfConsensusGroupIndex returns the index of self public key in current consensus group
func (rcns *roundConsensus) SelfConsensusGroupIndex() (int, error) {
    return rcns.ConsensusGroupIndex(rcns.selfPubKey)
}

// EligibleList returns the eligible list ID's
func (rcns *roundConsensus) EligibleList() []string {
    return rcns.eligibleList
}

// SetEligibleList sets the eligible list ID's
func (rcns *roundConsensus) SetEligibleList(eligibleList []string) {
    rcns.eligibleList = eligibleList
}

// ConsensusGroup returns the consensus group ID's
func (rcns *roundConsensus) ConsensusGroup() []string {
    return rcns.consensusGroup
}

// SetConsensusGroup sets the consensus group ID's
func (rcns *roundConsensus) SetConsensusGroup(consensusGroup []string) {
    rcns.consensusGroup = consensusGroup

    rcns.mut.Lock()

    rcns.validatorRoundStates = make(map[string]*roundState)

    for i := 0; i < len(consensusGroup); i++ {
        rcns.validatorRoundStates[rcns.consensusGroup[i]] = NewRoundState()
    }

    rcns.mut.Unlock()
}

// ConsensusGroupSize returns the consensus group size
func (rcns *roundConsensus) ConsensusGroupSize() int {
    return rcns.consensusGroupSize
}

// SetConsensusGroupSize sets the consensus group size
func (rcns *roundConsensus) SetConsensusGroupSize(consensusGroudpSize int) {
    rcns.consensusGroupSize = consensusGroudpSize
}

// SelfPubKey returns selfPubKey ID
func (rcns *roundConsensus) SelfPubKey() string {
    return rcns.selfPubKey
}

// SetSelfPubKey sets selfPubKey ID
func (rcns *roundConsensus) SetSelfPubKey(selfPubKey string) {
    rcns.selfPubKey = selfPubKey
}

// JobDone returns the state of the action done, by the node represented by the key parameter,
// in subround given by the subroundId parameter
func (rcns *roundConsensus) JobDone(key string, subroundId int) (bool, error) {
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
func (rcns *roundConsensus) SetJobDone(key string, subroundId int, value bool) error {
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

// SelfJobDone returns the self state of the action done in subround given by the subroundId parameter
func (rcns *roundConsensus) SelfJobDone(subroundId int) (bool, error) {
    return rcns.JobDone(rcns.selfPubKey, subroundId)
}

// SetSelfJobDone set the self state of the action done in subround given by the subroundId parameter
func (rcns *roundConsensus) SetSelfJobDone(subroundId int, value bool) error {
    return rcns.SetJobDone(rcns.selfPubKey, subroundId, value)
}

// IsNodeInConsensusGroup method checks if the node is part of consensus group of the current round
func (rcns *roundConsensus) IsNodeInConsensusGroup(node string) bool {
    for i := 0; i < len(rcns.consensusGroup); i++ {
        if rcns.consensusGroup[i] == node {
            return true
        }
    }

    return false
}

// IsNodeInEligibleList method checks if the node is part of the eligible list
func (rcns *roundConsensus) IsNodeInEligibleList(node string) bool {
    for i := 0; i < len(rcns.eligibleList); i++ {
        if rcns.eligibleList[i] == node {
            return true
        }
    }

    return false
}

// ComputeSize method returns the number of messages received from the nodes belonging to the current jobDone group
// related to this subround
func (rcns *roundConsensus) ComputeSize(subroundId int) int {
    n := 0

    for i := 0; i < len(rcns.consensusGroup); i++ {
        isJobDone, err := rcns.JobDone(rcns.consensusGroup[i], subroundId)

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
func (rcns *roundConsensus) ResetRoundState() {
    rcns.mut.Lock()

    for i := 0; i < len(rcns.consensusGroup); i++ {
        roundState := rcns.validatorRoundStates[rcns.consensusGroup[i]]

        if roundState == nil {
            log.Error(ErrNilRoundState.Error())
            continue
        }

        roundState.ResetJobsDone()

    }

    rcns.mut.Unlock()
}
