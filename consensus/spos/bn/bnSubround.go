package bn

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// subround struct contains the needed data for one subround and the subround properties. It defines a subround
// with it's properties (it's ID, next subround ID, it's duration, it's name) and also it has some handler functions
// which should be set. job function will be the main function of this subround, extend function will handle the overtime
// situation of the subround and check function will decide if in this subround the consensus is achieved
type subround struct {
	previous  int
	current   int
	next      int
	startTime int64
	endTime   int64
	name      string

	consensusStateChangedChannel chan bool

	job    func() bool          // method does the subround job and send the result to the peers
	check  func() bool          // method checks if the consensus of the subround is done
	extend func(subroundId int) // method is called when round time is out
}

// NewSubround creates a new SubroundId object
func NewSubround(
	previous int,
	current int,
	next int,
	startTime int64,
	endTime int64,
	name string,
	consensusStateChangedChannel chan bool,
) (*subround, error) {

	err := checkNewSubroundParams(
		consensusStateChangedChannel,
	)

	if err != nil {
		return nil, err
	}

	sr := subround{
		previous:                     previous,
		current:                      current,
		next:                         next,
		endTime:                      endTime,
		startTime:                    startTime,
		name:                         name,
		consensusStateChangedChannel: consensusStateChangedChannel,
	}

	return &sr, nil
}

func checkNewSubroundParams(
	consensusStateChangedChannel chan bool,
) error {
	if consensusStateChangedChannel == nil {
		return spos.ErrNilChannel
	}

	return nil
}

// DoWork method actually does the work of this subround. First it tries to do the job of the subround then it will
// check the consensus. If the upper time limit of this subround is reached, the extend method will be called before
// returning. If this method returns true the chronology will advance to the next subround.
func (sr *subround) DoWork(rounder consensus.Rounder) bool {
	if sr.job == nil || sr.check == nil {
		return false
	}

	sr.job()

	if sr.check() {
		return true
	}

	for {
		select {
		case <-sr.consensusStateChangedChannel:
			if sr.check() {
				return true
			}
		case <-time.After(rounder.RemainingTimeInRound(maxThresholdPercent)):
			if sr.extend != nil {
				sr.extend(sr.current)
			}

			return false
		}
	}
}

// Previous method returns the ID of the previous subround
func (sr *subround) Previous() int {
	return sr.previous
}

// Current method returns the ID of the current subround
func (sr *subround) Current() int {
	return sr.current
}

// Next method returns the ID of the next subround
func (sr *subround) Next() int {
	return sr.next
}

// StartTime method returns the start time of the subround
func (sr *subround) StartTime() int64 {
	return int64(sr.startTime)
}

// EndTime method returns the upper time limit of the subround
func (sr *subround) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of the subround
func (sr *subround) Name() string {
	return sr.name
}
