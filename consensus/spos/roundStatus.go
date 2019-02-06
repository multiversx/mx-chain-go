package spos

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SubroundStatus defines the type used to refer the state of the current subround
type SubroundStatus int

const (
	// SsNotFinished defines the un-finished state of the subround
	SsNotFinished SubroundStatus = iota
	// SsExtended defines the extended state of the subround
	SsExtended
	// SsFinished defines the finished state of the subround
	SsFinished
)

// RoundStatus defines the data needed by spos to know the state of each subround in the current round
type RoundStatus struct {
	status map[chronology.SubroundId]SubroundStatus
	mut    sync.RWMutex
}

// NewRoundStatus creates a new RoundStatus object
func NewRoundStatus() *RoundStatus {
	rstatus := RoundStatus{}
	rstatus.status = make(map[chronology.SubroundId]SubroundStatus)
	return &rstatus
}

// ResetRoundStatus method resets the state of each subround
func (rstatus *RoundStatus) ResetRoundStatus() {
	for k := range rstatus.status {
		rstatus.mut.Lock()
		rstatus.status[k] = SsNotFinished
		rstatus.mut.Unlock()
	}
}

// Status returns the status of the given subround id
func (rstatus *RoundStatus) Status(subroundId chronology.SubroundId) SubroundStatus {
	rstatus.mut.RLock()
	retcode := rstatus.status[subroundId]
	rstatus.mut.RUnlock()
	return retcode
}

// SetStatus sets the status of the given subround id
func (rstatus *RoundStatus) SetStatus(subroundId chronology.SubroundId, subroundStatus SubroundStatus) {
	rstatus.mut.Lock()
	rstatus.status[subroundId] = subroundStatus
	rstatus.mut.Unlock()
}
