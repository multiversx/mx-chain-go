package spos

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// Subround defines the data needed by one subround. Actually it defines a subround with it's properties (it's ID,
// next subround ID, it's duration, it's name and also it has some handler functions which should be set. work funvtion
// will be the main function of this subround, extend function will handle the overtime situation of the subround and
// check function will decide if in this subround the consensus is achieved
type Subround struct {
	current chronology.SubroundId
	next    chronology.SubroundId
	endTime int64
	name    string

	work   func() bool // this is a pointer to a function which actually do the job of this subround
	extend func()      // this is a pointer to a function which put this subround in the extended mode
	check  func() bool // this is a pointer to a function which will check the consensus
}

// NewSubround creates a new SubroundId object
func NewSubround(
	current chronology.SubroundId,
	next chronology.SubroundId,
	endTime int64,
	name string,
	work func() bool,
	extend func(),
	check func() bool,
) *Subround {

	sr := Subround{
		current: current,
		next:    next,
		endTime: endTime,
		name:    name,
		work:    work,
		extend:  extend,
		check:   check,
	}

	return &sr
}

// DoWork method actually does the work of this subround. First it tries to do the job of the subround and than it will
// check the consensus. If the upper time limit of this subround is reached, the subround state is set to extended and
// the chronology will advance to the next subround. This method will iterate until this round will be done,
// put it into extended mode or in canceled mode
func (sr *Subround) DoWork(computeSubRoundId func() chronology.SubroundId, isCancelled func() bool) bool {
	for {
		time.Sleep(sleepTime)

		subRoundId := computeSubRoundId()

		if subRoundId == chronology.SubroundId(-1) {
			break
		}

		if subRoundId > sr.current {
			if sr.extend != nil {
				sr.extend()
			}
			return true // Try to give a chance to this round (extend the current subround)
		}

		if sr.work != nil {
			sr.work()
		}

		if sr.check != nil {
			if sr.check() {
				return true
			}
		}

		if isCancelled() {
			break
		}
	}

	return false
}

// Current method returns the ID of this subround
func (sr *Subround) Current() chronology.SubroundId {
	return sr.current
}

// Next method returns the ID of the next subround
func (sr *Subround) Next() chronology.SubroundId {
	return sr.next
}

// EndTime method returns the upper time limit of this subround
func (sr *Subround) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *Subround) Name() string {
	return sr.name
}

func (sr *Subround) Check() bool {
	return sr.check()
}
