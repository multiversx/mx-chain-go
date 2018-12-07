package spos

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"sync"
)

//TODO: create unit tests

// RoundState defines the data needed by spos to know the state of each node from the current jobDone group,
// regarding to the consensus validatorRoundStates in each subround of the current round
type RoundState struct {
	jobDone map[chronology.SubroundId]bool
	mut     sync.RWMutex
}

// NewRoundState creates a new RoundState object
func NewRoundState() *RoundState {
	rv := RoundState{}
	rv.jobDone = make(map[chronology.SubroundId]bool)
	return &rv
}

// ResetJobsDone method resets the consensus validatorRoundStates of each subround
func (rv *RoundState) ResetJobsDone() {
	for k := range rv.jobDone {
		rv.mut.Lock()
		rv.jobDone[k] = false
		rv.mut.Unlock()
	}
}

// JobDone returns the consensus validatorRoundStates of the given subroundId
func (rv *RoundState) JobDone(subroundId chronology.SubroundId) bool {
	rv.mut.RLock()
	retcode := rv.jobDone[subroundId]
	rv.mut.RUnlock()
	return retcode
}

// SetJobDone sets the consensus validatorRoundStates of the given subroundId
func (rv *RoundState) SetJobDone(subroundId chronology.SubroundId, value bool) {
	rv.mut.Lock()
	rv.jobDone[subroundId] = value
	rv.mut.Unlock()
}
