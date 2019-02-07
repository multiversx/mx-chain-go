package spos

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
)

// sleepTime defines the time in milliseconds between each iteration made in DoWork methods of the subrounds
const sleepTime = time.Duration(5 * time.Millisecond)

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
	rs := RoundStatus{}
	rs.status = make(map[chronology.SubroundId]SubroundStatus)
	return &rs
}

// ResetRoundStatus method resets the state of each subround
func (rs *RoundStatus) ResetRoundStatus() {
	for k := range rs.status {
		rs.mut.Lock()
		rs.status[k] = SsNotFinished
		rs.mut.Unlock()
	}
}

// Status returns the status of the given subround id
func (rs *RoundStatus) Status(subroundId chronology.SubroundId) SubroundStatus {
	rs.mut.RLock()
	retcode := rs.status[subroundId]
	rs.mut.RUnlock()
	return retcode
}

// SetStatus sets the status of the given subround id
func (rs *RoundStatus) SetStatus(subroundId chronology.SubroundId, subroundStatus SubroundStatus) {
	rs.mut.Lock()
	rs.status[subroundId] = subroundStatus
	rs.mut.Unlock()
}

// RoundThreshold defines the minimum agreements needed for each subround to consider the subround finished.
// (Ex: PBFT threshold has 2 / 3 + 1 agreements)
type RoundThreshold struct {
	threshold map[chronology.SubroundId]int
	mut       sync.RWMutex
}

// NewRoundThreshold creates a new RoundThreshold object
func NewRoundThreshold() *RoundThreshold {
	rt := RoundThreshold{}
	rt.threshold = make(map[chronology.SubroundId]int)
	return &rt
}

// Threshold returns the threshold of agreements needed in the given subround id
func (rt *RoundThreshold) Threshold(subroundId chronology.SubroundId) int {
	rt.mut.RLock()
	retcode := rt.threshold[subroundId]
	rt.mut.RUnlock()
	return retcode
}

// SetThreshold sets the threshold of agreements needed in the given subround id
func (rt *RoundThreshold) SetThreshold(subroundId chronology.SubroundId, threshold int) {
	rt.mut.Lock()
	rt.threshold[subroundId] = threshold
	rt.mut.Unlock()
}

// Consensus defines the data needed by spos to do the consensus in each round
type Consensus struct {
	Data []byte // hold the data on which validators do the consensus
	// (could be for example a hash of the block header proposed by the leader)
	*RoundConsensus
	*RoundThreshold
	*RoundStatus

	shouldCheckConsensus bool

	Chr *chronology.Chronology
	mut sync.Mutex
}

// NewConsensus creates a new Consensus object
func NewConsensus(
	data []byte,
	vld *RoundConsensus,
	thr *RoundThreshold,
	rs *RoundStatus,
	chr *chronology.Chronology,
) *Consensus {

	cns := Consensus{
		Data:           data,
		RoundConsensus: vld,
		RoundThreshold: thr,
		RoundStatus:    rs,
		Chr:            chr,
	}

	return &cns
}

// IsSelfLeaderInCurrentRound method checks if the current node is leader in the current round
func (cns *Consensus) IsSelfLeaderInCurrentRound() bool {
	return cns.IsNodeLeaderInCurrentRound(cns.selfPubKey)
}

// IsNodeLeaderInCurrentRound method checks if the given node is leader in the current round
func (cns *Consensus) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := cns.GetLeader()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return leader == node
}

// GetLeader method gets the leader of the current round
func (cns *Consensus) GetLeader() (string, error) {
	if cns.Chr == nil {
		return "", ErrNilChronology
	}

	if cns.Chr.Round() == nil {
		return "", ErrNilRound
	}

	if cns.Chr.Round().Index() < 0 {
		return "", ErrNegativeRoundIndex
	}

	if cns.consensusGroup == nil {
		return "", ErrNilConsensusGroup
	}

	if len(cns.consensusGroup) == 0 {
		return "", ErrEmptyConsensusGroup
	}

	index := cns.Chr.Round().Index() % int32(len(cns.consensusGroup))
	return cns.consensusGroup[index], nil
}

func (cns *Consensus) getFormattedTime() string {
	return cns.Chr.SyncTime().FormattedCurrentTime(cns.Chr.ClockOffset())
}
