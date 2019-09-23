package spos

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
)

// Subround struct contains the needed data for one Subround and the Subround properties. It defines a Subround
// with it's properties (it's ID, next Subround ID, it's duration, it's name) and also it has some handler functions
// which should be set. Job function will be the main function of this Subround, Extend function will handle the overtime
// situation of the Subround and Check function will decide if in this Subround the consensus is achieved
type Subround struct {
	ConsensusCoreHandler
	*ConsensusState

	previous  int
	current   int
	next      int
	startTime int64
	endTime   int64
	name      string

	consensusStateChangedChannel chan bool
	executeStoredMessages        func()

	Job    func() bool          // method does the Subround Job and send the result to the peers
	Check  func() bool          // method checks if the consensus of the Subround is done
	Extend func(subroundId int) // method is called when round time is out
}

// NewSubround creates a new SubroundId object
func NewSubround(
	previous int,
	current int,
	next int,
	startTime int64,
	endTime int64,
	name string,
	consensusState *ConsensusState,
	consensusStateChangedChannel chan bool,
	executeStoredMessages func(),
	container ConsensusCoreHandler,
) (*Subround, error) {
	err := checkNewSubroundParams(
		consensusState,
		consensusStateChangedChannel,
		executeStoredMessages,
		container,
	)
	if err != nil {
		return nil, err
	}

	sr := Subround{
		container,
		consensusState,
		previous,
		current,
		next,
		startTime,
		endTime,
		name,
		consensusStateChangedChannel,
		executeStoredMessages,
		nil,
		nil,
		nil,
	}

	return &sr, nil
}

func checkNewSubroundParams(
	state *ConsensusState,
	consensusStateChangedChannel chan bool,
	executeStoredMessages func(),
	container ConsensusCoreHandler,
) error {
	err := ValidateConsensusCore(container)
	if err != nil {
		return err
	}
	if consensusStateChangedChannel == nil {
		return ErrNilChannel
	}
	if state == nil {
		return ErrNilConsensusState
	}
	if executeStoredMessages == nil {
		return ErrNilExecuteStoredMessages
	}

	return nil
}

// DoWork method actually does the work of this Subround. First it tries to do the Job of the Subround then it will
// Check the consensus. If the upper time limit of this Subround is reached, the Extend method will be called before
// returning. If this method returns true the chronology will advance to the next Subround.
func (sr *Subround) DoWork(rounder consensus.Rounder) bool {
	if sr.Job == nil || sr.Check == nil {
		return false
	}

	// execute stored messages which were received in this new round but before this initialisation
	go sr.executeStoredMessages()

	startTime := time.Time{}
	startTime = rounder.TimeStamp()
	maxTime := rounder.TimeDuration() * maxThresholdPercent / 100

	sr.Job()
	if sr.Check() {
		return true
	}

	for {
		select {
		case <-sr.consensusStateChangedChannel:
			if sr.Check() {
				return true
			}
		case <-time.After(rounder.RemainingTime(startTime, maxTime)):
			if sr.Extend != nil {
				sr.Extend(sr.current)
			}

			return false
		}
	}
}

// Previous method returns the ID of the previous Subround
func (sr *Subround) Previous() int {
	return sr.previous
}

// Current method returns the ID of the current Subround
func (sr *Subround) Current() int {
	return sr.current
}

// Next method returns the ID of the next Subround
func (sr *Subround) Next() int {
	return sr.next
}

// StartTime method returns the start time of the Subround
func (sr *Subround) StartTime() int64 {
	return int64(sr.startTime)
}

// EndTime method returns the upper time limit of the Subround
func (sr *Subround) EndTime() int64 {
	return int64(sr.endTime)
}

// Name method returns the name of the Subround
func (sr *Subround) Name() string {
	return sr.name
}

// IsInterfaceNil returns true if there is no value under the interface
func (sr *Subround) IsInterfaceNil() bool {
	if sr == nil {
		return true
	}
	return false
}
