package chronology

import (
	"time"

	"github.com/davecgh/go-spew/spew"
)

type Round struct {
	Index        int
	TimeStamp    time.Time
	TimeDuration time.Duration
	TimeDivision []time.Duration
	State        RoundState
	Subround
}

// A RoundState specifies in which state is the current round
type RoundState int

const (
	RS_BEFORE_ROUND RoundState = iota
	RS_START_ROUND
	RS_BLOCK
	RS_COMITMENT_HASH
	RS_BITMAP
	RS_COMITMENT
	RS_SIGNATURE
	RS_END_ROUND
	RS_AFTER_ROUND
	RS_ABORDED
	RS_UNKNOWN
)

type Subround struct {
	Block         SubroundState
	ComitmentHash SubroundState
	Bitmap        SubroundState
	Comitment     SubroundState
	Signature     SubroundState
}

// A SubroundState specifies what kind of state could have a subround
type SubroundState int

const (
	SS_NOTFINISHED SubroundState = iota
	SS_EXTENDED
	SS_FINISHED
)

func NewRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration, roundTimeDivision []time.Duration, subround Subround) Round {

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	var r Round

	r.Index = int(delta / roundTimeDuration.Nanoseconds())
	r.TimeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(r.Index) * roundTimeDuration.Nanoseconds()))
	r.TimeDuration = roundTimeDuration
	r.TimeDivision = roundTimeDivision
	r.State = r.GetRoundStateFromDateTime(timeStamp)
	r.Subround = subround

	return r
}

func (r *Round) UpdateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time) {

	delta := timeStamp.Sub(genesisRoundTimeStamp).Nanoseconds()

	index := int(delta / r.TimeDuration.Nanoseconds())

	if r.Index != index {
		r.Index = index
		r.TimeStamp = genesisRoundTimeStamp.Add(time.Duration(int64(index) * r.TimeDuration.Nanoseconds()))
	}

	r.State = r.GetRoundStateFromDateTime(timeStamp)
}

func (r *Round) GetRoundStateFromDateTime(timeStamp time.Time) RoundState {

	delta := timeStamp.Sub(r.TimeStamp).Nanoseconds()

	if delta < 0 {
		return RS_BEFORE_ROUND
	}

	if delta > r.TimeDuration.Nanoseconds() {
		return RS_AFTER_ROUND
	}

	for i, v := range r.TimeDivision {
		if delta <= v.Nanoseconds() {
			return RS_START_ROUND + RoundState(i)
		}
	}

	return RS_UNKNOWN
}

func (r *Round) GetRoundStateName(roundState RoundState) string {

	switch roundState {
	case RS_BEFORE_ROUND:
		return ("<BEFORE_ROUND>")
	case RS_START_ROUND:
		return ("<START_ROUND>")
	case RS_BLOCK:
		return ("<BLOCK>")
	case RS_COMITMENT_HASH:
		return ("<COMITMENT_HASH>")
	case RS_BITMAP:
		return ("<BITMAP>")
	case RS_COMITMENT:
		return ("<COMITMENT>")
	case RS_SIGNATURE:
		return ("<SIGNATURE>")
	case RS_END_ROUND:
		return ("<END_ROUND>")
	case RS_AFTER_ROUND:
		return ("<AFTER_ROUND>")
	case RS_ABORDED:
		return ("<ABORDED>")
	case RS_UNKNOWN:
		return ("<UNKNOWN>")
	default:
		return ("Undifined round state")
	}
}

func (r *Round) Print() {
	spew.Dump(r)
}

func (r *Round) ResetSubround() {
	r.Subround = Subround{SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED}
}
