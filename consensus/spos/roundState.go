package spos

import (
	"sync"
)

// RoundState defines the data needed by spos to know the state of each node from the current jobDone group,
// regarding to the consensus validatorRoundStates in each subround of the current round
type RoundState struct {
	jobDone map[int]bool
	mut     sync.RWMutex
}

// NewRoundState creates a new RoundState object
func NewRoundState() *RoundState {
	rstate := RoundState{}
	rstate.jobDone = make(map[int]bool)
	return &rstate
}

// JobDone returns the consensus validatorRoundStates of the given subroundId
func (rstate *RoundState) JobDone(subroundId int) bool {
	rstate.mut.RLock()
	retcode := rstate.jobDone[subroundId]
	rstate.mut.RUnlock()
	return retcode
}

// SetJobDone sets the consensus validatorRoundStates of the given subroundId
func (rstate *RoundState) SetJobDone(subroundId int, value bool) {
	rstate.mut.Lock()
	rstate.jobDone[subroundId] = value
	rstate.mut.Unlock()
}

// ResetJobsDone method resets the consensus validatorRoundStates of each subround
func (rstate *RoundState) ResetJobsDone() {
	for k := range rstate.jobDone {
		rstate.mut.Lock()
		rstate.jobDone[k] = false
		rstate.mut.Unlock()
	}
}
