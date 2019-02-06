package spos

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// RoundState defines the data needed by spos to know the state of each node from the current jobDone group,
// regarding to the consensus validatorRoundStates in each subround of the current round
type RoundState struct {
	jobDone map[chronology.SubroundId]bool
	mut     sync.RWMutex
}

// NewRoundState creates a new RoundState object
func NewRoundState() *RoundState {
	rstate := RoundState{}
	rstate.jobDone = make(map[chronology.SubroundId]bool)
	return &rstate
}

// SetJobDone sets the consensus validatorRoundStates of the given subroundId
func (rstate *RoundState) SetJobDone(subroundId chronology.SubroundId, value bool) {
	rstate.mut.Lock()
	rstate.jobDone[subroundId] = value
	rstate.mut.Unlock()
}

// JobDone returns the consensus validatorRoundStates of the given subroundId
func (rstate *RoundState) JobDone(subroundId chronology.SubroundId) bool {
	rstate.mut.RLock()
	retcode := rstate.jobDone[subroundId]
	rstate.mut.RUnlock()
	return retcode
}

// ResetJobsDone method resets the consensus validatorRoundStates of each subround
func (rstate *RoundState) ResetJobsDone() {
	for k := range rstate.jobDone {
		rstate.mut.Lock()
		rstate.jobDone[k] = false
		rstate.mut.Unlock()
	}
}
