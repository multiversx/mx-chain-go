package spos

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SubroundImpl defines the data needed by one subround. Actually it defines a subround with it's properties (it's ID,
// next subround ID, it's duration, it's name and also it has some handler functions which should be set. work funvtion
// will be the main function of this subround, extend function will handle the overtime situation of the subround and
// check function will decide if in this subround the consensus is achieved
type SubroundImpl struct {
	current chronology.Subround
	next    chronology.Subround
	endTime int64
	name    string

	work   func() bool         // this is a pointer to a function which actually do the job of this subround
	extend func()              // this is a pointer to a function which put this subround in the extended mode
	check  func(Subround) bool // this is a pointer to a function which will check the consensus
}

// NewSubroundImpl creates a new SubroundImpl object
func NewSubroundImpl(current chronology.Subround,
	next chronology.Subround,
	endTime int64,
	name string,
	work func() bool,
	extend func(),
	check func(Subround) bool) *SubroundImpl {

	sr := SubroundImpl{
		current: current,
		next:    next,
		endTime: endTime,
		name:    name,
		work:    work,
		extend:  extend,
		check:   check}

	return &sr
}

// DoWork method actually does the work of this subround. First it tries to do the job of the subround and than it will
// check the consensus. If the upper time limit of this subround is reached, the subround state is set to extended and
// the chronology will advance to the next subround. This method will iterate until this round will be done,
// put it into extended mode or in canceled mode
func (sr *SubroundImpl) DoWork(chr *chronology.Chronology) bool {
	for {
		time.Sleep(sleepTime)

		timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

		if timeSubRound == chronology.Subround(-1) {
			break
		}

		if timeSubRound > sr.current {
			if sr.extend != nil {
				sr.extend()
			}
			return true // Try to give a chance to this round (extend the current subround)
		}

		if sr.work != nil {
			sr.work()
		}

		if sr.check != nil {
			if sr.check(Subround(sr.current)) {
				return true
			}
		}

		if chr.SelfSubround() == chronology.Subround(-1) {
			break
		}
	}

	return false
}

// Current method returns the ID of this subround
func (sr *SubroundImpl) Current() chronology.Subround {
	return sr.current
}

// Next method returns the ID of the next subround
func (sr *SubroundImpl) Next() chronology.Subround {
	return sr.next
}

// EndTime method returns the upper time limit of this subround
func (sr *SubroundImpl) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SubroundImpl) Name() string {
	return sr.name
}
