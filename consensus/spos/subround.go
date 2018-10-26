package spos

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// SRImpl defines the data needed by one subround
type SRImpl struct {
	current chronology.Subround
	next    chronology.Subround
	endTime int64
	name    string

	doSubroundJob    func() bool         // this is a pointer to a function which actually do the job of this subround
	doExtendSubround func()              // this is a pointer to a function which put this subround in the extended mode
	doCheckConsensus func(Subround) bool // this is a pointer to a function which will check the consensus
}

// NewSRImpl creates a new SRImpl object
func NewSRImpl(current chronology.Subround,
	next chronology.Subround,
	endTime int64,
	name string,
	doSubroundJob func() bool,
	doExtendSubround func(),
	doCheckConsensus func(Subround) bool) *SRImpl {

	sr := SRImpl{
		current:          current,
		next:             next,
		endTime:          endTime,
		name:             name,
		doSubroundJob:    doSubroundJob,
		doExtendSubround: doExtendSubround,
		doCheckConsensus: doCheckConsensus}

	return &sr
}

// DoWork method actually does the work of this subround. First it tries to do the job of the subround and than it will
// check the consensus. If the upper time limit of this subround is reached, the subround state is set to extended and
// the chronology will advance to the next subround. This method will iterate until this round will be done,
// put it into extended mode or in canceled mode
func (sr *SRImpl) DoWork(chr *chronology.Chronology) bool {
	for {
		time.Sleep(sleepTime * time.Millisecond)

		timeSubRound := chr.GetSubroundFromDateTime(chr.SyncTime().CurrentTime(chr.ClockOffset()))

		if timeSubRound > sr.current {
			if sr.doExtendSubround != nil {
				sr.doExtendSubround()
			}
			return true // Try to give a chance to this round (extend the current subround)
		}

		if sr.doSubroundJob != nil {
			sr.doSubroundJob()
		}

		if sr.doCheckConsensus != nil {
			if sr.doCheckConsensus(Subround(sr.current)) {
				return true
			}
		}

		if chr.SelfSubround() == chronology.SrCanceled {
			break
		}
	}

	return false
}

// Current method returns the ID of this subround
func (sr *SRImpl) Current() chronology.Subround {
	return sr.current
}

// Next method returns the ID of the next subround
func (sr *SRImpl) Next() chronology.Subround {
	return sr.next
}

// EndTime method returns the upper time limit of this subround
func (sr *SRImpl) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of this subround
func (sr *SRImpl) Name() string {
	return sr.name
}
